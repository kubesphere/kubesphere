// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import "github.com/open-policy-agent/opa/ast"

type compareFunc func(a, b ast.Value) bool

func compareGreaterThan(a, b ast.Value) bool {
	return ast.Compare(a, b) > 0
}

func compareGreaterThanEq(a, b ast.Value) bool {
	return ast.Compare(a, b) >= 0
}

func compareLessThan(a, b ast.Value) bool {
	return ast.Compare(a, b) < 0
}

func compareLessThanEq(a, b ast.Value) bool {
	return ast.Compare(a, b) <= 0
}

func compareNotEq(a, b ast.Value) bool {
	return ast.Compare(a, b) != 0
}

func compareEq(a, b ast.Value) bool {
	return ast.Compare(a, b) == 0
}

func builtinCompare(cmp compareFunc) BuiltinFunc {
	return func(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
		return iter(ast.BooleanTerm(cmp(operands[0].Value, operands[1].Value)))
	}
}

func init() {
	RegisterBuiltinFunc(ast.GreaterThan.Name, builtinCompare(compareGreaterThan))
	RegisterBuiltinFunc(ast.GreaterThanEq.Name, builtinCompare(compareGreaterThanEq))
	RegisterBuiltinFunc(ast.LessThan.Name, builtinCompare(compareLessThan))
	RegisterBuiltinFunc(ast.LessThanEq.Name, builtinCompare(compareLessThanEq))
	RegisterBuiltinFunc(ast.NotEqual.Name, builtinCompare(compareNotEq))
	RegisterBuiltinFunc(ast.Equal.Name, builtinCompare(compareEq))
}
