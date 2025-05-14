// Copyright 2019 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// CheckPathConflicts returns a set of errors indicating paths that
// are in conflict with the result of the provided callable.
func CheckPathConflicts(c *Compiler, exists func([]string) (bool, error)) Errors {
	return v1.CheckPathConflicts(c, exists)
}
