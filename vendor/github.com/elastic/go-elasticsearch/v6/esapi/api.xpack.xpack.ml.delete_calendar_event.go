// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackMLDeleteCalendarEventFunc(t Transport) XPackMLDeleteCalendarEvent {
	return func(calendar_id string, event_id string, o ...func(*XPackMLDeleteCalendarEventRequest)) (*Response, error) {
		var r = XPackMLDeleteCalendarEventRequest{CalendarID: calendar_id, EventID: event_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLDeleteCalendarEvent -
//
type XPackMLDeleteCalendarEvent func(calendar_id string, event_id string, o ...func(*XPackMLDeleteCalendarEventRequest)) (*Response, error)

// XPackMLDeleteCalendarEventRequest configures the X PackML Delete Calendar Event API request.
//
type XPackMLDeleteCalendarEventRequest struct {
	CalendarID string
	EventID    string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLDeleteCalendarEventRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("calendars") + 1 + len(r.CalendarID) + 1 + len("events") + 1 + len(r.EventID))
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
	path.WriteString("/")
	path.WriteString(r.EventID)

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
func (f XPackMLDeleteCalendarEvent) WithContext(v context.Context) func(*XPackMLDeleteCalendarEventRequest) {
	return func(r *XPackMLDeleteCalendarEventRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLDeleteCalendarEvent) WithPretty() func(*XPackMLDeleteCalendarEventRequest) {
	return func(r *XPackMLDeleteCalendarEventRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLDeleteCalendarEvent) WithHuman() func(*XPackMLDeleteCalendarEventRequest) {
	return func(r *XPackMLDeleteCalendarEventRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLDeleteCalendarEvent) WithErrorTrace() func(*XPackMLDeleteCalendarEventRequest) {
	return func(r *XPackMLDeleteCalendarEventRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLDeleteCalendarEvent) WithFilterPath(v ...string) func(*XPackMLDeleteCalendarEventRequest) {
	return func(r *XPackMLDeleteCalendarEventRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLDeleteCalendarEvent) WithHeader(h map[string]string) func(*XPackMLDeleteCalendarEventRequest) {
	return func(r *XPackMLDeleteCalendarEventRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
