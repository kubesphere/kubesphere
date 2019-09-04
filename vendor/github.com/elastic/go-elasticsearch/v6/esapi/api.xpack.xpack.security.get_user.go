// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackSecurityGetUserFunc(t Transport) XPackSecurityGetUser {
	return func(o ...func(*XPackSecurityGetUserRequest)) (*Response, error) {
		var r = XPackSecurityGetUserRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSecurityGetUser - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-user.html
//
type XPackSecurityGetUser func(o ...func(*XPackSecurityGetUserRequest)) (*Response, error)

// XPackSecurityGetUserRequest configures the X Pack Security Get User API request.
//
type XPackSecurityGetUserRequest struct {
	Username []string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackSecurityGetUserRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("security") + 1 + len("user") + 1 + len(strings.Join(r.Username, ",")))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("security")
	path.WriteString("/")
	path.WriteString("user")
	if len(r.Username) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Username, ","))
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
func (f XPackSecurityGetUser) WithContext(v context.Context) func(*XPackSecurityGetUserRequest) {
	return func(r *XPackSecurityGetUserRequest) {
		r.ctx = v
	}
}

// WithUsername - a list of usernames.
//
func (f XPackSecurityGetUser) WithUsername(v ...string) func(*XPackSecurityGetUserRequest) {
	return func(r *XPackSecurityGetUserRequest) {
		r.Username = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSecurityGetUser) WithPretty() func(*XPackSecurityGetUserRequest) {
	return func(r *XPackSecurityGetUserRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSecurityGetUser) WithHuman() func(*XPackSecurityGetUserRequest) {
	return func(r *XPackSecurityGetUserRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSecurityGetUser) WithErrorTrace() func(*XPackSecurityGetUserRequest) {
	return func(r *XPackSecurityGetUserRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSecurityGetUser) WithFilterPath(v ...string) func(*XPackSecurityGetUserRequest) {
	return func(r *XPackSecurityGetUserRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSecurityGetUser) WithHeader(h map[string]string) func(*XPackSecurityGetUserRequest) {
	return func(r *XPackSecurityGetUserRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
