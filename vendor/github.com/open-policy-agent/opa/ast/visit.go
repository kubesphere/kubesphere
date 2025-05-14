// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import v1 "github.com/open-policy-agent/opa/v1/ast"

// Visitor defines the interface for iterating AST elements. The Visit function
// can return a Visitor w which will be used to visit the children of the AST
// element v. If the Visit function returns nil, the children will not be
// visited.
// Deprecated: use GenericVisitor or another visitor implementation
type Visitor = v1.Visitor

// BeforeAndAfterVisitor wraps Visitor to provide hooks for being called before
// and after the AST has been visited.
// Deprecated: use GenericVisitor or another visitor implementation
type BeforeAndAfterVisitor = v1.BeforeAndAfterVisitor

// Walk iterates the AST by calling the Visit function on the Visitor
// v for x before recursing.
// Deprecated: use GenericVisitor.Walk
func Walk(v Visitor, x interface{}) {
	v1.Walk(v, x)
}

// WalkBeforeAndAfter iterates the AST by calling the Visit function on the
// Visitor v for x before recursing.
// Deprecated: use GenericVisitor.Walk
func WalkBeforeAndAfter(v BeforeAndAfterVisitor, x interface{}) {
	v1.WalkBeforeAndAfter(v, x)
}

// WalkVars calls the function f on all vars under x. If the function f
// returns true, AST nodes under the last node will not be visited.
func WalkVars(x interface{}, f func(Var) bool) {
	v1.WalkVars(x, f)
}

// WalkClosures calls the function f on all closures under x. If the function f
// returns true, AST nodes under the last node will not be visited.
func WalkClosures(x interface{}, f func(interface{}) bool) {
	v1.WalkClosures(x, f)
}

// WalkRefs calls the function f on all references under x. If the function f
// returns true, AST nodes under the last node will not be visited.
func WalkRefs(x interface{}, f func(Ref) bool) {
	v1.WalkRefs(x, f)
}

// WalkTerms calls the function f on all terms under x. If the function f
// returns true, AST nodes under the last node will not be visited.
func WalkTerms(x interface{}, f func(*Term) bool) {
	v1.WalkTerms(x, f)
}

// WalkWiths calls the function f on all with modifiers under x. If the function f
// returns true, AST nodes under the last node will not be visited.
func WalkWiths(x interface{}, f func(*With) bool) {
	v1.WalkWiths(x, f)
}

// WalkExprs calls the function f on all expressions under x. If the function f
// returns true, AST nodes under the last node will not be visited.
func WalkExprs(x interface{}, f func(*Expr) bool) {
	v1.WalkExprs(x, f)
}

// WalkBodies calls the function f on all bodies under x. If the function f
// returns true, AST nodes under the last node will not be visited.
func WalkBodies(x interface{}, f func(Body) bool) {
	v1.WalkBodies(x, f)
}

// WalkRules calls the function f on all rules under x. If the function f
// returns true, AST nodes under the last node will not be visited.
func WalkRules(x interface{}, f func(*Rule) bool) {
	v1.WalkRules(x, f)
}

// WalkNodes calls the function f on all nodes under x. If the function f
// returns true, AST nodes under the last node will not be visited.
func WalkNodes(x interface{}, f func(Node) bool) {
	v1.WalkNodes(x, f)
}

// GenericVisitor provides a utility to walk over AST nodes using a
// closure. If the closure returns true, the visitor will not walk
// over AST nodes under x.
type GenericVisitor = v1.GenericVisitor

// NewGenericVisitor returns a new GenericVisitor that will invoke the function
// f on AST nodes.
func NewGenericVisitor(f func(x interface{}) bool) *GenericVisitor {
	return v1.NewGenericVisitor(f)
}

// BeforeAfterVisitor provides a utility to walk over AST nodes using
// closures. If the before closure returns true, the visitor will not
// walk over AST nodes under x. The after closure is invoked always
// after visiting a node.
type BeforeAfterVisitor = v1.BeforeAfterVisitor

// NewBeforeAfterVisitor returns a new BeforeAndAfterVisitor that
// will invoke the functions before and after AST nodes.
func NewBeforeAfterVisitor(before func(x interface{}) bool, after func(x interface{})) *BeforeAfterVisitor {
	return v1.NewBeforeAfterVisitor(before, after)
}

// VarVisitor walks AST nodes under a given node and collects all encountered
// variables. The collected variables can be controlled by specifying
// VarVisitorParams when creating the visitor.
type VarVisitor = v1.VarVisitor

// VarVisitorParams contains settings for a VarVisitor.
type VarVisitorParams = v1.VarVisitorParams

// NewVarVisitor returns a new VarVisitor object.
func NewVarVisitor() *VarVisitor {
	return v1.NewVarVisitor()
}
