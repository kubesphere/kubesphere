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
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
)

const (
	DefaultSelectLimit = 200
)

func GetLimit(n uint64) uint64 {
	if n < 0 {
		n = 0
	}
	if n > DefaultSelectLimit {
		n = DefaultSelectLimit
	}
	return n
}

func GetOffset(n uint64) uint64 {
	if n < 0 {
		n = 0
	}
	return n
}

type InsertHook func(query *InsertQuery)
type UpdateHook func(query *UpdateQuery)
type DeleteHook func(query *DeleteQuery)

type Database struct {
	*dbr.Session
	InsertHook InsertHook
	UpdateHook UpdateHook
	DeleteHook DeleteHook
}

type SelectQuery struct {
	*dbr.SelectBuilder
	JoinCount int // for join filter
}

type InsertQuery struct {
	*dbr.InsertBuilder
	Hook InsertHook
}

type DeleteQuery struct {
	*dbr.DeleteBuilder
	Hook DeleteHook
}

type UpdateQuery struct {
	*dbr.UpdateBuilder
	Hook UpdateHook
}

type UpsertQuery struct {
	table string
	*dbr.Session
	whereConds   map[string]string
	upsertValues map[string]interface{}
}

// SelectQuery
// Example: Select().From().Where().Limit().Offset().OrderDir().Load()
//          Select().From().Where().Limit().Offset().OrderDir().LoadOne()
//          Select().From().Where().Count()
//          SelectAll().From().Where().Limit().Offset().OrderDir().Load()
//          SelectAll().From().Where().Limit().Offset().OrderDir().LoadOne()
//          SelectAll().From().Where().Count()

func (db *Database) Select(columns ...string) *SelectQuery {
	return &SelectQuery{db.Session.Select(columns...), 0}
}

func (db *Database) SelectBySql(query string, value ...interface{}) *SelectQuery {
	return &SelectQuery{db.Session.SelectBySql(query, value...), 0}
}

func (db *Database) SelectAll(columns ...string) *SelectQuery {
	return &SelectQuery{db.Session.Select("*"), 0}
}

func (b *SelectQuery) Join(table, on interface{}) *SelectQuery {
	b.SelectBuilder.Join(table, on)
	return b
}

func (b *SelectQuery) JoinAs(table string, alias string, on interface{}) *SelectQuery {
	b.SelectBuilder.Join(dbr.I(table).As(alias), on)
	return b
}

func (b *SelectQuery) From(table string) *SelectQuery {
	b.SelectBuilder.From(table)
	return b
}

func (b *SelectQuery) Where(query interface{}, value ...interface{}) *SelectQuery {
	b.SelectBuilder.Where(query, value...)
	return b
}

func (b *SelectQuery) GroupBy(col ...string) *SelectQuery {
	b.SelectBuilder.GroupBy(col...)
	return b
}

func (b *SelectQuery) Distinct() *SelectQuery {
	b.SelectBuilder.Distinct()
	return b
}

func (b *SelectQuery) Limit(n uint64) *SelectQuery {
	n = GetLimit(n)
	b.SelectBuilder.Limit(n)
	return b
}

func (b *SelectQuery) Offset(n uint64) *SelectQuery {
	n = GetLimit(n)
	b.SelectBuilder.Offset(n)
	return b
}

func (b *SelectQuery) OrderDir(col string, isAsc bool) *SelectQuery {
	b.SelectBuilder.OrderDir(col, isAsc)
	return b
}

func (b *SelectQuery) Load(value interface{}) (int, error) {
	return b.SelectBuilder.Load(value)
}

func (b *SelectQuery) LoadOne(value interface{}) error {
	return b.SelectBuilder.LoadOne(value)
}

func getColumns(dbrColumns []interface{}) string {
	var columns []string
	for _, column := range dbrColumns {
		if c, ok := column.(string); ok {
			columns = append(columns, c)
		}
	}
	return strings.Join(columns, ", ")
}

func (b *SelectQuery) Count() (count uint32, err error) {
	// cache SelectStmt
	selectStmt := b.SelectStmt

	limit := selectStmt.LimitCount
	offset := selectStmt.OffsetCount
	column := selectStmt.Column
	isDistinct := selectStmt.IsDistinct
	order := selectStmt.Order

	b.SelectStmt.LimitCount = -1
	b.SelectStmt.OffsetCount = -1
	b.SelectStmt.Column = []interface{}{"COUNT(*)"}
	b.SelectStmt.Order = []dbr.Builder{}

	if isDistinct {
		b.SelectStmt.Column = []interface{}{fmt.Sprintf("COUNT(DISTINCT %s)", getColumns(column))}
		b.SelectStmt.IsDistinct = false
	}

	err = b.LoadOne(&count)
	// fallback SelectStmt
	selectStmt.LimitCount = limit
	selectStmt.OffsetCount = offset
	selectStmt.Column = column
	selectStmt.IsDistinct = isDistinct
	selectStmt.Order = order
	b.SelectStmt = selectStmt
	return
}

// InsertQuery
// Example: InsertInto().Columns().Record().Exec()

func (db *Database) InsertInto(table string) *InsertQuery {
	return &InsertQuery{db.Session.InsertInto(table), db.InsertHook}
}

func (b *InsertQuery) Exec() (sql.Result, error) {
	result, err := b.InsertBuilder.Exec()
	if b.Hook != nil && err == nil {
		defer b.Hook(b)
	}
	return result, err
}

func (b *InsertQuery) Columns(columns ...string) *InsertQuery {
	b.InsertBuilder.Columns(columns...)
	return b
}

func (b *InsertQuery) Record(structValue interface{}) *InsertQuery {
	b.InsertBuilder.Record(structValue)
	return b
}

// DeleteQuery
// Example: DeleteFrom().Where().Limit().Exec()

func (db *Database) DeleteFrom(table string) *DeleteQuery {
	return &DeleteQuery{db.Session.DeleteFrom(table), db.DeleteHook}
}

func (b *DeleteQuery) Where(query interface{}, value ...interface{}) *DeleteQuery {
	b.DeleteBuilder.Where(query, value...)
	return b
}

func (b *DeleteQuery) Limit(n uint64) *DeleteQuery {
	b.DeleteBuilder.Limit(n)
	return b
}

func (b *DeleteQuery) Exec() (sql.Result, error) {
	result, err := b.DeleteBuilder.Exec()
	if b.Hook != nil && err == nil {
		defer b.Hook(b)
	}
	return result, err
}

// UpdateQuery
// Example: Update().Set().Where().Exec()

func (db *Database) Update(table string) *UpdateQuery {
	return &UpdateQuery{db.Session.Update(table), db.UpdateHook}
}

func (b *UpdateQuery) Exec() (sql.Result, error) {
	result, err := b.UpdateBuilder.Exec()
	if b.Hook != nil && err == nil {
		defer b.Hook(b)
	}
	return result, err
}

func (b *UpdateQuery) Set(column string, value interface{}) *UpdateQuery {
	b.UpdateBuilder.Set(column, value)
	return b
}

func (b *UpdateQuery) SetMap(m map[string]interface{}) *UpdateQuery {
	b.UpdateBuilder.SetMap(m)
	return b
}

func (b *UpdateQuery) Where(query interface{}, value ...interface{}) *UpdateQuery {
	b.UpdateBuilder.Where(query, value...)
	return b
}

func (b *UpdateQuery) Limit(n uint64) *UpdateQuery {
	b.UpdateBuilder.Limit(n)
	return b
}
