package dbr

// identifier is a type of string
type I string

func (i I) Build(d Dialect, buf Buffer) error {
	buf.WriteString(d.QuoteIdent(string(i)))
	return nil
}

// As creates an alias for expr. e.g. SELECT `a1` AS `a2`
func (i I) As(alias string) Builder {
	return as(i, alias)
}

func as(expr interface{}, alias string) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		buf.WriteString(placeholder)
		buf.WriteValue(expr)
		buf.WriteString(" AS ")
		buf.WriteString(d.QuoteIdent(alias))
		return nil
	})
}
