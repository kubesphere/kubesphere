// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newSecurityClearCachedRealmsFunc(t Transport) SecurityClearCachedRealms {
	return func(realms []string, o ...func(*SecurityClearCachedRealmsRequest)) (*Response, error) {
		var r = SecurityClearCachedRealmsRequest{Realms: realms}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SecurityClearCachedRealms - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-clear-cache.html
//
type SecurityClearCachedRealms func(realms []string, o ...func(*SecurityClearCachedRealmsRequest)) (*Response, error)

// SecurityClearCachedRealmsRequest configures the Security Clear Cached Realms API request.
//
type SecurityClearCachedRealmsRequest struct {
	Realms []string

	Usernames []string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SecurityClearCachedRealmsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_security") + 1 + len("realm") + 1 + len(strings.Join(r.Realms, ",")) + 1 + len("_clear_cache"))
	path.WriteString("/")
	path.WriteString("_security")
	path.WriteString("/")
	path.WriteString("realm")
	path.WriteString("/")
	path.WriteString(strings.Join(r.Realms, ","))
	path.WriteString("/")
	path.WriteString("_clear_cache")

	params = make(map[string]string)

	if len(r.Usernames) > 0 {
		params["usernames"] = strings.Join(r.Usernames, ",")
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
func (f SecurityClearCachedRealms) WithContext(v context.Context) func(*SecurityClearCachedRealmsRequest) {
	return func(r *SecurityClearCachedRealmsRequest) {
		r.ctx = v
	}
}

// WithUsernames - comma-separated list of usernames to clear from the cache.
//
func (f SecurityClearCachedRealms) WithUsernames(v ...string) func(*SecurityClearCachedRealmsRequest) {
	return func(r *SecurityClearCachedRealmsRequest) {
		r.Usernames = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SecurityClearCachedRealms) WithPretty() func(*SecurityClearCachedRealmsRequest) {
	return func(r *SecurityClearCachedRealmsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SecurityClearCachedRealms) WithHuman() func(*SecurityClearCachedRealmsRequest) {
	return func(r *SecurityClearCachedRealmsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SecurityClearCachedRealms) WithErrorTrace() func(*SecurityClearCachedRealmsRequest) {
	return func(r *SecurityClearCachedRealmsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SecurityClearCachedRealms) WithFilterPath(v ...string) func(*SecurityClearCachedRealmsRequest) {
	return func(r *SecurityClearCachedRealmsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SecurityClearCachedRealms) WithHeader(h map[string]string) func(*SecurityClearCachedRealmsRequest) {
	return func(r *SecurityClearCachedRealmsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
