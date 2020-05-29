// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/open-policy-agent/opa/metrics"
	"github.com/open-policy-agent/opa/util"
)

// CompileErrorLimitDefault is the default number errors a compiler will allow before
// exiting.
const CompileErrorLimitDefault = 10

var errLimitReached = NewError(CompileErr, nil, "error limit reached")

// Compiler contains the state of a compilation process.
type Compiler struct {

	// Errors contains errors that occurred during the compilation process.
	// If there are one or more errors, the compilation process is considered
	// "failed".
	Errors Errors

	// Modules contains the compiled modules. The compiled modules are the
	// output of the compilation process. If the compilation process failed,
	// there is no guarantee about the state of the modules.
	Modules map[string]*Module

	// ModuleTree organizes the modules into a tree where each node is keyed by
	// an element in the module's package path. E.g., given modules containing
	// the following package directives: "a", "a.b", "a.c", and "a.b", the
	// resulting module tree would be:
	//
	//  root
	//    |
	//    +--- data (no modules)
	//           |
	//           +--- a (1 module)
	//                |
	//                +--- b (2 modules)
	//                |
	//                +--- c (1 module)
	//
	ModuleTree *ModuleTreeNode

	// RuleTree organizes rules into a tree where each node is keyed by an
	// element in the rule's path. The rule path is the concatenation of the
	// containing package and the stringified rule name. E.g., given the
	// following module:
	//
	//  package ex
	//  p[1] { true }
	//  p[2] { true }
	//  q = true
	//
	//  root
	//    |
	//    +--- data (no rules)
	//           |
	//           +--- ex (no rules)
	//                |
	//                +--- p (2 rules)
	//                |
	//                +--- q (1 rule)
	RuleTree *TreeNode

	// Graph contains dependencies between rules. An edge (u,v) is added to the
	// graph if rule 'u' refers to the virtual document defined by 'v'.
	Graph *Graph

	// TypeEnv holds type information for values inferred by the compiler.
	TypeEnv *TypeEnv

	// RewrittenVars is a mapping of variables that have been rewritten
	// with the key being the generated name and value being the original.
	RewrittenVars map[Var]Var

	localvargen  *localVarGenerator
	moduleLoader ModuleLoader
	ruleIndices  *util.HashMap
	stages       []struct {
		name       string
		metricName string
		f          func()
	}
	maxErrs           int
	sorted            []string // list of sorted module names
	pathExists        func([]string) (bool, error)
	after             map[string][]CompilerStageDefinition
	metrics           metrics.Metrics
	builtins          map[string]*Builtin
	unsafeBuiltinsMap map[string]struct{}
}

// CompilerStage defines the interface for stages in the compiler.
type CompilerStage func(*Compiler) *Error

// CompilerStageDefinition defines a compiler stage
type CompilerStageDefinition struct {
	Name       string
	MetricName string
	Stage      CompilerStage
}

// QueryContext contains contextual information for running an ad-hoc query.
//
// Ad-hoc queries can be run in the context of a package and imports may be
// included to provide concise access to data.
type QueryContext struct {
	Package *Package
	Imports []*Import
}

// NewQueryContext returns a new QueryContext object.
func NewQueryContext() *QueryContext {
	return &QueryContext{}
}

// WithPackage sets the pkg on qc.
func (qc *QueryContext) WithPackage(pkg *Package) *QueryContext {
	if qc == nil {
		qc = NewQueryContext()
	}
	qc.Package = pkg
	return qc
}

// WithImports sets the imports on qc.
func (qc *QueryContext) WithImports(imports []*Import) *QueryContext {
	if qc == nil {
		qc = NewQueryContext()
	}
	qc.Imports = imports
	return qc
}

// Copy returns a deep copy of qc.
func (qc *QueryContext) Copy() *QueryContext {
	if qc == nil {
		return nil
	}
	cpy := *qc
	if cpy.Package != nil {
		cpy.Package = qc.Package.Copy()
	}
	cpy.Imports = make([]*Import, len(qc.Imports))
	for i := range qc.Imports {
		cpy.Imports[i] = qc.Imports[i].Copy()
	}
	return &cpy
}

// QueryCompiler defines the interface for compiling ad-hoc queries.
type QueryCompiler interface {

	// Compile should be called to compile ad-hoc queries. The return value is
	// the compiled version of the query.
	Compile(q Body) (Body, error)

	// TypeEnv returns the type environment built after running type checking
	// on the query.
	TypeEnv() *TypeEnv

	// WithContext sets the QueryContext on the QueryCompiler. Subsequent calls
	// to Compile will take the QueryContext into account.
	WithContext(qctx *QueryContext) QueryCompiler

	// WithUnsafeBuiltins sets the built-in functions to treat as unsafe and not
	// allow inside of queries. By default the query compiler inherits the
	// compiler's unsafe built-in functions. This function allows callers to
	// override that set. If an empty (non-nil) map is provided, all built-ins
	// are allowed.
	WithUnsafeBuiltins(unsafe map[string]struct{}) QueryCompiler

	// WithStageAfter registers a stage to run during query compilation after
	// the named stage.
	WithStageAfter(after string, stage QueryCompilerStageDefinition) QueryCompiler

	// RewrittenVars maps generated vars in the compiled query to vars from the
	// parsed query. For example, given the query "input := 1" the rewritten
	// query would be "__local0__ = 1". The mapping would then be {__local0__: input}.
	RewrittenVars() map[Var]Var
}

// QueryCompilerStage defines the interface for stages in the query compiler.
type QueryCompilerStage func(QueryCompiler, Body) (Body, error)

// QueryCompilerStageDefinition defines a QueryCompiler stage
type QueryCompilerStageDefinition struct {
	Name       string
	MetricName string
	Stage      QueryCompilerStage
}

const compileStageMetricPrefex = "ast_compile_stage_"

// NewCompiler returns a new empty compiler.
func NewCompiler() *Compiler {

	c := &Compiler{
		Modules:       map[string]*Module{},
		TypeEnv:       NewTypeEnv(),
		RewrittenVars: map[Var]Var{},
		ruleIndices: util.NewHashMap(func(a, b util.T) bool {
			r1, r2 := a.(Ref), b.(Ref)
			return r1.Equal(r2)
		}, func(x util.T) int {
			return x.(Ref).Hash()
		}),
		maxErrs:           CompileErrorLimitDefault,
		after:             map[string][]CompilerStageDefinition{},
		unsafeBuiltinsMap: map[string]struct{}{},
	}

	c.ModuleTree = NewModuleTree(nil)
	c.RuleTree = NewRuleTree(c.ModuleTree)

	// Initialize the compiler with the statically compiled built-in functions.
	// If the caller customizes the compiler, a copy will be made.
	c.builtins = BuiltinMap
	checker := newTypeChecker()
	c.TypeEnv = checker.checkLanguageBuiltins(nil, c.builtins)

	c.stages = []struct {
		name       string
		metricName string
		f          func()
	}{
		// Reference resolution should run first as it may be used to lazily
		// load additional modules. If any stages run before resolution, they
		// need to be re-run after resolution.
		{"ResolveRefs", "compile_stage_resolve_refs", c.resolveAllRefs},

		// The local variable generator must be initialized after references are
		// resolved and the dynamic module loader has run but before subsequent
		// stages that need to generate variables.
		{"InitLocalVarGen", "compile_stage_init_local_var_gen", c.initLocalVarGen},

		{"RewriteLocalVars", "compile_stage_rewrite_local_vars", c.rewriteLocalVars},
		{"RewriteExprTerms", "compile_stage_rewrite_expr_terms", c.rewriteExprTerms},
		{"SetModuleTree", "compile_stage_set_module_tree", c.setModuleTree},
		{"SetRuleTree", "compile_stage_set_rule_tree", c.setRuleTree},
		{"SetGraph", "compile_stage_set_graph", c.setGraph},
		{"RewriteComprehensionTerms", "compile_stage_rewrite_comprehension_terms", c.rewriteComprehensionTerms},
		{"RewriteRefsInHead", "compile_stage_rewrite_refs_in_head", c.rewriteRefsInHead},
		{"RewriteWithValues", "compile_stage_rewrite_with_values", c.rewriteWithModifiers},
		{"CheckRuleConflicts", "compile_stage_check_rule_conflicts", c.checkRuleConflicts},
		{"CheckUndefinedFuncs", "compile_stage_check_undefined_funcs", c.checkUndefinedFuncs},
		{"CheckSafetyRuleHeads", "compile_stage_check_safety_rule_heads", c.checkSafetyRuleHeads},
		{"CheckSafetyRuleBodies", "compile_stage_check_safety_rule_bodies", c.checkSafetyRuleBodies},
		{"RewriteEquals", "compile_stage_rewrite_equals", c.rewriteEquals},
		{"RewriteDynamicTerms", "compile_stage_rewrite_dynamic_terms", c.rewriteDynamicTerms},
		{"CheckRecursion", "compile_stage_check_recursion", c.checkRecursion},
		{"CheckTypes", "compile_stage_check_types", c.checkTypes},
		{"CheckUnsafeBuiltins", "compile_state_check_unsafe_builtins", c.checkUnsafeBuiltins},
		{"BuildRuleIndices", "compile_stage_rebuild_indices", c.buildRuleIndices},
	}

	return c
}

// SetErrorLimit sets the number of errors the compiler can encounter before it
// quits. Zero or a negative number indicates no limit.
func (c *Compiler) SetErrorLimit(limit int) *Compiler {
	c.maxErrs = limit
	return c
}

// WithPathConflictsCheck enables base-virtual document conflict
// detection. The compiler will check that rules don't overlap with
// paths that exist as determined by the provided callable.
func (c *Compiler) WithPathConflictsCheck(fn func([]string) (bool, error)) *Compiler {
	c.pathExists = fn
	return c
}

// WithStageAfter registers a stage to run during compilation after
// the named stage.
func (c *Compiler) WithStageAfter(after string, stage CompilerStageDefinition) *Compiler {
	c.after[after] = append(c.after[after], stage)
	return c
}

// WithMetrics will set a metrics.Metrics and be used for profiling
// the Compiler instance.
func (c *Compiler) WithMetrics(metrics metrics.Metrics) *Compiler {
	c.metrics = metrics
	return c
}

// WithBuiltins adds a set of custom built-in functions to the compiler.
func (c *Compiler) WithBuiltins(builtins map[string]*Builtin) *Compiler {
	if len(builtins) == 0 {
		return c
	}
	cpy := make(map[string]*Builtin, len(c.builtins)+len(builtins))
	for k, v := range c.builtins {
		cpy[k] = v
	}
	for k, v := range builtins {
		cpy[k] = v
	}
	c.builtins = cpy
	// Build type env for custom functions and wrap existing one.
	checker := newTypeChecker()
	c.TypeEnv = checker.checkLanguageBuiltins(c.TypeEnv, builtins)
	return c
}

// WithUnsafeBuiltins will add all built-ins in the map to the "blacklist".
func (c *Compiler) WithUnsafeBuiltins(unsafeBuiltins map[string]struct{}) *Compiler {
	for name := range unsafeBuiltins {
		c.unsafeBuiltinsMap[name] = struct{}{}
	}
	return c
}

// QueryCompiler returns a new QueryCompiler object.
func (c *Compiler) QueryCompiler() QueryCompiler {
	return newQueryCompiler(c)
}

// Compile runs the compilation process on the input modules. The compiled
// version of the modules and associated data structures are stored on the
// compiler. If the compilation process fails for any reason, the compiler will
// contain a slice of errors.
func (c *Compiler) Compile(modules map[string]*Module) {

	c.Modules = make(map[string]*Module, len(modules))

	for k, v := range modules {
		c.Modules[k] = v.Copy()
		c.sorted = append(c.sorted, k)
	}

	sort.Strings(c.sorted)

	c.compile()
}

// Failed returns true if a compilation error has been encountered.
func (c *Compiler) Failed() bool {
	return len(c.Errors) > 0
}

// GetArity returns the number of args a function referred to by ref takes. If
// ref refers to built-in function, the built-in declaration is consulted,
// otherwise, the ref is used to perform a ruleset lookup.
func (c *Compiler) GetArity(ref Ref) int {
	if bi := c.builtins[ref.String()]; bi != nil {
		return len(bi.Decl.Args())
	}
	rules := c.GetRulesExact(ref)
	if len(rules) == 0 {
		return -1
	}
	return len(rules[0].Head.Args)
}

// GetRulesExact returns a slice of rules referred to by the reference.
//
// E.g., given the following module:
//
//	package a.b.c
//
//	p[k] = v { ... }    # rule1
//  p[k1] = v1 { ... }  # rule2
//
// The following calls yield the rules on the right.
//
//  GetRulesExact("data.a.b.c.p")   => [rule1, rule2]
//  GetRulesExact("data.a.b.c.p.x") => nil
//  GetRulesExact("data.a.b.c")     => nil
func (c *Compiler) GetRulesExact(ref Ref) (rules []*Rule) {
	node := c.RuleTree

	for _, x := range ref {
		if node = node.Child(x.Value); node == nil {
			return nil
		}
	}

	return extractRules(node.Values)
}

// GetRulesForVirtualDocument returns a slice of rules that produce the virtual
// document referred to by the reference.
//
// E.g., given the following module:
//
//	package a.b.c
//
//	p[k] = v { ... }    # rule1
//  p[k1] = v1 { ... }  # rule2
//
// The following calls yield the rules on the right.
//
//  GetRulesForVirtualDocument("data.a.b.c.p")   => [rule1, rule2]
//  GetRulesForVirtualDocument("data.a.b.c.p.x") => [rule1, rule2]
//  GetRulesForVirtualDocument("data.a.b.c")     => nil
func (c *Compiler) GetRulesForVirtualDocument(ref Ref) (rules []*Rule) {

	node := c.RuleTree

	for _, x := range ref {
		if node = node.Child(x.Value); node == nil {
			return nil
		}
		if len(node.Values) > 0 {
			return extractRules(node.Values)
		}
	}

	return extractRules(node.Values)
}

// GetRulesWithPrefix returns a slice of rules that share the prefix ref.
//
// E.g., given the following module:
//
//  package a.b.c
//
//  p[x] = y { ... }  # rule1
//  p[k] = v { ... }  # rule2
//  q { ... }         # rule3
//
// The following calls yield the rules on the right.
//
//  GetRulesWithPrefix("data.a.b.c.p")   => [rule1, rule2]
//  GetRulesWithPrefix("data.a.b.c.p.a") => nil
//  GetRulesWithPrefix("data.a.b.c")     => [rule1, rule2, rule3]
func (c *Compiler) GetRulesWithPrefix(ref Ref) (rules []*Rule) {

	node := c.RuleTree

	for _, x := range ref {
		if node = node.Child(x.Value); node == nil {
			return nil
		}
	}

	var acc func(node *TreeNode)

	acc = func(node *TreeNode) {
		rules = append(rules, extractRules(node.Values)...)
		for _, child := range node.Children {
			if child.Hide {
				continue
			}
			acc(child)
		}
	}

	acc(node)

	return rules
}

func extractRules(s []util.T) (rules []*Rule) {
	for _, r := range s {
		rules = append(rules, r.(*Rule))
	}
	return rules
}

// GetRules returns a slice of rules that are referred to by ref.
//
// E.g., given the following module:
//
//  package a.b.c
//
//  p[x] = y { q[x] = y; ... } # rule1
//  q[x] = y { ... }           # rule2
//
// The following calls yield the rules on the right.
//
//  GetRules("data.a.b.c.p")	=> [rule1]
//  GetRules("data.a.b.c.p.x")	=> [rule1]
//  GetRules("data.a.b.c.q")	=> [rule2]
//  GetRules("data.a.b.c")		=> [rule1, rule2]
//  GetRules("data.a.b.d")		=> nil
func (c *Compiler) GetRules(ref Ref) (rules []*Rule) {

	set := map[*Rule]struct{}{}

	for _, rule := range c.GetRulesForVirtualDocument(ref) {
		set[rule] = struct{}{}
	}

	for _, rule := range c.GetRulesWithPrefix(ref) {
		set[rule] = struct{}{}
	}

	for rule := range set {
		rules = append(rules, rule)
	}

	return rules
}

// GetRulesDynamic returns a slice of rules that could be referred to by a ref.
// When parts of the ref are statically known, we use that information to narrow
// down which rules the ref could refer to, but in the most general case this
// will be an over-approximation.
//
// E.g., given the following modules:
//
//  package a.b.c
//
//  r1 = 1  # rule1
//
// and:
//
//  package a.d.c
//
//  r2 = 2  # rule2
//
// The following calls yield the rules on the right.
//
//  GetRulesDynamic("data.a[x].c[y]") => [rule1, rule2]
//  GetRulesDynamic("data.a[x].c.r2") => [rule2]
//  GetRulesDynamic("data.a.b[x][y]") => [rule1]
func (c *Compiler) GetRulesDynamic(ref Ref) (rules []*Rule) {
	node := c.RuleTree

	set := map[*Rule]struct{}{}
	var walk func(node *TreeNode, i int)
	walk = func(node *TreeNode, i int) {
		if i >= len(ref) {
			// We've reached the end of the reference and want to collect everything
			// under this "prefix".
			node.DepthFirst(func(descendant *TreeNode) bool {
				insertRules(set, descendant.Values)
				return descendant.Hide
			})
		} else if i == 0 || IsConstant(ref[i].Value) {
			// The head of the ref is always grounded.  In case another part of the
			// ref is also grounded, we can lookup the exact child.  If it's not found
			// we can immediately return...
			if child := node.Child(ref[i].Value); child == nil {
				return
			} else if len(child.Values) > 0 {
				// If there are any rules at this position, it's what the ref would
				// refer to.  We can just append those and stop here.
				insertRules(set, child.Values)
			} else {
				// Otherwise, we continue using the child node.
				walk(child, i+1)
			}
		} else {
			// This part of the ref is a dynamic term.  We can't know what it refers
			// to and will just need to try all of the children.
			for _, child := range node.Children {
				if child.Hide {
					continue
				}
				insertRules(set, child.Values)
				walk(child, i+1)
			}
		}
	}

	walk(node, 0)
	for rule := range set {
		rules = append(rules, rule)
	}
	return rules
}

// Utility: add all rule values to the set.
func insertRules(set map[*Rule]struct{}, rules []util.T) {
	for _, rule := range rules {
		set[rule.(*Rule)] = struct{}{}
	}
}

// RuleIndex returns a RuleIndex built for the rule set referred to by path.
// The path must refer to the rule set exactly, i.e., given a rule set at path
// data.a.b.c.p, refs data.a.b.c.p.x and data.a.b.c would not return a
// RuleIndex built for the rule.
func (c *Compiler) RuleIndex(path Ref) RuleIndex {
	r, ok := c.ruleIndices.Get(path)
	if !ok {
		return nil
	}
	return r.(RuleIndex)
}

// ModuleLoader defines the interface that callers can implement to enable lazy
// loading of modules during compilation.
type ModuleLoader func(resolved map[string]*Module) (parsed map[string]*Module, err error)

// WithModuleLoader sets f as the ModuleLoader on the compiler.
//
// The compiler will invoke the ModuleLoader after resolving all references in
// the current set of input modules. The ModuleLoader can return a new
// collection of parsed modules that are to be included in the compilation
// process. This process will repeat until the ModuleLoader returns an empty
// collection or an error. If an error is returned, compilation will stop
// immediately.
func (c *Compiler) WithModuleLoader(f ModuleLoader) *Compiler {
	c.moduleLoader = f
	return c
}

// buildRuleIndices constructs indices for rules.
func (c *Compiler) buildRuleIndices() {

	c.RuleTree.DepthFirst(func(node *TreeNode) bool {
		if len(node.Values) == 0 {
			return false
		}
		index := newBaseDocEqIndex(func(ref Ref) bool {
			return isVirtual(c.RuleTree, ref.GroundPrefix())
		})
		if rules := extractRules(node.Values); index.Build(rules) {
			c.ruleIndices.Put(rules[0].Path(), index)
		}
		return false
	})

}

// checkRecursion ensures that there are no recursive definitions, i.e., there are
// no cycles in the Graph.
func (c *Compiler) checkRecursion() {
	eq := func(a, b util.T) bool {
		return a.(*Rule) == b.(*Rule)
	}

	c.RuleTree.DepthFirst(func(node *TreeNode) bool {
		for _, rule := range node.Values {
			for node := rule.(*Rule); node != nil; node = node.Else {
				c.checkSelfPath(node.Loc(), eq, node, node)
			}
		}
		return false
	})
}

func (c *Compiler) checkSelfPath(loc *Location, eq func(a, b util.T) bool, a, b util.T) {
	tr := NewGraphTraversal(c.Graph)
	if p := util.DFSPath(tr, eq, a, b); len(p) > 0 {
		n := []string{}
		for _, x := range p {
			n = append(n, astNodeToString(x))
		}
		c.err(NewError(RecursionErr, loc, "rule %v is recursive: %v", astNodeToString(a), strings.Join(n, " -> ")))
	}
}

func astNodeToString(x interface{}) string {
	switch x := x.(type) {
	case *Rule:
		return string(x.Head.Name)
	default:
		panic("not reached")
	}
}

// checkRuleConflicts ensures that rules definitions are not in conflict.
func (c *Compiler) checkRuleConflicts() {
	c.RuleTree.DepthFirst(func(node *TreeNode) bool {
		if len(node.Values) == 0 {
			return false
		}

		kinds := map[DocKind]struct{}{}
		defaultRules := 0
		arities := map[int]struct{}{}
		declared := false

		for _, rule := range node.Values {
			r := rule.(*Rule)
			kinds[r.Head.DocKind()] = struct{}{}
			arities[len(r.Head.Args)] = struct{}{}
			if r.Head.Assign {
				declared = true
			}
			if r.Default {
				defaultRules++
			}
		}

		name := Var(node.Key.(String))

		if declared && len(node.Values) > 1 {
			c.err(NewError(TypeErr, node.Values[0].(*Rule).Loc(), "rule named %v redeclared at %v", name, node.Values[1].(*Rule).Loc()))
		} else if len(kinds) > 1 || len(arities) > 1 {
			c.err(NewError(TypeErr, node.Values[0].(*Rule).Loc(), "conflicting rules named %v found", name))
		} else if defaultRules > 1 {
			c.err(NewError(TypeErr, node.Values[0].(*Rule).Loc(), "multiple default rules named %s found", name))
		}

		return false
	})

	if c.pathExists != nil {
		for _, err := range CheckPathConflicts(c, c.pathExists) {
			c.err(err)
		}
	}

	c.ModuleTree.DepthFirst(func(node *ModuleTreeNode) bool {
		for _, mod := range node.Modules {
			for _, rule := range mod.Rules {
				if childNode, ok := node.Children[String(rule.Head.Name)]; ok {
					for _, childMod := range childNode.Modules {
						msg := fmt.Sprintf("%v conflicts with rule defined at %v", childMod.Package, rule.Loc())
						c.err(NewError(TypeErr, mod.Package.Loc(), msg))
					}
				}
			}
		}
		return false
	})
}

func (c *Compiler) checkUndefinedFuncs() {
	for _, name := range c.sorted {
		m := c.Modules[name]
		for _, err := range checkUndefinedFuncs(m, c.GetArity) {
			c.err(err)
		}
	}
}

func checkUndefinedFuncs(x interface{}, arity func(Ref) int) Errors {

	var errs Errors

	WalkExprs(x, func(expr *Expr) bool {
		if !expr.IsCall() {
			return false
		}
		ref := expr.Operator()
		if arity(ref) >= 0 {
			return false
		}
		errs = append(errs, NewError(TypeErr, expr.Loc(), "undefined function %v", ref))
		return true
	})

	return errs
}

// checkSafetyRuleBodies ensures that variables appearing in negated expressions or non-target
// positions of built-in expressions will be bound when evaluating the rule from left
// to right, re-ordering as necessary.
func (c *Compiler) checkSafetyRuleBodies() {
	for _, name := range c.sorted {
		m := c.Modules[name]
		WalkRules(m, func(r *Rule) bool {
			safe := ReservedVars.Copy()
			safe.Update(r.Head.Args.Vars())
			r.Body = c.checkBodySafety(safe, m, r.Body)
			return false
		})
	}
}

func (c *Compiler) checkBodySafety(safe VarSet, m *Module, b Body) Body {
	reordered, unsafe := reorderBodyForSafety(c.builtins, c.GetArity, safe, b)
	if errs := safetyErrorSlice(unsafe); len(errs) > 0 {
		for _, err := range errs {
			c.err(err)
		}
		return b
	}
	return reordered
}

var safetyCheckVarVisitorParams = VarVisitorParams{
	SkipRefCallHead: true,
	SkipClosures:    true,
}

// checkSafetyRuleHeads ensures that variables appearing in the head of a
// rule also appear in the body.
func (c *Compiler) checkSafetyRuleHeads() {

	for _, name := range c.sorted {
		m := c.Modules[name]
		WalkRules(m, func(r *Rule) bool {
			safe := r.Body.Vars(safetyCheckVarVisitorParams)
			safe.Update(r.Head.Args.Vars())
			unsafe := r.Head.Vars().Diff(safe)
			for v := range unsafe {
				if !v.IsGenerated() {
					c.err(NewError(UnsafeVarErr, r.Loc(), "var %v is unsafe", v))
				}
			}
			return false
		})
	}
}

// checkTypes runs the type checker on all rules. The type checker builds a
// TypeEnv that is stored on the compiler.
func (c *Compiler) checkTypes() {
	// Recursion is caught in earlier step, so this cannot fail.
	sorted, _ := c.Graph.Sort()
	checker := newTypeChecker().WithVarRewriter(rewriteVarsInRef(c.RewrittenVars))
	env, errs := checker.CheckTypes(c.TypeEnv, sorted)
	for _, err := range errs {
		c.err(err)
	}
	c.TypeEnv = env
}

func (c *Compiler) checkUnsafeBuiltins() {
	for _, name := range c.sorted {
		errs := checkUnsafeBuiltins(c.unsafeBuiltinsMap, c.Modules[name])
		for _, err := range errs {
			c.err(err)
		}
	}
}

func (c *Compiler) runStage(metricName string, f func()) {
	if c.metrics != nil {
		c.metrics.Timer(metricName).Start()
		defer c.metrics.Timer(metricName).Stop()
	}
	f()
}

func (c *Compiler) runStageAfter(metricName string, s CompilerStage) *Error {
	if c.metrics != nil {
		c.metrics.Timer(metricName).Start()
		defer c.metrics.Timer(metricName).Stop()
	}
	return s(c)
}

func (c *Compiler) compile() {
	defer func() {
		if r := recover(); r != nil && r != errLimitReached {
			panic(r)
		}
	}()

	for _, s := range c.stages {
		c.runStage(s.metricName, s.f)
		if c.Failed() {
			return
		}
		for _, s := range c.after[s.name] {
			err := c.runStageAfter(s.MetricName, s.Stage)
			if err != nil {
				c.err(err)
			}
		}
	}
}

func (c *Compiler) err(err *Error) {
	if c.maxErrs > 0 && len(c.Errors) >= c.maxErrs {
		c.Errors = append(c.Errors, errLimitReached)
		panic(errLimitReached)
	}
	c.Errors = append(c.Errors, err)
}

func (c *Compiler) getExports() *util.HashMap {

	rules := util.NewHashMap(func(a, b util.T) bool {
		r1 := a.(Ref)
		r2 := a.(Ref)
		return r1.Equal(r2)
	}, func(v util.T) int {
		return v.(Ref).Hash()
	})

	for _, name := range c.sorted {
		mod := c.Modules[name]
		rv, ok := rules.Get(mod.Package.Path)
		if !ok {
			rv = []Var{}
		}
		rvs := rv.([]Var)

		for _, rule := range mod.Rules {
			rvs = append(rvs, rule.Head.Name)
		}
		rules.Put(mod.Package.Path, rvs)
	}

	return rules
}

// resolveAllRefs resolves references in expressions to their fully qualified values.
//
// For instance, given the following module:
//
// package a.b
// import data.foo.bar
// p[x] { bar[_] = x }
//
// The reference "bar[_]" would be resolved to "data.foo.bar[_]".
func (c *Compiler) resolveAllRefs() {

	rules := c.getExports()

	for _, name := range c.sorted {
		mod := c.Modules[name]

		var ruleExports []Var
		if x, ok := rules.Get(mod.Package.Path); ok {
			ruleExports = x.([]Var)
		}

		globals := getGlobals(mod.Package, ruleExports, mod.Imports)

		WalkRules(mod, func(rule *Rule) bool {
			err := resolveRefsInRule(globals, rule)
			if err != nil {
				c.err(NewError(CompileErr, rule.Location, err.Error()))
			}
			return false
		})

		// Once imports have been resolved, they are no longer needed.
		mod.Imports = nil
	}

	if c.moduleLoader != nil {

		parsed, err := c.moduleLoader(c.Modules)
		if err != nil {
			c.err(NewError(CompileErr, nil, err.Error()))
			return
		}

		if len(parsed) == 0 {
			return
		}

		for id, module := range parsed {
			c.Modules[id] = module.Copy()
			c.sorted = append(c.sorted, id)
		}

		sort.Strings(c.sorted)
		c.resolveAllRefs()
	}
}

func (c *Compiler) initLocalVarGen() {
	c.localvargen = newLocalVarGeneratorForModuleSet(c.sorted, c.Modules)
}

func (c *Compiler) rewriteComprehensionTerms() {
	f := newEqualityFactory(c.localvargen)
	for _, name := range c.sorted {
		mod := c.Modules[name]
		rewriteComprehensionTerms(f, mod)
	}
}

func (c *Compiler) rewriteExprTerms() {
	for _, name := range c.sorted {
		mod := c.Modules[name]
		WalkRules(mod, func(rule *Rule) bool {
			rewriteExprTermsInHead(c.localvargen, rule)
			rule.Body = rewriteExprTermsInBody(c.localvargen, rule.Body)
			return false
		})
	}
}

// rewriteTermsInHead will rewrite rules so that the head does not contain any
// terms that require evaluation (e.g., refs or comprehensions). If the key or
// value contains or more of these terms, the key or value will be moved into
// the body and assigned to a new variable. The new variable will replace the
// key or value in the head.
//
// For instance, given the following rule:
//
// p[{"foo": data.foo[i]}] { i < 100 }
//
// The rule would be re-written as:
//
// p[__local0__] { i < 100; __local0__ = {"foo": data.foo[i]} }
func (c *Compiler) rewriteRefsInHead() {
	f := newEqualityFactory(c.localvargen)
	for _, name := range c.sorted {
		mod := c.Modules[name]
		WalkRules(mod, func(rule *Rule) bool {
			if requiresEval(rule.Head.Key) {
				expr := f.Generate(rule.Head.Key)
				rule.Head.Key = expr.Operand(0)
				rule.Body.Append(expr)
			}
			if requiresEval(rule.Head.Value) {
				expr := f.Generate(rule.Head.Value)
				rule.Head.Value = expr.Operand(0)
				rule.Body.Append(expr)
			}
			for i := 0; i < len(rule.Head.Args); i++ {
				if requiresEval(rule.Head.Args[i]) {
					expr := f.Generate(rule.Head.Args[i])
					rule.Head.Args[i] = expr.Operand(0)
					rule.Body.Append(expr)
				}
			}
			return false
		})
	}
}

func (c *Compiler) rewriteEquals() {
	for _, name := range c.sorted {
		mod := c.Modules[name]
		rewriteEquals(mod)
	}
}

func (c *Compiler) rewriteDynamicTerms() {
	f := newEqualityFactory(c.localvargen)
	for _, name := range c.sorted {
		mod := c.Modules[name]
		WalkRules(mod, func(rule *Rule) bool {
			rule.Body = rewriteDynamics(f, rule.Body)
			return false
		})
	}
}

func (c *Compiler) rewriteLocalVars() {

	for _, name := range c.sorted {
		mod := c.Modules[name]
		gen := c.localvargen

		WalkRules(mod, func(rule *Rule) bool {

			var errs Errors

			// Rewrite assignments contained in head of rule. Assignments can
			// occur in rule head if they're inside a comprehension. Note,
			// assigned vars in comprehensions in the head will be rewritten
			// first to preserve scoping rules. For example:
			//
			// p = [x | x := 1] { x := 2 } becomes p = [__local0__ | __local0__ = 1] { __local1__ = 2 }
			//
			// This behaviour is consistent scoping inside the body. For example:
			//
			// p = xs { x := 2; xs = [x | x := 1] } becomes p = xs { __local0__ = 2; xs = [__local1__ | __local1__ = 1] }
			WalkTerms(rule.Head, func(term *Term) bool {
				stop := false
				stack := newLocalDeclaredVars()
				switch v := term.Value.(type) {
				case *ArrayComprehension:
					errs = rewriteDeclaredVarsInArrayComprehension(gen, stack, v, errs)
					stop = true
				case *SetComprehension:
					errs = rewriteDeclaredVarsInSetComprehension(gen, stack, v, errs)
					stop = true
				case *ObjectComprehension:
					errs = rewriteDeclaredVarsInObjectComprehension(gen, stack, v, errs)
					stop = true
				}

				for k, v := range stack.rewritten {
					c.RewrittenVars[k] = v
				}

				return stop
			})

			for _, err := range errs {
				c.err(err)
			}

			// Rewrite assignments in body.
			used := NewVarSet()

			if rule.Head.Key != nil {
				used.Update(rule.Head.Key.Vars())
			}

			if rule.Head.Value != nil {
				used.Update(rule.Head.Value.Vars())
			}

			stack := newLocalDeclaredVars()

			c.rewriteLocalArgVars(gen, stack, rule)

			body, declared, errs := rewriteLocalVars(gen, stack, used, rule.Body)
			for _, err := range errs {
				c.err(err)
			}

			// For rewritten vars use the collection of all variables that
			// were in the stack at some point in time.
			for k, v := range stack.rewritten {
				c.RewrittenVars[k] = v
			}

			rule.Body = body

			// Rewrite vars in head that refer to locally declared vars in the body.
			vis := NewGenericVisitor(func(x interface{}) bool {

				term, ok := x.(*Term)
				if !ok {
					return false
				}

				switch v := term.Value.(type) {
				case Object:
					// Make a copy of the object because the keys may be mutated.
					cpy, _ := v.Map(func(k, v *Term) (*Term, *Term, error) {
						if vark, ok := k.Value.(Var); ok {
							if gv, ok := declared[vark]; ok {
								k = k.Copy()
								k.Value = gv
							}
						}
						return k, v, nil
					})
					term.Value = cpy
				case Var:
					if gv, ok := declared[v]; ok {
						term.Value = gv
						return true
					}
				}

				return false
			})

			vis.Walk(rule.Head.Args)

			if rule.Head.Key != nil {
				vis.Walk(rule.Head.Key)
			}

			if rule.Head.Value != nil {
				vis.Walk(rule.Head.Value)
			}

			return false
		})
	}
}

func (c *Compiler) rewriteLocalArgVars(gen *localVarGenerator, stack *localDeclaredVars, rule *Rule) {

	vis := &ruleArgLocalRewriter{
		stack: stack,
		gen:   gen,
	}

	for i := range rule.Head.Args {
		Walk(vis, rule.Head.Args[i])
	}

	for i := range vis.errs {
		c.err(vis.errs[i])
	}
}

type ruleArgLocalRewriter struct {
	stack *localDeclaredVars
	gen   *localVarGenerator
	errs  []*Error
}

func (vis *ruleArgLocalRewriter) Visit(x interface{}) Visitor {

	t, ok := x.(*Term)
	if !ok {
		return vis
	}

	switch v := t.Value.(type) {
	case Var:
		gv, ok := vis.stack.Declared(v)
		if !ok {
			gv = vis.gen.Generate()
			vis.stack.Insert(v, gv, argVar)
		}
		t.Value = gv
		return nil
	case Object:
		if cpy, err := v.Map(func(k, v *Term) (*Term, *Term, error) {
			vcpy := v.Copy()
			Walk(vis, vcpy)
			return k, vcpy, nil
		}); err != nil {
			vis.errs = append(vis.errs, NewError(CompileErr, t.Location, err.Error()))
		} else {
			t.Value = cpy
		}
		return nil
	case Null, Boolean, Number, String, *ArrayComprehension, *SetComprehension, *ObjectComprehension, Set:
		// Scalars are no-ops. Comprehensions are handled above. Sets must not
		// contain variables.
		return nil
	default:
		// Recurse on refs, arrays, and calls. Any embedded
		// variables can be rewritten.
		return vis
	}
}

func (c *Compiler) rewriteWithModifiers() {
	f := newEqualityFactory(c.localvargen)
	for _, name := range c.sorted {
		mod := c.Modules[name]
		t := NewGenericTransformer(func(x interface{}) (interface{}, error) {
			body, ok := x.(Body)
			if !ok {
				return x, nil
			}
			body, err := rewriteWithModifiersInBody(c, f, body)
			if err != nil {
				c.err(err)
			}

			return body, nil
		})
		Transform(t, mod)
	}
}

func (c *Compiler) setModuleTree() {
	c.ModuleTree = NewModuleTree(c.Modules)
}

func (c *Compiler) setRuleTree() {
	c.RuleTree = NewRuleTree(c.ModuleTree)
}

func (c *Compiler) setGraph() {
	c.Graph = NewGraph(c.Modules, c.GetRulesDynamic)
}

type queryCompiler struct {
	compiler       *Compiler
	qctx           *QueryContext
	typeEnv        *TypeEnv
	rewritten      map[Var]Var
	after          map[string][]QueryCompilerStageDefinition
	unsafeBuiltins map[string]struct{}
}

func newQueryCompiler(compiler *Compiler) QueryCompiler {
	qc := &queryCompiler{
		compiler: compiler,
		qctx:     nil,
		after:    map[string][]QueryCompilerStageDefinition{},
	}
	return qc
}

func (qc *queryCompiler) WithContext(qctx *QueryContext) QueryCompiler {
	qc.qctx = qctx
	return qc
}

func (qc *queryCompiler) WithStageAfter(after string, stage QueryCompilerStageDefinition) QueryCompiler {
	qc.after[after] = append(qc.after[after], stage)
	return qc
}

func (qc *queryCompiler) WithUnsafeBuiltins(unsafe map[string]struct{}) QueryCompiler {
	qc.unsafeBuiltins = unsafe
	return qc
}

func (qc *queryCompiler) RewrittenVars() map[Var]Var {
	return qc.rewritten
}

func (qc *queryCompiler) runStage(metricName string, qctx *QueryContext, query Body, s func(*QueryContext, Body) (Body, error)) (Body, error) {
	if qc.compiler.metrics != nil {
		qc.compiler.metrics.Timer(metricName).Start()
		defer qc.compiler.metrics.Timer(metricName).Stop()
	}
	return s(qctx, query)
}

func (qc *queryCompiler) runStageAfter(metricName string, query Body, s QueryCompilerStage) (Body, error) {
	if qc.compiler.metrics != nil {
		qc.compiler.metrics.Timer(metricName).Start()
		defer qc.compiler.metrics.Timer(metricName).Stop()
	}
	return s(qc, query)
}

func (qc *queryCompiler) Compile(query Body) (Body, error) {

	query = query.Copy()

	stages := []struct {
		name       string
		metricName string
		f          func(*QueryContext, Body) (Body, error)
	}{
		{"ResolveRefs", "query_compile_stage_resolve_refs", qc.resolveRefs},
		{"RewriteLocalVars", "query_compile_stage_rewrite_local_vars", qc.rewriteLocalVars},
		{"RewriteExprTerms", "query_compile_stage_rewrite_expr_terms", qc.rewriteExprTerms},
		{"RewriteComprehensionTerms", "query_compile_stage_rewrite_comprehension_terms", qc.rewriteComprehensionTerms},
		{"RewriteWithValues", "query_compile_stage_rewrite_with_values", qc.rewriteWithModifiers},
		{"CheckUndefinedFuncs", "query_compile_stage_check_undefined_funcs", qc.checkUndefinedFuncs},
		{"CheckSafety", "query_compile_stage_check_safety", qc.checkSafety},
		{"RewriteDynamicTerms", "query_compile_stage_rewrite_dynamic_terms", qc.rewriteDynamicTerms},
		{"CheckTypes", "query_compile_stage_check_types", qc.checkTypes},
		{"CheckUnsafeBuiltins", "query_compile_stage_check_unsafe_builtins", qc.checkUnsafeBuiltins},
	}

	qctx := qc.qctx.Copy()

	for _, s := range stages {
		var err error
		query, err = qc.runStage(s.metricName, qctx, query, s.f)
		if err != nil {
			return nil, qc.applyErrorLimit(err)
		}
		for _, s := range qc.after[s.name] {
			query, err = qc.runStageAfter(s.MetricName, query, s.Stage)
			if err != nil {
				return nil, qc.applyErrorLimit(err)
			}
		}
	}

	return query, nil
}

func (qc *queryCompiler) TypeEnv() *TypeEnv {
	return qc.typeEnv
}

func (qc *queryCompiler) applyErrorLimit(err error) error {
	if errs, ok := err.(Errors); ok {
		if qc.compiler.maxErrs > 0 && len(errs) > qc.compiler.maxErrs {
			err = append(errs[:qc.compiler.maxErrs], errLimitReached)
		}
	}
	return err
}

func (qc *queryCompiler) resolveRefs(qctx *QueryContext, body Body) (Body, error) {

	var globals map[Var]Ref

	if qctx != nil && qctx.Package != nil {
		var ruleExports []Var
		rules := qc.compiler.getExports()
		if exist, ok := rules.Get(qctx.Package.Path); ok {
			ruleExports = exist.([]Var)
		}

		globals = getGlobals(qctx.Package, ruleExports, qc.qctx.Imports)
		qctx.Imports = nil
	}

	ignore := &declaredVarStack{declaredVars(body)}

	return resolveRefsInBody(globals, ignore, body), nil
}

func (qc *queryCompiler) rewriteComprehensionTerms(_ *QueryContext, body Body) (Body, error) {
	gen := newLocalVarGenerator("q", body)
	f := newEqualityFactory(gen)
	node, err := rewriteComprehensionTerms(f, body)
	if err != nil {
		return nil, err
	}
	return node.(Body), nil
}

func (qc *queryCompiler) rewriteDynamicTerms(_ *QueryContext, body Body) (Body, error) {
	gen := newLocalVarGenerator("q", body)
	f := newEqualityFactory(gen)
	return rewriteDynamics(f, body), nil
}

func (qc *queryCompiler) rewriteExprTerms(_ *QueryContext, body Body) (Body, error) {
	gen := newLocalVarGenerator("q", body)
	return rewriteExprTermsInBody(gen, body), nil
}

func (qc *queryCompiler) rewriteLocalVars(_ *QueryContext, body Body) (Body, error) {
	gen := newLocalVarGenerator("q", body)
	stack := newLocalDeclaredVars()
	body, _, err := rewriteLocalVars(gen, stack, nil, body)
	if len(err) != 0 {
		return nil, err
	}
	qc.rewritten = make(map[Var]Var, len(stack.rewritten))
	for k, v := range stack.rewritten {
		// The vars returned during the rewrite will include all seen vars,
		// even if they're not declared with an assignment operation. We don't
		// want to include these inside the rewritten set though.
		qc.rewritten[k] = v
	}
	return body, nil
}

func (qc *queryCompiler) checkUndefinedFuncs(_ *QueryContext, body Body) (Body, error) {
	if errs := checkUndefinedFuncs(body, qc.compiler.GetArity); len(errs) > 0 {
		return nil, errs
	}
	return body, nil
}

func (qc *queryCompiler) checkSafety(_ *QueryContext, body Body) (Body, error) {
	safe := ReservedVars.Copy()
	reordered, unsafe := reorderBodyForSafety(qc.compiler.builtins, qc.compiler.GetArity, safe, body)
	if errs := safetyErrorSlice(unsafe); len(errs) > 0 {
		return nil, errs
	}
	return reordered, nil
}

func (qc *queryCompiler) checkTypes(qctx *QueryContext, body Body) (Body, error) {
	var errs Errors
	checker := newTypeChecker().WithVarRewriter(rewriteVarsInRef(qc.rewritten, qc.compiler.RewrittenVars))
	qc.typeEnv, errs = checker.CheckBody(qc.compiler.TypeEnv, body)
	if len(errs) > 0 {
		return nil, errs
	}
	return body, nil
}

func (qc *queryCompiler) checkUnsafeBuiltins(qctx *QueryContext, body Body) (Body, error) {
	var unsafe map[string]struct{}
	if qc.unsafeBuiltins != nil {
		unsafe = qc.unsafeBuiltins
	} else {
		unsafe = qc.compiler.unsafeBuiltinsMap
	}
	errs := checkUnsafeBuiltins(unsafe, body)
	if len(errs) > 0 {
		return nil, errs
	}
	return body, nil
}

func (qc *queryCompiler) rewriteWithModifiers(qctx *QueryContext, body Body) (Body, error) {
	f := newEqualityFactory(newLocalVarGenerator("q", body))
	body, err := rewriteWithModifiersInBody(qc.compiler, f, body)
	if err != nil {
		return nil, Errors{err}
	}
	return body, nil
}

// ModuleTreeNode represents a node in the module tree. The module
// tree is keyed by the package path.
type ModuleTreeNode struct {
	Key      Value
	Modules  []*Module
	Children map[Value]*ModuleTreeNode
	Hide     bool
}

// NewModuleTree returns a new ModuleTreeNode that represents the root
// of the module tree populated with the given modules.
func NewModuleTree(mods map[string]*Module) *ModuleTreeNode {
	root := &ModuleTreeNode{
		Children: map[Value]*ModuleTreeNode{},
	}
	for _, m := range mods {
		node := root
		for i, x := range m.Package.Path {
			c, ok := node.Children[x.Value]
			if !ok {
				var hide bool
				if i == 1 && x.Value.Compare(SystemDocumentKey) == 0 {
					hide = true
				}
				c = &ModuleTreeNode{
					Key:      x.Value,
					Children: map[Value]*ModuleTreeNode{},
					Hide:     hide,
				}
				node.Children[x.Value] = c
			}
			node = c
		}
		node.Modules = append(node.Modules, m)
	}
	return root
}

// Size returns the number of modules in the tree.
func (n *ModuleTreeNode) Size() int {
	s := len(n.Modules)
	for _, c := range n.Children {
		s += c.Size()
	}
	return s
}

// DepthFirst performs a depth-first traversal of the module tree rooted at n.
// If f returns true, traversal will not continue to the children of n.
func (n *ModuleTreeNode) DepthFirst(f func(node *ModuleTreeNode) bool) {
	if !f(n) {
		for _, node := range n.Children {
			node.DepthFirst(f)
		}
	}
}

// TreeNode represents a node in the rule tree. The rule tree is keyed by
// rule path.
type TreeNode struct {
	Key      Value
	Values   []util.T
	Children map[Value]*TreeNode
	Hide     bool
}

// NewRuleTree returns a new TreeNode that represents the root
// of the rule tree populated with the given rules.
func NewRuleTree(mtree *ModuleTreeNode) *TreeNode {

	ruleSets := map[String][]util.T{}

	// Build rule sets for this package.
	for _, mod := range mtree.Modules {
		for _, rule := range mod.Rules {
			key := String(rule.Head.Name)
			ruleSets[key] = append(ruleSets[key], rule)
		}
	}

	// Each rule set becomes a leaf node.
	children := map[Value]*TreeNode{}

	for key, rules := range ruleSets {
		children[key] = &TreeNode{
			Key:      key,
			Children: nil,
			Values:   rules,
		}
	}

	// Each module in subpackage becomes child node.
	for _, child := range mtree.Children {
		children[child.Key] = NewRuleTree(child)
	}

	return &TreeNode{
		Key:      mtree.Key,
		Values:   nil,
		Children: children,
		Hide:     mtree.Hide,
	}
}

// Size returns the number of rules in the tree.
func (n *TreeNode) Size() int {
	s := len(n.Values)
	for _, c := range n.Children {
		s += c.Size()
	}
	return s
}

// Child returns n's child with key k.
func (n *TreeNode) Child(k Value) *TreeNode {
	switch k.(type) {
	case String, Var:
		return n.Children[k]
	}
	return nil
}

// DepthFirst performs a depth-first traversal of the rule tree rooted at n. If
// f returns true, traversal will not continue to the children of n.
func (n *TreeNode) DepthFirst(f func(node *TreeNode) bool) {
	if !f(n) {
		for _, node := range n.Children {
			node.DepthFirst(f)
		}
	}
}

// Graph represents the graph of dependencies between rules.
type Graph struct {
	adj    map[util.T]map[util.T]struct{}
	nodes  map[util.T]struct{}
	sorted []util.T
}

// NewGraph returns a new Graph based on modules. The list function must return
// the rules referred to directly by the ref.
func NewGraph(modules map[string]*Module, list func(Ref) []*Rule) *Graph {

	graph := &Graph{
		adj:    map[util.T]map[util.T]struct{}{},
		nodes:  map[util.T]struct{}{},
		sorted: nil,
	}

	// Create visitor to walk a rule AST and add edges to the rule graph for
	// each dependency.
	vis := func(a *Rule) *GenericVisitor {
		stop := false
		return NewGenericVisitor(func(x interface{}) bool {
			switch x := x.(type) {
			case Ref:
				for _, b := range list(x) {
					for node := b; node != nil; node = node.Else {
						graph.addDependency(a, node)
					}
				}
			case *Rule:
				if stop {
					// Do not recurse into else clauses (which will be handled
					// by the outer visitor.)
					return true
				}
				stop = true
			}
			return false
		})
	}

	// Walk over all rules, add them to graph, and build adjencency lists.
	for _, module := range modules {
		WalkRules(module, func(a *Rule) bool {
			graph.addNode(a)
			vis(a).Walk(a)
			return false
		})
	}

	return graph
}

// Dependencies returns the set of rules that x depends on.
func (g *Graph) Dependencies(x util.T) map[util.T]struct{} {
	return g.adj[x]
}

// Sort returns a slice of rules sorted by dependencies. If a cycle is found,
// ok is set to false.
func (g *Graph) Sort() (sorted []util.T, ok bool) {
	if g.sorted != nil {
		return g.sorted, true
	}

	sort := &graphSort{
		sorted: make([]util.T, 0, len(g.nodes)),
		deps:   g.Dependencies,
		marked: map[util.T]struct{}{},
		temp:   map[util.T]struct{}{},
	}

	for node := range g.nodes {
		if !sort.Visit(node) {
			return nil, false
		}
	}

	g.sorted = sort.sorted
	return g.sorted, true
}

func (g *Graph) addDependency(u util.T, v util.T) {

	if _, ok := g.nodes[u]; !ok {
		g.addNode(u)
	}

	if _, ok := g.nodes[v]; !ok {
		g.addNode(v)
	}

	edges, ok := g.adj[u]
	if !ok {
		edges = map[util.T]struct{}{}
		g.adj[u] = edges
	}

	edges[v] = struct{}{}
}

func (g *Graph) addNode(n util.T) {
	g.nodes[n] = struct{}{}
}

type graphSort struct {
	sorted []util.T
	deps   func(util.T) map[util.T]struct{}
	marked map[util.T]struct{}
	temp   map[util.T]struct{}
}

func (sort *graphSort) Marked(node util.T) bool {
	_, marked := sort.marked[node]
	return marked
}

func (sort *graphSort) Visit(node util.T) (ok bool) {
	if _, ok := sort.temp[node]; ok {
		return false
	}
	if sort.Marked(node) {
		return true
	}
	sort.temp[node] = struct{}{}
	for other := range sort.deps(node) {
		if !sort.Visit(other) {
			return false
		}
	}
	sort.marked[node] = struct{}{}
	delete(sort.temp, node)
	sort.sorted = append(sort.sorted, node)
	return true
}

// GraphTraversal is a Traversal that understands the dependency graph
type GraphTraversal struct {
	graph   *Graph
	visited map[util.T]struct{}
}

// NewGraphTraversal returns a Traversal for the dependency graph
func NewGraphTraversal(graph *Graph) *GraphTraversal {
	return &GraphTraversal{
		graph:   graph,
		visited: map[util.T]struct{}{},
	}
}

// Edges lists all dependency connections for a given node
func (g *GraphTraversal) Edges(x util.T) []util.T {
	r := []util.T{}
	for v := range g.graph.Dependencies(x) {
		r = append(r, v)
	}
	return r
}

// Visited returns whether a node has been visited, setting a node to visited if not
func (g *GraphTraversal) Visited(u util.T) bool {
	_, ok := g.visited[u]
	g.visited[u] = struct{}{}
	return ok
}

type unsafePair struct {
	Expr *Expr
	Vars VarSet
}

type unsafeVarLoc struct {
	Var Var
	Loc *Location
}

type unsafeVars map[*Expr]VarSet

func (vs unsafeVars) Add(e *Expr, v Var) {
	if u, ok := vs[e]; ok {
		u[v] = struct{}{}
	} else {
		vs[e] = VarSet{v: struct{}{}}
	}
}

func (vs unsafeVars) Set(e *Expr, s VarSet) {
	vs[e] = s
}

func (vs unsafeVars) Update(o unsafeVars) {
	for k, v := range o {
		if _, ok := vs[k]; !ok {
			vs[k] = VarSet{}
		}
		vs[k].Update(v)
	}
}

func (vs unsafeVars) Vars() (result []unsafeVarLoc) {

	locs := map[Var]*Location{}

	// If var appears in multiple sets then pick first by location.
	for expr, vars := range vs {
		for v := range vars {
			if locs[v].Compare(expr.Location) > 0 {
				locs[v] = expr.Location
			}
		}
	}

	for v, loc := range locs {
		result = append(result, unsafeVarLoc{
			Var: v,
			Loc: loc,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Loc.Compare(result[j].Loc) < 0
	})

	return result
}

func (vs unsafeVars) Slice() (result []unsafePair) {
	for expr, vs := range vs {
		result = append(result, unsafePair{
			Expr: expr,
			Vars: vs,
		})
	}
	return
}

// reorderBodyForSafety returns a copy of the body ordered such that
// left to right evaluation of the body will not encounter unbound variables
// in input positions or negated expressions.
//
// Expressions are added to the re-ordered body as soon as they are considered
// safe. If multiple expressions become safe in the same pass, they are added
// in their original order. This results in minimal re-ordering of the body.
//
// If the body cannot be reordered to ensure safety, the second return value
// contains a mapping of expressions to unsafe variables in those expressions.
func reorderBodyForSafety(builtins map[string]*Builtin, arity func(Ref) int, globals VarSet, body Body) (Body, unsafeVars) {

	body, unsafe := reorderBodyForClosures(builtins, arity, globals, body)
	if len(unsafe) != 0 {
		return nil, unsafe
	}

	reordered := Body{}
	safe := VarSet{}

	for _, e := range body {
		for v := range e.Vars(safetyCheckVarVisitorParams) {
			if globals.Contains(v) {
				safe.Add(v)
			} else {
				unsafe.Add(e, v)
			}
		}
	}

	for {
		n := len(reordered)

		for _, e := range body {
			if reordered.Contains(e) {
				continue
			}

			safe.Update(outputVarsForExpr(e, builtins, arity, safe))

			for v := range unsafe[e] {
				if safe.Contains(v) {
					delete(unsafe[e], v)
				}
			}

			if len(unsafe[e]) == 0 {
				delete(unsafe, e)
				reordered = append(reordered, e)
			}
		}

		if len(reordered) == n {
			break
		}
	}

	// Recursively visit closures and perform the safety checks on them.
	// Update the globals at each expression to include the variables that could
	// be closed over.
	g := globals.Copy()
	for i, e := range reordered {
		if i > 0 {
			g.Update(reordered[i-1].Vars(safetyCheckVarVisitorParams))
		}
		vis := &bodySafetyVisitor{
			builtins: builtins,
			arity:    arity,
			current:  e,
			globals:  g,
			unsafe:   unsafe,
		}
		NewGenericVisitor(vis.Visit).Walk(e)
	}

	// Need to reset expression indices as re-ordering may have
	// changed them.
	setExprIndices(reordered)

	return reordered, unsafe
}

type bodySafetyVisitor struct {
	builtins map[string]*Builtin
	arity    func(Ref) int
	current  *Expr
	globals  VarSet
	unsafe   unsafeVars
}

func (vis *bodySafetyVisitor) Visit(x interface{}) bool {
	switch x := x.(type) {
	case *Expr:
		cpy := *vis
		cpy.current = x

		switch ts := x.Terms.(type) {
		case *SomeDecl:
			NewGenericVisitor(cpy.Visit).Walk(ts)
		case []*Term:
			for _, t := range ts {
				NewGenericVisitor(cpy.Visit).Walk(t)
			}
		case *Term:
			NewGenericVisitor(cpy.Visit).Walk(ts)
		}
		for i := range x.With {
			NewGenericVisitor(cpy.Visit).Walk(x.With[i])
		}
		return true
	case *ArrayComprehension:
		vis.checkArrayComprehensionSafety(x)
		return true
	case *ObjectComprehension:
		vis.checkObjectComprehensionSafety(x)
		return true
	case *SetComprehension:
		vis.checkSetComprehensionSafety(x)
		return true
	}
	return false
}

// Check term for safety. This is analogous to the rule head safety check.
func (vis *bodySafetyVisitor) checkComprehensionSafety(tv VarSet, body Body) Body {
	bv := body.Vars(safetyCheckVarVisitorParams)
	bv.Update(vis.globals)
	uv := tv.Diff(bv)
	for v := range uv {
		vis.unsafe.Add(vis.current, v)
	}

	// Check body for safety, reordering as necessary.
	r, u := reorderBodyForSafety(vis.builtins, vis.arity, vis.globals, body)
	if len(u) == 0 {
		return r
	}

	vis.unsafe.Update(u)
	return body
}

func (vis *bodySafetyVisitor) checkArrayComprehensionSafety(ac *ArrayComprehension) {
	ac.Body = vis.checkComprehensionSafety(ac.Term.Vars(), ac.Body)
}

func (vis *bodySafetyVisitor) checkObjectComprehensionSafety(oc *ObjectComprehension) {
	tv := oc.Key.Vars()
	tv.Update(oc.Value.Vars())
	oc.Body = vis.checkComprehensionSafety(tv, oc.Body)
}

func (vis *bodySafetyVisitor) checkSetComprehensionSafety(sc *SetComprehension) {
	sc.Body = vis.checkComprehensionSafety(sc.Term.Vars(), sc.Body)
}

// reorderBodyForClosures returns a copy of the body ordered such that
// expressions (such as array comprehensions) that close over variables are ordered
// after other expressions that contain the same variable in an output position.
func reorderBodyForClosures(builtins map[string]*Builtin, arity func(Ref) int, globals VarSet, body Body) (Body, unsafeVars) {

	reordered := Body{}
	unsafe := unsafeVars{}

	for {
		n := len(reordered)

		for _, e := range body {
			if reordered.Contains(e) {
				continue
			}

			// Collect vars that are contained in closures within this
			// expression.
			vs := VarSet{}
			WalkClosures(e, func(x interface{}) bool {
				vis := &VarVisitor{vars: vs}
				vis.Walk(x)
				return true
			})

			// Compute vars that are closed over from the body but not yet
			// contained in the output position of an expression in the reordered
			// body. These vars are considered unsafe.
			cv := vs.Intersect(body.Vars(safetyCheckVarVisitorParams)).Diff(globals)
			uv := cv.Diff(outputVarsForBody(reordered, builtins, arity, globals))

			if len(uv) == 0 {
				reordered = append(reordered, e)
				delete(unsafe, e)
			} else {
				unsafe.Set(e, uv)
			}
		}

		if len(reordered) == n {
			break
		}
	}

	return reordered, unsafe
}

func outputVarsForBody(body Body, builtins map[string]*Builtin, arity func(Ref) int, safe VarSet) VarSet {
	o := safe.Copy()
	for _, e := range body {
		o.Update(outputVarsForExpr(e, builtins, arity, o))
	}
	return o.Diff(safe)
}

func outputVarsForExpr(expr *Expr, builtins map[string]*Builtin, arity func(Ref) int, safe VarSet) VarSet {

	// Negated expressions must be safe.
	if expr.Negated {
		return VarSet{}
	}

	// With modifier inputs must be safe.
	for _, with := range expr.With {
		unsafe := false
		WalkVars(with, func(v Var) bool {
			if !safe.Contains(v) {
				unsafe = true
				return true
			}
			return false
		})
		if unsafe {
			return VarSet{}
		}
	}

	if !expr.IsCall() {
		return outputVarsForExprRefs(expr, safe)
	}

	terms := expr.Terms.([]*Term)
	name := terms[0].String()

	if b := builtins[name]; b != nil {
		if b.Name == Equality.Name {
			return outputVarsForExprEq(expr, safe)
		}
		return outputVarsForExprBuiltin(expr, b, safe)
	}

	return outputVarsForExprCall(expr, builtins, arity, safe, terms)
}

func outputVarsForExprBuiltin(expr *Expr, b *Builtin, safe VarSet) VarSet {

	output := outputVarsForExprRefs(expr, safe)
	terms := expr.Terms.([]*Term)

	// Check that all input terms are safe.
	for i, t := range terms[1:] {
		if b.IsTargetPos(i) {
			continue
		}
		vis := NewVarVisitor().WithParams(VarVisitorParams{
			SkipClosures:   true,
			SkipSets:       true,
			SkipObjectKeys: true,
			SkipRefHead:    true,
		})
		vis.Walk(t)
		unsafe := vis.Vars().Diff(output).Diff(safe)
		if len(unsafe) > 0 {
			return VarSet{}
		}
	}

	// Add vars in target positions to result.
	for i, t := range terms[1:] {
		if b.IsTargetPos(i) {
			vis := NewVarVisitor().WithParams(VarVisitorParams{
				SkipRefHead:    true,
				SkipSets:       true,
				SkipObjectKeys: true,
				SkipClosures:   true,
			})
			vis.Walk(t)
			output.Update(vis.vars)
		}
	}

	return output
}

func outputVarsForExprEq(expr *Expr, safe VarSet) VarSet {
	if !validEqAssignArgCount(expr) {
		return safe
	}
	output := outputVarsForExprRefs(expr, safe)
	output.Update(safe)
	output.Update(Unify(output, expr.Operand(0), expr.Operand(1)))
	return output.Diff(safe)
}

func outputVarsForExprCall(expr *Expr, builtins map[string]*Builtin, arity func(Ref) int, safe VarSet, terms []*Term) VarSet {

	output := outputVarsForExprRefs(expr, safe)

	ref, ok := terms[0].Value.(Ref)
	if !ok {
		return VarSet{}
	}

	numArgs := arity(ref)
	if numArgs == -1 {
		return VarSet{}
	}

	numInputTerms := numArgs + 1

	if numInputTerms >= len(terms) {
		return output
	}

	vis := NewVarVisitor().WithParams(VarVisitorParams{
		SkipClosures:   true,
		SkipSets:       true,
		SkipObjectKeys: true,
		SkipRefHead:    true,
	})

	vis.Walk(Args(terms[:numInputTerms]))
	unsafe := vis.Vars().Diff(output).Diff(safe)

	if len(unsafe) > 0 {
		return VarSet{}
	}

	vis = NewVarVisitor().WithParams(VarVisitorParams{
		SkipRefHead:    true,
		SkipSets:       true,
		SkipObjectKeys: true,
		SkipClosures:   true,
	})

	vis.Walk(Args(terms[numInputTerms:]))
	output.Update(vis.vars)
	return output
}

func outputVarsForExprRefs(expr *Expr, safe VarSet) VarSet {
	output := VarSet{}
	WalkRefs(expr, func(r Ref) bool {
		if safe.Contains(r[0].Value.(Var)) {
			output.Update(r.OutputVars())
			return false
		}
		return true
	})
	return output
}

type equalityFactory struct {
	gen *localVarGenerator
}

func newEqualityFactory(gen *localVarGenerator) *equalityFactory {
	return &equalityFactory{gen}
}

func (f *equalityFactory) Generate(other *Term) *Expr {
	term := NewTerm(f.gen.Generate()).SetLocation(other.Location)
	expr := Equality.Expr(term, other)
	expr.Generated = true
	expr.Location = other.Location
	return expr
}

type localVarGenerator struct {
	exclude VarSet
	suffix  string
	next    int
}

func newLocalVarGeneratorForModuleSet(sorted []string, modules map[string]*Module) *localVarGenerator {
	exclude := NewVarSet()
	vis := &VarVisitor{vars: exclude}
	for _, key := range sorted {
		vis.Walk(modules[key])
	}
	return &localVarGenerator{exclude: exclude, next: 0}
}

func newLocalVarGenerator(suffix string, node interface{}) *localVarGenerator {
	exclude := NewVarSet()
	vis := &VarVisitor{vars: exclude}
	vis.Walk(node)
	return &localVarGenerator{exclude: exclude, suffix: suffix, next: 0}
}

func (l *localVarGenerator) Generate() Var {
	for {
		result := Var("__local" + l.suffix + strconv.Itoa(l.next) + "__")
		l.next++
		if !l.exclude.Contains(result) {
			return result
		}
	}
}

func getGlobals(pkg *Package, rules []Var, imports []*Import) map[Var]Ref {

	globals := map[Var]Ref{}

	// Populate globals with exports within the package.
	for _, v := range rules {
		global := append(Ref{}, pkg.Path...)
		global = append(global, &Term{Value: String(v)})
		globals[v] = global
	}

	// Populate globals with imports.
	for _, i := range imports {
		if len(i.Alias) > 0 {
			path := i.Path.Value.(Ref)
			globals[i.Alias] = path
		} else {
			path := i.Path.Value.(Ref)
			if len(path) == 1 {
				globals[path[0].Value.(Var)] = path
			} else {
				v := path[len(path)-1].Value.(String)
				globals[Var(v)] = path
			}
		}
	}

	return globals
}

func requiresEval(x *Term) bool {
	if x == nil {
		return false
	}
	return ContainsRefs(x) || ContainsComprehensions(x)
}

func resolveRef(globals map[Var]Ref, ignore *declaredVarStack, ref Ref) Ref {

	r := Ref{}
	for i, x := range ref {
		switch v := x.Value.(type) {
		case Var:
			if g, ok := globals[v]; ok && !ignore.Contains(v) {
				cpy := g.Copy()
				for i := range cpy {
					cpy[i].SetLocation(x.Location)
				}
				if i == 0 {
					r = cpy
				} else {
					r = append(r, NewTerm(cpy).SetLocation(x.Location))
				}
			} else {
				r = append(r, x)
			}
		case Ref, Array, Object, Set, *ArrayComprehension, *SetComprehension, *ObjectComprehension, Call:
			r = append(r, resolveRefsInTerm(globals, ignore, x))
		default:
			r = append(r, x)
		}
	}

	return r
}

func resolveRefsInRule(globals map[Var]Ref, rule *Rule) error {
	ignore := &declaredVarStack{}

	vars := NewVarSet()
	var vis *GenericVisitor
	var err error

	// Walk args to collect vars and transform body so that callers can shadow
	// root documents.
	vis = NewGenericVisitor(func(x interface{}) bool {
		if err != nil {
			return true
		}
		switch x := x.(type) {
		case Var:
			vars.Add(x)

		// Object keys cannot be pattern matched so only walk values.
		case Object:
			for _, k := range x.Keys() {
				vis.Walk(x.Get(k))
			}

		// Skip terms that could contain vars that cannot be pattern matched.
		case Set, *ArrayComprehension, *SetComprehension, *ObjectComprehension, Call:
			return true

		case *Term:
			if _, ok := x.Value.(Ref); ok {
				if RootDocumentRefs.Contains(x) {
					// We could support args named input, data, etc. however
					// this would require rewriting terms in the head and body.
					// Preventing root document shadowing is simpler, and
					// arguably, will prevent confusing names from being used.
					err = fmt.Errorf("args must not shadow %v (use a different variable name)", x)
					return true
				}
			}
		}
		return false
	})

	vis.Walk(rule.Head.Args)

	if err != nil {
		return err
	}

	ignore.Push(vars)
	ignore.Push(declaredVars(rule.Body))

	if rule.Head.Key != nil {
		rule.Head.Key = resolveRefsInTerm(globals, ignore, rule.Head.Key)
	}

	if rule.Head.Value != nil {
		rule.Head.Value = resolveRefsInTerm(globals, ignore, rule.Head.Value)
	}

	rule.Body = resolveRefsInBody(globals, ignore, rule.Body)
	return nil
}

func resolveRefsInBody(globals map[Var]Ref, ignore *declaredVarStack, body Body) Body {
	r := Body{}
	for _, expr := range body {
		r = append(r, resolveRefsInExpr(globals, ignore, expr))
	}
	return r
}

func resolveRefsInExpr(globals map[Var]Ref, ignore *declaredVarStack, expr *Expr) *Expr {
	cpy := *expr
	switch ts := expr.Terms.(type) {
	case *Term:
		cpy.Terms = resolveRefsInTerm(globals, ignore, ts)
	case []*Term:
		buf := make([]*Term, len(ts))
		for i := 0; i < len(ts); i++ {
			buf[i] = resolveRefsInTerm(globals, ignore, ts[i])
		}
		cpy.Terms = buf
	}
	for _, w := range cpy.With {
		w.Target = resolveRefsInTerm(globals, ignore, w.Target)
		w.Value = resolveRefsInTerm(globals, ignore, w.Value)
	}
	return &cpy
}

func resolveRefsInTerm(globals map[Var]Ref, ignore *declaredVarStack, term *Term) *Term {
	switch v := term.Value.(type) {
	case Var:
		if g, ok := globals[v]; ok && !ignore.Contains(v) {
			cpy := g.Copy()
			for i := range cpy {
				cpy[i].SetLocation(term.Location)
			}
			return NewTerm(cpy).SetLocation(term.Location)
		}
		return term
	case Ref:
		fqn := resolveRef(globals, ignore, v)
		cpy := *term
		cpy.Value = fqn
		return &cpy
	case Object:
		cpy := *term
		cpy.Value, _ = v.Map(func(k, v *Term) (*Term, *Term, error) {
			k = resolveRefsInTerm(globals, ignore, k)
			v = resolveRefsInTerm(globals, ignore, v)
			return k, v, nil
		})
		return &cpy
	case Array:
		cpy := *term
		cpy.Value = Array(resolveRefsInTermSlice(globals, ignore, v))
		return &cpy
	case Call:
		cpy := *term
		cpy.Value = Call(resolveRefsInTermSlice(globals, ignore, v))
		return &cpy
	case Set:
		s, _ := v.Map(func(e *Term) (*Term, error) {
			return resolveRefsInTerm(globals, ignore, e), nil
		})
		cpy := *term
		cpy.Value = s
		return &cpy
	case *ArrayComprehension:
		ac := &ArrayComprehension{}
		ignore.Push(declaredVars(v.Body))
		ac.Term = resolveRefsInTerm(globals, ignore, v.Term)
		ac.Body = resolveRefsInBody(globals, ignore, v.Body)
		cpy := *term
		cpy.Value = ac
		ignore.Pop()
		return &cpy
	case *ObjectComprehension:
		oc := &ObjectComprehension{}
		ignore.Push(declaredVars(v.Body))
		oc.Key = resolveRefsInTerm(globals, ignore, v.Key)
		oc.Value = resolveRefsInTerm(globals, ignore, v.Value)
		oc.Body = resolveRefsInBody(globals, ignore, v.Body)
		cpy := *term
		cpy.Value = oc
		ignore.Pop()
		return &cpy
	case *SetComprehension:
		sc := &SetComprehension{}
		ignore.Push(declaredVars(v.Body))
		sc.Term = resolveRefsInTerm(globals, ignore, v.Term)
		sc.Body = resolveRefsInBody(globals, ignore, v.Body)
		cpy := *term
		cpy.Value = sc
		ignore.Pop()
		return &cpy
	default:
		return term
	}
}

func resolveRefsInTermSlice(globals map[Var]Ref, ignore *declaredVarStack, terms []*Term) []*Term {
	cpy := make([]*Term, len(terms))
	for i := 0; i < len(terms); i++ {
		cpy[i] = resolveRefsInTerm(globals, ignore, terms[i])
	}
	return cpy
}

type declaredVarStack []VarSet

func (s declaredVarStack) Contains(v Var) bool {
	for i := len(s) - 1; i >= 0; i-- {
		if _, ok := s[i][v]; ok {
			return ok
		}
	}
	return false
}

func (s declaredVarStack) Add(v Var) {
	s[len(s)-1].Add(v)
}

func (s *declaredVarStack) Push(vs VarSet) {
	*s = append(*s, vs)
}

func (s *declaredVarStack) Pop() {
	curr := *s
	*s = curr[:len(curr)-1]
}

func declaredVars(x interface{}) VarSet {
	vars := NewVarSet()
	vis := NewGenericVisitor(func(x interface{}) bool {
		switch x := x.(type) {
		case *Expr:
			if x.IsAssignment() && validEqAssignArgCount(x) {
				WalkVars(x.Operand(0), func(v Var) bool {
					vars.Add(v)
					return false
				})
			} else if decl, ok := x.Terms.(*SomeDecl); ok {
				for i := range decl.Symbols {
					vars.Add(decl.Symbols[i].Value.(Var))
				}
			}
		case *ArrayComprehension, *SetComprehension, *ObjectComprehension:
			return true
		}
		return false
	})
	vis.Walk(x)
	return vars
}

// rewriteComprehensionTerms will rewrite comprehensions so that the term part
// is bound to a variable in the body. This allows any type of term to be used
// in the term part (even if the term requires evaluation.)
//
// For instance, given the following comprehension:
//
// [x[0] | x = y[_]; y = [1,2,3]]
//
// The comprehension would be rewritten as:
//
// [__local0__ | x = y[_]; y = [1,2,3]; __local0__ = x[0]]
func rewriteComprehensionTerms(f *equalityFactory, node interface{}) (interface{}, error) {
	return TransformComprehensions(node, func(x interface{}) (Value, error) {
		switch x := x.(type) {
		case *ArrayComprehension:
			if requiresEval(x.Term) {
				expr := f.Generate(x.Term)
				x.Term = expr.Operand(0)
				x.Body.Append(expr)
			}
			return x, nil
		case *SetComprehension:
			if requiresEval(x.Term) {
				expr := f.Generate(x.Term)
				x.Term = expr.Operand(0)
				x.Body.Append(expr)
			}
			return x, nil
		case *ObjectComprehension:
			if requiresEval(x.Key) {
				expr := f.Generate(x.Key)
				x.Key = expr.Operand(0)
				x.Body.Append(expr)
			}
			if requiresEval(x.Value) {
				expr := f.Generate(x.Value)
				x.Value = expr.Operand(0)
				x.Body.Append(expr)
			}
			return x, nil
		}
		panic("illegal type")
	})
}

// rewriteEquals will rewrite exprs under x as unification calls instead of ==
// calls. For example:
//
// data.foo == data.bar is rewritten as data.foo = data.bar
//
// This stage should only run the safety check (since == is a built-in with no
// outputs, so the inputs must not be marked as safe.)
//
// This stage is not executed by the query compiler by default because when
// callers specify == instead of = they expect to receive a true/false/undefined
// result back whereas with = the result is only ever true/undefined. For
// partial evaluation cases we do want to rewrite == to = to simplify the
// result.
func rewriteEquals(x interface{}) {
	doubleEq := Equal.Ref()
	unifyOp := Equality.Ref()
	WalkExprs(x, func(x *Expr) bool {
		if x.IsCall() {
			operator := x.Operator()
			if operator.Equal(doubleEq) && len(x.Operands()) == 2 {
				x.SetOperator(NewTerm(unifyOp))
			}
		}
		return false
	})
}

// rewriteDynamics will rewrite the body so that dynamic terms (i.e., refs and
// comprehensions) are bound to vars earlier in the query. This translation
// results in eager evaluation.
//
// For instance, given the following query:
//
// foo(data.bar) = 1
//
// The rewritten version will be:
//
// __local0__ = data.bar; foo(__local0__) = 1
func rewriteDynamics(f *equalityFactory, body Body) Body {
	result := make(Body, 0, len(body))
	for _, expr := range body {
		if expr.IsEquality() {
			result = rewriteDynamicsEqExpr(f, expr, result)
		} else if expr.IsCall() {
			result = rewriteDynamicsCallExpr(f, expr, result)
		} else {
			result = rewriteDynamicsTermExpr(f, expr, result)
		}
	}
	return result
}

func appendExpr(body Body, expr *Expr) Body {
	body.Append(expr)
	return body
}

func rewriteDynamicsEqExpr(f *equalityFactory, expr *Expr, result Body) Body {
	if !validEqAssignArgCount(expr) {
		return appendExpr(result, expr)
	}
	terms := expr.Terms.([]*Term)
	result, terms[1] = rewriteDynamicsInTerm(expr, f, terms[1], result)
	result, terms[2] = rewriteDynamicsInTerm(expr, f, terms[2], result)
	return appendExpr(result, expr)
}

func rewriteDynamicsCallExpr(f *equalityFactory, expr *Expr, result Body) Body {
	terms := expr.Terms.([]*Term)
	for i := 1; i < len(terms); i++ {
		result, terms[i] = rewriteDynamicsOne(expr, f, terms[i], result)
	}
	return appendExpr(result, expr)
}

func rewriteDynamicsTermExpr(f *equalityFactory, expr *Expr, result Body) Body {
	term := expr.Terms.(*Term)
	result, expr.Terms = rewriteDynamicsInTerm(expr, f, term, result)
	return appendExpr(result, expr)
}

func rewriteDynamicsInTerm(original *Expr, f *equalityFactory, term *Term, result Body) (Body, *Term) {
	switch v := term.Value.(type) {
	case Ref:
		for i := 1; i < len(v); i++ {
			result, v[i] = rewriteDynamicsOne(original, f, v[i], result)
		}
	case *ArrayComprehension:
		v.Body = rewriteDynamics(f, v.Body)
	case *SetComprehension:
		v.Body = rewriteDynamics(f, v.Body)
	case *ObjectComprehension:
		v.Body = rewriteDynamics(f, v.Body)
	default:
		result, term = rewriteDynamicsOne(original, f, term, result)
	}
	return result, term
}

func rewriteDynamicsOne(original *Expr, f *equalityFactory, term *Term, result Body) (Body, *Term) {
	switch v := term.Value.(type) {
	case Ref:
		for i := 1; i < len(v); i++ {
			result, v[i] = rewriteDynamicsOne(original, f, v[i], result)
		}
		generated := f.Generate(term)
		generated.With = original.With
		result.Append(generated)
		return result, result[len(result)-1].Operand(0)
	case Array:
		for i := 0; i < len(v); i++ {
			result, v[i] = rewriteDynamicsOne(original, f, v[i], result)
		}
		return result, term
	case Object:
		cpy := NewObject()
		for _, key := range v.Keys() {
			value := v.Get(key)
			result, key = rewriteDynamicsOne(original, f, key, result)
			result, value = rewriteDynamicsOne(original, f, value, result)
			cpy.Insert(key, value)
		}
		return result, NewTerm(cpy).SetLocation(term.Location)
	case Set:
		cpy := NewSet()
		for _, term := range v.Slice() {
			var rw *Term
			result, rw = rewriteDynamicsOne(original, f, term, result)
			cpy.Add(rw)
		}
		return result, NewTerm(cpy).SetLocation(term.Location)
	case *ArrayComprehension:
		var extra *Expr
		v.Body, extra = rewriteDynamicsComprehensionBody(original, f, v.Body, term)
		result.Append(extra)
		return result, result[len(result)-1].Operand(0)
	case *SetComprehension:
		var extra *Expr
		v.Body, extra = rewriteDynamicsComprehensionBody(original, f, v.Body, term)
		result.Append(extra)
		return result, result[len(result)-1].Operand(0)
	case *ObjectComprehension:
		var extra *Expr
		v.Body, extra = rewriteDynamicsComprehensionBody(original, f, v.Body, term)
		result.Append(extra)
		return result, result[len(result)-1].Operand(0)
	}
	return result, term
}

func rewriteDynamicsComprehensionBody(original *Expr, f *equalityFactory, body Body, term *Term) (Body, *Expr) {
	body = rewriteDynamics(f, body)
	generated := f.Generate(term)
	generated.With = original.With
	return body, generated
}

func rewriteExprTermsInHead(gen *localVarGenerator, rule *Rule) {
	if rule.Head.Key != nil {
		support, output := expandExprTerm(gen, rule.Head.Key)
		for i := range support {
			rule.Body.Append(support[i])
		}
		rule.Head.Key = output
	}
	if rule.Head.Value != nil {
		support, output := expandExprTerm(gen, rule.Head.Value)
		for i := range support {
			rule.Body.Append(support[i])
		}
		rule.Head.Value = output
	}
}

func rewriteExprTermsInBody(gen *localVarGenerator, body Body) Body {
	cpy := make(Body, 0, len(body))
	for i := 0; i < len(body); i++ {
		for _, expr := range expandExpr(gen, body[i]) {
			cpy.Append(expr)
		}
	}
	return cpy
}

func expandExpr(gen *localVarGenerator, expr *Expr) (result []*Expr) {
	for i := range expr.With {
		extras, value := expandExprTerm(gen, expr.With[i].Value)
		expr.With[i].Value = value
		result = append(result, extras...)
	}
	switch terms := expr.Terms.(type) {
	case *Term:
		extras, term := expandExprTerm(gen, terms)
		if len(expr.With) > 0 {
			for i := range extras {
				extras[i].With = expr.With
			}
		}
		result = append(result, extras...)
		expr.Terms = term
		result = append(result, expr)
	case []*Term:
		for i := 1; i < len(terms); i++ {
			var extras []*Expr
			extras, terms[i] = expandExprTerm(gen, terms[i])
			if len(expr.With) > 0 {
				for i := range extras {
					extras[i].With = expr.With
				}
			}
			result = append(result, extras...)
		}
		result = append(result, expr)
	}
	return
}

func expandExprTerm(gen *localVarGenerator, term *Term) (support []*Expr, output *Term) {
	output = term
	switch v := term.Value.(type) {
	case Call:
		for i := 1; i < len(v); i++ {
			var extras []*Expr
			extras, v[i] = expandExprTerm(gen, v[i])
			support = append(support, extras...)
		}
		output = NewTerm(gen.Generate()).SetLocation(term.Location)
		expr := v.MakeExpr(output).SetLocation(term.Location)
		expr.Generated = true
		support = append(support, expr)
	case Ref:
		support = expandExprRef(gen, v)
	case Array:
		support = expandExprTermSlice(gen, v)
	case Object:
		cpy, _ := v.Map(func(k, v *Term) (*Term, *Term, error) {
			extras1, expandedKey := expandExprTerm(gen, k)
			extras2, expandedValue := expandExprTerm(gen, v)
			support = append(support, extras1...)
			support = append(support, extras2...)
			return expandedKey, expandedValue, nil
		})
		output = NewTerm(cpy).SetLocation(term.Location)
	case Set:
		cpy, _ := v.Map(func(x *Term) (*Term, error) {
			extras, expanded := expandExprTerm(gen, x)
			support = append(support, extras...)
			return expanded, nil
		})
		output = NewTerm(cpy).SetLocation(term.Location)
	case *ArrayComprehension:
		support, term := expandExprTerm(gen, v.Term)
		for i := range support {
			v.Body.Append(support[i])
		}
		v.Term = term
		v.Body = rewriteExprTermsInBody(gen, v.Body)
	case *SetComprehension:
		support, term := expandExprTerm(gen, v.Term)
		for i := range support {
			v.Body.Append(support[i])
		}
		v.Term = term
		v.Body = rewriteExprTermsInBody(gen, v.Body)
	case *ObjectComprehension:
		support, key := expandExprTerm(gen, v.Key)
		for i := range support {
			v.Body.Append(support[i])
		}
		v.Key = key
		support, value := expandExprTerm(gen, v.Value)
		for i := range support {
			v.Body.Append(support[i])
		}
		v.Value = value
		v.Body = rewriteExprTermsInBody(gen, v.Body)
	}
	return
}

func expandExprRef(gen *localVarGenerator, v []*Term) (support []*Expr) {
	// Start by calling a normal expandExprTerm on all terms.
	support = expandExprTermSlice(gen, v)

	// Rewrite references in order to support indirect references.  We rewrite
	// e.g.
	//
	//     [1, 2, 3][i]
	//
	// to
	//
	//     __local_var = [1, 2, 3]
	//     __local_var[i]
	//
	// to support these.  This only impacts the reference subject, i.e. the
	// first item in the slice.
	var subject = v[0]
	switch subject.Value.(type) {
	case Array, Object, Set, *ArrayComprehension, *SetComprehension, *ObjectComprehension, Call:
		f := newEqualityFactory(gen)
		assignToLocal := f.Generate(subject)
		support = append(support, assignToLocal)
		v[0] = assignToLocal.Operand(0)
	}
	return
}

func expandExprTermSlice(gen *localVarGenerator, v []*Term) (support []*Expr) {
	for i := 0; i < len(v); i++ {
		var extras []*Expr
		extras, v[i] = expandExprTerm(gen, v[i])
		support = append(support, extras...)
	}
	return
}

type localDeclaredVars struct {
	vars []*declaredVarSet

	// rewritten contains a mapping of *all* user-defined variables
	// that have been rewritten whereas vars contains the state
	// from the current query (not not any nested queries, and all
	// vars seen).
	rewritten map[Var]Var
}

type varOccurrence int

const (
	newVar varOccurrence = iota
	argVar
	seenVar
	assignedVar
	declaredVar
)

type declaredVarSet struct {
	vs         map[Var]Var
	reverse    map[Var]Var
	occurrence map[Var]varOccurrence
}

func newDeclaredVarSet() *declaredVarSet {
	return &declaredVarSet{
		vs:         map[Var]Var{},
		reverse:    map[Var]Var{},
		occurrence: map[Var]varOccurrence{},
	}
}

func newLocalDeclaredVars() *localDeclaredVars {
	return &localDeclaredVars{
		vars:      []*declaredVarSet{newDeclaredVarSet()},
		rewritten: map[Var]Var{},
	}
}

func (s *localDeclaredVars) Push() {
	s.vars = append(s.vars, newDeclaredVarSet())
}

func (s *localDeclaredVars) Pop() *declaredVarSet {
	sl := s.vars
	curr := sl[len(sl)-1]
	s.vars = sl[:len(sl)-1]
	return curr
}

func (s localDeclaredVars) Peek() *declaredVarSet {
	return s.vars[len(s.vars)-1]
}

func (s localDeclaredVars) Insert(x, y Var, occurrence varOccurrence) {
	elem := s.vars[len(s.vars)-1]
	elem.vs[x] = y
	elem.reverse[y] = x
	elem.occurrence[x] = occurrence

	// If the variable has been rewritten (where x != y, with y being
	// the generated value), store it in the map of rewritten vars.
	// Assume that the generated values are unique for the compilation.
	if !x.Equal(y) {
		s.rewritten[y] = x
	}
}

func (s localDeclaredVars) Declared(x Var) (y Var, ok bool) {
	for i := len(s.vars) - 1; i >= 0; i-- {
		if y, ok = s.vars[i].vs[x]; ok {
			return
		}
	}
	return
}

// Occurrence returns a flag that indicates whether x has occurred in the
// current scope.
func (s localDeclaredVars) Occurrence(x Var) varOccurrence {
	return s.vars[len(s.vars)-1].occurrence[x]
}

// rewriteLocalVars rewrites bodies to remove assignment/declaration
// expressions. For example:
//
// a := 1; p[a]
//
// Is rewritten to:
//
// __local0__ = 1; p[__local0__]
//
// During rewriting, assignees are validated to prevent use before declaration.
func rewriteLocalVars(g *localVarGenerator, stack *localDeclaredVars, used VarSet, body Body) (Body, map[Var]Var, Errors) {
	var errs Errors
	body, errs = rewriteDeclaredVarsInBody(g, stack, used, body, errs)
	return body, stack.Pop().vs, errs
}

func rewriteDeclaredVarsInBody(g *localVarGenerator, stack *localDeclaredVars, used VarSet, body Body, errs Errors) (Body, Errors) {

	var cpy Body

	for i := range body {
		var expr *Expr
		if body[i].IsAssignment() {
			expr, errs = rewriteDeclaredAssignment(g, stack, body[i], errs)
		} else if decl, ok := body[i].Terms.(*SomeDecl); ok {
			errs = rewriteSomeDeclStatement(g, stack, decl, errs)
		} else {
			expr, errs = rewriteDeclaredVarsInExpr(g, stack, body[i], errs)
		}
		if expr != nil {
			cpy.Append(expr)
		}
	}

	// If the body only contained a var statement it will be empty at this
	// point. Append true to the body to ensure that it's non-empty (zero length
	// bodies are not supported.)
	if len(cpy) == 0 {
		cpy.Append(NewExpr(BooleanTerm(true)))
	}

	return cpy, checkUnusedDeclaredVars(body[0].Loc(), stack, used, cpy, errs)
}

func checkUnusedDeclaredVars(loc *Location, stack *localDeclaredVars, used VarSet, cpy Body, errs Errors) Errors {

	// NOTE(tsandall): Do not generate more errors if there are existing
	// declaration errors.
	if len(errs) > 0 {
		return errs
	}

	dvs := stack.Peek()
	declared := NewVarSet()

	for v, occ := range dvs.occurrence {
		if occ == declaredVar {
			declared.Add(dvs.vs[v])
		}
	}

	bodyvars := cpy.Vars(VarVisitorParams{})

	for v := range used {
		if gv, ok := stack.Declared(v); ok {
			bodyvars.Add(gv)
		} else {
			bodyvars.Add(v)
		}
	}

	unused := declared.Diff(bodyvars).Diff(used)

	for _, gv := range unused.Sorted() {
		errs = append(errs, NewError(CompileErr, loc, "declared var %v unused", dvs.reverse[gv]))
	}

	return errs
}

func rewriteSomeDeclStatement(g *localVarGenerator, stack *localDeclaredVars, decl *SomeDecl, errs Errors) Errors {
	for i := range decl.Symbols {
		v := decl.Symbols[i].Value.(Var)
		if _, err := rewriteDeclaredVar(g, stack, v, declaredVar); err != nil {
			errs = append(errs, NewError(CompileErr, decl.Loc(), err.Error()))
		}
	}
	return errs
}

func rewriteDeclaredVarsInExpr(g *localVarGenerator, stack *localDeclaredVars, expr *Expr, errs Errors) (*Expr, Errors) {
	vis := NewGenericVisitor(func(x interface{}) bool {
		var stop bool
		switch x := x.(type) {
		case *Term:
			stop, errs = rewriteDeclaredVarsInTerm(g, stack, x, errs)
		case *With:
			_, errs = rewriteDeclaredVarsInTerm(g, stack, x.Value, errs)
			stop = true
		}
		return stop
	})
	vis.Walk(expr)
	return expr, errs
}

func rewriteDeclaredAssignment(g *localVarGenerator, stack *localDeclaredVars, expr *Expr, errs Errors) (*Expr, Errors) {

	if expr.Negated {
		errs = append(errs, NewError(CompileErr, expr.Location, "cannot assign vars inside negated expression"))
		return expr, errs
	}

	numErrsBefore := len(errs)

	if !validEqAssignArgCount(expr) {
		return expr, errs
	}

	// Rewrite terms on right hand side capture seen vars and recursively
	// process comprehensions before left hand side is processed. Also
	// rewrite with modifier.
	errs = rewriteDeclaredVarsInTermRecursive(g, stack, expr.Operand(1), errs)

	for _, w := range expr.With {
		errs = rewriteDeclaredVarsInTermRecursive(g, stack, w.Value, errs)
	}

	// Rewrite vars on left hand side with unique names. Catch redeclaration
	// and invalid term types here.
	var vis func(t *Term) bool

	vis = func(t *Term) bool {
		switch v := t.Value.(type) {
		case Var:
			if gv, err := rewriteDeclaredVar(g, stack, v, assignedVar); err != nil {
				errs = append(errs, NewError(CompileErr, t.Location, err.Error()))
			} else {
				t.Value = gv
			}
			return true
		case Array:
			return false
		case Object:
			v.Foreach(func(_, v *Term) {
				WalkTerms(v, vis)
			})
			return true
		case Ref:
			if RootDocumentRefs.Contains(t) {
				if gv, err := rewriteDeclaredVar(g, stack, v[0].Value.(Var), assignedVar); err != nil {
					errs = append(errs, NewError(CompileErr, t.Location, err.Error()))
				} else {
					t.Value = gv
				}
				return true
			}
		}
		errs = append(errs, NewError(CompileErr, t.Location, "cannot assign to %v", TypeName(t.Value)))
		return true
	}

	WalkTerms(expr.Operand(0), vis)

	if len(errs) == numErrsBefore {
		loc := expr.Operator()[0].Location
		expr.SetOperator(RefTerm(VarTerm(Equality.Name).SetLocation(loc)).SetLocation(loc))
	}

	return expr, errs
}

func rewriteDeclaredVarsInTerm(g *localVarGenerator, stack *localDeclaredVars, term *Term, errs Errors) (bool, Errors) {
	switch v := term.Value.(type) {
	case Var:
		if gv, ok := stack.Declared(v); ok {
			term.Value = gv
		} else if stack.Occurrence(v) == newVar {
			stack.Insert(v, v, seenVar)
		}
	case Ref:
		if RootDocumentRefs.Contains(term) {
			if gv, ok := stack.Declared(v[0].Value.(Var)); ok {
				term.Value = gv
			}
			return true, errs
		}
		return false, errs
	case Object:
		cpy, _ := v.Map(func(k, v *Term) (*Term, *Term, error) {
			kcpy := k.Copy()
			errs = rewriteDeclaredVarsInTermRecursive(g, stack, kcpy, errs)
			errs = rewriteDeclaredVarsInTermRecursive(g, stack, v, errs)
			return kcpy, v, nil
		})
		term.Value = cpy
	case Set:
		cpy, _ := v.Map(func(elem *Term) (*Term, error) {
			elemcpy := elem.Copy()
			errs = rewriteDeclaredVarsInTermRecursive(g, stack, elemcpy, errs)
			return elemcpy, nil
		})
		term.Value = cpy
	case *ArrayComprehension:
		errs = rewriteDeclaredVarsInArrayComprehension(g, stack, v, errs)
	case *SetComprehension:
		errs = rewriteDeclaredVarsInSetComprehension(g, stack, v, errs)
	case *ObjectComprehension:
		errs = rewriteDeclaredVarsInObjectComprehension(g, stack, v, errs)
	default:
		return false, errs
	}
	return true, errs
}

func rewriteDeclaredVarsInTermRecursive(g *localVarGenerator, stack *localDeclaredVars, term *Term, errs Errors) Errors {
	WalkNodes(term, func(n Node) bool {
		var stop bool
		switch n := n.(type) {
		case *With:
			_, errs = rewriteDeclaredVarsInTerm(g, stack, n.Value, errs)
			stop = true
		case *Term:
			stop, errs = rewriteDeclaredVarsInTerm(g, stack, n, errs)
		}
		return stop
	})
	return errs
}

func rewriteDeclaredVarsInArrayComprehension(g *localVarGenerator, stack *localDeclaredVars, v *ArrayComprehension, errs Errors) Errors {
	stack.Push()
	v.Body, errs = rewriteDeclaredVarsInBody(g, stack, nil, v.Body, errs)
	errs = rewriteDeclaredVarsInTermRecursive(g, stack, v.Term, errs)
	stack.Pop()
	return errs
}

func rewriteDeclaredVarsInSetComprehension(g *localVarGenerator, stack *localDeclaredVars, v *SetComprehension, errs Errors) Errors {
	stack.Push()
	v.Body, errs = rewriteDeclaredVarsInBody(g, stack, nil, v.Body, errs)
	errs = rewriteDeclaredVarsInTermRecursive(g, stack, v.Term, errs)
	stack.Pop()
	return errs
}

func rewriteDeclaredVarsInObjectComprehension(g *localVarGenerator, stack *localDeclaredVars, v *ObjectComprehension, errs Errors) Errors {
	stack.Push()
	v.Body, errs = rewriteDeclaredVarsInBody(g, stack, nil, v.Body, errs)
	errs = rewriteDeclaredVarsInTermRecursive(g, stack, v.Key, errs)
	errs = rewriteDeclaredVarsInTermRecursive(g, stack, v.Value, errs)
	stack.Pop()
	return errs
}

func rewriteDeclaredVar(g *localVarGenerator, stack *localDeclaredVars, v Var, occ varOccurrence) (gv Var, err error) {
	switch stack.Occurrence(v) {
	case seenVar:
		return gv, fmt.Errorf("var %v referenced above", v)
	case assignedVar:
		return gv, fmt.Errorf("var %v assigned above", v)
	case declaredVar:
		return gv, fmt.Errorf("var %v declared above", v)
	case argVar:
		return gv, fmt.Errorf("arg %v redeclared", v)
	}
	gv = g.Generate()
	stack.Insert(v, gv, occ)
	return
}

// rewriteWithModifiersInBody will rewrite the body so that with modifiers do
// not contain terms that require evaluation as values. If this function
// encounters an invalid with modifier target then it will raise an error.
func rewriteWithModifiersInBody(c *Compiler, f *equalityFactory, body Body) (Body, *Error) {
	var result Body
	for i := range body {
		exprs, err := rewriteWithModifier(c, f, body[i])
		if err != nil {
			return nil, err
		}
		if len(exprs) > 0 {
			for _, expr := range exprs {
				result.Append(expr)
			}
		} else {
			result.Append(body[i])
		}
	}
	return result, nil
}

func rewriteWithModifier(c *Compiler, f *equalityFactory, expr *Expr) ([]*Expr, *Error) {

	var result []*Expr
	for i := range expr.With {
		err := validateTarget(c, expr.With[i].Target)
		if err != nil {
			return nil, err
		}

		if requiresEval(expr.With[i].Value) {
			eq := f.Generate(expr.With[i].Value)
			result = append(result, eq)
			expr.With[i].Value = eq.Operand(0)
		}
	}

	// If any of the with modifiers in this expression were rewritten then result
	// will be non-empty. In this case, the expression will have been modified and
	// it should also be added to the result.
	if len(result) > 0 {
		result = append(result, expr)
	}
	return result, nil
}

func validateTarget(c *Compiler, term *Term) *Error {
	if !isInputRef(term) && !isDataRef(term) {
		return NewError(TypeErr, term.Location, "with keyword target must start with %v or %v", InputRootDocument, DefaultRootDocument)
	}

	if isDataRef(term) {
		ref := term.Value.(Ref)
		node := c.RuleTree
		for i := 0; i < len(ref)-1; i++ {
			child := node.Child(ref[i].Value)
			if child == nil {
				break
			} else if len(child.Values) > 0 {
				return NewError(CompileErr, term.Loc(), "with keyword cannot partially replace virtual document(s)")
			}
			node = child
		}

		if node != nil {
			if child := node.Child(ref[len(ref)-1].Value); child != nil {
				for _, value := range child.Values {
					if len(value.(*Rule).Head.Args) > 0 {
						return NewError(CompileErr, term.Loc(), "with keyword cannot replace functions")
					}
				}
			}
		}

	}
	return nil
}

func isInputRef(term *Term) bool {
	if ref, ok := term.Value.(Ref); ok {
		if ref.HasPrefix(InputRootRef) {
			return true
		}
	}
	return false
}

func isDataRef(term *Term) bool {
	if ref, ok := term.Value.(Ref); ok {
		if ref.HasPrefix(DefaultRootRef) {
			return true
		}
	}
	return false
}

func isVirtual(node *TreeNode, ref Ref) bool {
	for i := 0; i < len(ref); i++ {
		child := node.Child(ref[i].Value)
		if child == nil {
			return false
		} else if len(child.Values) > 0 {
			return true
		}
		node = child
	}
	return true
}

func safetyErrorSlice(unsafe unsafeVars) (result Errors) {

	if len(unsafe) == 0 {
		return
	}

	for _, pair := range unsafe.Vars() {
		if !pair.Var.IsGenerated() {
			result = append(result, NewError(UnsafeVarErr, pair.Loc, "var %v is unsafe", pair.Var))
		}
	}

	if len(result) > 0 {
		return
	}

	// If the expression contains unsafe generated variables, report which
	// expressions are unsafe instead of the variables that are unsafe (since
	// the latter are not meaningful to the user.)
	pairs := unsafe.Slice()

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Expr.Location.Compare(pairs[j].Expr.Location) < 0
	})

	// Report at most one error per generated variable.
	seen := NewVarSet()

	for _, expr := range pairs {
		before := len(seen)
		for v := range expr.Vars {
			if v.IsGenerated() {
				seen.Add(v)
			}
		}
		if len(seen) > before {
			result = append(result, NewError(UnsafeVarErr, expr.Expr.Location, "expression is unsafe"))
		}
	}

	return
}

func checkUnsafeBuiltins(unsafeBuiltinsMap map[string]struct{}, node interface{}) Errors {
	errs := make(Errors, 0)
	WalkExprs(node, func(x *Expr) bool {
		if x.IsCall() {
			operator := x.Operator().String()
			if _, ok := unsafeBuiltinsMap[operator]; ok {
				errs = append(errs, NewError(TypeErr, x.Loc(), "unsafe built-in function calls in expression: %v", operator))
			}
		}
		return false
	})
	return errs
}

func rewriteVarsInRef(vars ...map[Var]Var) func(Ref) Ref {
	return func(node Ref) Ref {
		i, _ := TransformVars(node, func(v Var) (Value, error) {
			for _, m := range vars {
				if u, ok := m[v]; ok {
					return u, nil
				}
			}
			return v, nil
		})
		return i.(Ref)
	}
}

func rewriteVarsNop(node Ref) Ref {
	return node
}
