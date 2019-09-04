package esapi

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// Response represents the API response.
//
type Response struct {
	StatusCode int
	Header     http.Header
	Body       io.ReadCloser
}

// String returns the response as a string.
//
// The intended usage is for testing or debugging only.
//
func (r *Response) String() string {
	var (
		out = new(bytes.Buffer)
		b1  = bytes.NewBuffer([]byte{})
		b2  = bytes.NewBuffer([]byte{})
		tr  = io.TeeReader(r.Body, b1)
	)

	defer r.Body.Close()

	if _, err := io.Copy(b2, tr); err != nil {
		out.WriteString(fmt.Sprintf("<error reading response body: %v>", err))
		return out.String()
	}
	defer func() { r.Body = ioutil.NopCloser(b1) }()

	out.WriteString(fmt.Sprintf("[%d %s]", r.StatusCode, http.StatusText(r.StatusCode)))
	out.WriteRune(' ')
	out.ReadFrom(b2) // errcheck exclude (*bytes.Buffer).ReadFrom

	return out.String()
}

// Status returns the response status as a string.
//
func (r *Response) Status() string {
	var b strings.Builder
	b.WriteString(strconv.Itoa(r.StatusCode))
	b.WriteString(" ")
	b.WriteString(http.StatusText(r.StatusCode))
	return b.String()
}

// IsError returns true when the response status indicates failure.
//
func (r *Response) IsError() bool {
	return r.StatusCode > 299
}
