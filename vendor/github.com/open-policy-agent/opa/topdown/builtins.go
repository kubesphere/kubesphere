// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

type (
	// FunctionalBuiltin1 is deprecated. Use BuiltinFunc instead.
	FunctionalBuiltin1 func(op1 ast.Value) (output ast.Value, err error)

	// FunctionalBuiltin2 is deprecated. Use BuiltinFunc instead.
	FunctionalBuiltin2 func(op1, op2 ast.Value) (output ast.Value, err error)

	// FunctionalBuiltin3 is deprecated. Use BuiltinFunc instead.
	FunctionalBuiltin3 func(op1, op2, op3 ast.Value) (output ast.Value, err error)

	// FunctionalBuiltin4 is deprecated. Use BuiltinFunc instead.
	FunctionalBuiltin4 func(op1, op2, op3, op4 ast.Value) (output ast.Value, err error)

	// BuiltinContext contains context from the evaluator that may be used by
	// built-in functions.
	BuiltinContext struct {
		Context  context.Context // request context that was passed when query started
		Cancel   Cancel          // atomic value that signals evaluation to halt
		Runtime  *ast.Term       // runtime information on the OPA instance
		Cache    builtins.Cache  // built-in function state cache
		Location *ast.Location   // location of built-in call
		Tracers  []Tracer        // tracer objects for trace() built-in function
		QueryID  uint64          // identifies query being evaluated
		ParentID uint64          // identifies parent of query being evaluated
	}

	// BuiltinFunc defines an interface for implementing built-in functions.
	// The built-in function is called with the plugged operands from the call
	// (including the output operands.) The implementation should evaluate the
	// operands and invoke the iterator for each successful/defined output
	// value.
	BuiltinFunc func(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error
)

// RegisterBuiltinFunc adds a new built-in function to the evaluation engine.
func RegisterBuiltinFunc(name string, f BuiltinFunc) {
	builtinFunctions[name] = builtinErrorWrapper(name, f)
}

// RegisterFunctionalBuiltin1 is deprecated use RegisterBuiltinFunc instead.
func RegisterFunctionalBuiltin1(name string, fun FunctionalBuiltin1) {
	builtinFunctions[name] = functionalWrapper1(name, fun)
}

// RegisterFunctionalBuiltin2 is deprecated use RegisterBuiltinFunc instead.
func RegisterFunctionalBuiltin2(name string, fun FunctionalBuiltin2) {
	builtinFunctions[name] = functionalWrapper2(name, fun)
}

// RegisterFunctionalBuiltin3 is deprecated use RegisterBuiltinFunc instead.
func RegisterFunctionalBuiltin3(name string, fun FunctionalBuiltin3) {
	builtinFunctions[name] = functionalWrapper3(name, fun)
}

// RegisterFunctionalBuiltin4 is deprecated use RegisterBuiltinFunc instead.
func RegisterFunctionalBuiltin4(name string, fun FunctionalBuiltin4) {
	builtinFunctions[name] = functionalWrapper4(name, fun)
}

// GetBuiltin returns a built-in function implementation, nil if no built-in found.
func GetBuiltin(name string) BuiltinFunc {
	return builtinFunctions[name]
}

// BuiltinEmpty is deprecated.
type BuiltinEmpty struct{}

func (BuiltinEmpty) Error() string {
	return "<empty>"
}

var builtinFunctions = map[string]BuiltinFunc{}

func builtinErrorWrapper(name string, fn BuiltinFunc) BuiltinFunc {
	return func(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
		err := fn(bctx, args, iter)
		if err == nil {
			return nil
		}
		return handleBuiltinErr(name, bctx.Location, err)
	}
}

func functionalWrapper1(name string, fn FunctionalBuiltin1) BuiltinFunc {
	return func(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
		result, err := fn(args[0].Value)
		if err == nil {
			return iter(ast.NewTerm(result))
		}
		return handleBuiltinErr(name, bctx.Location, err)
	}
}

func functionalWrapper2(name string, fn FunctionalBuiltin2) BuiltinFunc {
	return func(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
		result, err := fn(args[0].Value, args[1].Value)
		if err == nil {
			return iter(ast.NewTerm(result))
		}
		return handleBuiltinErr(name, bctx.Location, err)
	}
}

func functionalWrapper3(name string, fn FunctionalBuiltin3) BuiltinFunc {
	return func(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
		result, err := fn(args[0].Value, args[1].Value, args[2].Value)
		if err == nil {
			return iter(ast.NewTerm(result))
		}
		return handleBuiltinErr(name, bctx.Location, err)
	}
}

func functionalWrapper4(name string, fn FunctionalBuiltin4) BuiltinFunc {
	return func(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
		result, err := fn(args[0].Value, args[1].Value, args[2].Value, args[3].Value)
		if err == nil {
			return iter(ast.NewTerm(result))
		}
		if _, empty := err.(BuiltinEmpty); empty {
			return nil
		}
		return handleBuiltinErr(name, bctx.Location, err)
	}
}

func handleBuiltinErr(name string, loc *ast.Location, err error) error {
	switch err := err.(type) {
	case BuiltinEmpty:
		return nil
	case *Error:
		return err
	case builtins.ErrOperand:
		return &Error{
			Code:     TypeErr,
			Message:  fmt.Sprintf("%v: %v", string(name), err.Error()),
			Location: loc,
		}
	default:
		return &Error{
			Code:     BuiltinErr,
			Message:  fmt.Sprintf("%v: %v", string(name), err.Error()),
			Location: loc,
		}
	}
}
