// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackMLDeleteCalendarJobFunc(t Transport) XPackMLDeleteCalendarJob {
	return func(calendar_id string, job_id string, o ...func(*XPackMLDeleteCalendarJobRequest)) (*Response, error) {
		var r = XPackMLDeleteCalendarJobRequest{CalendarID: calendar_id, JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLDeleteCalendarJob -
//
type XPackMLDeleteCalendarJob func(calendar_id string, job_id string, o ...func(*XPackMLDeleteCalendarJobRequest)) (*Response, error)

// XPackMLDeleteCalendarJobRequest configures the X PackML Delete Calendar Job API request.
//
type XPackMLDeleteCalendarJobRequest struct {
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
func (r XPackMLDeleteCalendarJobRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("calendars") + 1 + len(r.CalendarID) + 1 + len("jobs") + 1 + len(r.JobID))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
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
func (f XPackMLDeleteCalendarJob) WithContext(v context.Context) func(*XPackMLDeleteCalendarJobRequest) {
	return func(r *XPackMLDeleteCalendarJobRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLDeleteCalendarJob) WithPretty() func(*XPackMLDeleteCalendarJobRequest) {
	return func(r *XPackMLDeleteCalendarJobRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLDeleteCalendarJob) WithHuman() func(*XPackMLDeleteCalendarJobRequest) {
	return func(r *XPackMLDeleteCalendarJobRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLDeleteCalendarJob) WithErrorTrace() func(*XPackMLDeleteCalendarJobRequest) {
	return func(r *XPackMLDeleteCalendarJobRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLDeleteCalendarJob) WithFilterPath(v ...string) func(*XPackMLDeleteCalendarJobRequest) {
	return func(r *XPackMLDeleteCalendarJobRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLDeleteCalendarJob) WithHeader(h map[string]string) func(*XPackMLDeleteCalendarJobRequest) {
	return func(r *XPackMLDeleteCalendarJobRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
