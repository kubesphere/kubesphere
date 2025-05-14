// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown/builtins"
)

type randIntCachingKey string

var zero = big.NewInt(0)
var one = big.NewInt(1)

func builtinNumbersRange(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	if canGenerateCheapRange(operands) {
		return generateCheapRange(operands, iter)
	}

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

	if step.Cmp(zero) <= 0 {
		return errors.New("numbers.range_step: step must be a positive number above zero")
	}

	ast, err := generateRange(bctx, x, y, step, "numbers.range_step")
	if err != nil {
		return err
	}

	return iter(ast)
}

func canGenerateCheapRange(operands []*ast.Term) bool {
	x, err := builtins.IntOperand(operands[0].Value, 1)
	if err != nil || !ast.HasInternedIntNumberTerm(x) {
		return false
	}

	y, err := builtins.IntOperand(operands[1].Value, 2)
	if err != nil || !ast.HasInternedIntNumberTerm(y) {
		return false
	}

	return true
}

func generateCheapRange(operands []*ast.Term, iter func(*ast.Term) error) error {
	x, err := builtins.IntOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	y, err := builtins.IntOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	step := 1

	if len(operands) > 2 {
		stepOp, err := builtins.IntOperand(operands[2].Value, 3)
		if err == nil {
			step = stepOp
		}
	}

	if step <= 0 {
		return errors.New("numbers.range_step: step must be a positive number above zero")
	}

	terms := make([]*ast.Term, 0, y+1)

	if x <= y {
		for i := x; i <= y; i += step {
			terms = append(terms, ast.InternedIntNumberTerm(i))
		}
	} else {
		for i := x; i >= y; i -= step {
			terms = append(terms, ast.InternedIntNumberTerm(i))
		}
	}

	return iter(ast.ArrayTerm(terms...))
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
			Message: funcName + ": timed out before generating all numbers in range",
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
		return iter(ast.InternedIntNumberTerm(0))
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
	result := ast.InternedIntNumberTerm(r.Intn(n))
	bctx.Cache.Put(key, result)

	return iter(result)
}

func init() {
	RegisterBuiltinFunc(ast.NumbersRange.Name, builtinNumbersRange)
	RegisterBuiltinFunc(ast.NumbersRangeStep.Name, builtinNumbersRangeStep)
	RegisterBuiltinFunc(ast.RandIntn.Name, builtinRandIntn)
}
