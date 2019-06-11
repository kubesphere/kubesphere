package dialect

import "strings"

var (
	// MySQL dialect
	MySQL = mysql{}
	// PostgreSQL dialect
	PostgreSQL = postgreSQL{}
	// SQLite3 dialect
	SQLite3 = sqlite3{}
)

const (
	timeFormat = "2006-01-02 15:04:05.000000"
)

func quoteIdent(s, quote string) string {
	part := strings.SplitN(s, ".", 2)
	if len(part) == 2 {
		return quoteIdent(part[0], quote) + "." + quoteIdent(part[1], quote)
	}
	return quote + s + quote
}
