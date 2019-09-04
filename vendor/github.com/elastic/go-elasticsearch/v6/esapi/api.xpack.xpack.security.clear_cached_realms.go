// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackSecurityClearCachedRealmsFunc(t Transport) XPackSecurityClearCachedRealms {
	return func(realms []string, o ...func(*XPackSecurityClearCachedRealmsRequest)) (*Response, error) {
		var r = XPackSecurityClearCachedRealmsRequest{Realms: realms}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSecurityClearCachedRealms - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-clear-cache.html
//
type XPackSecurityClearCachedRealms func(realms []string, o ...func(*XPackSecurityClearCachedRealmsRequest)) (*Response, error)

// XPackSecurityClearCachedRealmsRequest configures the X Pack Security Clear Cached Realms API request.
//
type XPackSecurityClearCachedRealmsRequest struct {
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
func (r XPackSecurityClearCachedRealmsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_xpack") + 1 + len("security") + 1 + len("realm") + 1 + len(strings.Join(r.Realms, ",")) + 1 + len("_clear_cache"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("security")
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
func (f XPackSecurityClearCachedRealms) WithContext(v context.Context) func(*XPackSecurityClearCachedRealmsRequest) {
	return func(r *XPackSecurityClearCachedRealmsRequest) {
		r.ctx = v
	}
}

// WithUsernames - comma-separated list of usernames to clear from the cache.
//
func (f XPackSecurityClearCachedRealms) WithUsernames(v ...string) func(*XPackSecurityClearCachedRealmsRequest) {
	return func(r *XPackSecurityClearCachedRealmsRequest) {
		r.Usernames = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSecurityClearCachedRealms) WithPretty() func(*XPackSecurityClearCachedRealmsRequest) {
	return func(r *XPackSecurityClearCachedRealmsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSecurityClearCachedRealms) WithHuman() func(*XPackSecurityClearCachedRealmsRequest) {
	return func(r *XPackSecurityClearCachedRealmsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSecurityClearCachedRealms) WithErrorTrace() func(*XPackSecurityClearCachedRealmsRequest) {
	return func(r *XPackSecurityClearCachedRealmsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSecurityClearCachedRealms) WithFilterPath(v ...string) func(*XPackSecurityClearCachedRealmsRequest) {
	return func(r *XPackSecurityClearCachedRealmsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSecurityClearCachedRealms) WithHeader(h map[string]string) func(*XPackSecurityClearCachedRealmsRequest) {
	return func(r *XPackSecurityClearCachedRealmsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
