package dialect

import (
	"bytes"
	"fmt"
	"time"
)

type mysql struct{}

func (d mysql) QuoteIdent(s string) string {
	return quoteIdent(s, "`")
}

func (d mysql) EncodeString(s string) string {
	buf := new(bytes.Buffer)

	buf.WriteRune('\'')
	// https://dev.mysql.com/doc/refman/5.7/en/string-literals.html
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case 0:
			buf.WriteString(`\0`)
		case '\'':
			buf.WriteString(`\'`)
		case '"':
			buf.WriteString(`\"`)
		case '\b':
			buf.WriteString(`\b`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		case 26:
			buf.WriteString(`\Z`)
		case '\\':
			buf.WriteString(`\\`)
		default:
			buf.WriteByte(s[i])
		}
	}

	buf.WriteRune('\'')
	return buf.String()
}

func (d mysql) EncodeBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func (d mysql) EncodeTime(t time.Time) string {
	return `'` + t.UTC().Format(timeFormat) + `'`
}

func (d mysql) EncodeBytes(b []byte) string {
	return fmt.Sprintf(`0x%x`, b)
}

func (d mysql) Placeholder(_ int) string {
	return "?"
}
