// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newXPackMLOpenJobFunc(t Transport) XPackMLOpenJob {
	return func(job_id string, o ...func(*XPackMLOpenJobRequest)) (*Response, error) {
		var r = XPackMLOpenJobRequest{JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLOpenJob - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-open-job.html
//
type XPackMLOpenJob func(job_id string, o ...func(*XPackMLOpenJobRequest)) (*Response, error)

// XPackMLOpenJobRequest configures the X PackML Open Job API request.
//
type XPackMLOpenJobRequest struct {
	IgnoreDowntime *bool
	JobID          string
	Timeout        time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLOpenJobRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("_open"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
	path.WriteString("/")
	path.WriteString(r.JobID)
	path.WriteString("/")
	path.WriteString("_open")

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
func (f XPackMLOpenJob) WithContext(v context.Context) func(*XPackMLOpenJobRequest) {
	return func(r *XPackMLOpenJobRequest) {
		r.ctx = v
	}
}

// WithIgnoreDowntime - controls if gaps in data are treated as anomalous or as a maintenance window after a job re-start.
//
func (f XPackMLOpenJob) WithIgnoreDowntime(v bool) func(*XPackMLOpenJobRequest) {
	return func(r *XPackMLOpenJobRequest) {
		r.IgnoreDowntime = &v
	}
}

// WithTimeout - controls the time to wait until a job has opened. default to 30 minutes.
//
func (f XPackMLOpenJob) WithTimeout(v time.Duration) func(*XPackMLOpenJobRequest) {
	return func(r *XPackMLOpenJobRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLOpenJob) WithPretty() func(*XPackMLOpenJobRequest) {
	return func(r *XPackMLOpenJobRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLOpenJob) WithHuman() func(*XPackMLOpenJobRequest) {
	return func(r *XPackMLOpenJobRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLOpenJob) WithErrorTrace() func(*XPackMLOpenJobRequest) {
	return func(r *XPackMLOpenJobRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLOpenJob) WithFilterPath(v ...string) func(*XPackMLOpenJobRequest) {
	return func(r *XPackMLOpenJobRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLOpenJob) WithHeader(h map[string]string) func(*XPackMLOpenJobRequest) {
	return func(r *XPackMLOpenJobRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
