// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// ValueMap represents a key/value map between AST term values. Any type of term
// can be used as a key in the map.
type ValueMap = v1.ValueMap

// NewValueMap returns a new ValueMap.
func NewValueMap() *ValueMap {
	return v1.NewValueMap()
}
