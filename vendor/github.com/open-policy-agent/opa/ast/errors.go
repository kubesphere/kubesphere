// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// Errors represents a series of errors encountered during parsing, compiling,
// etc.
type Errors = v1.Errors

const (
	// ParseErr indicates an unclassified parse error occurred.
	ParseErr = v1.ParseErr

	// CompileErr indicates an unclassified compile error occurred.
	CompileErr = v1.CompileErr

	// TypeErr indicates a type error was caught.
	TypeErr = v1.TypeErr

	// UnsafeVarErr indicates an unsafe variable was found during compilation.
	UnsafeVarErr = v1.UnsafeVarErr

	// RecursionErr indicates recursion was found during compilation.
	RecursionErr = v1.RecursionErr
)

// IsError returns true if err is an AST error with code.
func IsError(code string, err error) bool {
	return v1.IsError(code, err)
}

// ErrorDetails defines the interface for detailed error messages.
type ErrorDetails = v1.ErrorDetails

// Error represents a single error caught during parsing, compiling, etc.
type Error = v1.Error

// NewError returns a new Error object.
func NewError(code string, loc *Location, f string, a ...interface{}) *Error {
	return v1.NewError(code, loc, f, a...)
}
