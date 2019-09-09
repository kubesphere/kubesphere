// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newXPackMLCloseJobFunc(t Transport) XPackMLCloseJob {
	return func(job_id string, o ...func(*XPackMLCloseJobRequest)) (*Response, error) {
		var r = XPackMLCloseJobRequest{JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLCloseJob - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-close-job.html
//
type XPackMLCloseJob func(job_id string, o ...func(*XPackMLCloseJobRequest)) (*Response, error)

// XPackMLCloseJobRequest configures the X PackML Close Job API request.
//
type XPackMLCloseJobRequest struct {
	Body io.Reader

	JobID string

	AllowNoJobs *bool
	Force       *bool
	Timeout     time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLCloseJobRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("_close"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
	path.WriteString("/")
	path.WriteString(r.JobID)
	path.WriteString("/")
	path.WriteString("_close")

	params = make(map[string]string)

	if r.AllowNoJobs != nil {
		params["allow_no_jobs"] = strconv.FormatBool(*r.AllowNoJobs)
	}

	if r.Force != nil {
		params["force"] = strconv.FormatBool(*r.Force)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
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
func (f XPackMLCloseJob) WithContext(v context.Context) func(*XPackMLCloseJobRequest) {
	return func(r *XPackMLCloseJobRequest) {
		r.ctx = v
	}
}

// WithBody - The URL params optionally sent in the body.
//
func (f XPackMLCloseJob) WithBody(v io.Reader) func(*XPackMLCloseJobRequest) {
	return func(r *XPackMLCloseJobRequest) {
		r.Body = v
	}
}

// WithAllowNoJobs - whether to ignore if a wildcard expression matches no jobs. (this includes `_all` string or when no jobs have been specified).
//
func (f XPackMLCloseJob) WithAllowNoJobs(v bool) func(*XPackMLCloseJobRequest) {
	return func(r *XPackMLCloseJobRequest) {
		r.AllowNoJobs = &v
	}
}

// WithForce - true if the job should be forcefully closed.
//
func (f XPackMLCloseJob) WithForce(v bool) func(*XPackMLCloseJobRequest) {
	return func(r *XPackMLCloseJobRequest) {
		r.Force = &v
	}
}

// WithTimeout - controls the time to wait until a job has closed. default to 30 minutes.
//
func (f XPackMLCloseJob) WithTimeout(v time.Duration) func(*XPackMLCloseJobRequest) {
	return func(r *XPackMLCloseJobRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLCloseJob) WithPretty() func(*XPackMLCloseJobRequest) {
	return func(r *XPackMLCloseJobRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLCloseJob) WithHuman() func(*XPackMLCloseJobRequest) {
	return func(r *XPackMLCloseJobRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLCloseJob) WithErrorTrace() func(*XPackMLCloseJobRequest) {
	return func(r *XPackMLCloseJobRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLCloseJob) WithFilterPath(v ...string) func(*XPackMLCloseJobRequest) {
	return func(r *XPackMLCloseJobRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLCloseJob) WithHeader(h map[string]string) func(*XPackMLCloseJobRequest) {
	return func(r *XPackMLCloseJobRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
