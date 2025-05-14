// Copyright 2019 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"slices"
	"strings"
)

// CheckPathConflicts returns a set of errors indicating paths that
// are in conflict with the result of the provided callable.
func CheckPathConflicts(c *Compiler, exists func([]string) (bool, error)) Errors {
	var errs Errors

	root := c.RuleTree.Child(DefaultRootDocument.Value)
	if root == nil {
		return nil
	}

	if len(c.pathConflictCheckRoots) == 0 || slices.Contains(c.pathConflictCheckRoots, "") {
		for _, child := range root.Children {
			errs = append(errs, checkDocumentConflicts(child, exists, nil)...)
		}
		return errs
	}

	for _, rootPath := range c.pathConflictCheckRoots {
		// traverse AST from `path` to go to the new root
		paths := strings.Split(rootPath, "/")
		node := root
		for _, key := range paths {
			node = node.Child(String(key))
			if node == nil {
				break
			}
		}

		if node == nil {
			// could not find the node from the AST (e.g. `path` is from a data file)
			// then no conflict is possible
			continue
		}

		for _, child := range node.Children {
			errs = append(errs, checkDocumentConflicts(child, exists, paths)...)
		}
	}

	return errs
}

func checkDocumentConflicts(node *TreeNode, exists func([]string) (bool, error), path []string) Errors {

	switch key := node.Key.(type) {
	case String:
		path = append(path, string(key))
	default: // other key types cannot conflict with data
		return nil
	}

	if len(node.Values) > 0 {
		s := strings.Join(path, "/")
		if ok, err := exists(path); err != nil {
			return Errors{NewError(CompileErr, node.Values[0].(*Rule).Loc(), "conflict check for data path %v: %v", s, err.Error())}
		} else if ok {
			return Errors{NewError(CompileErr, node.Values[0].(*Rule).Loc(), "conflicting rule for data path %v found", s)}
		}
	}

	var errs Errors

	for _, child := range node.Children {
		errs = append(errs, checkDocumentConflicts(child, exists, path)...)
	}

	return errs
}
