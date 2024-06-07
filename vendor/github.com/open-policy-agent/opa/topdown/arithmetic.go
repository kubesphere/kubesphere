// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"fmt"
	"math/big"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

type arithArity1 func(a *big.Float) (*big.Float, error)
type arithArity2 func(a, b *big.Float) (*big.Float, error)

func arithAbs(a *big.Float) (*big.Float, error) {
	return a.Abs(a), nil
}

var halfAwayFromZero = big.NewFloat(0.5)

func arithRound(a *big.Float) (*big.Float, error) {
	var i *big.Int
	if a.Signbit() {
		i, _ = new(big.Float).Sub(a, halfAwayFromZero).Int(nil)
	} else {
		i, _ = new(big.Float).Add(a, halfAwayFromZero).Int(nil)
	}
	return new(big.Float).SetInt(i), nil
}

func arithCeil(a *big.Float) (*big.Float, error) {
	i, _ := a.Int(nil)
	f := new(big.Float).SetInt(i)

	if f.Signbit() || a.Cmp(f) == 0 {
		return f, nil
	}

	return new(big.Float).Add(f, big.NewFloat(1.0)), nil
}

func arithFloor(a *big.Float) (*big.Float, error) {
	i, _ := a.Int(nil)
	f := new(big.Float).SetInt(i)

	if !f.Signbit() || a.Cmp(f) == 0 {
		return f, nil
	}

	return new(big.Float).Sub(f, big.NewFloat(1.0)), nil
}

func builtinPlus(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	n1, err := builtins.NumberOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	n2, err := builtins.NumberOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	x, ok1 := n1.Int()
	y, ok2 := n2.Int()

	if ok1 && ok2 && inSmallIntRange(x) && inSmallIntRange(y) {
		return iter(ast.IntNumberTerm(x + y))
	}

	f, err := arithPlus(builtins.NumberToFloat(n1), builtins.NumberToFloat(n2))
	if err != nil {
		return err
	}
	return iter(ast.NewTerm(builtins.FloatToNumber(f)))
}

func builtinMultiply(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	n1, err := builtins.NumberOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	n2, err := builtins.NumberOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	x, ok1 := n1.Int()
	y, ok2 := n2.Int()

	if ok1 && ok2 && inSmallIntRange(x) && inSmallIntRange(y) {
		return iter(ast.IntNumberTerm(x * y))
	}

	f, err := arithMultiply(builtins.NumberToFloat(n1), builtins.NumberToFloat(n2))
	if err != nil {
		return err
	}
	return iter(ast.NewTerm(builtins.FloatToNumber(f)))
}

func arithPlus(a, b *big.Float) (*big.Float, error) {
	return new(big.Float).Add(a, b), nil
}

func arithMinus(a, b *big.Float) (*big.Float, error) {
	return new(big.Float).Sub(a, b), nil
}

func arithMultiply(a, b *big.Float) (*big.Float, error) {
	return new(big.Float).Mul(a, b), nil
}

func arithDivide(a, b *big.Float) (*big.Float, error) {
	i, acc := b.Int64()
	if acc == big.Exact && i == 0 {
		return nil, fmt.Errorf("divide by zero")
	}
	return new(big.Float).Quo(a, b), nil
}

func arithRem(a, b *big.Int) (*big.Int, error) {
	if b.Int64() == 0 {
		return nil, fmt.Errorf("modulo by zero")
	}
	return new(big.Int).Rem(a, b), nil
}

func builtinArithArity1(fn arithArity1) BuiltinFunc {
	return func(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
		n, err := builtins.NumberOperand(operands[0].Value, 1)
		if err != nil {
			return err
		}
		f, err := fn(builtins.NumberToFloat(n))
		if err != nil {
			return err
		}
		return iter(ast.NewTerm(builtins.FloatToNumber(f)))
	}
}

func builtinArithArity2(fn arithArity2) BuiltinFunc {
	return func(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
		n1, err := builtins.NumberOperand(operands[0].Value, 1)
		if err != nil {
			return err
		}
		n2, err := builtins.NumberOperand(operands[1].Value, 2)
		if err != nil {
			return err
		}
		f, err := fn(builtins.NumberToFloat(n1), builtins.NumberToFloat(n2))
		if err != nil {
			return err
		}
		return iter(ast.NewTerm(builtins.FloatToNumber(f)))
	}
}

func builtinMinus(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	n1, ok1 := operands[0].Value.(ast.Number)
	n2, ok2 := operands[1].Value.(ast.Number)

	if ok1 && ok2 {

		x, okx := n1.Int()
		y, oky := n2.Int()

		if okx && oky && inSmallIntRange(x) && inSmallIntRange(y) {
			return iter(ast.IntNumberTerm(x - y))
		}

		f, err := arithMinus(builtins.NumberToFloat(n1), builtins.NumberToFloat(n2))
		if err != nil {
			return err
		}
		return iter(ast.NewTerm(builtins.FloatToNumber(f)))
	}

	s1, ok3 := operands[0].Value.(ast.Set)
	s2, ok4 := operands[1].Value.(ast.Set)

	if ok3 && ok4 {
		return iter(ast.NewTerm(s1.Diff(s2)))
	}

	if !ok1 && !ok3 {
		return builtins.NewOperandTypeErr(1, operands[0].Value, "number", "set")
	}

	if ok2 {
		return builtins.NewOperandTypeErr(2, operands[1].Value, "set")
	}

	return builtins.NewOperandTypeErr(2, operands[1].Value, "number")
}

func builtinRem(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	n1, ok1 := operands[0].Value.(ast.Number)
	n2, ok2 := operands[1].Value.(ast.Number)

	if ok1 && ok2 {

		x, okx := n1.Int()
		y, oky := n2.Int()

		if okx && oky && inSmallIntRange(x) && inSmallIntRange(y) {
			if y == 0 {
				return fmt.Errorf("modulo by zero")
			}

			return iter(ast.IntNumberTerm(x % y))
		}

		op1, err1 := builtins.NumberToInt(n1)
		op2, err2 := builtins.NumberToInt(n2)

		if err1 != nil || err2 != nil {
			return fmt.Errorf("modulo on floating-point number")
		}

		i, err := arithRem(op1, op2)
		if err != nil {
			return err
		}
		return iter(ast.NewTerm(builtins.IntToNumber(i)))
	}

	if !ok1 {
		return builtins.NewOperandTypeErr(1, operands[0].Value, "number")
	}

	return builtins.NewOperandTypeErr(2, operands[1].Value, "number")
}

func inSmallIntRange(num int) bool {
	return -1000 < num && num < 1000
}

func init() {
	RegisterBuiltinFunc(ast.Abs.Name, builtinArithArity1(arithAbs))
	RegisterBuiltinFunc(ast.Round.Name, builtinArithArity1(arithRound))
	RegisterBuiltinFunc(ast.Ceil.Name, builtinArithArity1(arithCeil))
	RegisterBuiltinFunc(ast.Floor.Name, builtinArithArity1(arithFloor))
	RegisterBuiltinFunc(ast.Plus.Name, builtinPlus)
	RegisterBuiltinFunc(ast.Minus.Name, builtinMinus)
	RegisterBuiltinFunc(ast.Multiply.Name, builtinMultiply)
	RegisterBuiltinFunc(ast.Divide.Name, builtinArithArity2(arithDivide))
	RegisterBuiltinFunc(ast.Rem.Name, builtinRem)
}
