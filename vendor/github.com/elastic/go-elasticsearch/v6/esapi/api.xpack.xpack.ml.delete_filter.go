// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackMLDeleteFilterFunc(t Transport) XPackMLDeleteFilter {
	return func(filter_id string, o ...func(*XPackMLDeleteFilterRequest)) (*Response, error) {
		var r = XPackMLDeleteFilterRequest{FilterID: filter_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLDeleteFilter -
//
type XPackMLDeleteFilter func(filter_id string, o ...func(*XPackMLDeleteFilterRequest)) (*Response, error)

// XPackMLDeleteFilterRequest configures the X PackML Delete Filter API request.
//
type XPackMLDeleteFilterRequest struct {
	FilterID string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLDeleteFilterRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("filters") + 1 + len(r.FilterID))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("filters")
	path.WriteString("/")
	path.WriteString(r.FilterID)

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
func (f XPackMLDeleteFilter) WithContext(v context.Context) func(*XPackMLDeleteFilterRequest) {
	return func(r *XPackMLDeleteFilterRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLDeleteFilter) WithPretty() func(*XPackMLDeleteFilterRequest) {
	return func(r *XPackMLDeleteFilterRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLDeleteFilter) WithHuman() func(*XPackMLDeleteFilterRequest) {
	return func(r *XPackMLDeleteFilterRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLDeleteFilter) WithErrorTrace() func(*XPackMLDeleteFilterRequest) {
	return func(r *XPackMLDeleteFilterRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLDeleteFilter) WithFilterPath(v ...string) func(*XPackMLDeleteFilterRequest) {
	return func(r *XPackMLDeleteFilterRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLDeleteFilter) WithHeader(h map[string]string) func(*XPackMLDeleteFilterRequest) {
	return func(r *XPackMLDeleteFilterRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
