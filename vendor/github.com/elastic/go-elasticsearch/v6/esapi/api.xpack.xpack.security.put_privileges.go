// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newXPackSecurityPutPrivilegesFunc(t Transport) XPackSecurityPutPrivileges {
	return func(body io.Reader, o ...func(*XPackSecurityPutPrivilegesRequest)) (*Response, error) {
		var r = XPackSecurityPutPrivilegesRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackSecurityPutPrivileges - TODO
//
type XPackSecurityPutPrivileges func(body io.Reader, o ...func(*XPackSecurityPutPrivilegesRequest)) (*Response, error)

// XPackSecurityPutPrivilegesRequest configures the X Pack Security Put Privileges API request.
//
type XPackSecurityPutPrivilegesRequest struct {
	Body io.Reader

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
func (r XPackSecurityPutPrivilegesRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(len("/_xpack/security/privilege/"))
	path.WriteString("/_xpack/security/privilege/")

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
func (f XPackSecurityPutPrivileges) WithContext(v context.Context) func(*XPackSecurityPutPrivilegesRequest) {
	return func(r *XPackSecurityPutPrivilegesRequest) {
		r.ctx = v
	}
}

// WithRefresh - if `true` (the default) then refresh the affected shards to make this operation visible to search, if `wait_for` then wait for a refresh to make this operation visible to search, if `false` then do nothing with refreshes..
//
func (f XPackSecurityPutPrivileges) WithRefresh(v string) func(*XPackSecurityPutPrivilegesRequest) {
	return func(r *XPackSecurityPutPrivilegesRequest) {
		r.Refresh = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackSecurityPutPrivileges) WithPretty() func(*XPackSecurityPutPrivilegesRequest) {
	return func(r *XPackSecurityPutPrivilegesRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackSecurityPutPrivileges) WithHuman() func(*XPackSecurityPutPrivilegesRequest) {
	return func(r *XPackSecurityPutPrivilegesRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackSecurityPutPrivileges) WithErrorTrace() func(*XPackSecurityPutPrivilegesRequest) {
	return func(r *XPackSecurityPutPrivilegesRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackSecurityPutPrivileges) WithFilterPath(v ...string) func(*XPackSecurityPutPrivilegesRequest) {
	return func(r *XPackSecurityPutPrivilegesRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackSecurityPutPrivileges) WithHeader(h map[string]string) func(*XPackSecurityPutPrivilegesRequest) {
	return func(r *XPackSecurityPutPrivilegesRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
