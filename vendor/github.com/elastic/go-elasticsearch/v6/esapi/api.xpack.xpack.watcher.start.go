// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackWatcherStartFunc(t Transport) XPackWatcherStart {
	return func(o ...func(*XPackWatcherStartRequest)) (*Response, error) {
		var r = XPackWatcherStartRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackWatcherStart - http://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-start.html
//
type XPackWatcherStart func(o ...func(*XPackWatcherStartRequest)) (*Response, error)

// XPackWatcherStartRequest configures the X Pack Watcher Start API request.
//
type XPackWatcherStartRequest struct {
	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackWatcherStartRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(len("/_xpack/watcher/_start"))
	path.WriteString("/_xpack/watcher/_start")

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
func (f XPackWatcherStart) WithContext(v context.Context) func(*XPackWatcherStartRequest) {
	return func(r *XPackWatcherStartRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackWatcherStart) WithPretty() func(*XPackWatcherStartRequest) {
	return func(r *XPackWatcherStartRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackWatcherStart) WithHuman() func(*XPackWatcherStartRequest) {
	return func(r *XPackWatcherStartRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackWatcherStart) WithErrorTrace() func(*XPackWatcherStartRequest) {
	return func(r *XPackWatcherStartRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackWatcherStart) WithFilterPath(v ...string) func(*XPackWatcherStartRequest) {
	return func(r *XPackWatcherStartRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackWatcherStart) WithHeader(h map[string]string) func(*XPackWatcherStartRequest) {
	return func(r *XPackWatcherStartRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
