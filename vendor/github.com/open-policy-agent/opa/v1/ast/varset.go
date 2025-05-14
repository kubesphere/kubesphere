// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"fmt"
	"slices"

	"github.com/open-policy-agent/opa/v1/util"
)

// VarSet represents a set of variables.
type VarSet map[Var]struct{}

// NewVarSet returns a new VarSet containing the specified variables.
func NewVarSet(vs ...Var) VarSet {
	s := make(VarSet, len(vs))
	for _, v := range vs {
		s.Add(v)
	}
	return s
}

// NewVarSet returns a new VarSet containing the specified variables.
func NewVarSetOfSize(size int) VarSet {
	return make(VarSet, size)
}

// Add updates the set to include the variable "v".
func (s VarSet) Add(v Var) {
	s[v] = struct{}{}
}

// Contains returns true if the set contains the variable "v".
func (s VarSet) Contains(v Var) bool {
	_, ok := s[v]
	return ok
}

// Copy returns a shallow copy of the VarSet.
func (s VarSet) Copy() VarSet {
	cpy := NewVarSetOfSize(len(s))
	for v := range s {
		cpy.Add(v)
	}
	return cpy
}

// Diff returns a VarSet containing variables in s that are not in vs.
func (s VarSet) Diff(vs VarSet) VarSet {
	i := 0
	for v := range s {
		if !vs.Contains(v) {
			i++
		}
	}
	r := NewVarSetOfSize(i)
	for v := range s {
		if !vs.Contains(v) {
			r.Add(v)
		}
	}
	return r
}

// Equal returns true if s contains exactly the same elements as vs.
func (s VarSet) Equal(vs VarSet) bool {
	if len(s) != len(vs) {
		return false
	}
	for v := range s {
		if !vs.Contains(v) {
			return false
		}
	}
	return true
}

// Intersect returns a VarSet containing variables in s that are in vs.
func (s VarSet) Intersect(vs VarSet) VarSet {
	i := 0
	for v := range s {
		if vs.Contains(v) {
			i++
		}
	}
	r := NewVarSetOfSize(i)
	for v := range s {
		if vs.Contains(v) {
			r.Add(v)
		}
	}
	return r
}

// Sorted returns a new sorted slice of vars from s.
func (s VarSet) Sorted() []Var {
	sorted := make([]Var, 0, len(s))
	for v := range s {
		sorted = append(sorted, v)
	}
	slices.SortFunc(sorted, VarCompare)
	return sorted
}

// Update merges the other VarSet into this VarSet.
func (s VarSet) Update(vs VarSet) {
	for v := range vs {
		s.Add(v)
	}
}

func (s VarSet) String() string {
	return fmt.Sprintf("%v", util.KeysSorted(s))
}
