// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"io"

	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// Pretty writes a pretty representation of the AST rooted at x to w.
//
// This is function is intended for debug purposes when inspecting ASTs.
func Pretty(w io.Writer, x interface{}) {
	v1.Pretty(w, x)
}
