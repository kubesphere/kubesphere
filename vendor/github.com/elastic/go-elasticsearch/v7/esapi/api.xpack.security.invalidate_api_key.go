// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newSecurityInvalidateAPIKeyFunc(t Transport) SecurityInvalidateAPIKey {
	return func(body io.Reader, o ...func(*SecurityInvalidateAPIKeyRequest)) (*Response, error) {
		var r = SecurityInvalidateAPIKeyRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SecurityInvalidateAPIKey - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-invalidate-api-key.html
//
type SecurityInvalidateAPIKey func(body io.Reader, o ...func(*SecurityInvalidateAPIKeyRequest)) (*Response, error)

// SecurityInvalidateAPIKeyRequest configures the Security InvalidateAPI Key API request.
//
type SecurityInvalidateAPIKeyRequest struct {
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
func (r SecurityInvalidateAPIKeyRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(len("/_security/api_key"))
	path.WriteString("/_security/api_key")

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
func (f SecurityInvalidateAPIKey) WithContext(v context.Context) func(*SecurityInvalidateAPIKeyRequest) {
	return func(r *SecurityInvalidateAPIKeyRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SecurityInvalidateAPIKey) WithPretty() func(*SecurityInvalidateAPIKeyRequest) {
	return func(r *SecurityInvalidateAPIKeyRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SecurityInvalidateAPIKey) WithHuman() func(*SecurityInvalidateAPIKeyRequest) {
	return func(r *SecurityInvalidateAPIKeyRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SecurityInvalidateAPIKey) WithErrorTrace() func(*SecurityInvalidateAPIKeyRequest) {
	return func(r *SecurityInvalidateAPIKeyRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SecurityInvalidateAPIKey) WithFilterPath(v ...string) func(*SecurityInvalidateAPIKeyRequest) {
	return func(r *SecurityInvalidateAPIKeyRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SecurityInvalidateAPIKey) WithHeader(h map[string]string) func(*SecurityInvalidateAPIKeyRequest) {
	return func(r *SecurityInvalidateAPIKeyRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
