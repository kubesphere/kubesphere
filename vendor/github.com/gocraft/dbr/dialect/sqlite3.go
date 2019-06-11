package dialect

import (
	"fmt"
	"strings"
	"time"
)

type sqlite3 struct{}

func (d sqlite3) QuoteIdent(s string) string {
	return quoteIdent(s, `"`)
}

func (d sqlite3) EncodeString(s string) string {
	// https://www.sqlite.org/faq.html
	return `'` + strings.Replace(s, `'`, `''`, -1) + `'`
}

func (d sqlite3) EncodeBool(b bool) string {
	// https://www.sqlite.org/lang_expr.html
	if b {
		return "1"
	}
	return "0"
}

func (d sqlite3) EncodeTime(t time.Time) string {
	// https://www.sqlite.org/lang_datefunc.html
	return MySQL.EncodeTime(t)
}

func (d sqlite3) EncodeBytes(b []byte) string {
	// https://www.sqlite.org/lang_expr.html
	return fmt.Sprintf(`X'%x'`, b)
}

func (d sqlite3) Placeholder(_ int) string {
	return "?"
}
