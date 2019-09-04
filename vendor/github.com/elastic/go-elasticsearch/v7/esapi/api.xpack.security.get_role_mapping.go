// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newSecurityGetRoleMappingFunc(t Transport) SecurityGetRoleMapping {
	return func(o ...func(*SecurityGetRoleMappingRequest)) (*Response, error) {
		var r = SecurityGetRoleMappingRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SecurityGetRoleMapping - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-role-mapping.html
//
type SecurityGetRoleMapping func(o ...func(*SecurityGetRoleMappingRequest)) (*Response, error)

// SecurityGetRoleMappingRequest configures the Security Get Role Mapping API request.
//
type SecurityGetRoleMappingRequest struct {
	Name string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SecurityGetRoleMappingRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_security") + 1 + len("role_mapping") + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString("_security")
	path.WriteString("/")
	path.WriteString("role_mapping")
	if r.Name != "" {
		path.WriteString("/")
		path.WriteString(r.Name)
	}

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
func (f SecurityGetRoleMapping) WithContext(v context.Context) func(*SecurityGetRoleMappingRequest) {
	return func(r *SecurityGetRoleMappingRequest) {
		r.ctx = v
	}
}

// WithName - role-mapping name.
//
func (f SecurityGetRoleMapping) WithName(v string) func(*SecurityGetRoleMappingRequest) {
	return func(r *SecurityGetRoleMappingRequest) {
		r.Name = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SecurityGetRoleMapping) WithPretty() func(*SecurityGetRoleMappingRequest) {
	return func(r *SecurityGetRoleMappingRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SecurityGetRoleMapping) WithHuman() func(*SecurityGetRoleMappingRequest) {
	return func(r *SecurityGetRoleMappingRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SecurityGetRoleMapping) WithErrorTrace() func(*SecurityGetRoleMappingRequest) {
	return func(r *SecurityGetRoleMappingRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SecurityGetRoleMapping) WithFilterPath(v ...string) func(*SecurityGetRoleMappingRequest) {
	return func(r *SecurityGetRoleMappingRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SecurityGetRoleMapping) WithHeader(h map[string]string) func(*SecurityGetRoleMappingRequest) {
	return func(r *SecurityGetRoleMappingRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
