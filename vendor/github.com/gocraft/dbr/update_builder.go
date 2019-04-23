package dbr

import (
	"context"
	"database/sql"
	"fmt"
)

type UpdateBuilder struct {
	runner
	EventReceiver
	Dialect Dialect

	*UpdateStmt

	LimitCount int64
}

func (sess *Session) Update(table string) *UpdateBuilder {
	return &UpdateBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		UpdateStmt:    Update(table),
		LimitCount:    -1,
	}
}

func (tx *Tx) Update(table string) *UpdateBuilder {
	return &UpdateBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		UpdateStmt:    Update(table),
		LimitCount:    -1,
	}
}

func (sess *Session) UpdateBySql(query string, value ...interface{}) *UpdateBuilder {
	return &UpdateBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		UpdateStmt:    UpdateBySql(query, value...),
		LimitCount:    -1,
	}
}

func (tx *Tx) UpdateBySql(query string, value ...interface{}) *UpdateBuilder {
	return &UpdateBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		UpdateStmt:    UpdateBySql(query, value...),
		LimitCount:    -1,
	}
}

func (b *UpdateBuilder) Exec() (sql.Result, error) {
	return b.ExecContext(context.Background())
}

func (b *UpdateBuilder) ExecContext(ctx context.Context) (sql.Result, error) {
	return exec(ctx, b.runner, b.EventReceiver, b, b.Dialect)
}

func (b *UpdateBuilder) Set(column string, value interface{}) *UpdateBuilder {
	b.UpdateStmt.Set(column, value)
	return b
}

func (b *UpdateBuilder) SetMap(m map[string]interface{}) *UpdateBuilder {
	b.UpdateStmt.SetMap(m)
	return b
}

func (b *UpdateBuilder) Where(query interface{}, value ...interface{}) *UpdateBuilder {
	b.UpdateStmt.Where(query, value...)
	return b
}

func (b *UpdateBuilder) Limit(n uint64) *UpdateBuilder {
	b.LimitCount = int64(n)
	return b
}

func (b *UpdateBuilder) Build(d Dialect, buf Buffer) error {
	err := b.UpdateStmt.Build(b.Dialect, buf)
	if err != nil {
		return err
	}
	if b.LimitCount >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(fmt.Sprint(b.LimitCount))
	}
	return nil
}
