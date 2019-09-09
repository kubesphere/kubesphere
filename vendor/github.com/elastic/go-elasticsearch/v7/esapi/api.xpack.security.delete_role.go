// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newSecurityDeleteRoleFunc(t Transport) SecurityDeleteRole {
	return func(name string, o ...func(*SecurityDeleteRoleRequest)) (*Response, error) {
		var r = SecurityDeleteRoleRequest{Name: name}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SecurityDeleteRole - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-delete-role.html
//
type SecurityDeleteRole func(name string, o ...func(*SecurityDeleteRoleRequest)) (*Response, error)

// SecurityDeleteRoleRequest configures the Security Delete Role API request.
//
type SecurityDeleteRoleRequest struct {
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
func (r SecurityDeleteRoleRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_security") + 1 + len("role") + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString("_security")
	path.WriteString("/")
	path.WriteString("role")
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
func (f SecurityDeleteRole) WithContext(v context.Context) func(*SecurityDeleteRoleRequest) {
	return func(r *SecurityDeleteRoleRequest) {
		r.ctx = v
	}
}

// WithRefresh - if `true` (the default) then refresh the affected shards to make this operation visible to search, if `wait_for` then wait for a refresh to make this operation visible to search, if `false` then do nothing with refreshes..
//
func (f SecurityDeleteRole) WithRefresh(v string) func(*SecurityDeleteRoleRequest) {
	return func(r *SecurityDeleteRoleRequest) {
		r.Refresh = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SecurityDeleteRole) WithPretty() func(*SecurityDeleteRoleRequest) {
	return func(r *SecurityDeleteRoleRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SecurityDeleteRole) WithHuman() func(*SecurityDeleteRoleRequest) {
	return func(r *SecurityDeleteRoleRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SecurityDeleteRole) WithErrorTrace() func(*SecurityDeleteRoleRequest) {
	return func(r *SecurityDeleteRoleRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SecurityDeleteRole) WithFilterPath(v ...string) func(*SecurityDeleteRoleRequest) {
	return func(r *SecurityDeleteRoleRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SecurityDeleteRole) WithHeader(h map[string]string) func(*SecurityDeleteRoleRequest) {
	return func(r *SecurityDeleteRoleRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
