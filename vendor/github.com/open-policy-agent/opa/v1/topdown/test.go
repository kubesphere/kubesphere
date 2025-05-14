// Copyright 2025 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import "github.com/open-policy-agent/opa/v1/ast"

const TestCaseOp Op = "TestCase"

func builtinTestCase(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	e := &Event{
		Op:      TestCaseOp,
		QueryID: bctx.QueryID,
		Node: ast.NewExpr([]*ast.Term{
			ast.NewTerm(ast.InternalTestCase.Ref()),
			ast.NewTerm(operands[0].Value),
		}),
	}

	for _, tracer := range bctx.QueryTracers {
		tracer.TraceEvent(*e)
	}

	return iter(ast.BooleanTerm(true))
}

func init() {
	RegisterBuiltinFunc(ast.InternalTestCase.Name, builtinTestCase)
}
