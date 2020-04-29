// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package copypropagation

import (
	"sort"

	"github.com/open-policy-agent/opa/ast"
)

// CopyPropagator implements a simple copy propagation optimization to remove
// intermediate variables in partial evaluation results.
//
// For example, given the query: input.x > 1 where 'input' is unknown, the
// compiled query would become input.x = a; a > 1 which would remain in the
// partial evaluation result. The CopyPropagator will remove the variable
// assignment so that partial evaluation simply outputs input.x > 1.
//
// In many cases, copy propagation can remove all variables from the result of
// partial evaluation which simplifies evaluation for non-OPA consumers.
//
// In some cases, copy propagation cannot remove all variables. If the output of
// a built-in call is subsequently used as a ref head, the output variable must
// be kept. For example. sort(input, x); x[0] == 1. In this case, copy
// propagation cannot replace x[0] == 1 with sort(input, x)[0] == 1 as this is
// not legal.
type CopyPropagator struct {
	livevars           ast.VarSet // vars that must be preserved in the resulting query
	sorted             []ast.Var  // sorted copy of vars to ensure deterministic result
	ensureNonEmptyBody bool
}

// New returns a new CopyPropagator that optimizes queries while preserving vars
// in the livevars set.
func New(livevars ast.VarSet) *CopyPropagator {

	sorted := make([]ast.Var, 0, len(livevars))
	for v := range livevars {
		sorted = append(sorted, v)
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Compare(sorted[j]) < 0
	})

	return &CopyPropagator{livevars: livevars, sorted: sorted}
}

// WithEnsureNonEmptyBody configures p to ensure that results are always non-empty.
func (p *CopyPropagator) WithEnsureNonEmptyBody(yes bool) *CopyPropagator {
	p.ensureNonEmptyBody = yes
	return p
}

// Apply executes the copy propagation optimization and returns a new query.
func (p *CopyPropagator) Apply(query ast.Body) (result ast.Body) {

	uf, ok := makeDisjointSets(p.livevars, query)
	if !ok {
		return query
	}

	// Compute set of vars that appear in the head of refs in the query. If a var
	// is dereferenced, we cannot plug it with a constant value so the constant on
	// the union-find root must be unset (e.g., [1][0] is not legal.)
	headvars := ast.NewVarSet()
	ast.WalkRefs(query, func(x ast.Ref) bool {
		if v, ok := x[0].Value.(ast.Var); ok {
			if root, ok := uf.Find(v); ok {
				root.constant = nil
				headvars.Add(root.key)
			} else {
				headvars.Add(v)
			}
		}
		return false
	})

	bindings := map[ast.Var]*binding{}

	for _, expr := range query {

		pctx := &plugContext{
			bindings: bindings,
			uf:       uf,
			negated:  expr.Negated,
			headvars: headvars,
		}

		if expr, keep := p.plugBindings(pctx, expr); keep {
			if p.updateBindings(pctx, expr) {
				result.Append(expr)
			}
		}
	}

	// Run post-processing step on the query to ensure that all live vars are bound
	// in the result. The plugging that happens above substitutes all vars in the
	// same set with the root.
	//
	// This step should run before the next step to prevent unnecessary bindings
	// from being added to the result. For example:
	//
	// - Given the following result: <empty>
	// - Given the following bindings: x/input.x and y/input
	// - Given the following liveset: {x}
	//
	// If this step were to run AFTER the following step, the output would be:
	//
	//	x = input.x; y = input
	//
	// Even though y = input is not required.
	for _, v := range p.sorted {
		if root, ok := uf.Find(v); ok {
			if root.constant != nil {
				result.Append(ast.Equality.Expr(ast.NewTerm(v), root.constant))
			} else if b, ok := bindings[root.key]; ok {
				result.Append(ast.Equality.Expr(ast.NewTerm(v), ast.NewTerm(b.v)))
			} else if root.key != v {
				result.Append(ast.Equality.Expr(ast.NewTerm(v), ast.NewTerm(root.key)))
			}
		}
	}

	// Run post-processing step on query to ensure that all killed exprs are
	// accounted for. If an expr is killed but the binding is never used, the query
	// must still include the expr. For example, given the query 'input.x = a' and
	// an empty livevar set, the result must include the ref input.x otherwise the
	// query could be satisfied without input.x being defined. When exprs are
	// killed we initialize the binding counter to zero and then increment it each
	// time the binding is substituted. if the binding was never substituted it
	// means the binding value must be added back into the query.
	for _, b := range sortbindings(bindings) {
		if !b.containedIn(result) {
			result.Append(ast.Equality.Expr(ast.NewTerm(b.k), ast.NewTerm(b.v)))
		}
	}

	if p.ensureNonEmptyBody && len(result) == 0 {
		result = append(result, ast.NewExpr(ast.BooleanTerm(true)))
	}

	return result
}

// plugBindings applies the binding list and union-find to x. This process
// removes as many variables as possible.
func (p *CopyPropagator) plugBindings(pctx *plugContext, expr *ast.Expr) (*ast.Expr, bool) {

	// Kill single term expressions that are in the binding list. They will be
	// re-added during post-processing if needed.
	if term, ok := expr.Terms.(*ast.Term); ok {
		if v, ok := term.Value.(ast.Var); ok {
			if root, ok := pctx.uf.Find(v); ok {
				if _, ok := pctx.bindings[root.key]; ok {
					return nil, false
				}
			}
		}
	}

	xform := bindingPlugTransform{
		pctx: pctx,
	}

	// Deep copy the expression as it may be mutated during the transform and
	// the caller running copy propagation may have references to the
	// expression. Note, the transform does not contain any error paths and
	// should never return a non-expression value for the root so consider
	// errors unreachable.
	x, err := ast.Transform(xform, expr.Copy())

	if expr, ok := x.(*ast.Expr); !ok || err != nil {
		panic("unreachable")
	} else {
		return expr, true
	}
}

type bindingPlugTransform struct {
	pctx *plugContext
}

func (t bindingPlugTransform) Transform(x interface{}) (interface{}, error) {
	switch x := x.(type) {
	case ast.Var:
		return t.plugBindingsVar(t.pctx, x), nil
	case ast.Ref:
		return t.plugBindingsRef(t.pctx, x), nil
	default:
		return x, nil
	}
}

func (t bindingPlugTransform) plugBindingsVar(pctx *plugContext, v ast.Var) (result ast.Value) {

	result = v

	// Apply union-find to remove redundant variables from input.
	if root, ok := pctx.uf.Find(v); ok {
		result = root.Value()
	}

	// Apply binding list to substitute remaining vars.
	if v, ok := result.(ast.Var); ok {
		if b, ok := pctx.bindings[v]; ok {
			if !pctx.negated || b.v.IsGround() {
				result = b.v
			}
		}
	}

	return result
}

func (t bindingPlugTransform) plugBindingsRef(pctx *plugContext, v ast.Ref) ast.Ref {

	// Apply union-find to remove redundant variables from input.
	if root, ok := pctx.uf.Find(v[0].Value.(ast.Var)); ok {
		v[0].Value = root.Value()
	}

	result := v

	// Refs require special handling. If the head of the ref was killed, then
	// the rest of the ref must be concatenated with the new base.
	//
	// Invariant: ref heads can only be replaced by refs (not calls).
	if b, ok := pctx.bindings[v[0].Value.(ast.Var)]; ok {
		if !pctx.negated || b.v.IsGround() {
			result = b.v.(ast.Ref).Concat(v[1:])
		}
	}

	return result
}

// updateBindings returns false if the expression can be killed. If the
// expression is killed, the binding list is updated to map a var to value.
func (p *CopyPropagator) updateBindings(pctx *plugContext, expr *ast.Expr) bool {
	if pctx.negated || len(expr.With) > 0 {
		return true
	}
	if expr.IsEquality() {
		a, b := expr.Operand(0), expr.Operand(1)
		if a.Equal(b) {
			return false
		}
		k, v, keep := p.updateBindingsEq(a, b)
		if !keep {
			if v != nil {
				pctx.bindings[k] = newbinding(k, v)
			}
			return false
		}
	} else if expr.IsCall() {
		terms := expr.Terms.([]*ast.Term)
		output := terms[len(terms)-1]
		if k, ok := output.Value.(ast.Var); ok && !p.livevars.Contains(k) && !pctx.headvars.Contains(k) {
			pctx.bindings[k] = newbinding(k, ast.CallTerm(terms[:len(terms)-1]...).Value)
			return false
		}
	}
	return !isNoop(expr)
}

func (p *CopyPropagator) updateBindingsEq(a, b *ast.Term) (ast.Var, ast.Value, bool) {
	k, v, keep := p.updateBindingsEqAsymmetric(a, b)
	if !keep {
		return k, v, keep
	}
	return p.updateBindingsEqAsymmetric(b, a)
}

func (p *CopyPropagator) updateBindingsEqAsymmetric(a, b *ast.Term) (ast.Var, ast.Value, bool) {
	k, ok := a.Value.(ast.Var)
	if !ok || p.livevars.Contains(k) {
		return "", nil, true
	}

	switch b.Value.(type) {
	case ast.Ref, ast.Call:
		return k, b.Value, false
	}

	return "", nil, true
}

type plugContext struct {
	bindings map[ast.Var]*binding
	uf       *unionFind
	headvars ast.VarSet
	negated  bool
}

type binding struct {
	k ast.Var
	v ast.Value
}

func newbinding(k ast.Var, v ast.Value) *binding {
	return &binding{k: k, v: v}
}

func (b *binding) containedIn(query ast.Body) bool {
	var stop bool
	switch v := b.v.(type) {
	case ast.Ref:
		ast.WalkRefs(query, func(other ast.Ref) bool {
			if stop || other.HasPrefix(v) {
				stop = true
				return stop
			}
			return false
		})
	default:
		ast.WalkTerms(query, func(other *ast.Term) bool {
			if stop || other.Value.Compare(v) == 0 {
				stop = true
				return stop
			}
			return false
		})
	}
	return stop
}

func sortbindings(bindings map[ast.Var]*binding) []*binding {
	sorted := make([]*binding, 0, len(bindings))
	for _, b := range bindings {
		sorted = append(sorted, b)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].k.Compare(sorted[j].k) < 0
	})
	return sorted
}

type unionFind struct {
	roots   map[ast.Var]*unionFindRoot
	parents map[ast.Var]ast.Var
	rank    rankFunc
}

// makeDisjointSets builds the union-find structure for the query. The structure
// is built by processing all of the equality exprs in the query. Sets represent
// vars that must be equal to each other. In addition to vars, each set can have
// at most one constant. If the query contains expressions that cannot be
// satisfied (e.g., because a set has multiple constants) this function returns
// false.
func makeDisjointSets(livevars ast.VarSet, query ast.Body) (*unionFind, bool) {
	uf := newUnionFind(func(r1, r2 *unionFindRoot) (*unionFindRoot, *unionFindRoot) {
		if livevars.Contains(r1.key) {
			return r1, r2
		}
		return r2, r1
	})
	for _, expr := range query {
		if expr.IsEquality() && !expr.Negated && len(expr.With) == 0 {
			a, b := expr.Operand(0), expr.Operand(1)
			varA, ok1 := a.Value.(ast.Var)
			varB, ok2 := b.Value.(ast.Var)
			if ok1 && ok2 {
				if _, ok := uf.Merge(varA, varB); !ok {
					return nil, false
				}
			} else if ok1 && ast.IsConstant(b.Value) {
				root := uf.MakeSet(varA)
				if root.constant != nil && !root.constant.Equal(b) {
					return nil, false
				}
				root.constant = b
			} else if ok2 && ast.IsConstant(a.Value) {
				root := uf.MakeSet(varB)
				if root.constant != nil && !root.constant.Equal(a) {
					return nil, false
				}
				root.constant = a
			}
		}
	}

	return uf, true
}

type rankFunc func(*unionFindRoot, *unionFindRoot) (*unionFindRoot, *unionFindRoot)

func newUnionFind(rank rankFunc) *unionFind {
	return &unionFind{
		roots:   map[ast.Var]*unionFindRoot{},
		parents: map[ast.Var]ast.Var{},
		rank:    rank,
	}
}

func (uf *unionFind) MakeSet(v ast.Var) *unionFindRoot {

	root, ok := uf.Find(v)
	if ok {
		return root
	}

	root = newUnionFindRoot(v)
	uf.parents[v] = v
	uf.roots[v] = root
	return uf.roots[v]
}

func (uf *unionFind) Find(v ast.Var) (*unionFindRoot, bool) {

	parent, ok := uf.parents[v]
	if !ok {
		return nil, false
	}

	if parent == v {
		return uf.roots[v], true
	}

	return uf.Find(parent)
}

func (uf *unionFind) Merge(a, b ast.Var) (*unionFindRoot, bool) {

	r1 := uf.MakeSet(a)
	r2 := uf.MakeSet(b)

	if r1 != r2 {

		r1, r2 = uf.rank(r1, r2)

		uf.parents[r2.key] = r1.key
		delete(uf.roots, r2.key)

		// Sets can have at most one constant value associated with them. When
		// unioning, we must preserve this invariant. If a set has two constants,
		// there will be no way to prove the query.
		if r1.constant != nil && r2.constant != nil && !r1.constant.Equal(r2.constant) {
			return nil, false
		} else if r1.constant == nil {
			r1.constant = r2.constant
		}
	}

	return r1, true
}

type unionFindRoot struct {
	key      ast.Var
	constant *ast.Term
}

func newUnionFindRoot(key ast.Var) *unionFindRoot {
	return &unionFindRoot{
		key: key,
	}
}

func (r *unionFindRoot) Value() ast.Value {
	if r.constant != nil {
		return r.constant.Value
	}
	return r.key
}

func isNoop(expr *ast.Expr) bool {

	if !expr.IsCall() {
		term := expr.Terms.(*ast.Term)
		if !ast.IsConstant(term.Value) {
			return false
		}
		return !ast.Boolean(false).Equal(term.Value)
	}

	// A==A can be ignored
	if expr.Operator().Equal(ast.Equal.Ref()) {
		return expr.Operand(0).Equal(expr.Operand(1))
	}

	return false
}
