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

func New[T comparable]() Typed[T] {
	return make(Typed[T])
}

func From[T comparable](members ...T) Typed[T] {
	s := New[T]()
	s.AddAll(members)
	return s
}

func FromArray[T comparable](membersArray []T) Typed[T] {
	s := New[T]()
	s.AddAll(membersArray)
	return s
}

type Typed[T comparable] map[T]v

func (set Typed[T]) String() string {
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

func (set Typed[T]) Len() int {
	return len(set)
}

func (set Typed[T]) Add(item T) {
	set[item] = emptyValue
}

func (set Typed[T]) AddAll(itemArray []T) {
	for _, v := range itemArray {
		set.Add(v)
	}
}

// AddSet adds the contents of set "other" into the set.
func (set Typed[T]) AddSet(other Set[T]) {
	other.Iter(func(item T) error {
		set.Add(item)
		return nil
	})
}

func (set Typed[T]) Discard(item T) {
	delete(set, item)
}

func (set Typed[T]) Clear() {
	for item := range set {
		delete(set, item)
	}
}

func (set Typed[T]) Contains(item T) bool {
	_, present := set[item]
	return present
}

func (set Typed[T]) Iter(visitor func(item T) error) {
loop:
	for item := range set {
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

func (set Typed[T]) Copy() Set[T] {
	cpy := New[T]()
	for item := range set {
		cpy.Add(item)
	}
	return cpy
}

func (set Typed[T]) Slice() (s []T) {
	for item := range set {
		s = append(s, item)
	}
	return
}

func (set Typed[T]) Equals(other Set[T]) bool {
	if set.Len() != other.Len() {
		return false
	}
	for item := range set {
		if !other.Contains(item) {
			return false
		}
	}
	return true
}

func (set Typed[T]) ContainsAll(other Set[T]) bool {
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
