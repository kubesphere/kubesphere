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
