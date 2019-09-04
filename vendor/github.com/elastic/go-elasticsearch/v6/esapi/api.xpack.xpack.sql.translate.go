// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newXPackSQLTranslateFunc(t Transport) XPackSQLTranslate {
	return func(body io.Reader, o ...func(*XPackSQLTranslateRequest)) (*Response, error) {
		var r = XPackSQLTranslateRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSQLTranslate - Translate SQL into Elasticsearch queries
//
type XPackSQLTranslate func(body io.Reader, o ...func(*XPackSQLTranslateRequest)) (*Response, error)

// XPackSQLTranslateRequest configures the X PackSQL Translate API request.
//
type XPackSQLTranslateRequest struct {
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
func (r XPackSQLTranslateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(len("/_xpack/sql/translate"))
	path.WriteString("/_xpack/sql/translate")

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
func (f XPackSQLTranslate) WithContext(v context.Context) func(*XPackSQLTranslateRequest) {
	return func(r *XPackSQLTranslateRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSQLTranslate) WithPretty() func(*XPackSQLTranslateRequest) {
	return func(r *XPackSQLTranslateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSQLTranslate) WithHuman() func(*XPackSQLTranslateRequest) {
	return func(r *XPackSQLTranslateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSQLTranslate) WithErrorTrace() func(*XPackSQLTranslateRequest) {
	return func(r *XPackSQLTranslateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSQLTranslate) WithFilterPath(v ...string) func(*XPackSQLTranslateRequest) {
	return func(r *XPackSQLTranslateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSQLTranslate) WithHeader(h map[string]string) func(*XPackSQLTranslateRequest) {
	return func(r *XPackSQLTranslateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
