// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func newMLGetCalendarEventsFunc(t Transport) MLGetCalendarEvents {
	return func(calendar_id string, o ...func(*MLGetCalendarEventsRequest)) (*Response, error) {
		var r = MLGetCalendarEventsRequest{CalendarID: calendar_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLGetCalendarEvents -
//
type MLGetCalendarEvents func(calendar_id string, o ...func(*MLGetCalendarEventsRequest)) (*Response, error)

// MLGetCalendarEventsRequest configures the ML Get Calendar Events API request.
//
type MLGetCalendarEventsRequest struct {
	CalendarID string

	End   interface{}
	From  *int
	JobID string
	Size  *int
	Start string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLGetCalendarEventsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_ml") + 1 + len("calendars") + 1 + len(r.CalendarID) + 1 + len("events"))
	path.WriteString("/")
	path.WriteString("_ml")
	path.WriteString("/")
	path.WriteString("calendars")
	path.WriteString("/")
	path.WriteString(r.CalendarID)
	path.WriteString("/")
	path.WriteString("events")

	params = make(map[string]string)

	if r.End != nil {
		params["end"] = fmt.Sprintf("%v", r.End)
	}

	if r.From != nil {
		params["from"] = strconv.FormatInt(int64(*r.From), 10)
	}

	if r.JobID != "" {
		params["job_id"] = r.JobID
	}

	if r.Size != nil {
		params["size"] = strconv.FormatInt(int64(*r.Size), 10)
	}

	if r.Start != "" {
		params["start"] = r.Start
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
func (f MLGetCalendarEvents) WithContext(v context.Context) func(*MLGetCalendarEventsRequest) {
	return func(r *MLGetCalendarEventsRequest) {
		r.ctx = v
	}
}

// WithEnd - get events before this time.
//
func (f MLGetCalendarEvents) WithEnd(v interface{}) func(*MLGetCalendarEventsRequest) {
	return func(r *MLGetCalendarEventsRequest) {
		r.End = v
	}
}

// WithFrom - skips a number of events.
//
func (f MLGetCalendarEvents) WithFrom(v int) func(*MLGetCalendarEventsRequest) {
	return func(r *MLGetCalendarEventsRequest) {
		r.From = &v
	}
}

// WithJobID - get events for the job. when this option is used calendar_id must be '_all'.
//
func (f MLGetCalendarEvents) WithJobID(v string) func(*MLGetCalendarEventsRequest) {
	return func(r *MLGetCalendarEventsRequest) {
		r.JobID = v
	}
}

// WithSize - specifies a max number of events to get.
//
func (f MLGetCalendarEvents) WithSize(v int) func(*MLGetCalendarEventsRequest) {
	return func(r *MLGetCalendarEventsRequest) {
		r.Size = &v
	}
}

// WithStart - get events after this time.
//
func (f MLGetCalendarEvents) WithStart(v string) func(*MLGetCalendarEventsRequest) {
	return func(r *MLGetCalendarEventsRequest) {
		r.Start = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLGetCalendarEvents) WithPretty() func(*MLGetCalendarEventsRequest) {
	return func(r *MLGetCalendarEventsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLGetCalendarEvents) WithHuman() func(*MLGetCalendarEventsRequest) {
	return func(r *MLGetCalendarEventsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLGetCalendarEvents) WithErrorTrace() func(*MLGetCalendarEventsRequest) {
	return func(r *MLGetCalendarEventsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLGetCalendarEvents) WithFilterPath(v ...string) func(*MLGetCalendarEventsRequest) {
	return func(r *MLGetCalendarEventsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLGetCalendarEvents) WithHeader(h map[string]string) func(*MLGetCalendarEventsRequest) {
	return func(r *MLGetCalendarEventsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
