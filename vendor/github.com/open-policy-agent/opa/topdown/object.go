// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/internal/ref"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

func builtinObjectUnion(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	objA, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	objB, err := builtins.ObjectOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	r := mergeWithOverwrite(objA, objB)

	return iter(ast.NewTerm(r))
}

func builtinObjectUnionN(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	arr, err := builtins.ArrayOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// Because we need merge-with-overwrite behavior, we can iterate
	// back-to-front, and get a mostly correct set of key assignments that
	// give us the "last assignment wins, with merges" behavior we want.
	// However, if a non-object overwrites an object value anywhere in the
	// chain of assignments for a key, we have to "freeze" that key to
	// prevent accidentally picking up nested objects that could merge with
	// it from earlier in the input array.
	// Example:
	//   Input: [{"a": {"b": 2}}, {"a": 4}, {"a": {"c": 3}}]
	//   Want Output: {"a": {"c": 3}}
	result := ast.NewObject()
	frozenKeys := map[*ast.Term]struct{}{}
	for i := arr.Len() - 1; i >= 0; i-- {
		o, ok := arr.Elem(i).Value.(ast.Object)
		if !ok {
			return builtins.NewOperandElementErr(1, arr, arr.Elem(i).Value, "object")
		}
		mergewithOverwriteInPlace(result, o, frozenKeys)
		if err != nil {
			return err
		}
	}

	return iter(ast.NewTerm(result))
}

func builtinObjectRemove(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	// Expect an object and an array/set/object of keys
	obj, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// Build a set of keys to remove
	keysToRemove, err := getObjectKeysParam(operands[1].Value)
	if err != nil {
		return err
	}
	r := ast.NewObject()
	obj.Foreach(func(key *ast.Term, value *ast.Term) {
		if !keysToRemove.Contains(key) {
			r.Insert(key, value)
		}
	})

	return iter(ast.NewTerm(r))
}

func builtinObjectFilter(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	// Expect an object and an array/set/object of keys
	obj, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// Build a new object from the supplied filter keys
	keys, err := getObjectKeysParam(operands[1].Value)
	if err != nil {
		return err
	}

	filterObj := ast.NewObject()
	keys.Foreach(func(key *ast.Term) {
		filterObj.Insert(key, ast.NullTerm())
	})

	// Actually do the filtering
	r, err := obj.Filter(filterObj)
	if err != nil {
		return err
	}

	return iter(ast.NewTerm(r))
}

func builtinObjectGet(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	object, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// if the get key is not an array, attempt to get the top level key for the operand value in the object
	path, err := builtins.ArrayOperand(operands[1].Value, 2)
	if err != nil {
		if ret := object.Get(operands[1]); ret != nil {
			return iter(ret)
		}

		return iter(operands[2])
	}

	// if the path is empty, then we skip selecting nested keys and return the whole object
	if path.Len() == 0 {
		return iter(operands[0])
	}

	// build an ast.Ref from the array and see if it matches within the object
	pathRef := ref.ArrayPath(path)
	value, err := object.Find(pathRef)
	if err != nil {
		return iter(operands[2])
	}

	return iter(ast.NewTerm(value))
}

// getObjectKeysParam returns a set of key values
// from a supplied ast array, object, set value
func getObjectKeysParam(arrayOrSet ast.Value) (ast.Set, error) {
	keys := ast.NewSet()

	switch v := arrayOrSet.(type) {
	case *ast.Array:
		_ = v.Iter(func(f *ast.Term) error {
			keys.Add(f)
			return nil
		})
	case ast.Set:
		_ = v.Iter(func(f *ast.Term) error {
			keys.Add(f)
			return nil
		})
	case ast.Object:
		_ = v.Iter(func(k *ast.Term, _ *ast.Term) error {
			keys.Add(k)
			return nil
		})
	default:
		return nil, builtins.NewOperandTypeErr(2, arrayOrSet, "object", "set", "array")
	}

	return keys, nil
}

func mergeWithOverwrite(objA, objB ast.Object) ast.Object {
	merged, _ := objA.MergeWith(objB, func(v1, v2 *ast.Term) (*ast.Term, bool) {
		originalValueObj, ok2 := v1.Value.(ast.Object)
		updateValueObj, ok1 := v2.Value.(ast.Object)
		if !ok1 || !ok2 {
			// If we can't merge, stick with the right-hand value
			return v2, false
		}

		// Recursively update the existing value
		merged := mergeWithOverwrite(originalValueObj, updateValueObj)
		return ast.NewTerm(merged), false
	})
	return merged
}

// Modifies obj with any new keys from other, and recursively
// merges any keys where the values are both objects.
func mergewithOverwriteInPlace(obj, other ast.Object, frozenKeys map[*ast.Term]struct{}) {
	other.Foreach(func(k, v *ast.Term) {
		v2 := obj.Get(k)
		// The key didn't exist in other, keep the original value.
		if v2 == nil {
			obj.Insert(k, v)
			return
		}
		// The key exists in both. Merge or reject change.
		updateValueObj, ok2 := v.Value.(ast.Object)
		originalValueObj, ok1 := v2.Value.(ast.Object)
		// Both are objects? Merge recursively.
		if ok1 && ok2 {
			// Check to make sure that this key isn't frozen before merging.
			if _, ok := frozenKeys[v2]; !ok {
				mergewithOverwriteInPlace(originalValueObj, updateValueObj, frozenKeys)
			}
		} else {
			// Else, original value wins. Freeze the key.
			frozenKeys[v2] = struct{}{}
		}
	})
}

func init() {
	RegisterBuiltinFunc(ast.ObjectUnion.Name, builtinObjectUnion)
	RegisterBuiltinFunc(ast.ObjectUnionN.Name, builtinObjectUnionN)
	RegisterBuiltinFunc(ast.ObjectRemove.Name, builtinObjectRemove)
	RegisterBuiltinFunc(ast.ObjectFilter.Name, builtinObjectFilter)
	RegisterBuiltinFunc(ast.ObjectGet.Name, builtinObjectGet)
}
