// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"errors"

	"github.com/open-policy-agent/opa/v1/ast"
)

var (
	nullStringTerm    = ast.StringTerm("null")
	booleanStringTerm = ast.StringTerm("boolean")
	numberStringTerm  = ast.StringTerm("number")
	stringStringTerm  = ast.StringTerm("string")
	arrayStringTerm   = ast.StringTerm("array")
	objectStringTerm  = ast.StringTerm("object")
	setStringTerm     = ast.StringTerm("set")
)

func builtinTypeName(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch operands[0].Value.(type) {
	case ast.Null:
		return iter(nullStringTerm)
	case ast.Boolean:
		return iter(booleanStringTerm)
	case ast.Number:
		return iter(numberStringTerm)
	case ast.String:
		return iter(stringStringTerm)
	case *ast.Array:
		return iter(arrayStringTerm)
	case ast.Object:
		return iter(objectStringTerm)
	case ast.Set:
		return iter(setStringTerm)
	}

	return errors.New("illegal value")
}

func init() {
	RegisterBuiltinFunc(ast.TypeNameBuiltin.Name, builtinTypeName)
}
