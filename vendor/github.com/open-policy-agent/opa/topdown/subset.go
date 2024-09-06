// Copyright 2022 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

func bothObjects(t1, t2 *ast.Term) (bool, ast.Object, ast.Object) {
	if (t1 == nil) || (t2 == nil) {
		return false, nil, nil
	}

	obj1, ok := t1.Value.(ast.Object)
	if !ok {
		return false, nil, nil
	}

	obj2, ok := t2.Value.(ast.Object)
	if !ok {
		return false, nil, nil
	}

	return true, obj1, obj2
}

func bothSets(t1, t2 *ast.Term) (bool, ast.Set, ast.Set) {
	if (t1 == nil) || (t2 == nil) {
		return false, nil, nil
	}

	set1, ok := t1.Value.(ast.Set)
	if !ok {
		return false, nil, nil
	}

	set2, ok := t2.Value.(ast.Set)
	if !ok {
		return false, nil, nil
	}

	return true, set1, set2
}

func bothArrays(t1, t2 *ast.Term) (bool, *ast.Array, *ast.Array) {
	if (t1 == nil) || (t2 == nil) {
		return false, nil, nil
	}

	array1, ok := t1.Value.(*ast.Array)
	if !ok {
		return false, nil, nil
	}

	array2, ok := t2.Value.(*ast.Array)
	if !ok {
		return false, nil, nil
	}

	return true, array1, array2
}

func arraySet(t1, t2 *ast.Term) (bool, *ast.Array, ast.Set) {
	if (t1 == nil) || (t2 == nil) {
		return false, nil, nil
	}

	array, ok := t1.Value.(*ast.Array)
	if !ok {
		return false, nil, nil
	}

	set, ok := t2.Value.(ast.Set)
	if !ok {
		return false, nil, nil
	}

	return true, array, set
}

// objectSubset implements the subset operation on a pair of objects.
//
// This function will try to recursively apply the subset operation where it
// can, such as if both super and sub have an object or set as the value
// associated with a key.
func objectSubset(super ast.Object, sub ast.Object) bool {
	var superTerm *ast.Term
	isSubset := true

	sub.Until(func(key, subTerm *ast.Term) bool {
		// This really wants to be a for loop, hence the somewhat
		// weird internal structure. However, using Until() in this
		// was is a performance optimization, as it avoids performing
		// any key hashing on the sub-object.

		superTerm = super.Get(key)

		// subTerm is can't be nil because we got it from Until(), so
		// we only need to verify that super is non-nil.
		if superTerm == nil {
			isSubset = false
			return true // break, not a subset
		}

		if subTerm.Equal(superTerm) {
			return false // continue
		}

		// If both of the terms are objects then we want to apply
		// the subset operation recursively, otherwise we just compare
		// them normally. If only one term is an object, then we
		// do a normal comparison which will come up false.
		if ok, superObj, subObj := bothObjects(superTerm, subTerm); ok {
			if !objectSubset(superObj, subObj) {
				isSubset = false
				return true // break, not a subset
			}

			return false // continue
		}

		if ok, superSet, subSet := bothSets(superTerm, subTerm); ok {
			if !setSubset(superSet, subSet) {
				isSubset = false
				return true // break, not a subset
			}

			return false // continue
		}

		if ok, superArray, subArray := bothArrays(superTerm, subTerm); ok {
			if !arraySubset(superArray, subArray) {
				isSubset = false
				return true // break, not a subset
			}

			return false // continue
		}

		// We have already checked for exact equality, as well as for
		// all of the types of nested subsets we care about, so if we
		// get here it means this isn't a subset.
		isSubset = false
		return true // break, not a subset
	})

	return isSubset
}

// setSubset implements the subset operation on sets.
//
// Unlike in the object case, this is not recursive, we just compare values
// using ast.Set.Contains() because we have no well defined way to "match up"
// objects that are in different sets.
func setSubset(super ast.Set, sub ast.Set) bool {
	isSubset := true
	sub.Until(func(t *ast.Term) bool {
		if !super.Contains(t) {
			isSubset = false
			return true
		}
		return false
	})

	return isSubset
}

// arraySubset implements the subset operation on arrays.
//
// This is defined to mean that the entire "sub" array must appear in
// the "super" array. For the same rationale as setSubset(), we do not attempt
// to recurse into values.
func arraySubset(super, sub *ast.Array) bool {
	// Notice that this is essentially string search. The naive approach
	// used here is O(n^2). This should probably be rewritten later to use
	// Boyer-Moore or something.

	if sub.Len() > super.Len() {
		return false
	}

	if sub.Equal(super) {
		return true
	}

	superCursor := 0
	subCursor := 0
	for {
		if subCursor == sub.Len() {
			return true
		}

		if superCursor+subCursor == super.Len() {
			return false
		}

		subElem := sub.Elem(subCursor)
		superElem := super.Elem(superCursor + subCursor)
		if superElem == nil {
			return false
		}

		if superElem.Value.Compare(subElem.Value) == 0 {
			subCursor++
		} else {
			superCursor++
			subCursor = 0
		}
	}
}

// arraySetSubset implements the subset operation on array and set.
//
// This is defined to mean that the entire "sub" set must appear in
// the "super" array with no consideration of ordering.
// For the same rationale as setSubset(), we do not attempt
// to recurse into values.
func arraySetSubset(super *ast.Array, sub ast.Set) bool {
	unmatched := sub.Len()
	return super.Until(func(t *ast.Term) bool {
		if sub.Contains(t) {
			unmatched--
		}
		if unmatched == 0 {
			return true
		}
		return false
	})
}

func builtinObjectSubset(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	superTerm := operands[0]
	subTerm := operands[1]

	if ok, superObj, subObj := bothObjects(superTerm, subTerm); ok {
		// Both operands are objects.
		return iter(ast.BooleanTerm(objectSubset(superObj, subObj)))
	}

	if ok, superSet, subSet := bothSets(superTerm, subTerm); ok {
		// Both operands are sets.
		return iter(ast.BooleanTerm(setSubset(superSet, subSet)))
	}

	if ok, superArray, subArray := bothArrays(superTerm, subTerm); ok {
		// Both operands are sets.
		return iter(ast.BooleanTerm(arraySubset(superArray, subArray)))
	}

	if ok, superArray, subSet := arraySet(superTerm, subTerm); ok {
		// Super operand is array and sub operand is set
		return iter(ast.BooleanTerm(arraySetSubset(superArray, subSet)))
	}

	return builtins.ErrOperand("both arguments object.subset must be of the same type or array and set")
}

func init() {
	RegisterBuiltinFunc(ast.ObjectSubset.Name, builtinObjectSubset)
}
