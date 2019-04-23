package dbr

type joinType uint8

const (
	inner joinType = iota
	left
	right
	full
)

func join(t joinType, table interface{}, on interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		buf.WriteString(" ")
		switch t {
		case left:
			buf.WriteString("LEFT ")
		case right:
			buf.WriteString("RIGHT ")
		case full:
			buf.WriteString("FULL ")
		}
		buf.WriteString("JOIN ")
		switch table := table.(type) {
		case string:
			buf.WriteString(d.QuoteIdent(table))
		default:
			buf.WriteString(placeholder)
			buf.WriteValue(table)
		}
		buf.WriteString(" ON ")
		switch on := on.(type) {
		case string:
			buf.WriteString(on)
		case Builder:
			buf.WriteString(placeholder)
			buf.WriteValue(on)
		}
		return nil
	})
}
