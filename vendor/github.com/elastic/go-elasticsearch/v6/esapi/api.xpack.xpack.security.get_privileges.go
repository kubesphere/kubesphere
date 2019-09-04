// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackSecurityGetPrivilegesFunc(t Transport) XPackSecurityGetPrivileges {
	return func(o ...func(*XPackSecurityGetPrivilegesRequest)) (*Response, error) {
		var r = XPackSecurityGetPrivilegesRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSecurityGetPrivileges - TODO
//
type XPackSecurityGetPrivileges func(o ...func(*XPackSecurityGetPrivilegesRequest)) (*Response, error)

// XPackSecurityGetPrivilegesRequest configures the X Pack Security Get Privileges API request.
//
type XPackSecurityGetPrivilegesRequest struct {
	Application string
	Name        string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackSecurityGetPrivilegesRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("security") + 1 + len("privilege") + 1 + len(r.Application) + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("security")
	path.WriteString("/")
	path.WriteString("privilege")
	if r.Application != "" {
		path.WriteString("/")
		path.WriteString(r.Application)
	}
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
func (f XPackSecurityGetPrivileges) WithContext(v context.Context) func(*XPackSecurityGetPrivilegesRequest) {
	return func(r *XPackSecurityGetPrivilegesRequest) {
		r.ctx = v
	}
}

// WithApplication - application name.
//
func (f XPackSecurityGetPrivileges) WithApplication(v string) func(*XPackSecurityGetPrivilegesRequest) {
	return func(r *XPackSecurityGetPrivilegesRequest) {
		r.Application = v
	}
}

// WithName - privilege name.
//
func (f XPackSecurityGetPrivileges) WithName(v string) func(*XPackSecurityGetPrivilegesRequest) {
	return func(r *XPackSecurityGetPrivilegesRequest) {
		r.Name = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSecurityGetPrivileges) WithPretty() func(*XPackSecurityGetPrivilegesRequest) {
	return func(r *XPackSecurityGetPrivilegesRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSecurityGetPrivileges) WithHuman() func(*XPackSecurityGetPrivilegesRequest) {
	return func(r *XPackSecurityGetPrivilegesRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSecurityGetPrivileges) WithErrorTrace() func(*XPackSecurityGetPrivilegesRequest) {
	return func(r *XPackSecurityGetPrivilegesRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSecurityGetPrivileges) WithFilterPath(v ...string) func(*XPackSecurityGetPrivilegesRequest) {
	return func(r *XPackSecurityGetPrivilegesRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSecurityGetPrivileges) WithHeader(h map[string]string) func(*XPackSecurityGetPrivilegesRequest) {
	return func(r *XPackSecurityGetPrivilegesRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
