package dbr

import "context"

type SelectBuilder struct {
	runner
	EventReceiver
	Dialect Dialect

	*SelectStmt
}

func prepareSelect(a []string) []interface{} {
	b := make([]interface{}, len(a))
	for i := range a {
		b[i] = a[i]
	}
	return b
}

func (sess *Session) Select(column ...string) *SelectBuilder {
	return &SelectBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		SelectStmt:    Select(prepareSelect(column)...),
	}
}

func (tx *Tx) Select(column ...string) *SelectBuilder {
	return &SelectBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		SelectStmt:    Select(prepareSelect(column)...),
	}
}

func (sess *Session) SelectBySql(query string, value ...interface{}) *SelectBuilder {
	return &SelectBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		SelectStmt:    SelectBySql(query, value...),
	}
}

func (tx *Tx) SelectBySql(query string, value ...interface{}) *SelectBuilder {
	return &SelectBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		SelectStmt:    SelectBySql(query, value...),
	}
}

// DEPRECATED: use LoadOne instead
func (b *SelectBuilder) LoadStruct(value interface{}) error {
	return b.LoadOne(value)
}

// DEPRECATED: use Load instead
func (b *SelectBuilder) LoadStructs(value interface{}) (int, error) {
	return b.Load(value)
}

// DEPRECATED: use LoadOne instead
func (b *SelectBuilder) LoadValue(value interface{}) error {
	return b.LoadOne(value)
}

// DEPRECATED: use Load instead
func (b *SelectBuilder) LoadValues(value interface{}) (int, error) {
	return b.Load(value)
}

func (b *SelectBuilder) LoadOneContext(ctx context.Context, value interface{}) error {
	count, err := query(ctx, b.runner, b.EventReceiver, b, b.Dialect, value)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (b *SelectBuilder) LoadOne(value interface{}) error {
	return b.LoadOneContext(context.Background(), value)
}

func (b *SelectBuilder) LoadContext(ctx context.Context, value interface{}) (int, error) {
	return query(ctx, b.runner, b.EventReceiver, b, b.Dialect, value)
}

func (b *SelectBuilder) Load(value interface{}) (int, error) {
	return b.LoadContext(context.Background(), value)
}

func (b *SelectBuilder) Join(table, on interface{}) *SelectBuilder {
	b.SelectStmt.Join(table, on)
	return b
}

func (b *SelectBuilder) LeftJoin(table, on interface{}) *SelectBuilder {
	b.SelectStmt.LeftJoin(table, on)
	return b
}

func (b *SelectBuilder) RightJoin(table, on interface{}) *SelectBuilder {
	b.SelectStmt.RightJoin(table, on)
	return b
}

func (b *SelectBuilder) FullJoin(table, on interface{}) *SelectBuilder {
	b.SelectStmt.FullJoin(table, on)
	return b
}

func (b *SelectBuilder) Distinct() *SelectBuilder {
	b.SelectStmt.Distinct()
	return b
}

func (b *SelectBuilder) From(table interface{}) *SelectBuilder {
	b.SelectStmt.From(table)
	return b
}

func (b *SelectBuilder) GroupBy(col ...string) *SelectBuilder {
	b.SelectStmt.GroupBy(col...)
	return b
}

func (b *SelectBuilder) Having(query interface{}, value ...interface{}) *SelectBuilder {
	b.SelectStmt.Having(query, value...)
	return b
}

func (b *SelectBuilder) Limit(n uint64) *SelectBuilder {
	b.SelectStmt.Limit(n)
	return b
}

func (b *SelectBuilder) Offset(n uint64) *SelectBuilder {
	b.SelectStmt.Offset(n)
	return b
}

func (b *SelectBuilder) OrderDir(col string, isAsc bool) *SelectBuilder {
	if isAsc {
		b.SelectStmt.OrderAsc(col)
	} else {
		b.SelectStmt.OrderDesc(col)
	}
	return b
}

func (b *SelectBuilder) Paginate(page, perPage uint64) *SelectBuilder {
	b.Limit(perPage)
	b.Offset((page - 1) * perPage)
	return b
}

func (b *SelectBuilder) OrderBy(col string) *SelectBuilder {
	b.SelectStmt.Order = append(b.SelectStmt.Order, Expr(col))
	return b
}

func (b *SelectBuilder) Where(query interface{}, value ...interface{}) *SelectBuilder {
	b.SelectStmt.Where(query, value...)
	return b
}
