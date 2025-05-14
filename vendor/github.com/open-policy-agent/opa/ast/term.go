// Copyright 2024 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"encoding/json"
	"io"

	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// Location records a position in source code.
type Location = v1.Location

// NewLocation returns a new Location object.
func NewLocation(text []byte, file string, row int, col int) *Location {
	return v1.NewLocation(text, file, row, col)
}

// Value declares the common interface for all Term values. Every kind of Term value
// in the language is represented as a type that implements this interface:
//
// - Null, Boolean, Number, String
// - Object, Array, Set
// - Variables, References
// - Array, Set, and Object Comprehensions
// - Calls
type Value = v1.Value

// InterfaceToValue converts a native Go value x to a Value.
func InterfaceToValue(x interface{}) (Value, error) {
	return v1.InterfaceToValue(x)
}

// ValueFromReader returns an AST value from a JSON serialized value in the reader.
func ValueFromReader(r io.Reader) (Value, error) {
	return v1.ValueFromReader(r)
}

// As converts v into a Go native type referred to by x.
func As(v Value, x interface{}) error {
	return v1.As(v, x)
}

// Resolver defines the interface for resolving references to native Go values.
type Resolver = v1.Resolver

// ValueResolver defines the interface for resolving references to AST values.
type ValueResolver = v1.ValueResolver

// UnknownValueErr indicates a ValueResolver was unable to resolve a reference
// because the reference refers to an unknown value.
type UnknownValueErr = v1.UnknownValueErr

// IsUnknownValueErr returns true if the err is an UnknownValueErr.
func IsUnknownValueErr(err error) bool {
	return v1.IsUnknownValueErr(err)
}

// ValueToInterface returns the Go representation of an AST value.  The AST
// value should not contain any values that require evaluation (e.g., vars,
// comprehensions, etc.)
func ValueToInterface(v Value, resolver Resolver) (interface{}, error) {
	return v1.ValueToInterface(v, resolver)
}

// JSON returns the JSON representation of v. The value must not contain any
// refs or terms that require evaluation (e.g., vars, comprehensions, etc.)
func JSON(v Value) (interface{}, error) {
	return v1.JSON(v)
}

// JSONOpt defines parameters for AST to JSON conversion.
type JSONOpt = v1.JSONOpt

// JSONWithOpt returns the JSON representation of v. The value must not contain any
// refs or terms that require evaluation (e.g., vars, comprehensions, etc.)
func JSONWithOpt(v Value, opt JSONOpt) (interface{}, error) {
	return v1.JSONWithOpt(v, opt)
}

// MustJSON returns the JSON representation of v. The value must not contain any
// refs or terms that require evaluation (e.g., vars, comprehensions, etc.) If
// the conversion fails, this function will panic. This function is mostly for
// test purposes.
func MustJSON(v Value) interface{} {
	return v1.MustJSON(v)
}

// MustInterfaceToValue converts a native Go value x to a Value. If the
// conversion fails, this function will panic. This function is mostly for test
// purposes.
func MustInterfaceToValue(x interface{}) Value {
	return v1.MustInterfaceToValue(x)
}

// Term is an argument to a function.
type Term = v1.Term

// NewTerm returns a new Term object.
func NewTerm(v Value) *Term {
	return v1.NewTerm(v)
}

// IsConstant returns true if the AST value is constant.
func IsConstant(v Value) bool {
	return v1.IsConstant(v)
}

// IsComprehension returns true if the supplied value is a comprehension.
func IsComprehension(x Value) bool {
	return v1.IsComprehension(x)
}

// ContainsRefs returns true if the Value v contains refs.
func ContainsRefs(v interface{}) bool {
	return v1.ContainsRefs(v)
}

// ContainsComprehensions returns true if the Value v contains comprehensions.
func ContainsComprehensions(v interface{}) bool {
	return v1.ContainsComprehensions(v)
}

// ContainsClosures returns true if the Value v contains closures.
func ContainsClosures(v interface{}) bool {
	return v1.ContainsClosures(v)
}

// IsScalar returns true if the AST value is a scalar.
func IsScalar(v Value) bool {
	return v1.IsScalar(v)
}

// Null represents the null value defined by JSON.
type Null = v1.Null

// NullTerm creates a new Term with a Null value.
func NullTerm() *Term {
	return v1.NullTerm()
}

// Boolean represents a boolean value defined by JSON.
type Boolean = v1.Boolean

// BooleanTerm creates a new Term with a Boolean value.
func BooleanTerm(b bool) *Term {
	return v1.BooleanTerm(b)
}

// Number represents a numeric value as defined by JSON.
type Number = v1.Number

// NumberTerm creates a new Term with a Number value.
func NumberTerm(n json.Number) *Term {
	return v1.NumberTerm(n)
}

// IntNumberTerm creates a new Term with an integer Number value.
func IntNumberTerm(i int) *Term {
	return v1.IntNumberTerm(i)
}

// UIntNumberTerm creates a new Term with an unsigned integer Number value.
func UIntNumberTerm(u uint64) *Term {
	return v1.UIntNumberTerm(u)
}

// FloatNumberTerm creates a new Term with a floating point Number value.
func FloatNumberTerm(f float64) *Term {
	return v1.FloatNumberTerm(f)
}

// String represents a string value as defined by JSON.
type String = v1.String

// StringTerm creates a new Term with a String value.
func StringTerm(s string) *Term {
	return v1.StringTerm(s)
}

// Var represents a variable as defined by the language.
type Var = v1.Var

// VarTerm creates a new Term with a Variable value.
func VarTerm(v string) *Term {
	return v1.VarTerm(v)
}

// Ref represents a reference as defined by the language.
type Ref = v1.Ref

// EmptyRef returns a new, empty reference.
func EmptyRef() Ref {
	return v1.EmptyRef()
}

// PtrRef returns a new reference against the head for the pointer
// s. Path components in the pointer are unescaped.
func PtrRef(head *Term, s string) (Ref, error) {
	return v1.PtrRef(head, s)
}

// RefTerm creates a new Term with a Ref value.
func RefTerm(r ...*Term) *Term {
	return v1.RefTerm(r...)
}

func IsVarCompatibleString(s string) bool {
	return v1.IsVarCompatibleString(s)
}

// QueryIterator defines the interface for querying AST documents with references.
type QueryIterator = v1.QueryIterator

// ArrayTerm creates a new Term with an Array value.
func ArrayTerm(a ...*Term) *Term {
	return v1.ArrayTerm(a...)
}

// NewArray creates an Array with the terms provided. The array will
// use the provided term slice.
func NewArray(a ...*Term) *Array {
	return v1.NewArray(a...)
}

// Array represents an array as defined by the language. Arrays are similar to the
// same types as defined by JSON with the exception that they can contain Vars
// and References.
type Array = v1.Array

// Set represents a set as defined by the language.
type Set = v1.Set

// NewSet returns a new Set containing t.
func NewSet(t ...*Term) Set {
	return v1.NewSet(t...)
}

func SetTerm(t ...*Term) *Term {
	return v1.SetTerm(t...)
}

// Object represents an object as defined by the language.
type Object = v1.Object

// NewObject creates a new Object with t.
func NewObject(t ...[2]*Term) Object {
	return v1.NewObject(t...)
}

// ObjectTerm creates a new Term with an Object value.
func ObjectTerm(o ...[2]*Term) *Term {
	return v1.ObjectTerm(o...)
}

func LazyObject(blob map[string]interface{}) Object {
	return v1.LazyObject(blob)
}

// Item is a helper for constructing an tuple containing two Terms
// representing a key/value pair in an Object.
func Item(key, value *Term) [2]*Term {
	return v1.Item(key, value)
}

// NOTE(philipc): The only way to get an ObjectKeyIterator should be
// from an Object. This ensures that the iterator can have implementation-
// specific details internally, with no contracts except to the very
// limited interface.
type ObjectKeysIterator = v1.ObjectKeysIterator

// ArrayComprehension represents an array comprehension as defined in the language.
type ArrayComprehension = v1.ArrayComprehension

// ArrayComprehensionTerm creates a new Term with an ArrayComprehension value.
func ArrayComprehensionTerm(term *Term, body Body) *Term {
	return v1.ArrayComprehensionTerm(term, body)
}

// ObjectComprehension represents an object comprehension as defined in the language.
type ObjectComprehension = v1.ObjectComprehension

// ObjectComprehensionTerm creates a new Term with an ObjectComprehension value.
func ObjectComprehensionTerm(key, value *Term, body Body) *Term {
	return v1.ObjectComprehensionTerm(key, value, body)
}

// SetComprehension represents a set comprehension as defined in the language.
type SetComprehension = v1.SetComprehension

// SetComprehensionTerm creates a new Term with an SetComprehension value.
func SetComprehensionTerm(term *Term, body Body) *Term {
	return v1.SetComprehensionTerm(term, body)
}

// Call represents as function call in the language.
type Call = v1.Call

// CallTerm returns a new Term with a Call value defined by terms. The first
// term is the operator and the rest are operands.
func CallTerm(terms ...*Term) *Term {
	return v1.CallTerm(terms...)
}
