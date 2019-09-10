// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newXPackRollupStopJobFunc(t Transport) XPackRollupStopJob {
	return func(id string, o ...func(*XPackRollupStopJobRequest)) (*Response, error) {
		var r = XPackRollupStopJobRequest{DocumentID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackRollupStopJob -
//
type XPackRollupStopJob func(id string, o ...func(*XPackRollupStopJobRequest)) (*Response, error)

// XPackRollupStopJobRequest configures the X Pack Rollup Stop Job API request.
//
type XPackRollupStopJobRequest struct {
	DocumentID string

	Timeout           time.Duration
	WaitForCompletion *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackRollupStopJobRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_xpack") + 1 + len("rollup") + 1 + len("job") + 1 + len(r.DocumentID) + 1 + len("_stop"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("rollup")
	path.WriteString("/")
	path.WriteString("job")
	path.WriteString("/")
	path.WriteString(r.DocumentID)
	path.WriteString("/")
	path.WriteString("_stop")

	params = make(map[string]string)

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
	}

	if r.WaitForCompletion != nil {
		params["wait_for_completion"] = strconv.FormatBool(*r.WaitForCompletion)
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
func (f XPackRollupStopJob) WithContext(v context.Context) func(*XPackRollupStopJobRequest) {
	return func(r *XPackRollupStopJobRequest) {
		r.ctx = v
	}
}

// WithTimeout - block for (at maximum) the specified duration while waiting for the job to stop.  defaults to 30s..
//
func (f XPackRollupStopJob) WithTimeout(v time.Duration) func(*XPackRollupStopJobRequest) {
	return func(r *XPackRollupStopJobRequest) {
		r.Timeout = v
	}
}

// WithWaitForCompletion - true if the api should block until the job has fully stopped, false if should be executed async. defaults to false..
//
func (f XPackRollupStopJob) WithWaitForCompletion(v bool) func(*XPackRollupStopJobRequest) {
	return func(r *XPackRollupStopJobRequest) {
		r.WaitForCompletion = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackRollupStopJob) WithPretty() func(*XPackRollupStopJobRequest) {
	return func(r *XPackRollupStopJobRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackRollupStopJob) WithHuman() func(*XPackRollupStopJobRequest) {
	return func(r *XPackRollupStopJobRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackRollupStopJob) WithErrorTrace() func(*XPackRollupStopJobRequest) {
	return func(r *XPackRollupStopJobRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackRollupStopJob) WithFilterPath(v ...string) func(*XPackRollupStopJobRequest) {
	return func(r *XPackRollupStopJobRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackRollupStopJob) WithHeader(h map[string]string) func(*XPackRollupStopJobRequest) {
	return func(r *XPackRollupStopJobRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
