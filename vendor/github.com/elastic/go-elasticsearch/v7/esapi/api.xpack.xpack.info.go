// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackInfoFunc(t Transport) XPackInfo {
	return func(o ...func(*XPackInfoRequest)) (*Response, error) {
		var r = XPackInfoRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackInfo - https://www.elastic.co/guide/en/elasticsearch/reference/current/info-api.html
//
type XPackInfo func(o ...func(*XPackInfoRequest)) (*Response, error)

// XPackInfoRequest configures the X Pack Info API request.
//
type XPackInfoRequest struct {
	Categories []string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackInfoRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_xpack"))
	path.WriteString("/_xpack")

	params = make(map[string]string)

	if len(r.Categories) > 0 {
		params["categories"] = strings.Join(r.Categories, ",")
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
func (f XPackInfo) WithContext(v context.Context) func(*XPackInfoRequest) {
	return func(r *XPackInfoRequest) {
		r.ctx = v
	}
}

// WithCategories - comma-separated list of info categories. can be any of: build, license, features.
//
func (f XPackInfo) WithCategories(v ...string) func(*XPackInfoRequest) {
	return func(r *XPackInfoRequest) {
		r.Categories = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackInfo) WithPretty() func(*XPackInfoRequest) {
	return func(r *XPackInfoRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackInfo) WithHuman() func(*XPackInfoRequest) {
	return func(r *XPackInfoRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackInfo) WithErrorTrace() func(*XPackInfoRequest) {
	return func(r *XPackInfoRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackInfo) WithFilterPath(v ...string) func(*XPackInfoRequest) {
	return func(r *XPackInfoRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackInfo) WithHeader(h map[string]string) func(*XPackInfoRequest) {
	return func(r *XPackInfoRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
