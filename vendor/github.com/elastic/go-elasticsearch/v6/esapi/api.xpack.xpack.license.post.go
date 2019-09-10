// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newXPackLicensePostFunc(t Transport) XPackLicensePost {
	return func(o ...func(*XPackLicensePostRequest)) (*Response, error) {
		var r = XPackLicensePostRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackLicensePost - https://www.elastic.co/guide/en/elasticsearch/reference/6.7/update-license.html
//
type XPackLicensePost func(o ...func(*XPackLicensePostRequest)) (*Response, error)

// XPackLicensePostRequest configures the X Pack License Post API request.
//
type XPackLicensePostRequest struct {
	Body io.Reader

	Acknowledge *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackLicensePostRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(len("/_xpack/license"))
	path.WriteString("/_xpack/license")

	params = make(map[string]string)

	if r.Acknowledge != nil {
		params["acknowledge"] = strconv.FormatBool(*r.Acknowledge)
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
func (f XPackLicensePost) WithContext(v context.Context) func(*XPackLicensePostRequest) {
	return func(r *XPackLicensePostRequest) {
		r.ctx = v
	}
}

// WithBody - licenses to be installed.
//
func (f XPackLicensePost) WithBody(v io.Reader) func(*XPackLicensePostRequest) {
	return func(r *XPackLicensePostRequest) {
		r.Body = v
	}
}

// WithAcknowledge - whether the user has acknowledged acknowledge messages (default: false).
//
func (f XPackLicensePost) WithAcknowledge(v bool) func(*XPackLicensePostRequest) {
	return func(r *XPackLicensePostRequest) {
		r.Acknowledge = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackLicensePost) WithPretty() func(*XPackLicensePostRequest) {
	return func(r *XPackLicensePostRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackLicensePost) WithHuman() func(*XPackLicensePostRequest) {
	return func(r *XPackLicensePostRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackLicensePost) WithErrorTrace() func(*XPackLicensePostRequest) {
	return func(r *XPackLicensePostRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackLicensePost) WithFilterPath(v ...string) func(*XPackLicensePostRequest) {
	return func(r *XPackLicensePostRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackLicensePost) WithHeader(h map[string]string) func(*XPackLicensePostRequest) {
	return func(r *XPackLicensePostRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
