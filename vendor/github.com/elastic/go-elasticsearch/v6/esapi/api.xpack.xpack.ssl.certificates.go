// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackSSLCertificatesFunc(t Transport) XPackSSLCertificates {
	return func(o ...func(*XPackSSLCertificatesRequest)) (*Response, error) {
		var r = XPackSSLCertificatesRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSSLCertificates - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-ssl.html
//
type XPackSSLCertificates func(o ...func(*XPackSSLCertificatesRequest)) (*Response, error)

// XPackSSLCertificatesRequest configures the X PackSSL Certificates API request.
//
type XPackSSLCertificatesRequest struct {
	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackSSLCertificatesRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_xpack/ssl/certificates"))
	path.WriteString("/_xpack/ssl/certificates")

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
func (f XPackSSLCertificates) WithContext(v context.Context) func(*XPackSSLCertificatesRequest) {
	return func(r *XPackSSLCertificatesRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSSLCertificates) WithPretty() func(*XPackSSLCertificatesRequest) {
	return func(r *XPackSSLCertificatesRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSSLCertificates) WithHuman() func(*XPackSSLCertificatesRequest) {
	return func(r *XPackSSLCertificatesRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSSLCertificates) WithErrorTrace() func(*XPackSSLCertificatesRequest) {
	return func(r *XPackSSLCertificatesRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSSLCertificates) WithFilterPath(v ...string) func(*XPackSSLCertificatesRequest) {
	return func(r *XPackSSLCertificatesRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSSLCertificates) WithHeader(h map[string]string) func(*XPackSSLCertificatesRequest) {
	return func(r *XPackSSLCertificatesRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
