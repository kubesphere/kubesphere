// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newXPackSecurityGetTokenFunc(t Transport) XPackSecurityGetToken {
	return func(body io.Reader, o ...func(*XPackSecurityGetTokenRequest)) (*Response, error) {
		var r = XPackSecurityGetTokenRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSecurityGetToken - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-token.html
//
type XPackSecurityGetToken func(body io.Reader, o ...func(*XPackSecurityGetTokenRequest)) (*Response, error)

// XPackSecurityGetTokenRequest configures the X Pack Security Get Token API request.
//
type XPackSecurityGetTokenRequest struct {
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
func (r XPackSecurityGetTokenRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(len("/_xpack/security/oauth2/token"))
	path.WriteString("/_xpack/security/oauth2/token")

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
func (f XPackSecurityGetToken) WithContext(v context.Context) func(*XPackSecurityGetTokenRequest) {
	return func(r *XPackSecurityGetTokenRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSecurityGetToken) WithPretty() func(*XPackSecurityGetTokenRequest) {
	return func(r *XPackSecurityGetTokenRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSecurityGetToken) WithHuman() func(*XPackSecurityGetTokenRequest) {
	return func(r *XPackSecurityGetTokenRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSecurityGetToken) WithErrorTrace() func(*XPackSecurityGetTokenRequest) {
	return func(r *XPackSecurityGetTokenRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSecurityGetToken) WithFilterPath(v ...string) func(*XPackSecurityGetTokenRequest) {
	return func(r *XPackSecurityGetTokenRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSecurityGetToken) WithHeader(h map[string]string) func(*XPackSecurityGetTokenRequest) {
	return func(r *XPackSecurityGetTokenRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
