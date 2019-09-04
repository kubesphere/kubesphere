// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newXPackUsageFunc(t Transport) XPackUsage {
	return func(o ...func(*XPackUsageRequest)) (*Response, error) {
		var r = XPackUsageRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackUsage - Retrieve information about xpack features usage
//
type XPackUsage func(o ...func(*XPackUsageRequest)) (*Response, error)

// XPackUsageRequest configures the X Pack Usage API request.
//
type XPackUsageRequest struct {
	MasterTimeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackUsageRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_xpack/usage"))
	path.WriteString("/_xpack/usage")

	params = make(map[string]string)

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
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
func (f XPackUsage) WithContext(v context.Context) func(*XPackUsageRequest) {
	return func(r *XPackUsageRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - specify timeout for watch write operation.
//
func (f XPackUsage) WithMasterTimeout(v time.Duration) func(*XPackUsageRequest) {
	return func(r *XPackUsageRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackUsage) WithPretty() func(*XPackUsageRequest) {
	return func(r *XPackUsageRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackUsage) WithHuman() func(*XPackUsageRequest) {
	return func(r *XPackUsageRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackUsage) WithErrorTrace() func(*XPackUsageRequest) {
	return func(r *XPackUsageRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackUsage) WithFilterPath(v ...string) func(*XPackUsageRequest) {
	return func(r *XPackUsageRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackUsage) WithHeader(h map[string]string) func(*XPackUsageRequest) {
	return func(r *XPackUsageRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
