// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"fmt"

	"github.com/open-policy-agent/opa/ast"
)

func builtinTypeName(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch operands[0].Value.(type) {
	case ast.Null:
		return iter(ast.StringTerm("null"))
	case ast.Boolean:
		return iter(ast.StringTerm("boolean"))
	case ast.Number:
		return iter(ast.StringTerm("number"))
	case ast.String:
		return iter(ast.StringTerm("string"))
	case *ast.Array:
		return iter(ast.StringTerm("array"))
	case ast.Object:
		return iter(ast.StringTerm("object"))
	case ast.Set:
		return iter(ast.StringTerm("set"))
	}

	return fmt.Errorf("illegal value")
}

func init() {
	RegisterBuiltinFunc(ast.TypeNameBuiltin.Name, builtinTypeName)
}
