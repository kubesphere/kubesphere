// Copyright (c) 2016-2022 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package set

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// NewBoxed creates a new "boxed" Set, where the items stored in the set are boxed inside an interface.  The values
// placed into the set must be comparable (i.e. suitable for use as a map key).  This is checked at runtime and the
// code will panic on trying to add a non-comparable entry.
//
// This implementation exists because Go's generics currently have a gap.  The type set of the "comparable"
// constraint currently doesn't include interface types, which under Go's normal rules _are_ comparable (but may
// panic at runtime if the interface happens to contain a non-comparable object).  If possible use a typed map
// via New() or From(); use this if you really need a Set[any] or Set[SomeInterface].
func NewBoxed[T any]() Boxed[T] {
	return make(Boxed[T])
}

func FromBoxed[T any](members ...T) Boxed[T] {
	s := NewBoxed[T]()
	s.AddAll(members)
	return s
}

func FromArrayBoxed[T any](membersArray []T) Boxed[T] {
	s := NewBoxed[T]()
	s.AddAll(membersArray)
	return s
}

func Empty[T any]() Set[T] {
	return (Boxed[T])(nil)
}

type Boxed[T any] map[any]v

func (set Boxed[T]) String() string {
	var buf bytes.Buffer
	_, _ = buf.WriteString("set.Set{")
	first := true
	set.Iter(func(item T) error {
		if !first {
			buf.WriteString(",")
		} else {
			first = false
		}
		_, _ = fmt.Fprint(&buf, item)
		return nil
	})
	_, _ = buf.WriteString("}")
	return buf.String()
}

func (set Boxed[T]) Len() int {
	return len(set)
}

func (set Boxed[T]) Add(item T) {
	set[item] = emptyValue
}

func (set Boxed[T]) AddAll(itemArray []T) {
	for _, v := range itemArray {
		set.Add(v)
	}
}

// AddSet adds the contents of set "other" into the set.
func (set Boxed[T]) AddSet(other Set[T]) {
	other.Iter(func(item T) error {
		set.Add(item)
		return nil
	})
}

func (set Boxed[T]) Discard(item T) {
	delete(set, item)
}

func (set Boxed[T]) Clear() {
	for item := range set {
		delete(set, item)
	}
}

func (set Boxed[T]) Contains(item T) bool {
	_, present := set[item]
	return present
}

func (set Boxed[T]) Iter(visitor func(item T) error) {
loop:
	for item := range set {
		item := item.(T)
		err := visitor(item)
		switch err {
		case StopIteration:
			break loop
		case RemoveItem:
			delete(set, item)
		case nil:
			break
		default:
			log.WithError(err).Panic("Unexpected iteration error")
		}
	}
}

func (set Boxed[T]) Copy() Set[T] {
	cpy := NewBoxed[T]()
	for item := range set {
		item := item.(T)
		cpy.Add(item)
	}
	return cpy
}

func (set Boxed[T]) Slice() (s []T) {
	for item := range set {
		item := item.(T)
		s = append(s, item)
	}
	return
}

func (set Boxed[T]) Equals(other Set[T]) bool {
	if set.Len() != other.Len() {
		return false
	}
	for item := range set {
		item := item.(T)
		if !other.Contains(item) {
			return false
		}
	}
	return true
}

func (set Boxed[T]) ContainsAll(other Set[T]) bool {
	result := true
	other.Iter(func(item T) error {
		if !set.Contains(item) {
			result = false
			return StopIteration
		}
		return nil
	})
	return result
}
