// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package storage

import (
	"github.com/open-policy-agent/opa/ast"
	v1 "github.com/open-policy-agent/opa/v1/storage"
)

// Path refers to a document in storage.
type Path = v1.Path

// ParsePath returns a new path for the given str.
func ParsePath(str string) (path Path, ok bool) {
	return v1.ParsePath(str)
}

// ParsePathEscaped returns a new path for the given escaped str.
func ParsePathEscaped(str string) (path Path, ok bool) {
	return v1.ParsePathEscaped(str)
}

// NewPathForRef returns a new path for the given ref.
func NewPathForRef(ref ast.Ref) (path Path, err error) {
	return v1.NewPathForRef(ref)
}

// MustParsePath returns a new Path for s. If s cannot be parsed, this function
// will panic. This is mostly for test purposes.
func MustParsePath(s string) Path {
	return v1.MustParsePath(s)
}
