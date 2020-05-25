package db

import (
	"context"
	"database/sql"
	"fmt"

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

type SelectQuery struct {
	*dbr.SelectBuilder
	ctx       context.Context
	JoinCount int // for join filter
}

type InsertQuery struct {
	*dbr.InsertBuilder
	ctx  context.Context
	Hook InsertHook
}

type DeleteQuery struct {
	*dbr.DeleteBuilder
	ctx  context.Context
	Hook DeleteHook
}

type UpdateQuery struct {
	*dbr.UpdateBuilder
	ctx  context.Context
	Hook UpdateHook
}

type Conn struct {
	*dbr.Session
	ctx        context.Context
	InsertHook InsertHook
	UpdateHook UpdateHook
	DeleteHook DeleteHook
}

// SelectQuery
// Example: Select().From().Where().Limit().Offset().OrderDir().Load()
//          Select().From().Where().Limit().Offset().OrderDir().LoadOne()
//          Select().From().Where().Count()
//          SelectAll().From().Where().Limit().Offset().OrderDir().Load()
//          SelectAll().From().Where().Limit().Offset().OrderDir().LoadOne()
//          SelectAll().From().Where().Count()

func (conn *Conn) Select(columns ...string) *SelectQuery {
	return &SelectQuery{conn.Session.Select(columns...), conn.ctx, 0}
}

func (conn *Conn) SelectBySql(query string, value ...interface{}) *SelectQuery {
	return &SelectQuery{conn.Session.SelectBySql(query, value...), conn.ctx, 0}
}

func (conn *Conn) SelectAll(columns ...string) *SelectQuery {
	return &SelectQuery{conn.Session.Select("*"), conn.ctx, 0}
}

func (b *SelectQuery) Join(table, on interface{}) *SelectQuery {
	b.SelectBuilder.Join(table, on)
	return b
}
func (b *SelectQuery) RightJoin(table, on interface{}) *SelectQuery {
	b.SelectBuilder.RightJoin(table, on)
	return b
}
func (b *SelectQuery) LeftJoin(table, on interface{}) *SelectQuery {
	b.SelectBuilder.LeftJoin(table, on)
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
	n = GetOffset(n)
	b.SelectBuilder.Offset(n)
	return b
}

func (b *SelectQuery) OrderDir(col string, isAsc bool) *SelectQuery {
	b.SelectBuilder.OrderDir(col, isAsc)
	return b
}

func (b *SelectQuery) Load(value interface{}) (int, error) {
	return b.SelectBuilder.LoadContext(b.ctx, value)
}

func (b *SelectQuery) LoadOne(value interface{}) error {
	return b.SelectBuilder.LoadOneContext(b.ctx, value)
}

func getColumns(dbrColumns []interface{}) string {
	for _, column := range dbrColumns {
		if c, ok := column.(string); ok {
			return c
		}
	}
	return "*"
}

func (b *SelectQuery) Count() (count uint32, err error) {
	// cache SelectStmt
	selectStmt := b.SelectBuilder

	limit := selectStmt.LimitCount
	offset := selectStmt.OffsetCount
	column := selectStmt.Column
	isDistinct := selectStmt.IsDistinct
	order := selectStmt.Order

	b.LimitCount = -1
	b.OffsetCount = -1
	b.Column = []interface{}{"COUNT(*)"}
	b.Order = []dbr.Builder{}

	if isDistinct {
		b.Column = []interface{}{fmt.Sprintf("COUNT(DISTINCT %s)", getColumns(column))}
		b.IsDistinct = false
	}

	err = b.LoadOne(&count)
	// fallback SelectStmt
	selectStmt.LimitCount = limit
	selectStmt.OffsetCount = offset
	selectStmt.Column = column
	selectStmt.IsDistinct = isDistinct
	selectStmt.Order = order
	b.SelectBuilder = selectStmt
	return
}

// InsertQuery
// Example: InsertInto().Columns().Record().Exec()

func (conn *Conn) InsertInto(table string) *InsertQuery {
	return &InsertQuery{conn.Session.InsertInto(table), conn.ctx, conn.InsertHook}
}

func (b *InsertQuery) Exec() (sql.Result, error) {
	result, err := b.InsertBuilder.ExecContext(b.ctx)
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
	if len(b.Column) == 0 {
		b.Columns(GetColumnsFromStruct(structValue)...)
	}
	b.InsertBuilder.Record(structValue)
	return b
}

// DeleteQuery
// Example: DeleteFrom().Where().Limit().Exec()

func (conn *Conn) DeleteFrom(table string) *DeleteQuery {
	return &DeleteQuery{conn.Session.DeleteFrom(table), conn.ctx, conn.DeleteHook}
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
	result, err := b.DeleteBuilder.ExecContext(b.ctx)
	if b.Hook != nil && err == nil {
		defer b.Hook(b)
	}
	return result, err
}

// UpdateQuery
// Example: Update().Set().Where().Exec()

func (conn *Conn) Update(table string) *UpdateQuery {
	return &UpdateQuery{conn.Session.Update(table), conn.ctx, conn.UpdateHook}
}

func (b *UpdateQuery) Exec() (sql.Result, error) {
	result, err := b.UpdateBuilder.ExecContext(b.ctx)
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
