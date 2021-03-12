// +build !go1.9

package csvutil

import (
	"encoding/csv"
	"io"
)

func newCSVReader(r io.Reader) *csv.Reader {
	return csv.NewReader(r)
}
