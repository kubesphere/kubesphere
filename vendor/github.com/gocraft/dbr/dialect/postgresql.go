package dialect

import (
	"fmt"
	"strings"
	"time"
)

type postgreSQL struct{}

func (d postgreSQL) QuoteIdent(s string) string {
	return quoteIdent(s, `"`)
}

func (d postgreSQL) EncodeString(s string) string {
	// http://www.postgresql.org/docs/9.2/static/sql-syntax-lexical.html
	return `'` + strings.Replace(s, `'`, `''`, -1) + `'`
}

func (d postgreSQL) EncodeBool(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}

func (d postgreSQL) EncodeTime(t time.Time) string {
	return MySQL.EncodeTime(t)
}

func (d postgreSQL) EncodeBytes(b []byte) string {
	return fmt.Sprintf(`E'\\x%x'`, b)
}

func (d postgreSQL) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n+1)
}
