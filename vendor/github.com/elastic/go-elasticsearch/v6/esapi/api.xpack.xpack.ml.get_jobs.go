// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newXPackMLGetJobsFunc(t Transport) XPackMLGetJobs {
	return func(o ...func(*XPackMLGetJobsRequest)) (*Response, error) {
		var r = XPackMLGetJobsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLGetJobs - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-job.html
//
type XPackMLGetJobs func(o ...func(*XPackMLGetJobsRequest)) (*Response, error)

// XPackMLGetJobsRequest configures the X PackML Get Jobs API request.
//
type XPackMLGetJobsRequest struct {
	JobID string

	AllowNoJobs *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLGetJobsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
	if r.JobID != "" {
		path.WriteString("/")
		path.WriteString(r.JobID)
	}

	params = make(map[string]string)

	if r.AllowNoJobs != nil {
		params["allow_no_jobs"] = strconv.FormatBool(*r.AllowNoJobs)
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
func (f XPackMLGetJobs) WithContext(v context.Context) func(*XPackMLGetJobsRequest) {
	return func(r *XPackMLGetJobsRequest) {
		r.ctx = v
	}
}

// WithJobID - the ID of the jobs to fetch.
//
func (f XPackMLGetJobs) WithJobID(v string) func(*XPackMLGetJobsRequest) {
	return func(r *XPackMLGetJobsRequest) {
		r.JobID = v
	}
}

// WithAllowNoJobs - whether to ignore if a wildcard expression matches no jobs. (this includes `_all` string or when no jobs have been specified).
//
func (f XPackMLGetJobs) WithAllowNoJobs(v bool) func(*XPackMLGetJobsRequest) {
	return func(r *XPackMLGetJobsRequest) {
		r.AllowNoJobs = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLGetJobs) WithPretty() func(*XPackMLGetJobsRequest) {
	return func(r *XPackMLGetJobsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLGetJobs) WithHuman() func(*XPackMLGetJobsRequest) {
	return func(r *XPackMLGetJobsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLGetJobs) WithErrorTrace() func(*XPackMLGetJobsRequest) {
	return func(r *XPackMLGetJobsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLGetJobs) WithFilterPath(v ...string) func(*XPackMLGetJobsRequest) {
	return func(r *XPackMLGetJobsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLGetJobs) WithHeader(h map[string]string) func(*XPackMLGetJobsRequest) {
	return func(r *XPackMLGetJobsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
