// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newSecurityHasPrivilegesFunc(t Transport) SecurityHasPrivileges {
	return func(body io.Reader, o ...func(*SecurityHasPrivilegesRequest)) (*Response, error) {
		var r = SecurityHasPrivilegesRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SecurityHasPrivileges - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-has-privileges.html
//
type SecurityHasPrivileges func(body io.Reader, o ...func(*SecurityHasPrivilegesRequest)) (*Response, error)

// SecurityHasPrivilegesRequest configures the Security Has Privileges API request.
//
type SecurityHasPrivilegesRequest struct {
	Body io.Reader

	User string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SecurityHasPrivilegesRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_security") + 1 + len("user") + 1 + len(r.User) + 1 + len("_has_privileges"))
	path.WriteString("/")
	path.WriteString("_security")
	path.WriteString("/")
	path.WriteString("user")
	if r.User != "" {
		path.WriteString("/")
		path.WriteString(r.User)
	}
	path.WriteString("/")
	path.WriteString("_has_privileges")

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
func (f SecurityHasPrivileges) WithContext(v context.Context) func(*SecurityHasPrivilegesRequest) {
	return func(r *SecurityHasPrivilegesRequest) {
		r.ctx = v
	}
}

// WithUser - username.
//
func (f SecurityHasPrivileges) WithUser(v string) func(*SecurityHasPrivilegesRequest) {
	return func(r *SecurityHasPrivilegesRequest) {
		r.User = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SecurityHasPrivileges) WithPretty() func(*SecurityHasPrivilegesRequest) {
	return func(r *SecurityHasPrivilegesRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SecurityHasPrivileges) WithHuman() func(*SecurityHasPrivilegesRequest) {
	return func(r *SecurityHasPrivilegesRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SecurityHasPrivileges) WithErrorTrace() func(*SecurityHasPrivilegesRequest) {
	return func(r *SecurityHasPrivilegesRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SecurityHasPrivileges) WithFilterPath(v ...string) func(*SecurityHasPrivilegesRequest) {
	return func(r *SecurityHasPrivilegesRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SecurityHasPrivileges) WithHeader(h map[string]string) func(*SecurityHasPrivilegesRequest) {
	return func(r *SecurityHasPrivilegesRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
