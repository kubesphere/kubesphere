// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newXPackSecurityHasPrivilegesFunc(t Transport) XPackSecurityHasPrivileges {
	return func(body io.Reader, o ...func(*XPackSecurityHasPrivilegesRequest)) (*Response, error) {
		var r = XPackSecurityHasPrivilegesRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSecurityHasPrivileges - https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-has-privileges.html
//
type XPackSecurityHasPrivileges func(body io.Reader, o ...func(*XPackSecurityHasPrivilegesRequest)) (*Response, error)

// XPackSecurityHasPrivilegesRequest configures the X Pack Security Has Privileges API request.
//
type XPackSecurityHasPrivilegesRequest struct {
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
func (r XPackSecurityHasPrivilegesRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("security") + 1 + len("user") + 1 + len(r.User) + 1 + len("_has_privileges"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("security")
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
func (f XPackSecurityHasPrivileges) WithContext(v context.Context) func(*XPackSecurityHasPrivilegesRequest) {
	return func(r *XPackSecurityHasPrivilegesRequest) {
		r.ctx = v
	}
}

// WithUser - username.
//
func (f XPackSecurityHasPrivileges) WithUser(v string) func(*XPackSecurityHasPrivilegesRequest) {
	return func(r *XPackSecurityHasPrivilegesRequest) {
		r.User = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSecurityHasPrivileges) WithPretty() func(*XPackSecurityHasPrivilegesRequest) {
	return func(r *XPackSecurityHasPrivilegesRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSecurityHasPrivileges) WithHuman() func(*XPackSecurityHasPrivilegesRequest) {
	return func(r *XPackSecurityHasPrivilegesRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSecurityHasPrivileges) WithErrorTrace() func(*XPackSecurityHasPrivilegesRequest) {
	return func(r *XPackSecurityHasPrivilegesRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSecurityHasPrivileges) WithFilterPath(v ...string) func(*XPackSecurityHasPrivilegesRequest) {
	return func(r *XPackSecurityHasPrivilegesRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSecurityHasPrivileges) WithHeader(h map[string]string) func(*XPackSecurityHasPrivilegesRequest) {
	return func(r *XPackSecurityHasPrivilegesRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
