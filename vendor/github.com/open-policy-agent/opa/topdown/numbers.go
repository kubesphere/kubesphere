// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"fmt"
	"math/big"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

type randIntCachingKey string

var one = big.NewInt(1)

func builtinNumbersRange(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	x, err := builtins.BigIntOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	y, err := builtins.BigIntOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	ast, err := generateRange(bctx, x, y, one, "numbers.range")
	if err != nil {
		return err
	}

	return iter(ast)
}

func builtinNumbersRangeStep(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	x, err := builtins.BigIntOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	y, err := builtins.BigIntOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	step, err := builtins.BigIntOperand(operands[2].Value, 3)
	if err != nil {
		return err
	}

	if step.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("numbers.range_step: step must be a positive number above zero")
	}

	ast, err := generateRange(bctx, x, y, step, "numbers.range_step")
	if err != nil {
		return err
	}

	return iter(ast)
}

func generateRange(bctx BuiltinContext, x *big.Int, y *big.Int, step *big.Int, funcName string) (*ast.Term, error) {

	cmp := x.Cmp(y)

	comp := func(i *big.Int, y *big.Int) bool { return i.Cmp(y) <= 0 }
	iter := func(i *big.Int) *big.Int { return i.Add(i, step) }

	if cmp > 0 {
		comp = func(i *big.Int, y *big.Int) bool { return i.Cmp(y) >= 0 }
		iter = func(i *big.Int) *big.Int { return i.Sub(i, step) }
	}

	result := ast.NewArray()
	haltErr := Halt{
		Err: &Error{
			Code:    CancelErr,
			Message: fmt.Sprintf("%s: timed out before generating all numbers in range", funcName),
		},
	}

	for i := new(big.Int).Set(x); comp(i, y); i = iter(i) {
		if bctx.Cancel != nil && bctx.Cancel.Cancelled() {
			return nil, haltErr
		}
		result = result.Append(ast.NewTerm(builtins.IntToNumber(i)))
	}

	return ast.NewTerm(result), nil
}

func builtinRandIntn(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	strOp, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err

	}

	n, err := builtins.IntOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	if n == 0 {
		return iter(ast.IntNumberTerm(0))
	}

	if n < 0 {
		n = -n
	}

	var key = randIntCachingKey(fmt.Sprintf("%s-%d", strOp, n))

	if val, ok := bctx.Cache.Get(key); ok {
		return iter(val.(*ast.Term))
	}

	r, err := bctx.Rand()
	if err != nil {
		return err
	}
	result := ast.IntNumberTerm(r.Intn(n))
	bctx.Cache.Put(key, result)

	return iter(result)
}

func init() {
	RegisterBuiltinFunc(ast.NumbersRange.Name, builtinNumbersRange)
	RegisterBuiltinFunc(ast.NumbersRangeStep.Name, builtinNumbersRangeStep)
	RegisterBuiltinFunc(ast.RandIntn.Name, builtinRandIntn)
}
