// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackSecurityClearCachedRolesFunc(t Transport) XPackSecurityClearCachedRoles {
	return func(name []string, o ...func(*XPackSecurityClearCachedRolesRequest)) (*Response, error) {
		var r = XPackSecurityClearCachedRolesRequest{Name: name}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSecurityClearCachedRoles - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-clear-role-cache.html
//
type XPackSecurityClearCachedRoles func(name []string, o ...func(*XPackSecurityClearCachedRolesRequest)) (*Response, error)

// XPackSecurityClearCachedRolesRequest configures the X Pack Security Clear Cached Roles API request.
//
type XPackSecurityClearCachedRolesRequest struct {
	Name []string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackSecurityClearCachedRolesRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_xpack") + 1 + len("security") + 1 + len("role") + 1 + len(strings.Join(r.Name, ",")) + 1 + len("_clear_cache"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("security")
	path.WriteString("/")
	path.WriteString("role")
	path.WriteString("/")
	path.WriteString(strings.Join(r.Name, ","))
	path.WriteString("/")
	path.WriteString("_clear_cache")

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
func (f XPackSecurityClearCachedRoles) WithContext(v context.Context) func(*XPackSecurityClearCachedRolesRequest) {
	return func(r *XPackSecurityClearCachedRolesRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSecurityClearCachedRoles) WithPretty() func(*XPackSecurityClearCachedRolesRequest) {
	return func(r *XPackSecurityClearCachedRolesRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSecurityClearCachedRoles) WithHuman() func(*XPackSecurityClearCachedRolesRequest) {
	return func(r *XPackSecurityClearCachedRolesRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSecurityClearCachedRoles) WithErrorTrace() func(*XPackSecurityClearCachedRolesRequest) {
	return func(r *XPackSecurityClearCachedRolesRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSecurityClearCachedRoles) WithFilterPath(v ...string) func(*XPackSecurityClearCachedRolesRequest) {
	return func(r *XPackSecurityClearCachedRolesRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSecurityClearCachedRoles) WithHeader(h map[string]string) func(*XPackSecurityClearCachedRolesRequest) {
	return func(r *XPackSecurityClearCachedRolesRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
