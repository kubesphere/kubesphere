// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newSecurityPutRoleMappingFunc(t Transport) SecurityPutRoleMapping {
	return func(name string, body io.Reader, o ...func(*SecurityPutRoleMappingRequest)) (*Response, error) {
		var r = SecurityPutRoleMappingRequest{Name: name, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SecurityPutRoleMapping - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role-mapping.html
//
type SecurityPutRoleMapping func(name string, body io.Reader, o ...func(*SecurityPutRoleMappingRequest)) (*Response, error)

// SecurityPutRoleMappingRequest configures the Security Put Role Mapping API request.
//
type SecurityPutRoleMappingRequest struct {
	Body io.Reader

	Name string

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
func (r SecurityPutRoleMappingRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_security") + 1 + len("role_mapping") + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString("_security")
	path.WriteString("/")
	path.WriteString("role_mapping")
	path.WriteString("/")
	path.WriteString(r.Name)

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
func (f SecurityPutRoleMapping) WithContext(v context.Context) func(*SecurityPutRoleMappingRequest) {
	return func(r *SecurityPutRoleMappingRequest) {
		r.ctx = v
	}
}

// WithRefresh - if `true` (the default) then refresh the affected shards to make this operation visible to search, if `wait_for` then wait for a refresh to make this operation visible to search, if `false` then do nothing with refreshes..
//
func (f SecurityPutRoleMapping) WithRefresh(v string) func(*SecurityPutRoleMappingRequest) {
	return func(r *SecurityPutRoleMappingRequest) {
		r.Refresh = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SecurityPutRoleMapping) WithPretty() func(*SecurityPutRoleMappingRequest) {
	return func(r *SecurityPutRoleMappingRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SecurityPutRoleMapping) WithHuman() func(*SecurityPutRoleMappingRequest) {
	return func(r *SecurityPutRoleMappingRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SecurityPutRoleMapping) WithErrorTrace() func(*SecurityPutRoleMappingRequest) {
	return func(r *SecurityPutRoleMappingRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SecurityPutRoleMapping) WithFilterPath(v ...string) func(*SecurityPutRoleMappingRequest) {
	return func(r *SecurityPutRoleMappingRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SecurityPutRoleMapping) WithHeader(h map[string]string) func(*SecurityPutRoleMappingRequest) {
	return func(r *SecurityPutRoleMappingRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
