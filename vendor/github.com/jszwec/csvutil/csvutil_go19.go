// +build go1.9

package csvutil

import (
	"encoding/csv"
	"io"
)

func newCSVReader(r io.Reader) *csv.Reader {
	rr := csv.NewReader(r)
	rr.ReuseRecord = true
	return rr
}
