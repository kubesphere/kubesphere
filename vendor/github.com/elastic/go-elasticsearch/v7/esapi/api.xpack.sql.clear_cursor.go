// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newSQLClearCursorFunc(t Transport) SQLClearCursor {
	return func(body io.Reader, o ...func(*SQLClearCursorRequest)) (*Response, error) {
		var r = SQLClearCursorRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SQLClearCursor - Clear SQL cursor
//
type SQLClearCursor func(body io.Reader, o ...func(*SQLClearCursorRequest)) (*Response, error)

// SQLClearCursorRequest configures the SQL Clear Cursor API request.
//
type SQLClearCursorRequest struct {
	Body io.Reader

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SQLClearCursorRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(len("/_sql/close"))
	path.WriteString("/_sql/close")

	params = make(map[string]string)

	if r.Pretty {
		params["pretty"] = "true"
	}

	if r.Human {
		params["human"] = "true"
	}

	if r.ErrorTrace {
		params["error_trace"] = "true"
	}

	if len(r.FilterPath) > 0 {
		params["filter_path"] = strings.Join(r.FilterPath, ",")
	}

	req, _ := newRequest(method, path.String(), r.Body)

	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	if r.Body != nil {
		req.Header[headerContentType] = headerContentTypeJSON
	}

	if len(r.Header) > 0 {
		if len(req.Header) == 0 {
			req.Header = r.Header
		} else {
			for k, vv := range r.Header {
				for _, v := range vv {
					req.Header.Add(k, v)
				}
			}
		}
	}

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	res, err := transport.Perform(req)
	if err != nil {
		return nil, err
	}

	response := Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	return &response, nil
}

// WithContext sets the request context.
//
func (f SQLClearCursor) WithContext(v context.Context) func(*SQLClearCursorRequest) {
	return func(r *SQLClearCursorRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SQLClearCursor) WithPretty() func(*SQLClearCursorRequest) {
	return func(r *SQLClearCursorRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SQLClearCursor) WithHuman() func(*SQLClearCursorRequest) {
	return func(r *SQLClearCursorRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SQLClearCursor) WithErrorTrace() func(*SQLClearCursorRequest) {
	return func(r *SQLClearCursorRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SQLClearCursor) WithFilterPath(v ...string) func(*SQLClearCursorRequest) {
	return func(r *SQLClearCursorRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SQLClearCursor) WithHeader(h map[string]string) func(*SQLClearCursorRequest) {
	return func(r *SQLClearCursorRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
