// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"strconv"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

func builtinToNumber(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch a := operands[0].Value.(type) {
	case ast.Null:
		return iter(ast.NumberTerm("0"))
	case ast.Boolean:
		if a {
			return iter(ast.NumberTerm("1"))
		}
		return iter(ast.NumberTerm("0"))
	case ast.Number:
		return iter(ast.NewTerm(a))
	case ast.String:
		_, err := strconv.ParseFloat(string(a), 64)
		if err != nil {
			return err
		}
		return iter(ast.NewTerm(ast.Number(a)))
	}
	return builtins.NewOperandTypeErr(1, operands[0].Value, "null", "boolean", "number", "string")
}

// Deprecated in v0.13.0.
func builtinToArray(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch val := operands[0].Value.(type) {
	case *ast.Array:
		return iter(ast.NewTerm(val))
	case ast.Set:
		arr := make([]*ast.Term, val.Len())
		i := 0
		val.Foreach(func(term *ast.Term) {
			arr[i] = term
			i++
		})
		return iter(ast.NewTerm(ast.NewArray(arr...)))
	default:
		return builtins.NewOperandTypeErr(1, operands[0].Value, "array", "set")
	}
}

// Deprecated in v0.13.0.
func builtinToSet(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch val := operands[0].Value.(type) {
	case *ast.Array:
		s := ast.NewSet()
		val.Foreach(func(v *ast.Term) {
			s.Add(v)
		})
		return iter(ast.NewTerm(s))
	case ast.Set:
		return iter(ast.NewTerm(val))
	default:
		return builtins.NewOperandTypeErr(1, operands[0].Value, "array", "set")
	}
}

// Deprecated in v0.13.0.
func builtinToString(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch val := operands[0].Value.(type) {
	case ast.String:
		return iter(ast.NewTerm(val))
	default:
		return builtins.NewOperandTypeErr(1, operands[0].Value, "string")
	}
}

// Deprecated in v0.13.0.
func builtinToBoolean(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch val := operands[0].Value.(type) {
	case ast.Boolean:
		return iter(ast.NewTerm(val))
	default:
		return builtins.NewOperandTypeErr(1, operands[0].Value, "boolean")
	}
}

// Deprecated in v0.13.0.
func builtinToNull(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch val := operands[0].Value.(type) {
	case ast.Null:
		return iter(ast.NewTerm(val))
	default:
		return builtins.NewOperandTypeErr(1, operands[0].Value, "null")
	}
}

// Deprecated in v0.13.0.
func builtinToObject(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch val := operands[0].Value.(type) {
	case ast.Object:
		return iter(ast.NewTerm(val))
	default:
		return builtins.NewOperandTypeErr(1, operands[0].Value, "object")
	}
}

func init() {
	RegisterBuiltinFunc(ast.ToNumber.Name, builtinToNumber)
	RegisterBuiltinFunc(ast.CastArray.Name, builtinToArray)
	RegisterBuiltinFunc(ast.CastSet.Name, builtinToSet)
	RegisterBuiltinFunc(ast.CastString.Name, builtinToString)
	RegisterBuiltinFunc(ast.CastBoolean.Name, builtinToBoolean)
	RegisterBuiltinFunc(ast.CastNull.Name, builtinToNull)
	RegisterBuiltinFunc(ast.CastObject.Name, builtinToObject)
}
