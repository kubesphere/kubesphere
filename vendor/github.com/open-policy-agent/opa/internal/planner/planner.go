// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package planner contains a query planner for Rego queries.
package planner

import (
	"errors"
	"fmt"
	"sort"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/internal/ir"
)

type planiter func() error
type binaryiter func(ir.Local, ir.Local) error

type wasmBuiltin struct {
	*ast.Builtin
	WasmFunction string
}

// Planner implements a query planner for Rego queries.
type Planner struct {
	policy    *ir.Policy              // result of planning
	queries   []ast.Body              // input query to plan
	modules   []*ast.Module           // input modules to support queries
	rewritten map[ast.Var]ast.Var     // rewritten query vars
	strings   map[string]int          // global string constant indices
	externs   map[string]struct{}     // built-in functions that are required in execution environment
	decls     map[string]*ast.Builtin // built-in functions that may be provided in execution environment
	rules     *ruletrie               // rules that may be planned
	funcs     *funcstack              // functions that have been planned
	curr      *ir.Block               // in-progress query block
	vars      *varstack               // in-scope variables
	ltarget   ir.Local                // target variable of last planned statement
	lnext     ir.Local                // next variable to use
}

// New returns a new Planner object.
func New() *Planner {
	return &Planner{
		policy: &ir.Policy{
			Static: &ir.Static{},
			Plan:   &ir.Plan{},
			Funcs:  &ir.Funcs{},
		},
		strings: map[string]int{},
		externs: map[string]struct{}{},
		lnext:   ir.Unused,
		vars: newVarstack(map[ast.Var]ir.Local{
			ast.InputRootDocument.Value.(ast.Var):   ir.Input,
			ast.DefaultRootDocument.Value.(ast.Var): ir.Data,
		}),
		rules: newRuletrie(),
		funcs: newFuncstack(),
	}
}

// WithBuiltinDecls tells the planner what built-in function may be available
// inside the execution environment.
func (p *Planner) WithBuiltinDecls(decls map[string]*ast.Builtin) *Planner {
	p.decls = decls
	return p
}

// WithQueries sets the query set to generate a plan for.
func (p *Planner) WithQueries(queries []ast.Body) *Planner {
	p.queries = queries
	return p
}

// WithModules sets the module set that contains query dependencies.
func (p *Planner) WithModules(modules []*ast.Module) *Planner {
	p.modules = modules
	return p
}

// WithRewrittenVars sets a mapping of rewritten query vars on the planner. The
// plan will use the rewritten variable name but the result set key will be the
// original variable name.
func (p *Planner) WithRewrittenVars(vs map[ast.Var]ast.Var) *Planner {
	p.rewritten = vs
	return p
}

// Plan returns a IR plan for the policy query.
func (p *Planner) Plan() (*ir.Policy, error) {

	if err := p.buildFunctrie(); err != nil {
		return nil, err
	}

	if err := p.planQueries(); err != nil {
		return nil, err
	}

	if err := p.planExterns(); err != nil {
		return nil, err
	}

	return p.policy, nil
}

func (p *Planner) buildFunctrie() error {

	for _, module := range p.modules {

		// Create functrie node for empty packages so that extent queries return
		// empty objects. For example:
		//
		// package x.y
		//
		// Query: data.x
		//
		// Expected result: {"y": {}}
		if len(module.Rules) == 0 {
			_ = p.rules.LookupOrInsert(module.Package.Path)
			continue
		}

		for _, rule := range module.Rules {
			val := p.rules.LookupOrInsert(rule.Path())
			val.rules = append(val.rules, rule)
		}
	}

	return nil
}

func (p *Planner) planRules(rules []*ast.Rule) (string, error) {

	path := rules[0].Path().String()

	if funcName, ok := p.funcs.Get(path); ok {
		return funcName, nil
	}

	// Save current state of planner.
	//
	// TODO(tsandall): perhaps we would be better off using stacks here or
	// splitting block planner into separate struct that could be instantiated
	// for rule and comprehension bodies.
	pvars := p.vars
	pcurr := p.curr
	pltarget := p.ltarget
	plnext := p.lnext

	// Reset the variable counter for the function plan.
	p.lnext = ir.Input

	// Create function definition for rules.
	fn := &ir.Func{
		Name: fmt.Sprintf("g%d.%s", p.funcs.gen, path),
		Params: []ir.Local{
			p.newLocal(), // input document
			p.newLocal(), // data document
		},
		Return: p.newLocal(),
	}

	// Initialize parameters for functions.
	for i := 0; i < len(rules[0].Head.Args); i++ {
		fn.Params = append(fn.Params, p.newLocal())
	}

	params := fn.Params[2:]

	// Initialize return value for partial set/object rules. Complete docs do
	// not require their return value to be initialized.
	if rules[0].Head.DocKind() == ast.PartialObjectDoc {
		fn.Blocks = append(fn.Blocks, &ir.Block{
			Stmts: []ir.Stmt{
				&ir.MakeObjectStmt{
					Target: fn.Return,
				},
			},
		})
	} else if rules[0].Head.DocKind() == ast.PartialSetDoc {
		fn.Blocks = append(fn.Blocks, &ir.Block{
			Stmts: []ir.Stmt{
				&ir.MakeSetStmt{
					Target: fn.Return,
				},
			},
		})
	}

	// At this point the locals for the params and return value have been
	// allocated. This will be the first local that can be used in each block.
	lnext := p.lnext

	var defaultRule *ast.Rule

	// Generate function blocks for rules.
	for i := range rules {

		// Save default rule for the end.
		if rules[i].Default {
			defaultRule = rules[i]
			continue
		}

		// Ordered rules are nested inside an additional block so that execution
		// can short-circuit. For unordered rules blocks can be added directly
		// to the function.
		var blocks *[]*ir.Block

		if rules[i].Else == nil {
			blocks = &fn.Blocks
		} else {
			stmt := &ir.BlockStmt{}
			block := &ir.Block{Stmts: []ir.Stmt{stmt}}
			fn.Blocks = append(fn.Blocks, block)
			blocks = &stmt.Blocks
		}

		// Unordered rules are treated as a special case of ordered rules.
		for rule := rules[i]; rule != nil; rule = rule.Else {

			// Setup planner for block.
			p.lnext = lnext
			p.vars = newVarstack(map[ast.Var]ir.Local{
				ast.InputRootDocument.Value.(ast.Var):   fn.Params[0],
				ast.DefaultRootDocument.Value.(ast.Var): fn.Params[1],
			})

			curr := &ir.Block{}
			*blocks = append(*blocks, curr)
			p.curr = curr

			// Complete and partial rules are treated as special cases of
			// functions. If there are args, the first step is a no-op.
			err := p.planFuncParams(params, rule.Head.Args, 0, func() error {

				// Run planner on the rule body.
				err := p.planQuery(rule.Body, 0, func() error {

					// Run planner on the result.
					switch rule.Head.DocKind() {
					case ast.CompleteDoc:
						return p.planTerm(rule.Head.Value, func() error {
							p.appendStmt(&ir.AssignVarOnceStmt{
								Target: fn.Return,
								Source: p.ltarget,
							})
							return nil
						})
					case ast.PartialSetDoc:
						return p.planTerm(rule.Head.Key, func() error {
							p.appendStmt(&ir.SetAddStmt{
								Set:   fn.Return,
								Value: p.ltarget,
							})
							return nil
						})
					case ast.PartialObjectDoc:
						return p.planTerm(rule.Head.Key, func() error {
							key := p.ltarget
							return p.planTerm(rule.Head.Value, func() error {
								value := p.ltarget
								p.appendStmt(&ir.ObjectInsertOnceStmt{
									Object: fn.Return,
									Key:    key,
									Value:  value,
								})
								return nil
							})
						})
					default:
						return fmt.Errorf("illegal rule kind")
					}
				})

				if err != nil {
					return err
				}

				// Ordered rules are handled by short circuiting execution. The
				// plan will jump out to the extra block that was planned above.
				if rule.Else != nil {
					p.appendStmt(&ir.IsDefinedStmt{Source: fn.Return})
					p.appendStmt(&ir.BreakStmt{Index: 1})
				}

				return nil
			})

			if err != nil {
				return "", err
			}
		}
	}

	// Default rules execute if the return is undefined.
	if defaultRule != nil {

		fn.Blocks = append(fn.Blocks, &ir.Block{
			Stmts: []ir.Stmt{
				&ir.IsUndefinedStmt{Source: fn.Return},
			},
		})

		p.curr = fn.Blocks[len(fn.Blocks)-1]

		err := p.planQuery(defaultRule.Body, 0, func() error {
			return p.planTerm(defaultRule.Head.Value, func() error {
				p.appendStmt(&ir.AssignVarOnceStmt{
					Target: fn.Return,
					Source: p.ltarget,
				})
				return nil
			})
		})

		if err != nil {
			return "", err
		}
	}

	// All rules return a value.
	fn.Blocks = append(fn.Blocks, &ir.Block{
		Stmts: []ir.Stmt{
			&ir.ReturnLocalStmt{
				Source: fn.Return,
			},
		},
	})

	p.appendFunc(fn)
	p.funcs.Add(path, fn.Name)

	// Restore the state of the planner.
	p.lnext = plnext
	p.ltarget = pltarget
	p.vars = pvars
	p.curr = pcurr

	return fn.Name, nil
}

func (p *Planner) planFuncParams(params []ir.Local, args ast.Args, idx int, iter planiter) error {
	if idx >= len(args) {
		return iter()
	}
	return p.planUnifyLocal(params[idx], args[idx], func() error {
		return p.planFuncParams(params, args, idx+1, iter)
	})
}

func (p *Planner) planQueries() error {

	// Initialize the plan with a block that prepares the query result.
	p.curr = &ir.Block{}

	// Build a set of variables appearing in the query and allocate strings for
	// each one. The strings will be used in the result set objects.
	qvs := ast.NewVarSet()

	for _, q := range p.queries {
		vs := q.Vars(ast.VarVisitorParams{SkipRefCallHead: true, SkipClosures: true}).Diff(ast.ReservedVars)
		qvs.Update(vs)
	}

	lvarnames := make(map[ast.Var]ir.Local, len(qvs))

	for _, qv := range qvs.Sorted() {
		qv = p.rewrittenVar(qv)
		if !qv.IsGenerated() && !qv.IsWildcard() {
			stmt := &ir.MakeStringStmt{
				Index:  p.getStringConst(string(qv)),
				Target: p.newLocal(),
			}
			p.appendStmt(stmt)
			lvarnames[qv] = stmt.Target
		}
	}

	if len(p.curr.Stmts) > 0 {
		p.appendBlock(p.curr)
	}

	lnext := p.lnext

	for _, q := range p.queries {
		p.lnext = lnext
		p.vars.Push(map[ast.Var]ir.Local{})
		p.curr = &ir.Block{}
		defined := false
		qvs := q.Vars(ast.VarVisitorParams{SkipRefCallHead: true, SkipClosures: true}).Diff(ast.ReservedVars).Sorted()

		if err := p.planQuery(q, 0, func() error {

			// Add an object containing variable bindings into the result set.
			lr := p.newLocal()

			p.appendStmt(&ir.MakeObjectStmt{
				Target: lr,
			})

			for _, qv := range qvs {
				rw := p.rewrittenVar(qv)
				if !rw.IsGenerated() && !rw.IsWildcard() {
					p.appendStmt(&ir.ObjectInsertStmt{
						Object: lr,
						Key:    lvarnames[rw],
						Value:  p.vars.GetOrEmpty(qv),
					})
				}
			}

			p.appendStmt(&ir.ResultSetAdd{
				Value: lr,
			})

			defined = true
			return nil
		}); err != nil {
			return err
		}

		p.vars.Pop()

		if defined {
			p.appendBlock(p.curr)
		}
	}

	return nil
}

func (p *Planner) planQuery(q ast.Body, index int, iter planiter) error {

	if index >= len(q) {
		return iter()
	}

	return p.planExpr(q[index], func() error {
		return p.planQuery(q, index+1, iter)
	})
}

// TODO(tsandall): improve errors to include location information.
func (p *Planner) planExpr(e *ast.Expr, iter planiter) error {
	if e.Negated {
		return p.planNot(e, iter)
	}

	if len(e.With) > 0 {
		return p.planWith(e, iter)
	}

	if e.IsCall() {
		return p.planExprCall(e, iter)
	}

	return p.planExprTerm(e, iter)
}

func (p *Planner) planNot(e *ast.Expr, iter planiter) error {

	not := &ir.NotStmt{
		Block: &ir.Block{},
	}

	prev := p.curr
	p.curr = not.Block

	if err := p.planExpr(e.Complement(), func() error {
		return nil
	}); err != nil {
		return err
	}

	p.curr = prev
	p.appendStmt(not)

	return iter()
}

func (p *Planner) planWith(e *ast.Expr, iter planiter) error {

	// Plan the values that will be applied by the with modifiers. All values
	// must be defined for the overall expression to evaluate.
	values := make([]*ast.Term, len(e.With))

	for i := range e.With {
		values[i] = e.With[i].Value
	}

	return p.planTermSlice(values, func(locals []ir.Local) error {

		paths := make([][]int, len(e.With))
		saveVars := ast.NewVarSet()
		dataRefs := []ast.Ref{}

		for i := range e.With {

			target := e.With[i].Target.Value.(ast.Ref)
			paths[i] = make([]int, len(target)-1)

			for j := 1; j < len(target); j++ {
				if s, ok := target[j].Value.(ast.String); ok {
					paths[i][j-1] = p.getStringConst(string(s))
				} else {
					return errors.New("invalid with target")
				}
			}

			head := target[0].Value.(ast.Var)
			saveVars.Add(head)

			if head.Equal(ast.DefaultRootDocument.Value) {
				dataRefs = append(dataRefs, target)
			}
		}

		restore := make([][2]ir.Local, len(saveVars))

		for i, v := range saveVars.Sorted() {
			lorig := p.vars.GetOrEmpty(v)
			lsave := p.newLocal()
			p.appendStmt(&ir.AssignVarStmt{Source: lorig, Target: lsave})
			restore[i] = [2]ir.Local{lorig, lsave}
		}

		// If any of the with statements targeted the data document we shadow
		// the existing planned functions during expression planning. This
		// causes the planner to re-plan any rules that may be required during
		// planning of this expression (transitively).
		if len(dataRefs) > 0 {
			p.funcs.Push(map[string]string{})
			for _, ref := range dataRefs {
				p.rules.Push(ref)
			}
		}

		err := p.planWithRec(e, paths, locals, 0, func() error {
			if len(dataRefs) > 0 {
				p.funcs.Pop()
				for i := len(dataRefs) - 1; i >= 0; i-- {
					p.rules.Pop(dataRefs[i])
				}
			}

			err := p.planWithUndoRec(restore, 0, func() error {

				err := iter()

				if len(dataRefs) > 0 {
					p.funcs.Push(map[string]string{})
					for _, ref := range dataRefs {
						p.rules.Push(ref)
					}
				}
				return err
			})

			return err
		})

		if len(dataRefs) > 0 {
			p.funcs.Pop()
			for i := len(dataRefs) - 1; i >= 0; i-- {
				p.rules.Pop(dataRefs[i])
			}
		}
		return err

	})
}

func (p *Planner) planWithRec(e *ast.Expr, targets [][]int, values []ir.Local, index int, iter planiter) error {
	if index >= len(e.With) {
		return p.planExpr(e.NoWith(), iter)
	}

	prev := p.curr
	p.curr = &ir.Block{}

	err := p.planWithRec(e, targets, values, index+1, iter)
	if err != nil {
		return err
	}

	block := p.curr
	p.curr = prev
	target := e.With[index].Target.Value.(ast.Ref)
	head := target[0].Value.(ast.Var)

	stmt := &ir.WithStmt{
		Local: p.vars.GetOrEmpty(head),
		Path:  targets[index],
		Value: values[index],
		Block: block,
	}

	p.appendStmt(stmt)

	return nil
}

func (p *Planner) planWithUndoRec(restore [][2]ir.Local, index int, iter planiter) error {

	if index >= len(restore) {
		return iter()
	}

	prev := p.curr
	p.curr = &ir.Block{}

	if err := p.planWithUndoRec(restore, index+1, iter); err != nil {
		return err
	}

	block := p.curr
	p.curr = prev
	lorig := restore[index][0]
	lsave := restore[index][1]

	p.appendStmt(&ir.WithStmt{
		Local: lorig,
		Value: lsave,
		Block: block,
	})

	return nil
}

func (p *Planner) planExprTerm(e *ast.Expr, iter planiter) error {
	return p.planTerm(e.Terms.(*ast.Term), func() error {
		falsy := p.newLocal()
		p.appendStmt(&ir.MakeBooleanStmt{
			Value:  false,
			Target: falsy,
		})
		p.appendStmt(&ir.NotEqualStmt{
			A: p.ltarget,
			B: falsy,
		})
		return iter()
	})
}

func (p *Planner) planExprCall(e *ast.Expr, iter planiter) error {
	operator := e.Operator().String()
	switch operator {
	case ast.Equality.Name:
		return p.planUnify(e.Operand(0), e.Operand(1), iter)
	case ast.Equal.Name:
		return p.planBinaryExpr(e, func(a, b ir.Local) error {
			p.appendStmt(&ir.EqualStmt{
				A: a,
				B: b,
			})
			return iter()
		})
	case ast.LessThan.Name:
		return p.planBinaryExpr(e, func(a, b ir.Local) error {
			p.appendStmt(&ir.LessThanStmt{
				A: a,
				B: b,
			})
			return iter()
		})
	case ast.LessThanEq.Name:
		return p.planBinaryExpr(e, func(a, b ir.Local) error {
			p.appendStmt(&ir.LessThanEqualStmt{
				A: a,
				B: b,
			})
			return iter()
		})
	case ast.GreaterThan.Name:
		return p.planBinaryExpr(e, func(a, b ir.Local) error {
			p.appendStmt(&ir.GreaterThanStmt{
				A: a,
				B: b,
			})
			return iter()
		})
	case ast.GreaterThanEq.Name:
		return p.planBinaryExpr(e, func(a, b ir.Local) error {
			p.appendStmt(&ir.GreaterThanEqualStmt{
				A: a,
				B: b,
			})
			return iter()
		})
	case ast.NotEqual.Name:
		return p.planBinaryExpr(e, func(a, b ir.Local) error {
			p.appendStmt(&ir.NotEqualStmt{
				A: a,
				B: b,
			})
			return iter()
		})
	default:

		var name string
		var arity int
		var args []ir.Local

		node := p.rules.Lookup(e.Operator())

		if node != nil {
			var err error
			name, err = p.planRules(node.Rules())
			if err != nil {
				return err
			}
			arity = node.Arity()
			args = []ir.Local{
				p.vars.GetOrEmpty(ast.InputRootDocument.Value.(ast.Var)),
				p.vars.GetOrEmpty(ast.DefaultRootDocument.Value.(ast.Var)),
			}
		} else if decl, ok := p.decls[operator]; ok {
			arity = len(decl.Decl.Args())
			name = operator
			p.externs[operator] = struct{}{}
		} else {
			return fmt.Errorf("illegal call: unknown operator %q", operator)
		}

		operands := e.Operands()

		if len(operands) == arity {
			// rule: f(x) = x { ... }
			// call: f(x) # result not captured
			return p.planCallArgs(operands, 0, args, func(args []ir.Local) error {
				p.ltarget = p.newLocal()
				p.appendStmt(&ir.CallStmt{
					Func:   name,
					Args:   args,
					Result: p.ltarget,
				})

				falsy := p.newLocal()

				p.appendStmt(&ir.MakeBooleanStmt{
					Value:  false,
					Target: falsy,
				})

				p.appendStmt(&ir.NotEqualStmt{
					A: p.ltarget,
					B: falsy,
				})

				return iter()
			})
		} else if len(operands) == arity+1 {
			// rule: f(x) = x { ... }
			// call: f(x, 1)  # caller captures result
			return p.planCallArgs(operands[:len(operands)-1], 0, args, func(args []ir.Local) error {
				result := p.newLocal()
				p.appendStmt(&ir.CallStmt{
					Func:   name,
					Args:   args,
					Result: result,
				})
				return p.planUnifyLocal(result, operands[len(operands)-1], iter)
			})
		}

		return fmt.Errorf("illegal call: wrong number of operands: got %v, want %v)", len(operands), arity)
	}
}

func (p *Planner) planCallArgs(terms []*ast.Term, idx int, args []ir.Local, iter func([]ir.Local) error) error {
	if idx >= len(terms) {
		return iter(args)
	}
	return p.planTerm(terms[idx], func() error {
		args = append(args, p.ltarget)
		return p.planCallArgs(terms, idx+1, args, iter)
	})
}

func (p *Planner) planUnify(a, b *ast.Term, iter planiter) error {

	switch va := a.Value.(type) {
	case ast.Null, ast.Boolean, ast.Number, ast.String, ast.Ref, ast.Set, *ast.SetComprehension, *ast.ArrayComprehension, *ast.ObjectComprehension:
		return p.planTerm(a, func() error {
			return p.planUnifyLocal(p.ltarget, b, iter)
		})
	case ast.Var:
		return p.planUnifyVar(va, b, iter)
	case *ast.Array:
		switch vb := b.Value.(type) {
		case ast.Var:
			return p.planUnifyVar(vb, a, iter)
		case ast.Ref:
			return p.planTerm(b, func() error {
				return p.planUnifyLocalArray(p.ltarget, va, iter)
			})
		case *ast.Array:
			if va.Len() == vb.Len() {
				return p.planUnifyArraysRec(va, vb, 0, iter)
			}
			return nil
		}
	case ast.Object:
		switch vb := b.Value.(type) {
		case ast.Var:
			return p.planUnifyVar(vb, a, iter)
		case ast.Ref:
			return p.planTerm(b, func() error {
				return p.planUnifyLocalObject(p.ltarget, va, iter)
			})
		case ast.Object:
			if va.Len() == vb.Len() {
				return p.planUnifyObjectsRec(va, vb, va.Keys(), 0, iter)
			}
			return nil
		}
	}

	return fmt.Errorf("not implemented: unify(%v, %v)", a, b)
}

func (p *Planner) planUnifyVar(a ast.Var, b *ast.Term, iter planiter) error {

	if la, ok := p.vars.Get(a); ok {
		return p.planUnifyLocal(la, b, iter)
	}

	return p.planTerm(b, func() error {
		target := p.newLocal()
		p.vars.Put(a, target)
		p.appendStmt(&ir.AssignVarStmt{
			Source: p.ltarget,
			Target: target,
		})
		return iter()
	})
}

func (p *Planner) planUnifyLocal(a ir.Local, b *ast.Term, iter planiter) error {
	switch vb := b.Value.(type) {
	case ast.Null, ast.Boolean, ast.Number, ast.String, ast.Ref, ast.Set, *ast.SetComprehension, *ast.ArrayComprehension, *ast.ObjectComprehension:
		return p.planTerm(b, func() error {
			p.appendStmt(&ir.EqualStmt{
				A: a,
				B: p.ltarget,
			})
			return iter()
		})
	case ast.Var:
		if lv, ok := p.vars.Get(vb); ok {
			p.appendStmt(&ir.EqualStmt{
				A: a,
				B: lv,
			})
			return iter()
		}
		lv := p.newLocal()
		p.vars.Put(vb, lv)
		p.appendStmt(&ir.AssignVarStmt{
			Source: a,
			Target: lv,
		})
		return iter()
	case *ast.Array:
		return p.planUnifyLocalArray(a, vb, iter)
	case ast.Object:
		return p.planUnifyLocalObject(a, vb, iter)
	}

	return fmt.Errorf("not implemented: unifyLocal(%v, %v)", a, b)
}

func (p *Planner) planUnifyLocalArray(a ir.Local, b *ast.Array, iter planiter) error {
	p.appendStmt(&ir.IsArrayStmt{
		Source: a,
	})

	blen := p.newLocal()
	alen := p.newLocal()

	p.appendStmt(&ir.LenStmt{
		Source: a,
		Target: alen,
	})

	p.appendStmt(&ir.MakeNumberIntStmt{
		Value:  int64(b.Len()),
		Target: blen,
	})

	p.appendStmt(&ir.EqualStmt{
		A: alen,
		B: blen,
	})

	lkey := p.newLocal()

	p.appendStmt(&ir.MakeNumberIntStmt{
		Target: lkey,
	})

	lval := p.newLocal()

	return p.planUnifyLocalArrayRec(a, 0, b, lkey, lval, iter)
}

func (p *Planner) planUnifyLocalArrayRec(a ir.Local, index int, b *ast.Array, lkey, lval ir.Local, iter planiter) error {
	if b.Len() == index {
		return iter()
	}

	p.appendStmt(&ir.AssignIntStmt{
		Value:  int64(index),
		Target: lkey,
	})

	p.appendStmt(&ir.DotStmt{
		Source: a,
		Key:    lkey,
		Target: lval,
	})

	return p.planUnifyLocal(lval, b.Elem(index), func() error {
		return p.planUnifyLocalArrayRec(a, index+1, b, lkey, lval, iter)
	})
}

func (p *Planner) planUnifyLocalObject(a ir.Local, b ast.Object, iter planiter) error {
	p.appendStmt(&ir.IsObjectStmt{
		Source: a,
	})

	blen := p.newLocal()
	alen := p.newLocal()

	p.appendStmt(&ir.LenStmt{
		Source: a,
		Target: alen,
	})

	p.appendStmt(&ir.MakeNumberIntStmt{
		Value:  int64(b.Len()),
		Target: blen,
	})

	p.appendStmt(&ir.EqualStmt{
		A: alen,
		B: blen,
	})

	lkey := p.newLocal()
	lval := p.newLocal()
	bkeys := b.Keys()

	return p.planUnifyLocalObjectRec(a, 0, bkeys, b, lkey, lval, iter)
}

func (p *Planner) planUnifyLocalObjectRec(a ir.Local, index int, keys []*ast.Term, b ast.Object, lkey, lval ir.Local, iter planiter) error {

	if index == len(keys) {
		return iter()
	}

	return p.planTerm(keys[index], func() error {
		p.appendStmt(&ir.AssignVarStmt{
			Source: p.ltarget,
			Target: lkey,
		})
		p.appendStmt(&ir.DotStmt{
			Source: a,
			Key:    lkey,
			Target: lval,
		})
		return p.planUnifyLocal(lval, b.Get(keys[index]), func() error {
			return p.planUnifyLocalObjectRec(a, index+1, keys, b, lkey, lval, iter)
		})
	})
}

func (p *Planner) planUnifyArraysRec(a, b *ast.Array, index int, iter planiter) error {
	if index == a.Len() {
		return iter()
	}
	return p.planUnify(a.Elem(index), b.Elem(index), func() error {
		return p.planUnifyArraysRec(a, b, index+1, iter)
	})
}

func (p *Planner) planUnifyObjectsRec(a, b ast.Object, keys []*ast.Term, index int, iter planiter) error {
	if index == len(keys) {
		return iter()
	}

	aval := a.Get(keys[index])
	bval := b.Get(keys[index])
	if aval == nil || bval == nil {
		return nil
	}

	return p.planUnify(aval, bval, func() error {
		return p.planUnifyObjectsRec(a, b, keys, index+1, iter)
	})
}

func (p *Planner) planBinaryExpr(e *ast.Expr, iter binaryiter) error {
	return p.planTerm(e.Operand(0), func() error {
		a := p.ltarget
		return p.planTerm(e.Operand(1), func() error {
			b := p.ltarget
			return iter(a, b)
		})
	})
}

func (p *Planner) planTerm(t *ast.Term, iter planiter) error {

	switch v := t.Value.(type) {
	case ast.Null:
		return p.planNull(v, iter)
	case ast.Boolean:
		return p.planBoolean(v, iter)
	case ast.Number:
		return p.planNumber(v, iter)
	case ast.String:
		return p.planString(v, iter)
	case ast.Var:
		return p.planVar(v, iter)
	case ast.Ref:
		return p.planRef(v, iter)
	case *ast.Array:
		return p.planArray(v, iter)
	case ast.Object:
		return p.planObject(v, iter)
	case ast.Set:
		return p.planSet(v, iter)
	case *ast.SetComprehension:
		return p.planSetComprehension(v, iter)
	case *ast.ArrayComprehension:
		return p.planArrayComprehension(v, iter)
	case *ast.ObjectComprehension:
		return p.planObjectComprehension(v, iter)
	default:
		return fmt.Errorf("%v term not implemented", ast.TypeName(v))
	}
}

func (p *Planner) planNull(null ast.Null, iter planiter) error {

	target := p.newLocal()

	p.appendStmt(&ir.MakeNullStmt{
		Target: target,
	})

	p.ltarget = target

	return iter()
}

func (p *Planner) planBoolean(b ast.Boolean, iter planiter) error {

	target := p.newLocal()

	p.appendStmt(&ir.MakeBooleanStmt{
		Value:  bool(b),
		Target: target,
	})

	p.ltarget = target

	return iter()
}

func (p *Planner) planNumber(num ast.Number, iter planiter) error {

	index := p.getStringConst(string(num))
	target := p.newLocal()

	p.appendStmt(&ir.MakeNumberRefStmt{
		Index:  index,
		Target: target,
	})

	p.ltarget = target
	return iter()
}

func (p *Planner) planNumberFloat(f float64, iter planiter) error {

	target := p.newLocal()

	p.appendStmt(&ir.MakeNumberFloatStmt{
		Value:  f,
		Target: target,
	})

	p.ltarget = target

	return iter()
}

func (p *Planner) planNumberInt(i int64, iter planiter) error {

	target := p.newLocal()

	p.appendStmt(&ir.MakeNumberIntStmt{
		Value:  i,
		Target: target,
	})

	p.ltarget = target

	return iter()
}

func (p *Planner) planString(str ast.String, iter planiter) error {

	index := p.getStringConst(string(str))
	target := p.newLocal()

	p.appendStmt(&ir.MakeStringStmt{
		Index:  index,
		Target: target,
	})

	p.ltarget = target

	return iter()
}

func (p *Planner) planVar(v ast.Var, iter planiter) error {
	p.ltarget = p.vars.GetOrElse(v, func() ir.Local {
		return p.newLocal()
	})
	return iter()
}

func (p *Planner) planArray(arr *ast.Array, iter planiter) error {

	larr := p.newLocal()

	p.appendStmt(&ir.MakeArrayStmt{
		Capacity: int32(arr.Len()),
		Target:   larr,
	})

	return p.planArrayRec(arr, 0, larr, iter)
}

func (p *Planner) planArrayRec(arr *ast.Array, index int, larr ir.Local, iter planiter) error {
	if index == arr.Len() {
		p.ltarget = larr
		return iter()
	}

	return p.planTerm(arr.Elem(index), func() error {

		p.appendStmt(&ir.ArrayAppendStmt{
			Value: p.ltarget,
			Array: larr,
		})

		return p.planArrayRec(arr, index+1, larr, iter)
	})
}

func (p *Planner) planObject(obj ast.Object, iter planiter) error {

	lobj := p.newLocal()

	p.appendStmt(&ir.MakeObjectStmt{
		Target: lobj,
	})

	return p.planObjectRec(obj, 0, obj.Keys(), lobj, iter)
}

func (p *Planner) planObjectRec(obj ast.Object, index int, keys []*ast.Term, lobj ir.Local, iter planiter) error {
	if index == len(keys) {
		p.ltarget = lobj
		return iter()
	}

	return p.planTerm(keys[index], func() error {
		lkey := p.ltarget

		return p.planTerm(obj.Get(keys[index]), func() error {
			lval := p.ltarget
			p.appendStmt(&ir.ObjectInsertStmt{
				Key:    lkey,
				Value:  lval,
				Object: lobj,
			})

			return p.planObjectRec(obj, index+1, keys, lobj, iter)
		})
	})
}

func (p *Planner) planSet(set ast.Set, iter planiter) error {
	lset := p.newLocal()

	p.appendStmt(&ir.MakeSetStmt{
		Target: lset,
	})

	return p.planSetRec(set, 0, set.Slice(), lset, iter)
}

func (p *Planner) planSetRec(set ast.Set, index int, elems []*ast.Term, lset ir.Local, iter planiter) error {
	if index == len(elems) {
		p.ltarget = lset
		return iter()
	}

	return p.planTerm(elems[index], func() error {
		p.appendStmt(&ir.SetAddStmt{
			Value: p.ltarget,
			Set:   lset,
		})
		return p.planSetRec(set, index+1, elems, lset, iter)
	})
}

func (p *Planner) planSetComprehension(sc *ast.SetComprehension, iter planiter) error {

	lset := p.newLocal()

	p.appendStmt(&ir.MakeSetStmt{
		Target: lset,
	})

	return p.planComprehension(sc.Body, func() error {
		return p.planTerm(sc.Term, func() error {
			p.appendStmt(&ir.SetAddStmt{
				Value: p.ltarget,
				Set:   lset,
			})
			return nil
		})
	}, lset, iter)
}

func (p *Planner) planArrayComprehension(ac *ast.ArrayComprehension, iter planiter) error {

	larr := p.newLocal()

	p.appendStmt(&ir.MakeArrayStmt{
		Target: larr,
	})

	return p.planComprehension(ac.Body, func() error {
		return p.planTerm(ac.Term, func() error {
			p.appendStmt(&ir.ArrayAppendStmt{
				Value: p.ltarget,
				Array: larr,
			})
			return nil
		})
	}, larr, iter)
}

func (p *Planner) planObjectComprehension(oc *ast.ObjectComprehension, iter planiter) error {

	lobj := p.newLocal()

	p.appendStmt(&ir.MakeObjectStmt{
		Target: lobj,
	})

	return p.planComprehension(oc.Body, func() error {
		return p.planTerm(oc.Key, func() error {
			lkey := p.ltarget
			return p.planTerm(oc.Value, func() error {
				p.appendStmt(&ir.ObjectInsertOnceStmt{
					Key:    lkey,
					Value:  p.ltarget,
					Object: lobj,
				})
				return nil
			})
		})
	}, lobj, iter)
}

func (p *Planner) planComprehension(body ast.Body, closureIter planiter, target ir.Local, iter planiter) error {

	prev := p.curr
	p.curr = &ir.Block{}

	if err := p.planQuery(body, 0, func() error {
		return closureIter()
	}); err != nil {
		return err
	}

	block := p.curr
	p.curr = prev

	p.appendStmt(&ir.BlockStmt{
		Blocks: []*ir.Block{
			block,
		},
	})

	p.ltarget = target
	return iter()
}

func (p *Planner) planRef(ref ast.Ref, iter planiter) error {

	head, ok := ref[0].Value.(ast.Var)
	if !ok {
		return fmt.Errorf("illegal ref: non-var head")
	}

	if head.Compare(ast.DefaultRootDocument.Value) == 0 {
		virtual := p.rules.Get(ref[0].Value)
		base := &baseptr{local: p.vars.GetOrEmpty(ast.DefaultRootDocument.Value.(ast.Var))}
		return p.planRefData(virtual, base, ref, 1, iter)
	}

	p.ltarget, ok = p.vars.Get(head)
	if !ok {
		return fmt.Errorf("illegal ref: unsafe head")
	}

	return p.planRefRec(ref, 1, iter)
}

func (p *Planner) planRefRec(ref ast.Ref, index int, iter planiter) error {

	if len(ref) == index {
		return iter()
	}

	scan := false

	ast.WalkVars(ref[index], func(v ast.Var) bool {
		if !scan {
			_, exists := p.vars.Get(v)
			if !exists {
				scan = true
			}
		}
		return scan
	})

	if !scan {
		return p.planDot(ref[index], func() error {
			return p.planRefRec(ref, index+1, iter)
		})
	}

	return p.planScan(ref[index], func(lkey ir.Local) error {
		return p.planRefRec(ref, index+1, iter)
	})
}

type baseptr struct {
	local ir.Local
	path  ast.Ref
}

// planRefData implements the virtual document model by generating the value of
// the ref parameter and invoking the iterator with the planner target set to
// the virtual document and all variables in the reference assigned.
func (p *Planner) planRefData(virtual *ruletrie, base *baseptr, ref ast.Ref, index int, iter planiter) error {

	// Early-exit if the end of the reference has been reached. In this case the
	// plan has to materialize the full extent of the referenced value.
	if index >= len(ref) {
		return p.planRefDataExtent(virtual, base, iter)
	}

	// If the reference operand is ground then either continue to the next
	// operand or invoke the function for the rule referred to by this operand.
	if ref[index].IsGround() {

		var vchild *ruletrie

		if virtual != nil {
			vchild = virtual.Get(ref[index].Value)
		}

		rules := vchild.Rules()

		if len(rules) > 0 {
			p.ltarget = p.newLocal()

			funcName, err := p.planRules(rules)
			if err != nil {
				return err
			}

			p.appendStmt(&ir.CallStmt{
				Func: funcName,
				Args: []ir.Local{
					p.vars.GetOrEmpty(ast.InputRootDocument.Value.(ast.Var)),
					p.vars.GetOrEmpty(ast.DefaultRootDocument.Value.(ast.Var)),
				},
				Result: p.ltarget,
			})

			return p.planRefRec(ref, index+1, iter)
		}

		bchild := *base
		bchild.path = append(bchild.path, ref[index])

		return p.planRefData(vchild, &bchild, ref, index+1, iter)
	}

	exclude := ast.NewSet()

	// The planner does not support dynamic dispatch so generate blocks to
	// evaluate each of the rulesets on the child nodes.
	if virtual != nil {

		stmt := &ir.BlockStmt{}

		for _, child := range virtual.Children() {

			block := &ir.Block{}
			prev := p.curr
			p.curr = block
			key := ast.NewTerm(child)
			exclude.Add(key)

			// Assignments in each block due to local unification must be undone
			// so create a new frame that will be popped after this key is
			// processed.
			p.vars.Push(map[ast.Var]ir.Local{})

			if err := p.planTerm(key, func() error {
				return p.planUnifyLocal(p.ltarget, ref[index], func() error {
					// Create a copy of the reference with this operand plugged.
					// This will result in evaluation of the rulesets on the
					// child node.
					cpy := ref.Copy()
					cpy[index] = key
					return p.planRefData(virtual, base, cpy, index, iter)
				})
			}); err != nil {
				return err
			}

			p.vars.Pop()
			p.curr = prev
			stmt.Blocks = append(stmt.Blocks, block)
		}

		p.appendStmt(stmt)
	}

	// If the virtual tree was enumerated then we do not want to enumerate base
	// trees that are rooted at the same key as any of the virtual sub trees. To
	// prevent this we build a set of keys that are to be excluded and check
	// below during the base scan.
	var lexclude *ir.Local

	if exclude.Len() > 0 {
		if err := p.planSet(exclude, func() error {
			v := p.ltarget
			lexclude = &v
			return nil
		}); err != nil {
			return err
		}
	}

	p.ltarget = base.local

	// Perform a scan of the base documents starting from the location referred
	// to by the data pointer. Use the set we built above to avoid revisiting
	// sub trees.
	return p.planRefRec(base.path, 0, func() error {
		return p.planScan(ref[index], func(lkey ir.Local) error {
			if lexclude != nil {
				lignore := p.newLocal()
				p.appendStmt(&ir.NotStmt{
					Block: &ir.Block{
						Stmts: []ir.Stmt{
							&ir.DotStmt{
								Source: *lexclude,
								Key:    lkey,
								Target: lignore,
							},
						},
					},
				})
			}

			// Assume that virtual sub trees have been visited already so
			// recurse without the virtual node.
			return p.planRefData(nil, &baseptr{local: p.ltarget}, ref, index+1, iter)
		})
	})
}

// planRefDataExtent generates the full extent (combined) of the base and
// virtual nodes and then invokes the iterator with the planner target set to
// the full extent.
func (p *Planner) planRefDataExtent(virtual *ruletrie, base *baseptr, iter planiter) error {

	vtarget := p.newLocal()

	// Generate the virtual document out of rules contained under the virtual
	// node (recursively). This document will _ONLY_ contain values generated by
	// rules. No base document values will be included.
	if virtual != nil {

		p.appendStmt(&ir.MakeObjectStmt{
			Target: vtarget,
		})

		for _, key := range virtual.Children() {
			child := virtual.Get(key)

			// Skip functions.
			if child.Arity() > 0 {
				continue
			}

			lkey := p.newLocal()
			idx := p.getStringConst(string(key.(ast.String)))
			p.appendStmt(&ir.MakeStringStmt{
				Index:  idx,
				Target: lkey,
			})

			rules := child.Rules()

			// Build object hierarchy depth-first.
			if len(rules) == 0 {
				err := p.planRefDataExtent(child, nil, func() error {
					p.appendStmt(&ir.ObjectInsertStmt{
						Object: vtarget,
						Key:    lkey,
						Value:  p.ltarget,
					})
					return nil
				})
				if err != nil {
					return err
				}
				continue
			}

			// Generate virtual document for leaf.
			lvalue := p.newLocal()

			funcName, err := p.planRules(rules)
			if err != nil {
				return err
			}

			// Add leaf to object if defined.
			p.appendStmt(&ir.BlockStmt{
				Blocks: []*ir.Block{
					&ir.Block{
						Stmts: []ir.Stmt{
							&ir.CallStmt{
								Func: funcName,
								Args: []ir.Local{
									p.vars.GetOrEmpty(ast.InputRootDocument.Value.(ast.Var)),
									p.vars.GetOrEmpty(ast.DefaultRootDocument.Value.(ast.Var)),
								},
								Result: lvalue,
							},
							&ir.ObjectInsertStmt{
								Object: vtarget,
								Key:    lkey,
								Value:  lvalue,
							},
						},
					},
				},
			})
		}

		// At this point vtarget refers to the full extent of the virtual
		// document at ref. If the base pointer is unset, no further processing
		// is required.
		if base == nil {
			p.ltarget = vtarget
			return iter()
		}
	}

	// Obtain the base document value and merge (recursively) with the virtual
	// document value above if needed.
	prev := p.curr
	p.curr = &ir.Block{}
	p.ltarget = base.local
	target := p.newLocal()

	err := p.planRefRec(base.path, 0, func() error {

		if virtual == nil {
			target = p.ltarget
		} else {
			stmt := &ir.ObjectMergeStmt{
				A:      p.ltarget,
				B:      vtarget,
				Target: target,
			}
			p.appendStmt(stmt)
		}

		p.appendStmt(&ir.BreakStmt{Index: 1})
		return nil
	})

	if err != nil {
		return err
	}

	inner := p.curr

	// Fallback to virtual document value if base document is undefined.
	// Otherwise, this block is undefined.
	p.curr = &ir.Block{}
	p.appendStmt(&ir.BlockStmt{Blocks: []*ir.Block{inner}})

	if virtual != nil {
		p.appendStmt(&ir.AssignVarStmt{
			Source: vtarget,
			Target: target,
		})
	} else {
		p.appendStmt(&ir.BreakStmt{Index: 1})
	}

	outer := p.curr
	p.curr = prev
	p.appendStmt(&ir.BlockStmt{Blocks: []*ir.Block{outer}})

	// At this point, target refers to either the full extent of the base and
	// virtual documents at ref or just the base document at ref.
	p.ltarget = target

	return iter()
}

func (p *Planner) planDot(key *ast.Term, iter planiter) error {

	source := p.ltarget

	return p.planTerm(key, func() error {

		target := p.newLocal()

		p.appendStmt(&ir.DotStmt{
			Source: source,
			Key:    p.ltarget,
			Target: target,
		})

		p.ltarget = target

		return iter()
	})
}

type scaniter func(ir.Local) error

func (p *Planner) planScan(key *ast.Term, iter scaniter) error {

	scan := &ir.ScanStmt{
		Source: p.ltarget,
		Key:    p.newLocal(),
		Value:  p.newLocal(),
		Block:  &ir.Block{},
	}

	prev := p.curr
	p.curr = scan.Block

	if err := p.planUnifyLocal(scan.Key, key, func() error {
		p.ltarget = scan.Value
		return iter(scan.Key)
	}); err != nil {
		return err
	}

	p.curr = prev
	p.appendStmt(scan)

	return nil

}

// planSaveLocals returns a slice of locals holding temporary variables that
// have been assigned from the supplied vars.
func (p *Planner) planSaveLocals(vars ...ir.Local) []ir.Local {

	lsaved := make([]ir.Local, len(vars))

	for i := range vars {

		lsaved[i] = p.newLocal()

		p.appendStmt(&ir.AssignVarStmt{
			Source: vars[i],
			Target: lsaved[i],
		})
	}

	return lsaved
}

type termsliceiter func([]ir.Local) error

func (p *Planner) planTermSlice(terms []*ast.Term, iter termsliceiter) error {
	return p.planTermSliceRec(terms, make([]ir.Local, len(terms)), 0, iter)
}

func (p *Planner) planTermSliceRec(terms []*ast.Term, locals []ir.Local, index int, iter termsliceiter) error {
	if index >= len(terms) {
		return iter(locals)
	}

	return p.planTerm(terms[index], func() error {
		locals[index] = p.ltarget
		return p.planTermSliceRec(terms, locals, index+1, iter)
	})
}

func (p *Planner) planExterns() error {

	p.policy.Static.BuiltinFuncs = make([]*ir.BuiltinFunc, 0, len(p.externs))

	for name := range p.externs {
		p.policy.Static.BuiltinFuncs = append(p.policy.Static.BuiltinFuncs, &ir.BuiltinFunc{Name: name})
	}

	sort.Slice(p.policy.Static.BuiltinFuncs, func(i, j int) bool {
		return p.policy.Static.BuiltinFuncs[i].Name < p.policy.Static.BuiltinFuncs[j].Name
	})

	return nil
}

func (p *Planner) getStringConst(s string) int {
	index, ok := p.strings[s]
	if !ok {
		index = len(p.policy.Static.Strings)
		p.policy.Static.Strings = append(p.policy.Static.Strings, &ir.StringConst{
			Value: s,
		})
		p.strings[s] = index
	}
	return index
}

func (p *Planner) appendStmt(s ir.Stmt) {
	p.curr.Stmts = append(p.curr.Stmts, s)
}

func (p *Planner) appendFunc(f *ir.Func) {
	p.policy.Funcs.Funcs = append(p.policy.Funcs.Funcs, f)
}

func (p *Planner) appendBlock(b *ir.Block) {
	p.policy.Plan.Blocks = append(p.policy.Plan.Blocks, b)
}

func (p *Planner) newLocal() ir.Local {
	x := p.lnext
	p.lnext++
	return x
}

func (p *Planner) rewrittenVar(k ast.Var) ast.Var {
	rw, ok := p.rewritten[k]
	if !ok {
		return k
	}
	return rw
}
