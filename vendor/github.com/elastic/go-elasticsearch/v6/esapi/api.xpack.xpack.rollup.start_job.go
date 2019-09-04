// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackRollupStartJobFunc(t Transport) XPackRollupStartJob {
	return func(id string, o ...func(*XPackRollupStartJobRequest)) (*Response, error) {
		var r = XPackRollupStartJobRequest{DocumentID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackRollupStartJob -
//
type XPackRollupStartJob func(id string, o ...func(*XPackRollupStartJobRequest)) (*Response, error)

// XPackRollupStartJobRequest configures the X Pack Rollup Start Job API request.
//
type XPackRollupStartJobRequest struct {
	DocumentID string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackRollupStartJobRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_xpack") + 1 + len("rollup") + 1 + len("job") + 1 + len(r.DocumentID) + 1 + len("_start"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("rollup")
	path.WriteString("/")
	path.WriteString("job")
	path.WriteString("/")
	path.WriteString(r.DocumentID)
	path.WriteString("/")
	path.WriteString("_start")

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
func (f XPackRollupStartJob) WithContext(v context.Context) func(*XPackRollupStartJobRequest) {
	return func(r *XPackRollupStartJobRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackRollupStartJob) WithPretty() func(*XPackRollupStartJobRequest) {
	return func(r *XPackRollupStartJobRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackRollupStartJob) WithHuman() func(*XPackRollupStartJobRequest) {
	return func(r *XPackRollupStartJobRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackRollupStartJob) WithErrorTrace() func(*XPackRollupStartJobRequest) {
	return func(r *XPackRollupStartJobRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackRollupStartJob) WithFilterPath(v ...string) func(*XPackRollupStartJobRequest) {
	return func(r *XPackRollupStartJobRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackRollupStartJob) WithHeader(h map[string]string) func(*XPackRollupStartJobRequest) {
	return func(r *XPackRollupStartJobRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
