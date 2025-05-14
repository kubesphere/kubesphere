// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// Compare returns an integer indicating whether two AST values are less than,
// equal to, or greater than each other.
//
// If a is less than b, the return value is negative. If a is greater than b,
// the return value is positive. If a is equal to b, the return value is zero.
//
// Different types are never equal to each other. For comparison purposes, types
// are sorted as follows:
//
// nil < Null < Boolean < Number < String < Var < Ref < Array < Object < Set <
// ArrayComprehension < ObjectComprehension < SetComprehension < Expr < SomeDecl
// < With < Body < Rule < Import < Package < Module.
//
// Arrays and Refs are equal if and only if both a and b have the same length
// and all corresponding elements are equal. If one element is not equal, the
// return value is the same as for the first differing element. If all elements
// are equal but a and b have different lengths, the shorter is considered less
// than the other.
//
// Objects are considered equal if and only if both a and b have the same sorted
// (key, value) pairs and are of the same length. Other comparisons are
// consistent but not defined.
//
// Sets are considered equal if and only if the symmetric difference of a and b
// is empty.
// Other comparisons are consistent but not defined.
func Compare(a, b interface{}) int {
	return v1.Compare(a, b)
}
