// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newMLUpdateFilterFunc(t Transport) MLUpdateFilter {
	return func(body io.Reader, filter_id string, o ...func(*MLUpdateFilterRequest)) (*Response, error) {
		var r = MLUpdateFilterRequest{Body: body, FilterID: filter_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLUpdateFilter -
//
type MLUpdateFilter func(body io.Reader, filter_id string, o ...func(*MLUpdateFilterRequest)) (*Response, error)

// MLUpdateFilterRequest configures the ML Update Filter API request.
//
type MLUpdateFilterRequest struct {
	Body io.Reader

	FilterID string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLUpdateFilterRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_ml") + 1 + len("filters") + 1 + len(r.FilterID) + 1 + len("_update"))
	path.WriteString("/")
	path.WriteString("_ml")
	path.WriteString("/")
	path.WriteString("filters")
	path.WriteString("/")
	path.WriteString(r.FilterID)
	path.WriteString("/")
	path.WriteString("_update")

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
func (f MLUpdateFilter) WithContext(v context.Context) func(*MLUpdateFilterRequest) {
	return func(r *MLUpdateFilterRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLUpdateFilter) WithPretty() func(*MLUpdateFilterRequest) {
	return func(r *MLUpdateFilterRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLUpdateFilter) WithHuman() func(*MLUpdateFilterRequest) {
	return func(r *MLUpdateFilterRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLUpdateFilter) WithErrorTrace() func(*MLUpdateFilterRequest) {
	return func(r *MLUpdateFilterRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLUpdateFilter) WithFilterPath(v ...string) func(*MLUpdateFilterRequest) {
	return func(r *MLUpdateFilterRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLUpdateFilter) WithHeader(h map[string]string) func(*MLUpdateFilterRequest) {
	return func(r *MLUpdateFilterRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
