// Copyright 2021 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package ptr provides utilities for pointer operations using storage layer paths.
package ptr

import (
	"strconv"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/storage"
	"github.com/open-policy-agent/opa/v1/storage/internal/errors"
)

func Ptr(data interface{}, path storage.Path) (interface{}, error) {
	node := data
	for i := range path {
		key := path[i]
		switch curr := node.(type) {
		case map[string]interface{}:
			var ok bool
			if node, ok = curr[key]; !ok {
				return nil, errors.NewNotFoundError(path)
			}
		case []interface{}:
			pos, err := ValidateArrayIndex(curr, key, path)
			if err != nil {
				return nil, err
			}
			node = curr[pos]
		default:
			return nil, errors.NewNotFoundError(path)
		}
	}

	return node, nil
}

func ValuePtr(data ast.Value, path storage.Path) (ast.Value, error) {
	node := data
	for i := range path {
		key := path[i]
		switch curr := node.(type) {
		case ast.Object:
			// This term is only created for the lookup, which is not.. ideal.
			// By using the pool, we can at least avoid allocating the term itself,
			// while still having to pay 1 allocation for the value. A better solution
			// would be dynamically interned string terms.
			keyTerm := ast.TermPtrPool.Get()
			keyTerm.Value = ast.String(key)

			val := curr.Get(keyTerm)
			ast.TermPtrPool.Put(keyTerm)
			if val == nil {
				return nil, errors.NewNotFoundError(path)
			}
			node = val.Value
		case *ast.Array:
			pos, err := ValidateASTArrayIndex(curr, key, path)
			if err != nil {
				return nil, err
			}
			node = curr.Elem(pos).Value
		default:
			return nil, errors.NewNotFoundError(path)
		}
	}

	return node, nil
}

func ValidateArrayIndex(arr []interface{}, s string, path storage.Path) (int, error) {
	idx, ok := isInt(s)
	if !ok {
		return 0, errors.NewNotFoundErrorWithHint(path, errors.ArrayIndexTypeMsg)
	}
	return inRange(idx, arr, path)
}

func ValidateASTArrayIndex(arr *ast.Array, s string, path storage.Path) (int, error) {
	idx, ok := isInt(s)
	if !ok {
		return 0, errors.NewNotFoundErrorWithHint(path, errors.ArrayIndexTypeMsg)
	}
	return inRange(idx, arr, path)
}

// ValidateArrayIndexForWrite also checks that `s` is a valid way to address an
// array element like `ValidateArrayIndex`, but returns a `resource_conflict` error
// if it is not.
func ValidateArrayIndexForWrite(arr []interface{}, s string, i int, path storage.Path) (int, error) {
	idx, ok := isInt(s)
	if !ok {
		return 0, errors.NewWriteConflictError(path[:i-1])
	}
	return inRange(idx, arr, path)
}

func isInt(s string) (int, bool) {
	idx, err := strconv.Atoi(s)
	return idx, err == nil
}

func inRange(i int, arr interface{}, path storage.Path) (int, error) {

	var arrLen int

	switch v := arr.(type) {
	case []interface{}:
		arrLen = len(v)
	case *ast.Array:
		arrLen = v.Len()
	}

	if i < 0 || i >= arrLen {
		return 0, errors.NewNotFoundErrorWithHint(path, errors.OutOfRangeMsg)
	}
	return i, nil
}
