// Copyright 2023 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package EditTree implements a specialized tree data structure that
// allows for cheap edits and modifications of nested Term structures.
//
// # Overview
//
// The EditTree data structure exists to solve an ugly problem in Rego:
// modification/deletion of a Term can be very expensive, because we have
// to rebuild the whole Term, sans the modified/deleted parts.
//
// To work around that problem, the EditTree allows simple add, modify, or
// delete operations on Term structures, and then at the end of a series of
// edits, the caller can pay the cost of generating a new Term from the
// tree of edits relatively efficiently. (Essentially a recursive, DFS
// traversal of the tree.)
//
// The data structure preserves basic type/safety properties, the same as
// working on the real underlying Term values. To do this, recursive
// lookups are used. On average, these are fairly straightforward and
// cheap to do.
//
// Basic Operations:
//   - Insert/update EditTree node
//   - Delete EditTree node
//   - Unfold EditTree nodes along a path
//   - Render EditTree node
//
// These operations provide all of the basic utilities required to
// recursively construct the tree. Ref-based convenience functions are also
// provided, to make rendering subtrees at a particular JSON path easier.
//
// Path-Based Convenience Functions:
//   - InsertAtPath
//   - DeleteAtPath
//   - RenderAtPath
//
// Additionally, a few "optional" (but nice to have) functions have been
// added, to allow replacing slower/less-efficient equivalents elsewhere.
//
// Optional Functions:
//   - Exists (a more efficient boolean alternative to Unfold)
//   - Filter (an alternative to Object.Filter that efficiently renders
//     paths out of an EditTree)
//
// # Storing scalar children "inline"
//
// The original design for the EditTree allocated a new EditTree node for
// each Term stored in the tree, but this was found to be inefficient when
// dealing with large arrays and objects. The current design of the
// EditTree separates children based on their types, with scalars stored in
// a hash -> Term map, and composites stored in a hash -> EditTree map.
//
// This results in dramatically fewer heap allocations and faster access
// times for "shallow" Term structures, without penalizing nested Term
// structures noticeably.
//
// # Object operations
//
// Objects are the most straightforward composite type, as their key-value
// structure maps naturally onto trees. Their inserts/deletes are recorded
// directly in the appropriate child maps, with almost no additional
// complexity required.
//
// Object EditTree nodes use the child key and value maps, and will not
// initialize the bit-vectors, since those are only used for Arrays.
//
// # Set operations
//
// Set data types have a major problem: they're *content-addressed*. This
// means that we often have to render/materialize the sub-terms before
// carrying out inserts or deletes, in order to know if the path to the
// destination Term exists. This forces a tree collapse at the Set's
// EditTree node, and is brutally inefficient.
//
// Example:
//
//	Source set: {[0], "a"}
//	{"op": "add", "path": [[0], 1], "value": 1}      -> result value: {[0, 1], "a"}
//	{"op": "add", "path": [[0, 1], 3], "value": 3}   -> result value: {[0, 1, 3], "a"}
//
// We mitigate this somewhat by only collapsing a Set when a composite
// value is being used for indexing. Scalars imply a shallow access, which
// we can look up directly in the appropriate child map.
//
// Set EditTree nodes use the child key and value maps, and will not
// initialize the bit-vectors, since those are only used for Arrays.
//
// # Array operations
//
// Arrays can have elements inserted/replaced/deleted, and this requires
// some bookkeeping work to keep everything straight. We do this
// bookkeeping work using two bit vectors to track all the
// insertions/deletions.
//
// One bit-vector tracks which indexes are preserved/eliminated from the
// original Array, and the second bit-vector tracks which indexes have
// insertions. We can record inserts and deletes *directly* on the second
// bit-vector, "bleeding through" deletions to the preserved/eliminated bit
// vector when there's not an insert to wipe out first.
//
// For bleed-through deletes, a linear scan is required to find the index
// of which original element will be knocked out. We then mark that bit in
// the preserved/eliminated bit-vector. This is a fair bit of bookkeeping,
// but greatly reduces the cost and complexity of tracking Array state.
// There can only be insertions, or original values present. Any other
// "deletion" is an error.
//
// Insert and Delete operations also imply a linear "index rewriting" pass
// for an Array's child maps, where indexes that occur above the affected
// index of the insertion/deletion must be incremented or decremented
// appropriately. This ensures that when rendered later, the
// original/inserted values will be spliced in at the correct offsets in
// the final Array value.
//
// Due to optimizations discussed later, Array EditTree nodes do not use
// the child key map (leaving it uninitialized), but will initialize and
// use the child value maps normally. Array EditTree nodes are the only
// types of EditTree nodes that should ever be expected to have initialized
// bit-vectors present.
//
// # Scalar operations
//
// Scalars are fairly simple: just a term stored in an EditTree node, or in
// the scalar child map of a composite type's EditTree node. They cannot
// have children, and normally do not exist as independent EditTree nodes,
// except to satisfy certain EditTree APIs.
//
// Scalar EditTree nodes can only be expected to have a valid Term value;
// all other fields will be left uninitialized.
//
// # Optimization: Direct Array Indexing with ints
//
// Arrays are unique in Rego, because the only valid Terms that can index
// into them are integer, numeric values. When processing the key Terms for
// Objects and Sets, we have to identify children by their hash values
// (which hash to integers). Because the only valid key Terms for Arrays
// work as ints as well, we can skip the hashing step entirely, and just
// use the int indexes *directly*.
//
// This provides a substantial CPU savings in benchmarks, because the
// "index rewriting" passes become much cheaper from not having to rehash
// every child's index.
package edittree

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/open-policy-agent/opa/internal/edittree/bitvector"
	"github.com/open-policy-agent/opa/v1/ast"
)

// Deletions are encoded with a nil value pointer.
type EditTree struct {
	value                *ast.Term
	childKeys            map[int]*ast.Term
	childScalarValues    map[int]*ast.Term
	childCompositeValues map[int]*EditTree
	eliminated           *bitvector.BitVector // Which original indexes have been knocked out?
	insertions           *bitvector.BitVector // Which indexes have a new value inserted at them? (also used for "live" bookkeeping)
}

// Creates a new EditTree node from term.
func NewEditTree(term *ast.Term) *EditTree {
	if term == nil {
		return nil
	}

	var tree EditTree
	switch x := term.Value.(type) {
	case ast.Object, ast.Set:
		tree = EditTree{
			value:                term,
			childKeys:            map[int]*ast.Term{},
			childScalarValues:    map[int]*ast.Term{},
			childCompositeValues: map[int]*EditTree{},
		}
	case *ast.Array:
		tree = EditTree{
			value:                term,
			childScalarValues:    map[int]*ast.Term{},
			childCompositeValues: map[int]*EditTree{},
		}
		bytesLength := ((x.Len() - 1) / 8) + 1 // How many bytes to use for the bit-vectors.
		tree.eliminated = bitvector.NewBitVector(make([]byte, bytesLength), x.Len())
		tree.insertions = bitvector.NewBitVector(make([]byte, bytesLength), x.Len())
	default:
		tree = EditTree{
			value: term,
		}
	}

	return &tree
}

// Returns correct (collision-resolved) hash for this term + whether or not
// it was found in the table already.
func (e *EditTree) getKeyHash(key *ast.Term) (int, bool) {
	hash := key.Hash()
	// This `equal` utility is duplicated and manually inlined a number of
	// time in this file.  Inlining it avoids heap allocations, so it makes
	// a big performance difference: some operations like lookup become twice
	// as slow without it.
	var equal func(v ast.Value) bool

	switch x := key.Value.(type) {
	case ast.Null, ast.Boolean, ast.String, ast.Var:
		equal = func(y ast.Value) bool { return x == y }
	case ast.Number:
		if xi, ok := x.Int64(); ok {
			equal = func(y ast.Value) bool {
				if y, ok := y.(ast.Number); ok {
					if yi, ok := y.Int64(); ok {
						return xi == yi
					}
				}

				return false
			}
			break
		}

		// We use big.Rat for comparing big numbers.
		// It replaces big.Float due to following reason:
		// big.Float comes with a default precision of 64, and setting a
		// larger precision results in more memory being allocated
		// (regardless of the actual number we are parsing with SetString).
		//
		// Note: If we're so close to zero that big.Float says we are zero, do
		// *not* big.Rat).SetString on the original string it'll potentially
		// take very long.
		var a *big.Rat
		fa, ok := new(big.Float).SetString(string(x))
		if !ok {
			panic("illegal value")
		}
		if fa.IsInt() {
			if i, _ := fa.Int64(); i == 0 {
				a = new(big.Rat).SetInt64(0)
			}
		}
		if a == nil {
			a, ok = new(big.Rat).SetString(string(x))
			if !ok {
				panic("illegal value")
			}
		}

		equal = func(b ast.Value) bool {
			if bNum, ok := b.(ast.Number); ok {
				var b *big.Rat
				fb, ok := new(big.Float).SetString(string(bNum))
				if !ok {
					panic("illegal value")
				}
				if fb.IsInt() {
					if i, _ := fb.Int64(); i == 0 {
						b = new(big.Rat).SetInt64(0)
					}
				}
				if b == nil {
					b, ok = new(big.Rat).SetString(string(bNum))
					if !ok {
						panic("illegal value")
					}
				}

				return a.Cmp(b) == 0
			}
			return false
		}

	default:
		equal = func(y ast.Value) bool { return ast.Compare(x, y) == 0 }
	}

	// Look through childKeys, looking up the original hash
	// value first, and then use linear-probing to iter
	// through the keys until we either find the Term we're
	// after, or run out of candidates.
	for curr, ok := e.childKeys[hash]; ok; {
		if equal(curr.Value) {
			return hash, true
		}

		hash++
		curr, ok = e.childKeys[hash]
	}

	// Didn't find any matches in childKeys. Hash will be
	// the first open slot after the linear probing loop.
	return hash, false
}

//gcassert:inline
func isComposite(t *ast.Term) bool {
	switch t.Value.(type) {
	case ast.Object, ast.Set, *ast.Array:
		return true
	default:
		return false
	}
}

//gcassert:inline
func (e *EditTree) setChildKey(hash int, key *ast.Term) {
	e.childKeys[hash] = key
}

//gcassert:inline
func (e *EditTree) setChildScalarValue(hash int, value *ast.Term) {
	e.childScalarValues[hash] = value
}

//gcassert:inline
func (e *EditTree) setChildCompositeValue(hash int, child *EditTree) {
	e.childCompositeValues[hash] = child
}

// We don't have a deleteChildKeys method, because once a key is inserted,
// it can only be replaced with either another value, or a delete node from
// then on.
//
//gcassert:inline
func (e *EditTree) deleteChildValue(hash int) {
	delete(e.childScalarValues, hash)
	delete(e.childCompositeValues, hash)
}

// Insert creates a new child of e, and returns the new child EditTree node.
func (e *EditTree) Insert(key, value *ast.Term) (*EditTree, error) {
	if e.value == nil {
		return nil, errors.New("deleted node encountered during insert operation")
	}
	if key == nil {
		return nil, errors.New("nil key provided for insert operation")
	}
	if value == nil {
		return nil, errors.New("nil value provided for insert operation")
	}

	switch x := e.value.Value.(type) {
	case ast.Object:
		return e.unsafeInsertObject(key, value), nil
	case ast.Set:
		if !key.Equal(value) {
			return nil, fmt.Errorf("set key %v does not equal value to be inserted %v", key, value)
		}
		// We only collapse this Set-typed node if a composite type is involved.
		if isComposite(key) {
			// TODO: Investigate re-rendering *only* the immediate composite children.
			collapsed := e.Render()
			e.value = collapsed
			e.childKeys = map[int]*ast.Term{}
			e.childScalarValues = map[int]*ast.Term{}
			e.childCompositeValues = map[int]*EditTree{}
		}
		return e.unsafeInsertSet(key, value), nil
	case *ast.Array:
		idx, err := toIndex(e.insertions.Length(), key)
		if err != nil {
			return nil, err
		}
		if idx < 0 || idx > e.insertions.Length() {
			return nil, errors.New("index for array insertion out of bounds")
		}
		return e.unsafeInsertArray(idx, value), nil
	default:
		// Catch all primitive types.
		return nil, fmt.Errorf("expected composite type, found value: %v (type: %T)", x, x)
	}
}

func (e *EditTree) unsafeInsertObject(key, value *ast.Term) *EditTree {
	child := NewEditTree(value)
	keyHash, found := e.getKeyHash(key)
	if found {
		e.deleteChildValue(keyHash)
	}
	e.setChildKey(keyHash, key)
	if isComposite(value) {
		e.setChildCompositeValue(keyHash, child)
	} else {
		e.setChildScalarValue(keyHash, value)
	}
	return child
}

func (e *EditTree) unsafeInsertSet(key, value *ast.Term) *EditTree {
	child := NewEditTree(value)
	keyHash, found := e.getKeyHash(key)
	if found {
		e.deleteChildValue(keyHash)
	}
	e.setChildKey(keyHash, key)
	if isComposite(value) {
		e.setChildCompositeValue(keyHash, child)
	} else {
		e.setChildScalarValue(keyHash, value)
	}
	return child
}

func (e *EditTree) unsafeInsertArray(idx int, value *ast.Term) *EditTree {
	child := NewEditTree(value)
	// Collect insertion indexes above the insertion site for rewriting.
	rewritesScalars := []int{}
	rewritesComposites := []int{}
	for i := idx; i < e.insertions.Length(); i++ {
		if e.insertions.Element(i) == 1 {
			if _, ok := e.childScalarValues[i]; ok {
				rewritesScalars = append(rewritesScalars, i)
				continue
			}
			if _, ok := e.childCompositeValues[i]; ok {
				rewritesComposites = append(rewritesComposites, i)
				continue
			}
			panic(fmt.Errorf("invalid index %d during Insert operation", i))
		}
	}
	// Do rewrites in reverse order to make room for the newly-inserted element.
	for i := len(rewritesScalars) - 1; i >= 0; i-- {
		originalIdx := rewritesScalars[i]
		rewriteIdx := rewritesScalars[i] + 1
		v := e.childScalarValues[originalIdx]
		e.deleteChildValue(originalIdx)
		e.setChildScalarValue(rewriteIdx, v)
	}
	for i := len(rewritesComposites) - 1; i >= 0; i-- {
		originalIdx := rewritesComposites[i]
		rewriteIdx := rewritesComposites[i] + 1
		v := e.childCompositeValues[originalIdx]
		e.deleteChildValue(originalIdx)
		e.setChildCompositeValue(rewriteIdx, v)
	}
	// Insert new element in children array, bump bit-vector over by 1.
	if idx == e.insertions.Length() {
		e.insertions.Append(1)
	} else {
		e.insertions.Insert(1, idx)
	}
	if isComposite(value) {
		e.setChildCompositeValue(idx, child)
	} else {
		e.setChildScalarValue(idx, value)
	}
	return child
}

// Delete removes a child of e, or else creates a delete node for a term
// already present in e. It then returns the deleted child EditTree node.
func (e *EditTree) Delete(key *ast.Term) (*EditTree, error) {
	if e.value == nil {
		return nil, errors.New("deleted node encountered during delete operation")
	}
	if key == nil {
		return nil, errors.New("nil key provided for delete operation")
	}

	switch e.value.Value.(type) {
	case ast.Object:
		keyHash, found := e.getKeyHash(key)
		// If child found, replace with delete node. If delete node already existed, error.
		if found {
			if child, ok := e.childScalarValues[keyHash]; ok {
				if child == nil {
					return nil, fmt.Errorf("cannot delete the already deleted scalar node for key %v", key)
				}
				e.setChildKey(keyHash, key)
				e.setChildScalarValue(keyHash, nil)
				return NewEditTree(child), nil
			}
			if child, ok := e.childCompositeValues[keyHash]; ok {
				if child == nil {
					return nil, fmt.Errorf("cannot delete the already deleted composite node for key %v", key)
				}
				e.setChildKey(keyHash, key)
				e.setChildCompositeValue(keyHash, nil)
				return child, nil
			}
			// Note(philipc): We panic here, because the only way to reach
			// this panic is to have broken the bookkeeping around the key
			// and child maps in a way that is not recoverable.
			// For example, if we have an Object EditTree node, and mess up
			// the bookkeeping elsewhere by deleting just the value from
			// the child maps, *without* also deleting the key from the key
			// map, we would reach this place, where the data structure
			// *expects* a value to exist, but nothing is present.
			panic(fmt.Errorf("hash value %d not found in scalar or composite child maps", keyHash))
		}
		// No child, lookup the key in e.value, and put in a delete if present.
		// Error if key does not exist in e.value.
		return e.fallbackDelete(key)
	case ast.Set:
		// We only collapse this Set-typed node if a composite type is involved.
		if isComposite(key) {
			// TODO: Investigate re-rendering *only* the immediate composite children.
			collapsed := e.Render()
			e.value = collapsed
			e.childKeys = map[int]*ast.Term{}
			e.childScalarValues = map[int]*ast.Term{}
			e.childCompositeValues = map[int]*EditTree{}
		} else {
			keyHash, found := e.getKeyHash(key)
			// If child found, replace with delete node. If delete node already existed, error.
			if found {
				if child, ok := e.childScalarValues[keyHash]; ok {
					if child == nil {
						return nil, fmt.Errorf("cannot delete the already deleted scalar node for key %v", key)
					}
					if key.Equal(child) {
						e.setChildKey(keyHash, key)
						e.setChildScalarValue(keyHash, nil)
						return NewEditTree(child), nil
					}
				}
			}
		}
		// No child, lookup the key in e.value, and put in a delete if present.
		// Error if key does not exist in e.value.
		return e.fallbackDelete(key)
	case *ast.Array:
		idx, err := toIndex(e.insertions.Length(), key)
		if err != nil {
			return nil, err
		}
		if idx < 0 || idx > e.insertions.Length()-1 {
			return nil, errors.New("index for array delete out of bounds")
		}

		// Collect insertion indexes above the delete site for rewriting.
		rewritesScalars := []int{}
		rewritesComposites := []int{}
		for i := idx + 1; i < e.insertions.Length(); i++ {
			if e.insertions.Element(i) == 1 {
				if _, ok := e.childScalarValues[i]; ok {
					rewritesScalars = append(rewritesScalars, i)
					continue
				}
				if _, ok := e.childCompositeValues[i]; ok {
					rewritesComposites = append(rewritesComposites, i)
					continue
				}
				panic(fmt.Errorf("invalid index %d during Insert operation", i))
			}
		}
		// Do rewrites to clear out the newly-removed element.
		e.deleteChildValue(idx)
		for i := range rewritesScalars {
			originalIdx := rewritesScalars[i]
			rewriteIdx := rewritesScalars[i] - 1
			v := e.childScalarValues[originalIdx]
			e.deleteChildValue(originalIdx)
			e.setChildScalarValue(rewriteIdx, v)
		}
		for i := range rewritesComposites {
			originalIdx := rewritesComposites[i]
			rewriteIdx := rewritesComposites[i] - 1
			v := e.childCompositeValues[originalIdx]
			e.deleteChildValue(originalIdx)
			e.setChildCompositeValue(rewriteIdx, v)
		}

		// "bleed through" to the underlying array if needed.
		// To do this, we sum up the zeroes below the current index, and use that value
		// to index through the `eliminated` array until we find a surviving index.
		if e.insertions.Element(idx) == 0 {
			zeroesSeen := 1 + sumZeroesBelowIndex(idx, e.insertions)
			// Mark appropriate position in `eliminated` array, or error.
			elimIdx, found := findIndexOfNthZero(zeroesSeen, e.eliminated)
			if !found {
				panic(fmt.Errorf("could not successfully eliminate index %d from array", idx))
			}
			e.eliminated.Set(1, elimIdx)
		}
		// Delete element from insertions array, bump bit-vec over by 1.
		e.insertions.Delete(idx)
		return e, nil
	default:
		// Catch all primitive types.
		return nil, fmt.Errorf("expected composite type, found value: %v (type: %T)", e.value.Value, e.value.Value)
	}
}

//gcassert:inline
func sumZeroesBelowIndex(index int, bv *bitvector.BitVector) int {
	zeroesSeen := 0
	for i := range index {
		if bv.Element(i) == 0 {
			zeroesSeen++
		}
	}
	return zeroesSeen
}

func findIndexOfNthZero(n int, bv *bitvector.BitVector) (int, bool) {
	zeroesSeen := 0
	for i := range bv.Length() {
		if bv.Element(i) == 0 {
			zeroesSeen++
		}
		if zeroesSeen == n {
			return i, true
		}
	}
	return 0, false
}

// Helper function for sets/objects when the key isn't present in either
// child map.
func (e *EditTree) fallbackDelete(key *ast.Term) (*EditTree, error) {
	value, err := e.value.Value.Find(ast.Ref{key})
	if err != nil {
		return nil, fmt.Errorf("cannot delete child key %v that does not exist", key)
	}
	keyHash, _ := e.getKeyHash(key)
	e.setChildKey(keyHash, key)
	if isComposite(ast.NewTerm(value)) {
		e.setChildCompositeValue(keyHash, nil)
	} else {
		e.setChildScalarValue(keyHash, nil)
	}
	return NewEditTree(ast.NewTerm(value)), nil
}

// Unfurls a chain of EditTree nodes down a given path, or else returns an error.
func (e *EditTree) Unfold(path ast.Ref) (*EditTree, error) {
	// 0 path segments base case. (Root hits this.)
	if len(path) == 0 {
		return e, nil
	}
	// 1+ path segment case.
	if e.value == nil {
		return nil, errors.New("nil value encountered where composite value was expected")
	}

	// Switch behavior based on types.
	key := path[0]
	switch x := e.value.Value.(type) {
	case ast.Object:
		keyHash, found := e.getKeyHash(key)
		if found {
			if term, ok := e.childScalarValues[keyHash]; ok {
				if term == nil {
					return nil, fmt.Errorf("cannot unfold the already deleted scalar node for key %v", key)
				}
				child := NewEditTree(term)
				return child.Unfold(path[1:])
			}
			if child, ok := e.childCompositeValues[keyHash]; ok {
				if child == nil {
					return nil, fmt.Errorf("cannot unfold the already deleted composite node for key %v", key)
				}
				return child.Unfold(path[1:])
			}
			// Note(philipc): We panic here, because the only way to reach
			// this panic is to have broken the bookkeeping around the key
			// and child maps in a way that is not recoverable.
			// For example, if we have an Object EditTree node, and mess up
			// the bookkeeping elsewhere by deleting just the value from
			// the child maps, *without* also deleting the key from the key
			// map, we would reach this place, where the data structure
			// *expects* a value to exist, but nothing is present.
			panic(fmt.Errorf("hash value %d not found in scalar or composite child maps", keyHash))
		}
		// Fall back to looking up the key in e.value.
		// Extend the tree if key is present. Error otherwise.
		if v, err := x.Find(ast.Ref{path[0]}); err == nil {
			child, err := e.Insert(path[0], ast.NewTerm(v))
			if err != nil {
				return nil, err
			}
			return child.Unfold(path[1:])
		}
		return nil, fmt.Errorf("path %v does not exist in object term %v", ast.Ref{path[0]}, e.value.Value)
	case ast.Set:
		// Sets' keys *are* their values, so in order to allow accurate
		// traversal, we have to collapse the tree beneath this node,
		// so that we can accurately unfold it again for an update,
		// once we know that the key we care about is present.
		if isComposite(key) {
			collapsed := e.Render()
			e.value = collapsed
			e.childKeys = map[int]*ast.Term{}
			e.childScalarValues = map[int]*ast.Term{}
			e.childCompositeValues = map[int]*EditTree{}
		} else {
			keyHash, found := e.getKeyHash(key)
			if found {
				if term, ok := e.childScalarValues[keyHash]; ok {
					child := NewEditTree(term)
					return child.Unfold(path[1:])
				}
			}
		}
		// Fall back to looking up the key in e.value.
		// Extend the tree if key is present. Error otherwise.
		if v, err := e.value.Value.Find(ast.Ref{path[0]}); err == nil {
			child, err := e.Insert(path[0], ast.NewTerm(v))
			if err != nil {
				return nil, err
			}
			return child.Unfold(path[1:])
		}
		return nil, fmt.Errorf("path %v does not exist in set term %v", ast.Ref{path[0]}, e.value.Value)
	case *ast.Array:
		idx, err := toIndex(e.insertions.Length(), path[0])
		if err != nil {
			return nil, err
		}
		if term, ok := e.childScalarValues[idx]; ok {
			child := NewEditTree(term)
			return child.Unfold(path[1:])
		}
		if child, ok := e.childCompositeValues[idx]; ok {
			return child.Unfold(path[1:])
		}

		// Fall back to looking up the key in e.value.
		// Extend the tree if key is present. Error otherwise.
		if v, err := x.Find(ast.Ref{ast.InternedIntNumberTerm(idx)}); err == nil {
			// TODO: Consider a more efficient "Replace" function that special-cases this for arrays instead?
			_, err := e.Delete(ast.InternedIntNumberTerm(idx))
			if err != nil {
				return nil, err
			}
			child, err := e.Insert(ast.IntNumberTerm(idx), ast.NewTerm(v))
			if err != nil {
				return nil, err
			}
			return child.Unfold(path[1:])
		}
		return nil, fmt.Errorf("path %v does not exist in array term %v", ast.Ref{ast.IntNumberTerm(idx)}, e.value.Value)
	default:
		// Catch all primitive types.
		return nil, fmt.Errorf("expected composite type for path %v, found value: %v (type: %T)", ast.Ref{path[0]}, x, x)
	}
}

// Render generates the effective value for the term at e by recursively
// rendering the children of e, and then copying over any leftover keys
// from the original term stored at e.
func (e *EditTree) Render() *ast.Term {
	if e.value == nil {
		return nil
	}

	switch x := e.value.Value.(type) {
	case ast.Object:
		// Early exit if no modifications.
		if len(e.childKeys) == 0 {
			return e.value
		}
		// Build a new Object with modified/deleted keys.
		// We do this by adding the modified/deleted keys first, then
		// skipping those keys later when iterating over the original base
		// term in e.value.
		skipKeysList := make([]*ast.Term, 0, len(e.childKeys))
		out := make([][2]*ast.Term, 0, len(e.childKeys)+x.Len())
		for hash, term := range e.childScalarValues {
			skipKeysList = append(skipKeysList, e.childKeys[hash])
			if term == nil {
				continue // Delete case.
			}
			// Normal value case.
			out = append(out, [2]*ast.Term{e.childKeys[hash], term})
		}
		for hash, child := range e.childCompositeValues {
			skipKeysList = append(skipKeysList, e.childKeys[hash])
			if child == nil {
				continue // Delete case.
			}
			// Normal value case.
			subtreeResult := child.Render()
			out = append(out, [2]*ast.Term{e.childKeys[hash], subtreeResult})
		}
		skipKeys := ast.NewSet(skipKeysList...)
		// Copy over all keys that weren't deleted/modified.
		x.Foreach(func(k, v *ast.Term) {
			if skipKeys.Contains(k) {
				return
			}
			out = append(out, [2]*ast.Term{k, v})
		})
		return ast.ObjectTerm(out...)
	case ast.Set:
		// Early exit if no modifications.
		if len(e.childKeys) == 0 {
			return e.value
		}
		// Build a new Set.
		// Sets only can have deletions/new insertions, because the value
		// *is* its own key.
		skipKeysList := make([]*ast.Term, 0, len(e.childKeys))
		out := make([]*ast.Term, 0, x.Len()+len(e.childKeys))
		for hash, term := range e.childScalarValues {
			skipKeysList = append(skipKeysList, e.childKeys[hash])
			if term == nil {
				continue // Delete case.
			}
			// Normal value case.
			out = append(out, e.childScalarValues[hash])
		}
		// Only happens when this set hasn't been collapsed yet.
		for hash, child := range e.childCompositeValues {
			skipKeysList = append(skipKeysList, e.childKeys[hash])
			if child == nil {
				continue // Delete case.
			}
			// Normal value case.
			subtreeResult := child.Render()
			out = append(out, subtreeResult)
		}
		skipKeys := ast.NewSet(skipKeysList...)
		// Copy over all keys that weren't deleted/modified.
		x.Foreach(func(key *ast.Term) {
			if skipKeys.Contains(key) {
				return
			}
			out = append(out, key)
		})
		return ast.SetTerm(out...)
	case *ast.Array:
		// No early exit here, because we might have just deletes on the
		// original array. We build a new Array with modified/deleted keys.
		out := make([]*ast.Term, 0, e.insertions.Length())
		eIdx := 0
		for i := range e.insertions.Length() {
			// If the index == 0, that indicates we should look up the next
			// surviving original element.
			// If the index == 1, that indicates we should look up that
			// index in the child maps.
			if e.insertions.Element(i) == 0 {
				// Scan through the e.eliminated bit-vec, pick the
				// first non-zero index. The then use that index to
				// look up the element we should append from e.value.
				foundIdx := false
				for j := eIdx; j < e.eliminated.Length(); j++ {
					if e.eliminated.Element(j) == 0 {
						foundIdx = true
						eIdx = j
						break
					}
				}
				if !foundIdx {
					panic(fmt.Errorf("too many eliminated indexes in array, expected to find uneliminated index %d", i))
				}
				// Append element from original term.
				out = append(out, x.Elem(eIdx))
				eIdx++ // Bump the counter, so that we monotonically advance through the array.
			} else {
				// Append value from rendered child index.
				// Since deletions are not possible as children for an Array,
				// we don't need to check for nils here.
				if t, ok := e.childScalarValues[i]; ok {
					out = append(out, t)
				} else if child, ok := e.childCompositeValues[i]; ok {
					t := child.Render()
					out = append(out, t)
				} else {
					panic(fmt.Errorf("invalid index %d does not exist in array", i))
				}
			}
		}
		return ast.ArrayTerm(out...)
	default:
		return e.value
	}
}

// InsertAtPath traverses down the tree from e and uses the last path
// segment as the key to insert value into the tree.
// Returns the inserted EditTree node.
func (e *EditTree) InsertAtPath(path ast.Ref, value *ast.Term) (*EditTree, error) {
	if value == nil {
		return nil, errors.New("cannot insert nil value into EditTree")
	}

	if len(path) == 0 {
		e.value = value
		e.childKeys = map[int]*ast.Term{}
		e.childScalarValues = map[int]*ast.Term{}
		e.childCompositeValues = map[int]*EditTree{}
		if v, ok := value.Value.(*ast.Array); ok {
			bytesLength := ((v.Len() - 1) / 8) + 1 // How many bytes to use for the bit-vectors.
			e.eliminated = bitvector.NewBitVector(make([]byte, bytesLength), v.Len())
			e.insertions = bitvector.NewBitVector(make([]byte, bytesLength), v.Len())
		}
		return e, nil
	}

	dest, err := e.Unfold(path[:len(path)-1])
	if err != nil {
		return nil, err
	}

	return dest.Insert(path[len(path)-1], value)
}

// DeleteAtPath traverses down the tree from e and uses the last path
// segment as the key to delete a node from the tree.
// Returns the deleted EditTree node.
func (e *EditTree) DeleteAtPath(path ast.Ref) (*EditTree, error) {
	// Root document case:
	if len(path) == 0 {
		if e.value == nil {
			return nil, errors.New("deleted node encountered during delete operation")
		}
		e.value = nil
		e.childKeys = nil
		e.childScalarValues = nil
		e.childCompositeValues = nil
		e.eliminated = nil
		e.insertions = nil
		return e, nil
	}

	dest, err := e.Unfold(path[:len(path)-1])
	if err != nil {
		return nil, err
	}

	return dest.Delete(path[len(path)-1])
}

// RenderAtPath traverses down the tree from e and renders the EditTree
// node at the end of path.
func (e *EditTree) RenderAtPath(path ast.Ref) (*ast.Term, error) {
	dest, err := e.Unfold(path)
	if err != nil {
		return nil, err
	}

	return dest.Render(), nil
}

func (e *EditTree) String() string {
	if t := e.Render(); t != nil {
		return "EditTree[" + t.String() + "]"
	}
	return ""
}

func (e *EditTree) Exists(path ast.Ref) bool {
	if e.value == nil {
		return false
	}

	switch {
	// 0 path segments base case. (Root hits this.)
	case len(path) == 0:
		return true
	// 1+ path segments case.
	case len(path) >= 1:
		// Switch behavior based on types.
		key := path[0]
		switch x := e.value.Value.(type) {
		case ast.Object:
			keyHash, found := e.getKeyHash(key)
			if found {
				if term, ok := e.childScalarValues[keyHash]; ok {
					if term == nil {
						return false
					}
					return len(path) == 1
				}
				if child, ok := e.childCompositeValues[keyHash]; ok {
					if child == nil {
						return false
					}
					return child.Exists(path[1:])
				}
				// Note(philipc): We panic here, because the only way to reach
				// this panic is to have broken the bookkeeping around the key
				// and child maps in a way that is not recoverable.
				// For example, if we have an Object EditTree node, and mess up
				// the bookkeeping elsewhere by deleting just the value from
				// the child maps, *without* also deleting the key from the key
				// map, we would reach this place, where the data structure
				// *expects* a value to exist, but nothing is present.
				panic(fmt.Errorf("hash value %d not found in scalar or composite child maps", keyHash))
			}
			// Fallback if child lookup failed.
			_, err := x.Find(path)
			return err == nil
		case ast.Set:
			// Sets' keys *are* their values, so in order to allow accurate
			// traversal, we have to collapse the tree beneath this node,
			// so that we can accurately unfold it again for an update,
			// once we know that the key we care about is present.
			if isComposite(key) {
				collapsed := e.Render()
				e.value = collapsed
				e.childKeys = map[int]*ast.Term{}
				e.childScalarValues = map[int]*ast.Term{}
				e.childCompositeValues = map[int]*EditTree{}
			} else {
				keyHash, found := e.getKeyHash(key)
				if found {
					if _, ok := e.childScalarValues[keyHash]; ok {
						return len(path) == 1
					}
				}
			}
			// Fallback if child lookup failed.
			_, err := e.value.Value.Find(path)
			return err == nil
		case *ast.Array:
			var idx int
			idx, err := toIndex(e.insertions.Length(), path[0])
			if err != nil {
				return false
			}
			if _, ok := e.childScalarValues[idx]; ok {
				return len(path) == 1
			}
			if child, ok := e.childCompositeValues[idx]; ok {
				return child.Exists(path[1:])
			}
			// Fallback if child lookup failed.
			// We have to ensure that the lookup term is a number here, or Find will fail.
			_, err = x.Find(ast.Ref{ast.InternedIntNumberTerm(idx)}.Concat(path[1:]))
			return err == nil
		default:
			// Catch all primitive types.
			return false
		}
	}
	return false
}

// --------------------------------------------------------------------
// Utility functions

// toIndex tries to convert path elements (that may be strings) into indexes
// into an array.
func toIndex(arrayLength int, term *ast.Term) (int, error) {
	i := 0
	var ok bool
	switch v := term.Value.(type) {
	case ast.Number:
		if i, ok = v.Int(); !ok {
			return 0, errors.New("invalid number type for indexing")
		}
	case ast.String:
		if v == "-" {
			return arrayLength, nil
		}
		num := ast.Number(v)
		if i, ok = num.Int(); !ok {
			return 0, errors.New("invalid string for indexing")
		}
		if v != "0" && strings.HasPrefix(string(v), "0") {
			return 0, errors.New("leading zeros are not allowed in JSON paths")
		}
	default:
		return 0, errors.New("invalid type for indexing")
	}

	return i, nil
}

// --------------------------------------------------------------------
// Term-level utility functions

// Filter pulls out only the values selected by paths.
// This is done recursively by plucking off one level from the paths each time we descend a level.
// Note: Values pulled from arrays will have the same approximate
// ordering in the final term.
func (e *EditTree) Filter(paths []ast.Ref) *ast.Term {
	if e.value == nil {
		return nil
	}

	// Separate out keys for this level.
	// In the event of paths like "a", "a/b", "a/b/c", the "a" path will win out.
	// Nil keys, such as "" or [], are not permitted. (legacy behavior)
	pathMap := make(map[ast.Value][]ast.Ref, len(paths))
	renderNowList := []*ast.Term{}
	for i := range paths {
		path := paths[i]
		switch {
		case len(path) == 0:
			continue // ignore nil paths, such as "" and [].
		case len(path) == 1:
			renderNowList = append(renderNowList, path[0])
		default: // len(path) > 1
			if _, ok := pathMap[path[0].Value]; !ok {
				pathMap[path[0].Value] = []ast.Ref{}
			}
			pathMap[path[0].Value] = append(pathMap[path[0].Value], path[1:])
		}
	}
	renderNow := ast.NewSet(renderNowList...)
	// Clear everything out of the pathMap that has a renderNow candidate.
	for k := range pathMap {
		if renderNow.Contains(ast.NewTerm(k)) {
			delete(pathMap, k)
		}
	}

	// Now that we've reached the target, we can start rendering everything beneath us in the tree.
	switch e.value.Value.(type) {
	case ast.Object:
		out := make([][2]*ast.Term, 0, renderNow.Len()+len(pathMap))
		// Render any finished paths.
		renderNow.Foreach(func(k *ast.Term) {
			if e.Exists(ast.Ref{k}) {
				subtreeResult, _ := e.RenderAtPath(ast.Ref{k})
				out = append(out, [2]*ast.Term{k, subtreeResult})
			}
		})
		// Recursively descend remaining paths.
		for k, p := range pathMap {
			if e.Exists(ast.Ref{ast.NewTerm(k)}) {
				child, _ := e.Unfold(ast.Ref{ast.NewTerm(k)})
				subtreeResult := child.Filter(p)
				out = append(out, [2]*ast.Term{ast.NewTerm(k), subtreeResult})
			}
		}
		return ast.ObjectTerm(out...)
	case ast.Set:
		out := make([]*ast.Term, 0, renderNow.Len()+len(pathMap))
		// Render any finished paths.
		renderNow.Foreach(func(k *ast.Term) {
			if e.Exists(ast.Ref{k}) {
				subtreeResult, _ := e.RenderAtPath(ast.Ref{k})
				out = append(out, subtreeResult)
			}
		})
		// Recursively descend remaining paths.
		for k, p := range pathMap {
			if e.Exists(ast.Ref{ast.NewTerm(k)}) {
				child, _ := e.Unfold(ast.Ref{ast.NewTerm(k)})
				subtreeResult := child.Filter(p)
				out = append(out, subtreeResult)
			}
		}
		return ast.SetTerm(out...)
	case *ast.Array:
		// No early exit here, because we might have just deletes on the
		// original array. We build a new Array with modified/deleted keys.
		out := make([]*ast.Term, 0, renderNow.Len()+len(pathMap))
		// Sort array indexes before descending.
		idxList := make([]*ast.Term, 0, len(pathMap))
		renderNow.Foreach(func(k *ast.Term) {
			idxList = append(idxList, k)
		})
		for k := range pathMap {
			idxList = append(idxList, ast.NewTerm(k))
		}
		sort.Sort(termSlice(idxList))
		// Render child or recursively descend as needed.
		for i := range idxList {
			k := idxList[i]
			if renderNow.Contains(k) {
				if e.Exists(ast.Ref{k}) {
					subtreeResult, _ := e.RenderAtPath(ast.Ref{k})
					out = append(out, subtreeResult)
				}
			} else if e.Exists(ast.Ref{k}) {
				child, _ := e.Unfold(ast.Ref{k})
				subtreeResult := child.Filter(pathMap[k.Value])
				out = append(out, subtreeResult)
			}
		}
		return ast.ArrayTerm(out...)
	default:
		return e.value
	}
}

type termSlice []*ast.Term

func (s termSlice) Less(i, j int) bool { return ast.Compare(s[i].Value, s[j].Value) < 0 }
func (s termSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s termSlice) Len() int           { return len(s) }
