// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newMLDeleteCalendarJobFunc(t Transport) MLDeleteCalendarJob {
	return func(calendar_id string, job_id string, o ...func(*MLDeleteCalendarJobRequest)) (*Response, error) {
		var r = MLDeleteCalendarJobRequest{CalendarID: calendar_id, JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLDeleteCalendarJob -
//
type MLDeleteCalendarJob func(calendar_id string, job_id string, o ...func(*MLDeleteCalendarJobRequest)) (*Response, error)

// MLDeleteCalendarJobRequest configures the ML Delete Calendar Job API request.
//
type MLDeleteCalendarJobRequest struct {
	CalendarID string
	JobID      string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLDeleteCalendarJobRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_ml") + 1 + len("calendars") + 1 + len(r.CalendarID) + 1 + len("jobs") + 1 + len(r.JobID))
	path.WriteString("/")
	path.WriteString("_ml")
	path.WriteString("/")
	path.WriteString("calendars")
	path.WriteString("/")
	path.WriteString(r.CalendarID)
	path.WriteString("/")
	path.WriteString("jobs")
	path.WriteString("/")
	path.WriteString(r.JobID)

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
func (f MLDeleteCalendarJob) WithContext(v context.Context) func(*MLDeleteCalendarJobRequest) {
	return func(r *MLDeleteCalendarJobRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLDeleteCalendarJob) WithPretty() func(*MLDeleteCalendarJobRequest) {
	return func(r *MLDeleteCalendarJobRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLDeleteCalendarJob) WithHuman() func(*MLDeleteCalendarJobRequest) {
	return func(r *MLDeleteCalendarJobRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLDeleteCalendarJob) WithErrorTrace() func(*MLDeleteCalendarJobRequest) {
	return func(r *MLDeleteCalendarJobRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLDeleteCalendarJob) WithFilterPath(v ...string) func(*MLDeleteCalendarJobRequest) {
	return func(r *MLDeleteCalendarJobRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLDeleteCalendarJob) WithHeader(h map[string]string) func(*MLDeleteCalendarJobRequest) {
	return func(r *MLDeleteCalendarJobRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
