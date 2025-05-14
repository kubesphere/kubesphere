// Copyright 2021 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// SchemaSet holds a map from a path to a schema.
type SchemaSet = v1.SchemaSet

// NewSchemaSet returns an empty SchemaSet.
func NewSchemaSet() *SchemaSet {
	return v1.NewSchemaSet()
}
