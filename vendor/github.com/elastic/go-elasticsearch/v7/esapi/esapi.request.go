package esapi

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	headerContentType = "Content-Type"
)

var (
	headerContentTypeJSON = []string{"application/json"}
)

// Request defines the API request.
//
type Request interface {
	Do(ctx context.Context, transport Transport) (*Response, error)
}

// newRequest creates an HTTP request.
//
func newRequest(method, path string, body io.Reader) (*http.Request, error) {
	r := http.Request{
		Method:     method,
		URL:        &url.URL{Path: path},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}

	if body != nil {
		switch b := body.(type) {
		case *bytes.Buffer:
			r.Body = ioutil.NopCloser(body)
			r.ContentLength = int64(b.Len())
		case *bytes.Reader:
			r.Body = ioutil.NopCloser(body)
			r.ContentLength = int64(b.Len())
		case *strings.Reader:
			r.Body = ioutil.NopCloser(body)
			r.ContentLength = int64(b.Len())
		default:
			r.Body = ioutil.NopCloser(body)
		}
	}

	return &r, nil
}
