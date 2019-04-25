/*
Copyright 2018 The KubeSphere Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package db

import (
	"testing"

	"github.com/gocraft/dbr"
	"github.com/gocraft/dbr/dialect"
	"github.com/stretchr/testify/assert"
)

// Ref: https://github.com/gocraft/dbr/blob/5d59a8b3aa915660960329efb3af5513e7a0db07/condition_test.go
func TestCondition(t *testing.T) {
	for _, test := range []struct {
		cond  dbr.Builder
		query string
		value []interface{}
	}{
		{
			cond:  Eq("col", 1),
			query: "`col` = ?",
			value: []interface{}{1},
		},
		{
			cond:  Eq("col", nil),
			query: "`col` IS NULL",
			value: nil,
		},
		{
			cond:  Eq("col", []int{}),
			query: "0",
			value: nil,
		},
		{
			cond:  Neq("col", 1),
			query: "`col` != ?",
			value: []interface{}{1},
		},
		{
			cond:  Neq("col", nil),
			query: "`col` IS NOT NULL",
			value: nil,
		},
		{
			cond:  Gt("col", 1),
			query: "`col` > ?",
			value: []interface{}{1},
		},
		{
			cond:  Gte("col", 1),
			query: "`col` >= ?",
			value: []interface{}{1},
		},
		{
			cond:  Lt("col", 1),
			query: "`col` < ?",
			value: []interface{}{1},
		},
		{
			cond:  Lte("col", 1),
			query: "`col` <= ?",
			value: []interface{}{1},
		},
		{
			cond:  And(Lt("a", 1), Or(Gt("b", 2), Neq("c", 3))),
			query: "(`a` < ?) AND ((`b` > ?) OR (`c` != ?))",
			value: []interface{}{1, 2, 3},
		},
	} {
		buf := dbr.NewBuffer()
		err := test.cond.Build(dialect.MySQL, buf)
		assert.NoError(t, err)
		assert.Equal(t, test.query, buf.String())
		assert.Equal(t, test.value, buf.Value())
	}
}
