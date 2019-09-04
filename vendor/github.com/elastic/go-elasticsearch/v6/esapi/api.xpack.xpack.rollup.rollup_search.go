// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newXPackRollupRollupSearchFunc(t Transport) XPackRollupRollupSearch {
	return func(index string, body io.Reader, o ...func(*XPackRollupRollupSearchRequest)) (*Response, error) {
		var r = XPackRollupRollupSearchRequest{Index: index, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackRollupRollupSearch -
//
type XPackRollupRollupSearch func(index string, body io.Reader, o ...func(*XPackRollupRollupSearchRequest)) (*Response, error)

// XPackRollupRollupSearchRequest configures the X Pack Rollup Rollup Search API request.
//
type XPackRollupRollupSearchRequest struct {
	Index        string
	DocumentType string

	Body io.Reader

	TypedKeys *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackRollupRollupSearchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(r.Index) + 1 + len(r.DocumentType) + 1 + len("_rollup_search"))
	path.WriteString("/")
	path.WriteString(r.Index)
	if r.DocumentType != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentType)
	}
	path.WriteString("/")
	path.WriteString("_rollup_search")

	params = make(map[string]string)

	if r.TypedKeys != nil {
		params["typed_keys"] = strconv.FormatBool(*r.TypedKeys)
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
func (f XPackRollupRollupSearch) WithContext(v context.Context) func(*XPackRollupRollupSearchRequest) {
	return func(r *XPackRollupRollupSearchRequest) {
		r.ctx = v
	}
}

// WithDocumentType - the doc type inside the index.
//
func (f XPackRollupRollupSearch) WithDocumentType(v string) func(*XPackRollupRollupSearchRequest) {
	return func(r *XPackRollupRollupSearchRequest) {
		r.DocumentType = v
	}
}

// WithTypedKeys - specify whether aggregation and suggester names should be prefixed by their respective types in the response.
//
func (f XPackRollupRollupSearch) WithTypedKeys(v bool) func(*XPackRollupRollupSearchRequest) {
	return func(r *XPackRollupRollupSearchRequest) {
		r.TypedKeys = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackRollupRollupSearch) WithPretty() func(*XPackRollupRollupSearchRequest) {
	return func(r *XPackRollupRollupSearchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackRollupRollupSearch) WithHuman() func(*XPackRollupRollupSearchRequest) {
	return func(r *XPackRollupRollupSearchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackRollupRollupSearch) WithErrorTrace() func(*XPackRollupRollupSearchRequest) {
	return func(r *XPackRollupRollupSearchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackRollupRollupSearch) WithFilterPath(v ...string) func(*XPackRollupRollupSearchRequest) {
	return func(r *XPackRollupRollupSearchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackRollupRollupSearch) WithHeader(h map[string]string) func(*XPackRollupRollupSearchRequest) {
	return func(r *XPackRollupRollupSearchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
