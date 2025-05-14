// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"strings"
)

// T is a concise way to refer to T.
type T interface{}

type Hasher interface {
	Hash() int
}

type hashEntry[K any, V any] struct {
	k    K
	v    V
	next *hashEntry[K, V]
}

// TypedHashMap represents a key/value map.
type TypedHashMap[K any, V any] struct {
	keq   func(K, K) bool
	veq   func(V, V) bool
	khash func(K) int
	vhash func(V) int
	def   V
	table map[int]*hashEntry[K, V]
	size  int
}

// NewTypedHashMap returns a new empty TypedHashMap.
func NewTypedHashMap[K any, V any](keq func(K, K) bool, veq func(V, V) bool, khash func(K) int, vhash func(V) int, def V) *TypedHashMap[K, V] {
	return &TypedHashMap[K, V]{
		keq:   keq,
		veq:   veq,
		khash: khash,
		vhash: vhash,
		def:   def,
		table: make(map[int]*hashEntry[K, V]),
		size:  0,
	}
}

// HashMap represents a key/value map.
type HashMap = TypedHashMap[T, T]

// NewHashMap returns a new empty HashMap.
func NewHashMap(eq func(T, T) bool, hash func(T) int) *HashMap {
	return &HashMap{
		keq:   eq,
		veq:   eq,
		khash: hash,
		vhash: hash,
		def:   nil,
		table: make(map[int]*hashEntry[T, T]),
		size:  0,
	}
}

// Copy returns a shallow copy of this HashMap.
func (h *TypedHashMap[K, V]) Copy() *TypedHashMap[K, V] {
	cpy := NewTypedHashMap[K, V](h.keq, h.veq, h.khash, h.vhash, h.def)
	h.Iter(func(k K, v V) bool {
		cpy.Put(k, v)
		return false
	})
	return cpy
}

// Equal returns true if this HashMap equals the other HashMap.
// Two hash maps are equal if they contain the same key/value pairs.
func (h *TypedHashMap[K, V]) Equal(other *TypedHashMap[K, V]) bool {
	if h.Len() != other.Len() {
		return false
	}
	return !h.Iter(func(k K, v V) bool {
		ov, ok := other.Get(k)
		if !ok {
			return true
		}
		return !h.veq(v, ov)
	})
}

// Get returns the value for k.
func (h *TypedHashMap[K, V]) Get(k K) (V, bool) {
	hash := h.khash(k)
	for entry := h.table[hash]; entry != nil; entry = entry.next {
		if h.keq(entry.k, k) {
			return entry.v, true
		}
	}
	return h.def, false
}

// Delete removes the key k.
func (h *TypedHashMap[K, V]) Delete(k K) {
	hash := h.khash(k)
	var prev *hashEntry[K, V]
	for entry := h.table[hash]; entry != nil; entry = entry.next {
		if h.keq(entry.k, k) {
			if prev != nil {
				prev.next = entry.next
			} else {
				h.table[hash] = entry.next
			}
			h.size--
			return
		}
		prev = entry
	}
}

// Hash returns the hash code for this hash map.
func (h *TypedHashMap[K, V]) Hash() int {
	var hash int
	h.Iter(func(k K, v V) bool {
		hash += h.khash(k) + h.vhash(v)
		return false
	})
	return hash
}

// Iter invokes the iter function for each element in the HashMap.
// If the iter function returns true, iteration stops and the return value is true.
// If the iter function never returns true, iteration proceeds through all elements
// and the return value is false.
func (h *TypedHashMap[K, V]) Iter(iter func(K, V) bool) bool {
	for _, entry := range h.table {
		for ; entry != nil; entry = entry.next {
			if iter(entry.k, entry.v) {
				return true
			}
		}
	}
	return false
}

// Len returns the current size of this HashMap.
func (h *TypedHashMap[K, V]) Len() int {
	return h.size
}

// Put inserts a key/value pair into this HashMap. If the key is already present, the existing
// value is overwritten.
func (h *TypedHashMap[K, V]) Put(k K, v V) {
	hash := h.khash(k)
	head := h.table[hash]
	for entry := head; entry != nil; entry = entry.next {
		if h.keq(entry.k, k) {
			entry.v = v
			return
		}
	}
	h.table[hash] = &hashEntry[K, V]{k: k, v: v, next: head}
	h.size++
}

func (h *TypedHashMap[K, V]) String() string {
	var buf []string
	h.Iter(func(k K, v V) bool {
		buf = append(buf, fmt.Sprintf("%v: %v", k, v))
		return false
	})
	return "{" + strings.Join(buf, ", ") + "}"
}

// Update returns a new HashMap with elements from the other HashMap put into this HashMap.
// If the other HashMap contains elements with the same key as this HashMap, the value
// from the other HashMap overwrites the value from this HashMap.
func (h *TypedHashMap[K, V]) Update(other *TypedHashMap[K, V]) *TypedHashMap[K, V] {
	updated := h.Copy()
	other.Iter(func(k K, v V) bool {
		updated.Put(k, v)
		return false
	})
	return updated
}

type hasherEntry[K Hasher, V any] struct {
	k    K
	v    V
	next *hasherEntry[K, V]
}

// HasherMap represents a simpler version of TypedHashMap that uses Hasher's
// for keys, and requires only an equality function for keys. Ideally we'd have
// and Equal method for all key types too, and we could get rid of that requirement.
type HasherMap[K Hasher, V any] struct {
	keq   func(K, K) bool
	table map[int]*hasherEntry[K, V]
	size  int
}

// NewHasherMap returns a new empty HasherMap.
func NewHasherMap[K Hasher, V any](keq func(K, K) bool) *HasherMap[K, V] {
	return &HasherMap[K, V]{
		keq:   keq,
		table: make(map[int]*hasherEntry[K, V]),
		size:  0,
	}
}

// Get returns the value for k.
func (h *HasherMap[K, V]) Get(k K) (V, bool) {
	for entry := h.table[k.Hash()]; entry != nil; entry = entry.next {
		if h.keq(entry.k, k) {
			return entry.v, true
		}
	}
	var zero V
	return zero, false
}

// Put inserts a key/value pair into this HashMap. If the key is already present, the existing
// value is overwritten.
func (h *HasherMap[K, V]) Put(k K, v V) {
	hash := k.Hash()
	head := h.table[hash]
	for entry := head; entry != nil; entry = entry.next {
		if h.keq(entry.k, k) {
			entry.v = v
			return
		}
	}
	h.table[hash] = &hasherEntry[K, V]{k: k, v: v, next: head}
	h.size++
}

// Delete removes the key k.
func (h *HasherMap[K, V]) Delete(k K) {
	hash := k.Hash()
	var prev *hasherEntry[K, V]
	for entry := h.table[hash]; entry != nil; entry = entry.next {
		if h.keq(entry.k, k) {
			if prev != nil {
				prev.next = entry.next
			} else {
				h.table[hash] = entry.next
			}
			h.size--
			return
		}
		prev = entry
	}
}

// Iter invokes the iter function for each element in the HasherMap.
// If the iter function returns true, iteration stops and the return value is true.
// If the iter function never returns true, iteration proceeds through all elements
// and the return value is false.
func (h *HasherMap[K, V]) Iter(iter func(K, V) bool) bool {
	for _, entry := range h.table {
		for ; entry != nil; entry = entry.next {
			if iter(entry.k, entry.v) {
				return true
			}
		}
	}
	return false
}

// Len returns the current size of this HashMap.
func (h *HasherMap[K, V]) Len() int {
	return h.size
}
