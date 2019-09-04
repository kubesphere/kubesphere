// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newLicenseDeleteFunc(t Transport) LicenseDelete {
	return func(o ...func(*LicenseDeleteRequest)) (*Response, error) {
		var r = LicenseDeleteRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// LicenseDelete - https://www.elastic.co/guide/en/elasticsearch/reference/master/delete-license.html
//
type LicenseDelete func(o ...func(*LicenseDeleteRequest)) (*Response, error)

// LicenseDeleteRequest configures the License Delete API request.
//
type LicenseDeleteRequest struct {
	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r LicenseDeleteRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(len("/_license"))
	path.WriteString("/_license")

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
func (f LicenseDelete) WithContext(v context.Context) func(*LicenseDeleteRequest) {
	return func(r *LicenseDeleteRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f LicenseDelete) WithPretty() func(*LicenseDeleteRequest) {
	return func(r *LicenseDeleteRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f LicenseDelete) WithHuman() func(*LicenseDeleteRequest) {
	return func(r *LicenseDeleteRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f LicenseDelete) WithErrorTrace() func(*LicenseDeleteRequest) {
	return func(r *LicenseDeleteRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f LicenseDelete) WithFilterPath(v ...string) func(*LicenseDeleteRequest) {
	return func(r *LicenseDeleteRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f LicenseDelete) WithHeader(h map[string]string) func(*LicenseDeleteRequest) {
	return func(r *LicenseDeleteRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
