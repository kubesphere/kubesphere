// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newSecurityDeleteRoleMappingFunc(t Transport) SecurityDeleteRoleMapping {
	return func(name string, o ...func(*SecurityDeleteRoleMappingRequest)) (*Response, error) {
		var r = SecurityDeleteRoleMappingRequest{Name: name}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SecurityDeleteRoleMapping - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-delete-role-mapping.html
//
type SecurityDeleteRoleMapping func(name string, o ...func(*SecurityDeleteRoleMappingRequest)) (*Response, error)

// SecurityDeleteRoleMappingRequest configures the Security Delete Role Mapping API request.
//
type SecurityDeleteRoleMappingRequest struct {
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
func (r SecurityDeleteRoleMappingRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

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
func (f SecurityDeleteRoleMapping) WithContext(v context.Context) func(*SecurityDeleteRoleMappingRequest) {
	return func(r *SecurityDeleteRoleMappingRequest) {
		r.ctx = v
	}
}

// WithRefresh - if `true` (the default) then refresh the affected shards to make this operation visible to search, if `wait_for` then wait for a refresh to make this operation visible to search, if `false` then do nothing with refreshes..
//
func (f SecurityDeleteRoleMapping) WithRefresh(v string) func(*SecurityDeleteRoleMappingRequest) {
	return func(r *SecurityDeleteRoleMappingRequest) {
		r.Refresh = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SecurityDeleteRoleMapping) WithPretty() func(*SecurityDeleteRoleMappingRequest) {
	return func(r *SecurityDeleteRoleMappingRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SecurityDeleteRoleMapping) WithHuman() func(*SecurityDeleteRoleMappingRequest) {
	return func(r *SecurityDeleteRoleMappingRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SecurityDeleteRoleMapping) WithErrorTrace() func(*SecurityDeleteRoleMappingRequest) {
	return func(r *SecurityDeleteRoleMappingRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SecurityDeleteRoleMapping) WithFilterPath(v ...string) func(*SecurityDeleteRoleMappingRequest) {
	return func(r *SecurityDeleteRoleMappingRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SecurityDeleteRoleMapping) WithHeader(h map[string]string) func(*SecurityDeleteRoleMappingRequest) {
	return func(r *SecurityDeleteRoleMappingRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
