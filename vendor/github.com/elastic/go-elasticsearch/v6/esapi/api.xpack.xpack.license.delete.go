// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackLicenseDeleteFunc(t Transport) XPackLicenseDelete {
	return func(o ...func(*XPackLicenseDeleteRequest)) (*Response, error) {
		var r = XPackLicenseDeleteRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackLicenseDelete - https://www.elastic.co/guide/en/elasticsearch/reference/6.7/delete-license.html
//
type XPackLicenseDelete func(o ...func(*XPackLicenseDeleteRequest)) (*Response, error)

// XPackLicenseDeleteRequest configures the X Pack License Delete API request.
//
type XPackLicenseDeleteRequest struct {
	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackLicenseDeleteRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(len("/_xpack/license"))
	path.WriteString("/_xpack/license")

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

	req, _ := newRequest(method, path.String(), nil)

	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
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
func (f XPackLicenseDelete) WithContext(v context.Context) func(*XPackLicenseDeleteRequest) {
	return func(r *XPackLicenseDeleteRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackLicenseDelete) WithPretty() func(*XPackLicenseDeleteRequest) {
	return func(r *XPackLicenseDeleteRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackLicenseDelete) WithHuman() func(*XPackLicenseDeleteRequest) {
	return func(r *XPackLicenseDeleteRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackLicenseDelete) WithErrorTrace() func(*XPackLicenseDeleteRequest) {
	return func(r *XPackLicenseDeleteRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackLicenseDelete) WithFilterPath(v ...string) func(*XPackLicenseDeleteRequest) {
	return func(r *XPackLicenseDeleteRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackLicenseDelete) WithHeader(h map[string]string) func(*XPackLicenseDeleteRequest) {
	return func(r *XPackLicenseDeleteRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
