// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newXPackSecurityPutRoleMappingFunc(t Transport) XPackSecurityPutRoleMapping {
	return func(name string, body io.Reader, o ...func(*XPackSecurityPutRoleMappingRequest)) (*Response, error) {
		var r = XPackSecurityPutRoleMappingRequest{Name: name, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSecurityPutRoleMapping - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role-mapping.html
//
type XPackSecurityPutRoleMapping func(name string, body io.Reader, o ...func(*XPackSecurityPutRoleMappingRequest)) (*Response, error)

// XPackSecurityPutRoleMappingRequest configures the X Pack Security Put Role Mapping API request.
//
type XPackSecurityPutRoleMappingRequest struct {
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
func (r XPackSecurityPutRoleMappingRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_xpack") + 1 + len("security") + 1 + len("role_mapping") + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("security")
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
func (f XPackSecurityPutRoleMapping) WithContext(v context.Context) func(*XPackSecurityPutRoleMappingRequest) {
	return func(r *XPackSecurityPutRoleMappingRequest) {
		r.ctx = v
	}
}

// WithRefresh - if `true` (the default) then refresh the affected shards to make this operation visible to search, if `wait_for` then wait for a refresh to make this operation visible to search, if `false` then do nothing with refreshes..
//
func (f XPackSecurityPutRoleMapping) WithRefresh(v string) func(*XPackSecurityPutRoleMappingRequest) {
	return func(r *XPackSecurityPutRoleMappingRequest) {
		r.Refresh = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSecurityPutRoleMapping) WithPretty() func(*XPackSecurityPutRoleMappingRequest) {
	return func(r *XPackSecurityPutRoleMappingRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSecurityPutRoleMapping) WithHuman() func(*XPackSecurityPutRoleMappingRequest) {
	return func(r *XPackSecurityPutRoleMappingRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSecurityPutRoleMapping) WithErrorTrace() func(*XPackSecurityPutRoleMappingRequest) {
	return func(r *XPackSecurityPutRoleMappingRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSecurityPutRoleMapping) WithFilterPath(v ...string) func(*XPackSecurityPutRoleMappingRequest) {
	return func(r *XPackSecurityPutRoleMappingRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSecurityPutRoleMapping) WithHeader(h map[string]string) func(*XPackSecurityPutRoleMappingRequest) {
	return func(r *XPackSecurityPutRoleMappingRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
