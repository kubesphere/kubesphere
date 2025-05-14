// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"reflect"
	"strings"
)

// TypeName returns a human readable name for the AST element type.
func TypeName(x interface{}) string {
	if _, ok := x.(*lazyObj); ok {
		return "object"
	}
	return strings.ToLower(reflect.Indirect(reflect.ValueOf(x)).Type().Name())
}

// ValueName returns a human readable name for the AST Value type.
// This is preferrable over calling TypeName when the argument is known to be
// a Value, as this doesn't require reflection (= heap allocations).
func ValueName(x Value) string {
	switch x.(type) {
	case String:
		return "string"
	case Boolean:
		return "boolean"
	case Number:
		return "number"
	case Null:
		return "null"
	case Var:
		return "var"
	case Object:
		return "object"
	case Set:
		return "set"
	case Ref:
		return "ref"
	case Call:
		return "call"
	case *Array:
		return "array"
	case *ArrayComprehension:
		return "arraycomprehension"
	case *ObjectComprehension:
		return "objectcomprehension"
	case *SetComprehension:
		return "setcomprehension"
	}

	return TypeName(x)
}
