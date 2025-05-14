// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// CompileErrorLimitDefault is the default number errors a compiler will allow before
// exiting.
const CompileErrorLimitDefault = 10

// Compiler contains the state of a compilation process.
type Compiler = v1.Compiler

// CompilerStage defines the interface for stages in the compiler.
type CompilerStage = v1.CompilerStage

// CompilerEvalMode allows toggling certain stages that are only
// needed for certain modes, Concretely, only "topdown" mode will
// have the compiler build comprehension and rule indices.
type CompilerEvalMode = v1.CompilerEvalMode

const (
	// EvalModeTopdown (default) instructs the compiler to build rule
	// and comprehension indices used by topdown evaluation.
	EvalModeTopdown = v1.EvalModeTopdown

	// EvalModeIR makes the compiler skip the stages for comprehension
	// and rule indices.
	EvalModeIR = v1.EvalModeIR
)

// CompilerStageDefinition defines a compiler stage
type CompilerStageDefinition = v1.CompilerStageDefinition

// RulesOptions defines the options for retrieving rules by Ref from the
// compiler.
type RulesOptions = v1.RulesOptions

// QueryContext contains contextual information for running an ad-hoc query.
//
// Ad-hoc queries can be run in the context of a package and imports may be
// included to provide concise access to data.
type QueryContext = v1.QueryContext

// NewQueryContext returns a new QueryContext object.
func NewQueryContext() *QueryContext {
	return v1.NewQueryContext()
}

// QueryCompiler defines the interface for compiling ad-hoc queries.
type QueryCompiler = v1.QueryCompiler

// QueryCompilerStage defines the interface for stages in the query compiler.
type QueryCompilerStage = v1.QueryCompilerStage

// QueryCompilerStageDefinition defines a QueryCompiler stage
type QueryCompilerStageDefinition = v1.QueryCompilerStageDefinition

// NewCompiler returns a new empty compiler.
func NewCompiler() *Compiler {
	return v1.NewCompiler().WithDefaultRegoVersion(DefaultRegoVersion)
}

// ModuleLoader defines the interface that callers can implement to enable lazy
// loading of modules during compilation.
type ModuleLoader = v1.ModuleLoader

// SafetyCheckVisitorParams defines the AST visitor parameters to use for collecting
// variables during the safety check. This has to be exported because it's relied on
// by the copy propagation implementation in topdown.
var SafetyCheckVisitorParams = v1.SafetyCheckVisitorParams

// ComprehensionIndex specifies how the comprehension term can be indexed. The keys
// tell the evaluator what variables to use for indexing. In the future, the index
// could be expanded with more information that would allow the evaluator to index
// a larger fragment of comprehensions (e.g., by closing over variables in the outer
// query.)
type ComprehensionIndex = v1.ComprehensionIndex

// ModuleTreeNode represents a node in the module tree. The module
// tree is keyed by the package path.
type ModuleTreeNode = v1.ModuleTreeNode

// TreeNode represents a node in the rule tree. The rule tree is keyed by
// rule path.
type TreeNode = v1.TreeNode

// NewRuleTree returns a new TreeNode that represents the root
// of the rule tree populated with the given rules.
func NewRuleTree(mtree *ModuleTreeNode) *TreeNode {
	return v1.NewRuleTree(mtree)
}

// Graph represents the graph of dependencies between rules.
type Graph = v1.Graph

// NewGraph returns a new Graph based on modules. The list function must return
// the rules referred to directly by the ref.
func NewGraph(modules map[string]*Module, list func(Ref) []*Rule) *Graph {
	return v1.NewGraph(modules, list)
}

// GraphTraversal is a Traversal that understands the dependency graph
type GraphTraversal = v1.GraphTraversal

// NewGraphTraversal returns a Traversal for the dependency graph
func NewGraphTraversal(graph *Graph) *GraphTraversal {
	return v1.NewGraphTraversal(graph)
}

// OutputVarsFromBody returns all variables which are the "output" for
// the given body. For safety checks this means that they would be
// made safe by the body.
func OutputVarsFromBody(c *Compiler, body Body, safe VarSet) VarSet {
	return v1.OutputVarsFromBody(c, body, safe)
}

// OutputVarsFromExpr returns all variables which are the "output" for
// the given expression. For safety checks this means that they would be
// made safe by the expr.
func OutputVarsFromExpr(c *Compiler, expr *Expr, safe VarSet) VarSet {
	return v1.OutputVarsFromExpr(c, expr, safe)
}
