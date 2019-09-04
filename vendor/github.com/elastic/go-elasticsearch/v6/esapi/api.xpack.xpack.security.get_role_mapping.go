// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackSecurityGetRoleMappingFunc(t Transport) XPackSecurityGetRoleMapping {
	return func(o ...func(*XPackSecurityGetRoleMappingRequest)) (*Response, error) {
		var r = XPackSecurityGetRoleMappingRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSecurityGetRoleMapping - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-role-mapping.html
//
type XPackSecurityGetRoleMapping func(o ...func(*XPackSecurityGetRoleMappingRequest)) (*Response, error)

// XPackSecurityGetRoleMappingRequest configures the X Pack Security Get Role Mapping API request.
//
type XPackSecurityGetRoleMappingRequest struct {
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
func (r XPackSecurityGetRoleMappingRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("security") + 1 + len("role_mapping") + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("security")
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
func (f XPackSecurityGetRoleMapping) WithContext(v context.Context) func(*XPackSecurityGetRoleMappingRequest) {
	return func(r *XPackSecurityGetRoleMappingRequest) {
		r.ctx = v
	}
}

// WithName - role-mapping name.
//
func (f XPackSecurityGetRoleMapping) WithName(v string) func(*XPackSecurityGetRoleMappingRequest) {
	return func(r *XPackSecurityGetRoleMappingRequest) {
		r.Name = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSecurityGetRoleMapping) WithPretty() func(*XPackSecurityGetRoleMappingRequest) {
	return func(r *XPackSecurityGetRoleMappingRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSecurityGetRoleMapping) WithHuman() func(*XPackSecurityGetRoleMappingRequest) {
	return func(r *XPackSecurityGetRoleMappingRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSecurityGetRoleMapping) WithErrorTrace() func(*XPackSecurityGetRoleMappingRequest) {
	return func(r *XPackSecurityGetRoleMappingRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSecurityGetRoleMapping) WithFilterPath(v ...string) func(*XPackSecurityGetRoleMappingRequest) {
	return func(r *XPackSecurityGetRoleMappingRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSecurityGetRoleMapping) WithHeader(h map[string]string) func(*XPackSecurityGetRoleMappingRequest) {
	return func(r *XPackSecurityGetRoleMappingRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
