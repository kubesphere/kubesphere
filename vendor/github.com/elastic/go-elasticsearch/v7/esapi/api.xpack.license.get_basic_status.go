// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newLicenseGetBasicStatusFunc(t Transport) LicenseGetBasicStatus {
	return func(o ...func(*LicenseGetBasicStatusRequest)) (*Response, error) {
		var r = LicenseGetBasicStatusRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// LicenseGetBasicStatus - https://www.elastic.co/guide/en/elasticsearch/reference/master/get-basic-status.html
//
type LicenseGetBasicStatus func(o ...func(*LicenseGetBasicStatusRequest)) (*Response, error)

// LicenseGetBasicStatusRequest configures the License Get Basic Status API request.
//
type LicenseGetBasicStatusRequest struct {
	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r LicenseGetBasicStatusRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_license/basic_status"))
	path.WriteString("/_license/basic_status")

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
func (f LicenseGetBasicStatus) WithContext(v context.Context) func(*LicenseGetBasicStatusRequest) {
	return func(r *LicenseGetBasicStatusRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f LicenseGetBasicStatus) WithPretty() func(*LicenseGetBasicStatusRequest) {
	return func(r *LicenseGetBasicStatusRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f LicenseGetBasicStatus) WithHuman() func(*LicenseGetBasicStatusRequest) {
	return func(r *LicenseGetBasicStatusRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f LicenseGetBasicStatus) WithErrorTrace() func(*LicenseGetBasicStatusRequest) {
	return func(r *LicenseGetBasicStatusRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f LicenseGetBasicStatus) WithFilterPath(v ...string) func(*LicenseGetBasicStatusRequest) {
	return func(r *LicenseGetBasicStatusRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f LicenseGetBasicStatus) WithHeader(h map[string]string) func(*LicenseGetBasicStatusRequest) {
	return func(r *LicenseGetBasicStatusRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
