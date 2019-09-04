// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newRollupGetJobsFunc(t Transport) RollupGetJobs {
	return func(o ...func(*RollupGetJobsRequest)) (*Response, error) {
		var r = RollupGetJobsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// RollupGetJobs -
//
type RollupGetJobs func(o ...func(*RollupGetJobsRequest)) (*Response, error)

// RollupGetJobsRequest configures the Rollup Get Jobs API request.
//
type RollupGetJobsRequest struct {
	JobID string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r RollupGetJobsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_rollup") + 1 + len("job") + 1 + len(r.JobID))
	path.WriteString("/")
	path.WriteString("_rollup")
	path.WriteString("/")
	path.WriteString("job")
	if r.JobID != "" {
		path.WriteString("/")
		path.WriteString(r.JobID)
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
func (f RollupGetJobs) WithContext(v context.Context) func(*RollupGetJobsRequest) {
	return func(r *RollupGetJobsRequest) {
		r.ctx = v
	}
}

// WithJobID - the ID of the job(s) to fetch. accepts glob patterns, or left blank for all jobs.
//
func (f RollupGetJobs) WithJobID(v string) func(*RollupGetJobsRequest) {
	return func(r *RollupGetJobsRequest) {
		r.JobID = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f RollupGetJobs) WithPretty() func(*RollupGetJobsRequest) {
	return func(r *RollupGetJobsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f RollupGetJobs) WithHuman() func(*RollupGetJobsRequest) {
	return func(r *RollupGetJobsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f RollupGetJobs) WithErrorTrace() func(*RollupGetJobsRequest) {
	return func(r *RollupGetJobsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f RollupGetJobs) WithFilterPath(v ...string) func(*RollupGetJobsRequest) {
	return func(r *RollupGetJobsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f RollupGetJobs) WithHeader(h map[string]string) func(*RollupGetJobsRequest) {
	return func(r *RollupGetJobsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
