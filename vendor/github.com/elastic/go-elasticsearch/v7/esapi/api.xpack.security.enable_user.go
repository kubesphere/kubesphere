// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newSecurityEnableUserFunc(t Transport) SecurityEnableUser {
	return func(username string, o ...func(*SecurityEnableUserRequest)) (*Response, error) {
		var r = SecurityEnableUserRequest{Username: username}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SecurityEnableUser - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-enable-user.html
//
type SecurityEnableUser func(username string, o ...func(*SecurityEnableUserRequest)) (*Response, error)

// SecurityEnableUserRequest configures the Security Enable User API request.
//
type SecurityEnableUserRequest struct {
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
func (r SecurityEnableUserRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_security") + 1 + len("user") + 1 + len(r.Username) + 1 + len("_enable"))
	path.WriteString("/")
	path.WriteString("_security")
	path.WriteString("/")
	path.WriteString("user")
	path.WriteString("/")
	path.WriteString(r.Username)
	path.WriteString("/")
	path.WriteString("_enable")

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
func (f SecurityEnableUser) WithContext(v context.Context) func(*SecurityEnableUserRequest) {
	return func(r *SecurityEnableUserRequest) {
		r.ctx = v
	}
}

// WithRefresh - if `true` (the default) then refresh the affected shards to make this operation visible to search, if `wait_for` then wait for a refresh to make this operation visible to search, if `false` then do nothing with refreshes..
//
func (f SecurityEnableUser) WithRefresh(v string) func(*SecurityEnableUserRequest) {
	return func(r *SecurityEnableUserRequest) {
		r.Refresh = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SecurityEnableUser) WithPretty() func(*SecurityEnableUserRequest) {
	return func(r *SecurityEnableUserRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SecurityEnableUser) WithHuman() func(*SecurityEnableUserRequest) {
	return func(r *SecurityEnableUserRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SecurityEnableUser) WithErrorTrace() func(*SecurityEnableUserRequest) {
	return func(r *SecurityEnableUserRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SecurityEnableUser) WithFilterPath(v ...string) func(*SecurityEnableUserRequest) {
	return func(r *SecurityEnableUserRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SecurityEnableUser) WithHeader(h map[string]string) func(*SecurityEnableUserRequest) {
	return func(r *SecurityEnableUserRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
