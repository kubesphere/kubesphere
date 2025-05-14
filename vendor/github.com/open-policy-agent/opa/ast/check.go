// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// UnificationErrDetail describes a type mismatch error when two values are
// unified (e.g., x = [1,2,y]).
type UnificationErrDetail = v1.UnificationErrDetail

// RefErrUnsupportedDetail describes an undefined reference error where the
// referenced value does not support dereferencing (e.g., scalars).
type RefErrUnsupportedDetail = v1.RefErrUnsupportedDetail

// RefErrInvalidDetail describes an undefined reference error where the referenced
// value does not support the reference operand (e.g., missing object key,
// invalid key type, etc.)
type RefErrInvalidDetail = v1.RefErrInvalidDetail
