// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func newXPackMLGetCalendarEventsFunc(t Transport) XPackMLGetCalendarEvents {
	return func(calendar_id string, o ...func(*XPackMLGetCalendarEventsRequest)) (*Response, error) {
		var r = XPackMLGetCalendarEventsRequest{CalendarID: calendar_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLGetCalendarEvents -
//
type XPackMLGetCalendarEvents func(calendar_id string, o ...func(*XPackMLGetCalendarEventsRequest)) (*Response, error)

// XPackMLGetCalendarEventsRequest configures the X PackML Get Calendar Events API request.
//
type XPackMLGetCalendarEventsRequest struct {
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
func (r XPackMLGetCalendarEventsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("calendars") + 1 + len(r.CalendarID) + 1 + len("events"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
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
func (f XPackMLGetCalendarEvents) WithContext(v context.Context) func(*XPackMLGetCalendarEventsRequest) {
	return func(r *XPackMLGetCalendarEventsRequest) {
		r.ctx = v
	}
}

// WithEnd - get events before this time.
//
func (f XPackMLGetCalendarEvents) WithEnd(v interface{}) func(*XPackMLGetCalendarEventsRequest) {
	return func(r *XPackMLGetCalendarEventsRequest) {
		r.End = v
	}
}

// WithFrom - skips a number of events.
//
func (f XPackMLGetCalendarEvents) WithFrom(v int) func(*XPackMLGetCalendarEventsRequest) {
	return func(r *XPackMLGetCalendarEventsRequest) {
		r.From = &v
	}
}

// WithJobID - get events for the job. when this option is used calendar_id must be '_all'.
//
func (f XPackMLGetCalendarEvents) WithJobID(v string) func(*XPackMLGetCalendarEventsRequest) {
	return func(r *XPackMLGetCalendarEventsRequest) {
		r.JobID = v
	}
}

// WithSize - specifies a max number of events to get.
//
func (f XPackMLGetCalendarEvents) WithSize(v int) func(*XPackMLGetCalendarEventsRequest) {
	return func(r *XPackMLGetCalendarEventsRequest) {
		r.Size = &v
	}
}

// WithStart - get events after this time.
//
func (f XPackMLGetCalendarEvents) WithStart(v string) func(*XPackMLGetCalendarEventsRequest) {
	return func(r *XPackMLGetCalendarEventsRequest) {
		r.Start = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLGetCalendarEvents) WithPretty() func(*XPackMLGetCalendarEventsRequest) {
	return func(r *XPackMLGetCalendarEventsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLGetCalendarEvents) WithHuman() func(*XPackMLGetCalendarEventsRequest) {
	return func(r *XPackMLGetCalendarEventsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLGetCalendarEvents) WithErrorTrace() func(*XPackMLGetCalendarEventsRequest) {
	return func(r *XPackMLGetCalendarEventsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLGetCalendarEvents) WithFilterPath(v ...string) func(*XPackMLGetCalendarEventsRequest) {
	return func(r *XPackMLGetCalendarEventsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLGetCalendarEvents) WithHeader(h map[string]string) func(*XPackMLGetCalendarEventsRequest) {
	return func(r *XPackMLGetCalendarEventsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
