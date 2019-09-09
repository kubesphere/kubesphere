// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackSecurityAuthenticateFunc(t Transport) XPackSecurityAuthenticate {
	return func(o ...func(*XPackSecurityAuthenticateRequest)) (*Response, error) {
		var r = XPackSecurityAuthenticateRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSecurityAuthenticate - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-authenticate.html
//
type XPackSecurityAuthenticate func(o ...func(*XPackSecurityAuthenticateRequest)) (*Response, error)

// XPackSecurityAuthenticateRequest configures the X Pack Security Authenticate API request.
//
type XPackSecurityAuthenticateRequest struct {
	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackSecurityAuthenticateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_xpack/security/_authenticate"))
	path.WriteString("/_xpack/security/_authenticate")

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
func (f XPackSecurityAuthenticate) WithContext(v context.Context) func(*XPackSecurityAuthenticateRequest) {
	return func(r *XPackSecurityAuthenticateRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSecurityAuthenticate) WithPretty() func(*XPackSecurityAuthenticateRequest) {
	return func(r *XPackSecurityAuthenticateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSecurityAuthenticate) WithHuman() func(*XPackSecurityAuthenticateRequest) {
	return func(r *XPackSecurityAuthenticateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSecurityAuthenticate) WithErrorTrace() func(*XPackSecurityAuthenticateRequest) {
	return func(r *XPackSecurityAuthenticateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSecurityAuthenticate) WithFilterPath(v ...string) func(*XPackSecurityAuthenticateRequest) {
	return func(r *XPackSecurityAuthenticateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSecurityAuthenticate) WithHeader(h map[string]string) func(*XPackSecurityAuthenticateRequest) {
	return func(r *XPackSecurityAuthenticateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
