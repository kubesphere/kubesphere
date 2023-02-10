// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

func builtinArrayConcat(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	arrA, err := builtins.ArrayOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	arrB, err := builtins.ArrayOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	arrC := make([]*ast.Term, arrA.Len()+arrB.Len())

	i := 0
	arrA.Foreach(func(elemA *ast.Term) {
		arrC[i] = elemA
		i++
	})

	arrB.Foreach(func(elemB *ast.Term) {
		arrC[i] = elemB
		i++
	})

	return iter(ast.NewTerm(ast.NewArray(arrC...)))
}

func builtinArraySlice(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	arr, err := builtins.ArrayOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	startIndex, err := builtins.IntOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	stopIndex, err := builtins.IntOperand(operands[2].Value, 3)
	if err != nil {
		return err
	}

	// Clamp stopIndex to avoid out-of-range errors. If negative, clamp to zero.
	// Otherwise, clamp to length of array.
	if stopIndex < 0 {
		stopIndex = 0
	} else if stopIndex > arr.Len() {
		stopIndex = arr.Len()
	}

	// Clamp startIndex to avoid out-of-range errors. If negative, clamp to zero.
	// Otherwise, clamp to stopIndex to avoid to avoid cases like arr[1:0].
	if startIndex < 0 {
		startIndex = 0
	} else if startIndex > stopIndex {
		startIndex = stopIndex
	}

	return iter(ast.NewTerm(arr.Slice(startIndex, stopIndex)))
}

func builtinArrayReverse(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	arr, err := builtins.ArrayOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	length := arr.Len()
	reversedArr := make([]*ast.Term, length)

	for index := 0; index < length; index++ {
		reversedArr[index] = arr.Elem(length - index - 1)
	}

	return iter(ast.ArrayTerm(reversedArr...))
}

func init() {
	RegisterBuiltinFunc(ast.ArrayConcat.Name, builtinArrayConcat)
	RegisterBuiltinFunc(ast.ArraySlice.Name, builtinArraySlice)
	RegisterBuiltinFunc(ast.ArrayReverse.Name, builtinArrayReverse)
}
