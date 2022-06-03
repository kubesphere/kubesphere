// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"math/big"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

var one = big.NewInt(1)

func builtinNumbersRange(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	x, err := builtins.BigIntOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	y, err := builtins.BigIntOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	result := ast.NewArray()
	cmp := x.Cmp(y)

	if cmp <= 0 {
		for i := new(big.Int).Set(x); i.Cmp(y) <= 0; i = i.Add(i, one) {
			result = result.Append(ast.NewTerm(builtins.IntToNumber(i)))
		}
	} else {
		for i := new(big.Int).Set(x); i.Cmp(y) >= 0; i = i.Sub(i, one) {
			result = result.Append(ast.NewTerm(builtins.IntToNumber(i)))
		}
	}

	return iter(ast.NewTerm(result))
}

func init() {
	RegisterBuiltinFunc(ast.NumbersRange.Name, builtinNumbersRange)
}
