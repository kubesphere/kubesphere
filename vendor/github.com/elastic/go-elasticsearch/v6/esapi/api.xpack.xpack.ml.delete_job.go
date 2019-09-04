// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newXPackMLDeleteJobFunc(t Transport) XPackMLDeleteJob {
	return func(job_id string, o ...func(*XPackMLDeleteJobRequest)) (*Response, error) {
		var r = XPackMLDeleteJobRequest{JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLDeleteJob - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-delete-job.html
//
type XPackMLDeleteJob func(job_id string, o ...func(*XPackMLDeleteJobRequest)) (*Response, error)

// XPackMLDeleteJobRequest configures the X PackML Delete Job API request.
//
type XPackMLDeleteJobRequest struct {
	JobID string

	Force             *bool
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
func (r XPackMLDeleteJobRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
	path.WriteString("/")
	path.WriteString(r.JobID)

	params = make(map[string]string)

	if r.Force != nil {
		params["force"] = strconv.FormatBool(*r.Force)
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
func (f XPackMLDeleteJob) WithContext(v context.Context) func(*XPackMLDeleteJobRequest) {
	return func(r *XPackMLDeleteJobRequest) {
		r.ctx = v
	}
}

// WithForce - true if the job should be forcefully deleted.
//
func (f XPackMLDeleteJob) WithForce(v bool) func(*XPackMLDeleteJobRequest) {
	return func(r *XPackMLDeleteJobRequest) {
		r.Force = &v
	}
}

// WithWaitForCompletion - should this request wait until the operation has completed before returning.
//
func (f XPackMLDeleteJob) WithWaitForCompletion(v bool) func(*XPackMLDeleteJobRequest) {
	return func(r *XPackMLDeleteJobRequest) {
		r.WaitForCompletion = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLDeleteJob) WithPretty() func(*XPackMLDeleteJobRequest) {
	return func(r *XPackMLDeleteJobRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLDeleteJob) WithHuman() func(*XPackMLDeleteJobRequest) {
	return func(r *XPackMLDeleteJobRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLDeleteJob) WithErrorTrace() func(*XPackMLDeleteJobRequest) {
	return func(r *XPackMLDeleteJobRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLDeleteJob) WithFilterPath(v ...string) func(*XPackMLDeleteJobRequest) {
	return func(r *XPackMLDeleteJobRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLDeleteJob) WithHeader(h map[string]string) func(*XPackMLDeleteJobRequest) {
	return func(r *XPackMLDeleteJobRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
