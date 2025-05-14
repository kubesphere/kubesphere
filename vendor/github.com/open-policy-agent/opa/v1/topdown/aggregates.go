// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"math/big"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown/builtins"
)

func builtinCount(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch a := operands[0].Value.(type) {
	case *ast.Array:
		return iter(ast.InternedIntNumberTerm(a.Len()))
	case ast.Object:
		return iter(ast.InternedIntNumberTerm(a.Len()))
	case ast.Set:
		return iter(ast.InternedIntNumberTerm(a.Len()))
	case ast.String:
		return iter(ast.InternedIntNumberTerm(len([]rune(a))))
	}
	return builtins.NewOperandTypeErr(1, operands[0].Value, "array", "object", "set", "string")
}

func builtinSum(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch a := operands[0].Value.(type) {
	case *ast.Array:
		sum := big.NewFloat(0)
		err := a.Iter(func(x *ast.Term) error {
			n, ok := x.Value.(ast.Number)
			if !ok {
				return builtins.NewOperandElementErr(1, a, x.Value, "number")
			}
			sum = new(big.Float).Add(sum, builtins.NumberToFloat(n))
			return nil
		})
		if err != nil {
			return err
		}
		return iter(ast.NewTerm(builtins.FloatToNumber(sum)))
	case ast.Set:
		sum := big.NewFloat(0)
		err := a.Iter(func(x *ast.Term) error {
			n, ok := x.Value.(ast.Number)
			if !ok {
				return builtins.NewOperandElementErr(1, a, x.Value, "number")
			}
			sum = new(big.Float).Add(sum, builtins.NumberToFloat(n))
			return nil
		})
		if err != nil {
			return err
		}
		return iter(ast.NewTerm(builtins.FloatToNumber(sum)))
	}
	return builtins.NewOperandTypeErr(1, operands[0].Value, "set", "array")
}

func builtinProduct(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch a := operands[0].Value.(type) {
	case *ast.Array:
		product := big.NewFloat(1)
		err := a.Iter(func(x *ast.Term) error {
			n, ok := x.Value.(ast.Number)
			if !ok {
				return builtins.NewOperandElementErr(1, a, x.Value, "number")
			}
			product = new(big.Float).Mul(product, builtins.NumberToFloat(n))
			return nil
		})
		if err != nil {
			return err
		}
		return iter(ast.NewTerm(builtins.FloatToNumber(product)))
	case ast.Set:
		product := big.NewFloat(1)
		err := a.Iter(func(x *ast.Term) error {
			n, ok := x.Value.(ast.Number)
			if !ok {
				return builtins.NewOperandElementErr(1, a, x.Value, "number")
			}
			product = new(big.Float).Mul(product, builtins.NumberToFloat(n))
			return nil
		})
		if err != nil {
			return err
		}
		return iter(ast.NewTerm(builtins.FloatToNumber(product)))
	}
	return builtins.NewOperandTypeErr(1, operands[0].Value, "set", "array")
}

func builtinMax(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch a := operands[0].Value.(type) {
	case *ast.Array:
		if a.Len() == 0 {
			return nil
		}
		max := ast.InternedNullTerm.Value
		a.Foreach(func(x *ast.Term) {
			if ast.Compare(max, x.Value) <= 0 {
				max = x.Value
			}
		})
		return iter(ast.NewTerm(max))
	case ast.Set:
		if a.Len() == 0 {
			return nil
		}
		max, err := a.Reduce(ast.InternedNullTerm, func(max *ast.Term, elem *ast.Term) (*ast.Term, error) {
			if ast.Compare(max, elem) <= 0 {
				return elem, nil
			}
			return max, nil
		})
		if err != nil {
			return err
		}
		return iter(max)
	}

	return builtins.NewOperandTypeErr(1, operands[0].Value, "set", "array")
}

func builtinMin(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch a := operands[0].Value.(type) {
	case *ast.Array:
		if a.Len() == 0 {
			return nil
		}
		min := a.Elem(0).Value
		a.Foreach(func(x *ast.Term) {
			if ast.Compare(min, x.Value) >= 0 {
				min = x.Value
			}
		})
		return iter(ast.NewTerm(min))
	case ast.Set:
		if a.Len() == 0 {
			return nil
		}
		min, err := a.Reduce(ast.InternedNullTerm, func(min *ast.Term, elem *ast.Term) (*ast.Term, error) {
			// The null term is considered to be less than any other term,
			// so in order for min of a set to make sense, we need to check
			// for it.
			if min.Value.Compare(ast.InternedNullTerm.Value) == 0 {
				return elem, nil
			}

			if ast.Compare(min, elem) >= 0 {
				return elem, nil
			}
			return min, nil
		})
		if err != nil {
			return err
		}
		return iter(min)
	}

	return builtins.NewOperandTypeErr(1, operands[0].Value, "set", "array")
}

func builtinSort(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch a := operands[0].Value.(type) {
	case *ast.Array:
		return iter(ast.NewTerm(a.Sorted()))
	case ast.Set:
		return iter(ast.NewTerm(a.Sorted()))
	}
	return builtins.NewOperandTypeErr(1, operands[0].Value, "set", "array")
}

func builtinAll(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch val := operands[0].Value.(type) {
	case ast.Set:
		res := true
		match := ast.InternedBooleanTerm(true)
		val.Until(func(term *ast.Term) bool {
			if !match.Equal(term) {
				res = false
				return true
			}
			return false
		})
		return iter(ast.InternedBooleanTerm(res))
	case *ast.Array:
		res := true
		match := ast.InternedBooleanTerm(true)
		val.Until(func(term *ast.Term) bool {
			if !match.Equal(term) {
				res = false
				return true
			}
			return false
		})
		return iter(ast.InternedBooleanTerm(res))
	default:
		return builtins.NewOperandTypeErr(1, operands[0].Value, "array", "set")
	}
}

func builtinAny(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	switch val := operands[0].Value.(type) {
	case ast.Set:
		res := val.Len() > 0 && val.Contains(ast.InternedBooleanTerm(true))
		return iter(ast.InternedBooleanTerm(res))
	case *ast.Array:
		res := false
		match := ast.InternedBooleanTerm(true)
		val.Until(func(term *ast.Term) bool {
			if match.Equal(term) {
				res = true
				return true
			}
			return false
		})
		return iter(ast.InternedBooleanTerm(res))
	default:
		return builtins.NewOperandTypeErr(1, operands[0].Value, "array", "set")
	}
}

func builtinMember(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	containee := operands[0]
	switch c := operands[1].Value.(type) {
	case ast.Set:
		return iter(ast.InternedBooleanTerm(c.Contains(containee)))
	case *ast.Array:
		for i := range c.Len() {
			if c.Elem(i).Value.Compare(containee.Value) == 0 {
				return iter(ast.InternedBooleanTerm(true))
			}
		}
		return iter(ast.InternedBooleanTerm(false))
	case ast.Object:
		return iter(ast.InternedBooleanTerm(c.Until(func(_, v *ast.Term) bool {
			return v.Value.Compare(containee.Value) == 0
		})))
	}
	return iter(ast.InternedBooleanTerm(false))
}

func builtinMemberWithKey(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	key, val := operands[0], operands[1]
	switch c := operands[2].Value.(type) {
	case interface{ Get(*ast.Term) *ast.Term }:
		ret := false
		if act := c.Get(key); act != nil {
			ret = act.Value.Compare(val.Value) == 0
		}
		return iter(ast.InternedBooleanTerm(ret))
	}
	return iter(ast.InternedBooleanTerm(false))
}

func init() {
	RegisterBuiltinFunc(ast.Count.Name, builtinCount)
	RegisterBuiltinFunc(ast.Sum.Name, builtinSum)
	RegisterBuiltinFunc(ast.Product.Name, builtinProduct)
	RegisterBuiltinFunc(ast.Max.Name, builtinMax)
	RegisterBuiltinFunc(ast.Min.Name, builtinMin)
	RegisterBuiltinFunc(ast.Sort.Name, builtinSort)
	RegisterBuiltinFunc(ast.Any.Name, builtinAny)
	RegisterBuiltinFunc(ast.All.Name, builtinAll)
	RegisterBuiltinFunc(ast.Member.Name, builtinMember)
	RegisterBuiltinFunc(ast.MemberWithKey.Name, builtinMemberWithKey)
}
