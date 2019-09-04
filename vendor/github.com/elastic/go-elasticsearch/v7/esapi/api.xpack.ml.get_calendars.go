// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newMLGetCalendarsFunc(t Transport) MLGetCalendars {
	return func(o ...func(*MLGetCalendarsRequest)) (*Response, error) {
		var r = MLGetCalendarsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLGetCalendars -
//
type MLGetCalendars func(o ...func(*MLGetCalendarsRequest)) (*Response, error)

// MLGetCalendarsRequest configures the ML Get Calendars API request.
//
type MLGetCalendarsRequest struct {
	Body io.Reader

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
func (r MLGetCalendarsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_ml") + 1 + len("calendars") + 1 + len(r.CalendarID))
	path.WriteString("/")
	path.WriteString("_ml")
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
func (f MLGetCalendars) WithContext(v context.Context) func(*MLGetCalendarsRequest) {
	return func(r *MLGetCalendarsRequest) {
		r.ctx = v
	}
}

// WithBody - The from and size parameters optionally sent in the body.
//
func (f MLGetCalendars) WithBody(v io.Reader) func(*MLGetCalendarsRequest) {
	return func(r *MLGetCalendarsRequest) {
		r.Body = v
	}
}

// WithCalendarID - the ID of the calendar to fetch.
//
func (f MLGetCalendars) WithCalendarID(v string) func(*MLGetCalendarsRequest) {
	return func(r *MLGetCalendarsRequest) {
		r.CalendarID = v
	}
}

// WithFrom - skips a number of calendars.
//
func (f MLGetCalendars) WithFrom(v int) func(*MLGetCalendarsRequest) {
	return func(r *MLGetCalendarsRequest) {
		r.From = &v
	}
}

// WithSize - specifies a max number of calendars to get.
//
func (f MLGetCalendars) WithSize(v int) func(*MLGetCalendarsRequest) {
	return func(r *MLGetCalendarsRequest) {
		r.Size = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLGetCalendars) WithPretty() func(*MLGetCalendarsRequest) {
	return func(r *MLGetCalendarsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLGetCalendars) WithHuman() func(*MLGetCalendarsRequest) {
	return func(r *MLGetCalendarsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLGetCalendars) WithErrorTrace() func(*MLGetCalendarsRequest) {
	return func(r *MLGetCalendarsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLGetCalendars) WithFilterPath(v ...string) func(*MLGetCalendarsRequest) {
	return func(r *MLGetCalendarsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLGetCalendars) WithHeader(h map[string]string) func(*MLGetCalendarsRequest) {
	return func(r *MLGetCalendarsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
