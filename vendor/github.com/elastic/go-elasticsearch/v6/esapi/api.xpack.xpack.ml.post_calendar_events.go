// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newXPackMLPostCalendarEventsFunc(t Transport) XPackMLPostCalendarEvents {
	return func(calendar_id string, body io.Reader, o ...func(*XPackMLPostCalendarEventsRequest)) (*Response, error) {
		var r = XPackMLPostCalendarEventsRequest{CalendarID: calendar_id, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLPostCalendarEvents -
//
type XPackMLPostCalendarEvents func(calendar_id string, body io.Reader, o ...func(*XPackMLPostCalendarEventsRequest)) (*Response, error)

// XPackMLPostCalendarEventsRequest configures the X PackML Post Calendar Events API request.
//
type XPackMLPostCalendarEventsRequest struct {
	Body io.Reader

	CalendarID string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLPostCalendarEventsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

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
func (f XPackMLPostCalendarEvents) WithContext(v context.Context) func(*XPackMLPostCalendarEventsRequest) {
	return func(r *XPackMLPostCalendarEventsRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLPostCalendarEvents) WithPretty() func(*XPackMLPostCalendarEventsRequest) {
	return func(r *XPackMLPostCalendarEventsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLPostCalendarEvents) WithHuman() func(*XPackMLPostCalendarEventsRequest) {
	return func(r *XPackMLPostCalendarEventsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLPostCalendarEvents) WithErrorTrace() func(*XPackMLPostCalendarEventsRequest) {
	return func(r *XPackMLPostCalendarEventsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLPostCalendarEvents) WithFilterPath(v ...string) func(*XPackMLPostCalendarEventsRequest) {
	return func(r *XPackMLPostCalendarEventsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLPostCalendarEvents) WithHeader(h map[string]string) func(*XPackMLPostCalendarEventsRequest) {
	return func(r *XPackMLPostCalendarEventsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
