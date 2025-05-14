// Copyright 2021 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"fmt"

	"github.com/open-policy-agent/opa/v1/types"
	"github.com/open-policy-agent/opa/v1/util"
)

// SchemaSet holds a map from a path to a schema.
type SchemaSet struct {
	m *util.HasherMap[Ref, any]
}

// NewSchemaSet returns an empty SchemaSet.
func NewSchemaSet() *SchemaSet {
	return &SchemaSet{
		m: util.NewHasherMap[Ref, any](RefEqual),
	}
}

// Put inserts a raw schema into the set.
func (ss *SchemaSet) Put(path Ref, raw any) {
	ss.m.Put(path, raw)
}

// Get returns the raw schema identified by the path.
func (ss *SchemaSet) Get(path Ref) any {
	if ss != nil {
		if x, ok := ss.m.Get(path); ok {
			return x
		}
	}
	return nil
}

func loadSchema(raw any, allowNet []string) (types.Type, error) {

	jsonSchema, err := compileSchema(raw, allowNet)
	if err != nil {
		return nil, err
	}

	tpe, err := newSchemaParser().parseSchema(jsonSchema.RootSchema)
	if err != nil {
		return nil, fmt.Errorf("type checking: %w", err)
	}

	return tpe, nil
}
