// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newXPackMLGetCalendarsFunc(t Transport) XPackMLGetCalendars {
	return func(o ...func(*XPackMLGetCalendarsRequest)) (*Response, error) {
		var r = XPackMLGetCalendarsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLGetCalendars -
//
type XPackMLGetCalendars func(o ...func(*XPackMLGetCalendarsRequest)) (*Response, error)

// XPackMLGetCalendarsRequest configures the X PackML Get Calendars API request.
//
type XPackMLGetCalendarsRequest struct {
	CalendarID string

	From *int
	Size *int

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLGetCalendarsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("calendars") + 1 + len(r.CalendarID))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("calendars")
	if r.CalendarID != "" {
		path.WriteString("/")
		path.WriteString(r.CalendarID)
	}

	params = make(map[string]string)

	if r.From != nil {
		params["from"] = strconv.FormatInt(int64(*r.From), 10)
	}

	if r.Size != nil {
		params["size"] = strconv.FormatInt(int64(*r.Size), 10)
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
func (f XPackMLGetCalendars) WithContext(v context.Context) func(*XPackMLGetCalendarsRequest) {
	return func(r *XPackMLGetCalendarsRequest) {
		r.ctx = v
	}
}

// WithCalendarID - the ID of the calendar to fetch.
//
func (f XPackMLGetCalendars) WithCalendarID(v string) func(*XPackMLGetCalendarsRequest) {
	return func(r *XPackMLGetCalendarsRequest) {
		r.CalendarID = v
	}
}

// WithFrom - skips a number of calendars.
//
func (f XPackMLGetCalendars) WithFrom(v int) func(*XPackMLGetCalendarsRequest) {
	return func(r *XPackMLGetCalendarsRequest) {
		r.From = &v
	}
}

// WithSize - specifies a max number of calendars to get.
//
func (f XPackMLGetCalendars) WithSize(v int) func(*XPackMLGetCalendarsRequest) {
	return func(r *XPackMLGetCalendarsRequest) {
		r.Size = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLGetCalendars) WithPretty() func(*XPackMLGetCalendarsRequest) {
	return func(r *XPackMLGetCalendarsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLGetCalendars) WithHuman() func(*XPackMLGetCalendarsRequest) {
	return func(r *XPackMLGetCalendarsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLGetCalendars) WithErrorTrace() func(*XPackMLGetCalendarsRequest) {
	return func(r *XPackMLGetCalendarsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLGetCalendars) WithFilterPath(v ...string) func(*XPackMLGetCalendarsRequest) {
	return func(r *XPackMLGetCalendarsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLGetCalendars) WithHeader(h map[string]string) func(*XPackMLGetCalendarsRequest) {
	return func(r *XPackMLGetCalendarsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
