// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newSecurityCreateAPIKeyFunc(t Transport) SecurityCreateAPIKey {
	return func(body io.Reader, o ...func(*SecurityCreateAPIKeyRequest)) (*Response, error) {
		var r = SecurityCreateAPIKeyRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SecurityCreateAPIKey - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-create-api-key.html
//
type SecurityCreateAPIKey func(body io.Reader, o ...func(*SecurityCreateAPIKeyRequest)) (*Response, error)

// SecurityCreateAPIKeyRequest configures the Security CreateAPI Key API request.
//
type SecurityCreateAPIKeyRequest struct {
	Body io.Reader

	Refresh string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SecurityCreateAPIKeyRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(len("/_security/api_key"))
	path.WriteString("/_security/api_key")

	params = make(map[string]string)

	if r.Refresh != "" {
		params["refresh"] = r.Refresh
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
func (f SecurityCreateAPIKey) WithContext(v context.Context) func(*SecurityCreateAPIKeyRequest) {
	return func(r *SecurityCreateAPIKeyRequest) {
		r.ctx = v
	}
}

// WithRefresh - if `true` (the default) then refresh the affected shards to make this operation visible to search, if `wait_for` then wait for a refresh to make this operation visible to search, if `false` then do nothing with refreshes..
//
func (f SecurityCreateAPIKey) WithRefresh(v string) func(*SecurityCreateAPIKeyRequest) {
	return func(r *SecurityCreateAPIKeyRequest) {
		r.Refresh = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SecurityCreateAPIKey) WithPretty() func(*SecurityCreateAPIKeyRequest) {
	return func(r *SecurityCreateAPIKeyRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SecurityCreateAPIKey) WithHuman() func(*SecurityCreateAPIKeyRequest) {
	return func(r *SecurityCreateAPIKeyRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SecurityCreateAPIKey) WithErrorTrace() func(*SecurityCreateAPIKeyRequest) {
	return func(r *SecurityCreateAPIKeyRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SecurityCreateAPIKey) WithFilterPath(v ...string) func(*SecurityCreateAPIKeyRequest) {
	return func(r *SecurityCreateAPIKeyRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SecurityCreateAPIKey) WithHeader(h map[string]string) func(*SecurityCreateAPIKeyRequest) {
	return func(r *SecurityCreateAPIKeyRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
