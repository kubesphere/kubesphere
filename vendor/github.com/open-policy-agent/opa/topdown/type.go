// Copyright 2022 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"github.com/open-policy-agent/opa/ast"
)

func builtinIsNumber(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch operands[0].Value.(type) {
	case ast.Number:
		return iter(ast.BooleanTerm(true))
	default:
		return iter(ast.BooleanTerm(false))
	}
}

func builtinIsString(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch operands[0].Value.(type) {
	case ast.String:
		return iter(ast.BooleanTerm(true))
	default:
		return iter(ast.BooleanTerm(false))
	}
}

func builtinIsBoolean(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch operands[0].Value.(type) {
	case ast.Boolean:
		return iter(ast.BooleanTerm(true))
	default:
		return iter(ast.BooleanTerm(false))
	}
}

func builtinIsArray(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch operands[0].Value.(type) {
	case *ast.Array:
		return iter(ast.BooleanTerm(true))
	default:
		return iter(ast.BooleanTerm(false))
	}
}

func builtinIsSet(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch operands[0].Value.(type) {
	case ast.Set:
		return iter(ast.BooleanTerm(true))
	default:
		return iter(ast.BooleanTerm(false))
	}
}

func builtinIsObject(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch operands[0].Value.(type) {
	case ast.Object:
		return iter(ast.BooleanTerm(true))
	default:
		return iter(ast.BooleanTerm(false))
	}
}

func builtinIsNull(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch operands[0].Value.(type) {
	case ast.Null:
		return iter(ast.BooleanTerm(true))
	default:
		return iter(ast.BooleanTerm(false))
	}
}

func init() {
	RegisterBuiltinFunc(ast.IsNumber.Name, builtinIsNumber)
	RegisterBuiltinFunc(ast.IsString.Name, builtinIsString)
	RegisterBuiltinFunc(ast.IsBoolean.Name, builtinIsBoolean)
	RegisterBuiltinFunc(ast.IsArray.Name, builtinIsArray)
	RegisterBuiltinFunc(ast.IsSet.Name, builtinIsSet)
	RegisterBuiltinFunc(ast.IsObject.Name, builtinIsObject)
	RegisterBuiltinFunc(ast.IsNull.Name, builtinIsNull)
}
