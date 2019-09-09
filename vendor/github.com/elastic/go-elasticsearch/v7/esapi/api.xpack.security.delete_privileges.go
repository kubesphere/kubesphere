// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newSecurityDeletePrivilegesFunc(t Transport) SecurityDeletePrivileges {
	return func(name string, application string, o ...func(*SecurityDeletePrivilegesRequest)) (*Response, error) {
		var r = SecurityDeletePrivilegesRequest{Name: name, Application: application}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SecurityDeletePrivileges - TODO
//
type SecurityDeletePrivileges func(name string, application string, o ...func(*SecurityDeletePrivilegesRequest)) (*Response, error)

// SecurityDeletePrivilegesRequest configures the Security Delete Privileges API request.
//
type SecurityDeletePrivilegesRequest struct {
	Application string
	Name        string

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
func (r SecurityDeletePrivilegesRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_security") + 1 + len("privilege") + 1 + len(r.Application) + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString("_security")
	path.WriteString("/")
	path.WriteString("privilege")
	path.WriteString("/")
	path.WriteString(r.Application)
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
func (f SecurityDeletePrivileges) WithContext(v context.Context) func(*SecurityDeletePrivilegesRequest) {
	return func(r *SecurityDeletePrivilegesRequest) {
		r.ctx = v
	}
}

// WithRefresh - if `true` (the default) then refresh the affected shards to make this operation visible to search, if `wait_for` then wait for a refresh to make this operation visible to search, if `false` then do nothing with refreshes..
//
func (f SecurityDeletePrivileges) WithRefresh(v string) func(*SecurityDeletePrivilegesRequest) {
	return func(r *SecurityDeletePrivilegesRequest) {
		r.Refresh = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SecurityDeletePrivileges) WithPretty() func(*SecurityDeletePrivilegesRequest) {
	return func(r *SecurityDeletePrivilegesRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SecurityDeletePrivileges) WithHuman() func(*SecurityDeletePrivilegesRequest) {
	return func(r *SecurityDeletePrivilegesRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SecurityDeletePrivileges) WithErrorTrace() func(*SecurityDeletePrivilegesRequest) {
	return func(r *SecurityDeletePrivilegesRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SecurityDeletePrivileges) WithFilterPath(v ...string) func(*SecurityDeletePrivilegesRequest) {
	return func(r *SecurityDeletePrivilegesRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SecurityDeletePrivileges) WithHeader(h map[string]string) func(*SecurityDeletePrivilegesRequest) {
	return func(r *SecurityDeletePrivilegesRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
