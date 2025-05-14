// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// RuleIndex defines the interface for rule indices.
type RuleIndex v1.RuleIndex

// IndexResult contains the result of an index lookup.
type IndexResult = v1.IndexResult

// NewIndexResult returns a new IndexResult object.
func NewIndexResult(kind RuleKind) *IndexResult {
	return v1.NewIndexResult(kind)
}
