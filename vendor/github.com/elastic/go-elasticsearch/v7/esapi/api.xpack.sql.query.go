// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newSQLQueryFunc(t Transport) SQLQuery {
	return func(body io.Reader, o ...func(*SQLQueryRequest)) (*Response, error) {
		var r = SQLQueryRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SQLQuery - Execute SQL
//
type SQLQuery func(body io.Reader, o ...func(*SQLQueryRequest)) (*Response, error)

// SQLQueryRequest configures the SQL Query API request.
//
type SQLQueryRequest struct {
	Body io.Reader

	Format string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SQLQueryRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(len("/_sql"))
	path.WriteString("/_sql")

	params = make(map[string]string)

	if r.Format != "" {
		params["format"] = r.Format
	}

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
func (f SQLQuery) WithContext(v context.Context) func(*SQLQueryRequest) {
	return func(r *SQLQueryRequest) {
		r.ctx = v
	}
}

// WithFormat - a short version of the accept header, e.g. json, yaml.
//
func (f SQLQuery) WithFormat(v string) func(*SQLQueryRequest) {
	return func(r *SQLQueryRequest) {
		r.Format = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SQLQuery) WithPretty() func(*SQLQueryRequest) {
	return func(r *SQLQueryRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SQLQuery) WithHuman() func(*SQLQueryRequest) {
	return func(r *SQLQueryRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SQLQuery) WithErrorTrace() func(*SQLQueryRequest) {
	return func(r *SQLQueryRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SQLQuery) WithFilterPath(v ...string) func(*SQLQueryRequest) {
	return func(r *SQLQueryRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SQLQuery) WithHeader(h map[string]string) func(*SQLQueryRequest) {
	return func(r *SQLQueryRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
