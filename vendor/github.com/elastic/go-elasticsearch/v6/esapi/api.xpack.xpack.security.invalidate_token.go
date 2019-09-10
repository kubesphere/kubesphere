// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newXPackSecurityInvalidateTokenFunc(t Transport) XPackSecurityInvalidateToken {
	return func(body io.Reader, o ...func(*XPackSecurityInvalidateTokenRequest)) (*Response, error) {
		var r = XPackSecurityInvalidateTokenRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSecurityInvalidateToken - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-invalidate-token.html
//
type XPackSecurityInvalidateToken func(body io.Reader, o ...func(*XPackSecurityInvalidateTokenRequest)) (*Response, error)

// XPackSecurityInvalidateTokenRequest configures the X Pack Security Invalidate Token API request.
//
type XPackSecurityInvalidateTokenRequest struct {
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
func (r XPackSecurityInvalidateTokenRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

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
func (f XPackSecurityInvalidateToken) WithContext(v context.Context) func(*XPackSecurityInvalidateTokenRequest) {
	return func(r *XPackSecurityInvalidateTokenRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSecurityInvalidateToken) WithPretty() func(*XPackSecurityInvalidateTokenRequest) {
	return func(r *XPackSecurityInvalidateTokenRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSecurityInvalidateToken) WithHuman() func(*XPackSecurityInvalidateTokenRequest) {
	return func(r *XPackSecurityInvalidateTokenRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSecurityInvalidateToken) WithErrorTrace() func(*XPackSecurityInvalidateTokenRequest) {
	return func(r *XPackSecurityInvalidateTokenRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSecurityInvalidateToken) WithFilterPath(v ...string) func(*XPackSecurityInvalidateTokenRequest) {
	return func(r *XPackSecurityInvalidateTokenRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSecurityInvalidateToken) WithHeader(h map[string]string) func(*XPackSecurityInvalidateTokenRequest) {
	return func(r *XPackSecurityInvalidateTokenRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
