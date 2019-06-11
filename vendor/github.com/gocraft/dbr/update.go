package dbr

// UpdateStmt builds `UPDATE ...`
type UpdateStmt struct {
	raw

	Table string
	Value map[string]interface{}

	WhereCond []Builder
}

// Build builds `UPDATE ...` in dialect
func (b *UpdateStmt) Build(d Dialect, buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.Build(d, buf)
	}

	if b.Table == "" {
		return ErrTableNotSpecified
	}

	if len(b.Value) == 0 {
		return ErrColumnNotSpecified
	}

	buf.WriteString("UPDATE ")
	buf.WriteString(d.QuoteIdent(b.Table))
	buf.WriteString(" SET ")

	i := 0
	for col, v := range b.Value {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(d.QuoteIdent(col))
		buf.WriteString(" = ")
		buf.WriteString(placeholder)

		buf.WriteValue(v)
		i++
	}

	if len(b.WhereCond) > 0 {
		buf.WriteString(" WHERE ")
		err := And(b.WhereCond...).Build(d, buf)
		if err != nil {
			return err
		}
	}
	return nil
}

// Update creates an UpdateStmt
func Update(table string) *UpdateStmt {
	return &UpdateStmt{
		Table: table,
		Value: make(map[string]interface{}),
	}
}

// UpdateBySql creates an UpdateStmt with raw query
func UpdateBySql(query string, value ...interface{}) *UpdateStmt {
	return &UpdateStmt{
		raw: raw{
			Query: query,
			Value: value,
		},
		Value: make(map[string]interface{}),
	}
}

// Where adds a where condition
func (b *UpdateStmt) Where(query interface{}, value ...interface{}) *UpdateStmt {
	switch query := query.(type) {
	case string:
		b.WhereCond = append(b.WhereCond, Expr(query, value...))
	case Builder:
		b.WhereCond = append(b.WhereCond, query)
	}
	return b
}

// Set specifies a key-value pair
func (b *UpdateStmt) Set(column string, value interface{}) *UpdateStmt {
	b.Value[column] = value
	return b
}

// SetMap specifies a list of key-value pair
func (b *UpdateStmt) SetMap(m map[string]interface{}) *UpdateStmt {
	for col, val := range m {
		b.Set(col, val)
	}
	return b
}
