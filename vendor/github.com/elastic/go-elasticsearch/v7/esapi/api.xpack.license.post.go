// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newLicensePostFunc(t Transport) LicensePost {
	return func(o ...func(*LicensePostRequest)) (*Response, error) {
		var r = LicensePostRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// LicensePost - https://www.elastic.co/guide/en/elasticsearch/reference/master/update-license.html
//
type LicensePost func(o ...func(*LicensePostRequest)) (*Response, error)

// LicensePostRequest configures the License Post API request.
//
type LicensePostRequest struct {
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
func (r LicensePostRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(len("/_license"))
	path.WriteString("/_license")

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
func (f LicensePost) WithContext(v context.Context) func(*LicensePostRequest) {
	return func(r *LicensePostRequest) {
		r.ctx = v
	}
}

// WithBody - licenses to be installed.
//
func (f LicensePost) WithBody(v io.Reader) func(*LicensePostRequest) {
	return func(r *LicensePostRequest) {
		r.Body = v
	}
}

// WithAcknowledge - whether the user has acknowledged acknowledge messages (default: false).
//
func (f LicensePost) WithAcknowledge(v bool) func(*LicensePostRequest) {
	return func(r *LicensePostRequest) {
		r.Acknowledge = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f LicensePost) WithPretty() func(*LicensePostRequest) {
	return func(r *LicensePostRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f LicensePost) WithHuman() func(*LicensePostRequest) {
	return func(r *LicensePostRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f LicensePost) WithErrorTrace() func(*LicensePostRequest) {
	return func(r *LicensePostRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f LicensePost) WithFilterPath(v ...string) func(*LicensePostRequest) {
	return func(r *LicensePostRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f LicensePost) WithHeader(h map[string]string) func(*LicensePostRequest) {
	return func(r *LicensePostRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
