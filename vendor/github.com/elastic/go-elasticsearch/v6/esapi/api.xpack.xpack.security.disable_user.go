// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackSecurityDisableUserFunc(t Transport) XPackSecurityDisableUser {
	return func(username string, o ...func(*XPackSecurityDisableUserRequest)) (*Response, error) {
		var r = XPackSecurityDisableUserRequest{Username: username}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSecurityDisableUser - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-disable-user.html
//
type XPackSecurityDisableUser func(username string, o ...func(*XPackSecurityDisableUserRequest)) (*Response, error)

// XPackSecurityDisableUserRequest configures the X Pack Security Disable User API request.
//
type XPackSecurityDisableUserRequest struct {
	Username string

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
func (r XPackSecurityDisableUserRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_xpack") + 1 + len("security") + 1 + len("user") + 1 + len(r.Username) + 1 + len("_disable"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("security")
	path.WriteString("/")
	path.WriteString("user")
	path.WriteString("/")
	path.WriteString(r.Username)
	path.WriteString("/")
	path.WriteString("_disable")

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
func (f XPackSecurityDisableUser) WithContext(v context.Context) func(*XPackSecurityDisableUserRequest) {
	return func(r *XPackSecurityDisableUserRequest) {
		r.ctx = v
	}
}

// WithRefresh - if `true` (the default) then refresh the affected shards to make this operation visible to search, if `wait_for` then wait for a refresh to make this operation visible to search, if `false` then do nothing with refreshes..
//
func (f XPackSecurityDisableUser) WithRefresh(v string) func(*XPackSecurityDisableUserRequest) {
	return func(r *XPackSecurityDisableUserRequest) {
		r.Refresh = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSecurityDisableUser) WithPretty() func(*XPackSecurityDisableUserRequest) {
	return func(r *XPackSecurityDisableUserRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSecurityDisableUser) WithHuman() func(*XPackSecurityDisableUserRequest) {
	return func(r *XPackSecurityDisableUserRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSecurityDisableUser) WithErrorTrace() func(*XPackSecurityDisableUserRequest) {
	return func(r *XPackSecurityDisableUserRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSecurityDisableUser) WithFilterPath(v ...string) func(*XPackSecurityDisableUserRequest) {
	return func(r *XPackSecurityDisableUserRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSecurityDisableUser) WithHeader(h map[string]string) func(*XPackSecurityDisableUserRequest) {
	return func(r *XPackSecurityDisableUserRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
