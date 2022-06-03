// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"fmt"

	"github.com/open-policy-agent/opa/ast"
)

var errConflictingDoc = fmt.Errorf("conflicting documents")
var errBadPath = fmt.Errorf("bad document path")

func mergeTermWithValues(exist *ast.Term, pairs [][2]*ast.Term) (*ast.Term, error) {

	var result *ast.Term
	var init bool

	for _, pair := range pairs {

		if err := ast.IsValidImportPath(pair[0].Value); err != nil {
			return nil, errBadPath
		}

		target := pair[0].Value.(ast.Ref)

		if len(target) == 1 {
			result = pair[1]
			init = true
		} else {
			if !init {
				result = exist.Copy()
				init = true
			}

			if result == nil {
				result = ast.NewTerm(makeTree(target[1:], pair[1]))
			} else {
				node := result
				done := false
				for i := 1; i < len(target)-1 && !done; i++ {
					if child := node.Get(target[i]); child == nil {
						obj, ok := node.Value.(ast.Object)
						if !ok {
							return nil, errConflictingDoc
						}
						obj.Insert(target[i], ast.NewTerm(makeTree(target[i+1:], pair[1])))
						done = true
					} else {
						node = child
					}
				}
				if !done {
					obj, ok := node.Value.(ast.Object)
					if !ok {
						return nil, errConflictingDoc
					}
					obj.Insert(target[len(target)-1], pair[1])
				}
			}
		}
	}

	if !init {
		result = exist
	}

	return result, nil
}

// makeTree returns an object that represents a document where the value v is
// the leaf and elements in k represent intermediate objects.
func makeTree(k ast.Ref, v *ast.Term) ast.Object {
	var obj ast.Object
	for i := len(k) - 1; i >= 1; i-- {
		obj = ast.NewObject(ast.Item(k[i], v))
		v = &ast.Term{Value: obj}
	}
	obj = ast.NewObject(ast.Item(k[0], v))
	return obj
}
