// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newXPackMLPutCalendarFunc(t Transport) XPackMLPutCalendar {
	return func(calendar_id string, o ...func(*XPackMLPutCalendarRequest)) (*Response, error) {
		var r = XPackMLPutCalendarRequest{CalendarID: calendar_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLPutCalendar -
//
type XPackMLPutCalendar func(calendar_id string, o ...func(*XPackMLPutCalendarRequest)) (*Response, error)

// XPackMLPutCalendarRequest configures the X PackML Put Calendar API request.
//
type XPackMLPutCalendarRequest struct {
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
func (r XPackMLPutCalendarRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("calendars") + 1 + len(r.CalendarID))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("calendars")
	path.WriteString("/")
	path.WriteString(r.CalendarID)

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
func (f XPackMLPutCalendar) WithContext(v context.Context) func(*XPackMLPutCalendarRequest) {
	return func(r *XPackMLPutCalendarRequest) {
		r.ctx = v
	}
}

// WithBody - The calendar details.
//
func (f XPackMLPutCalendar) WithBody(v io.Reader) func(*XPackMLPutCalendarRequest) {
	return func(r *XPackMLPutCalendarRequest) {
		r.Body = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLPutCalendar) WithPretty() func(*XPackMLPutCalendarRequest) {
	return func(r *XPackMLPutCalendarRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLPutCalendar) WithHuman() func(*XPackMLPutCalendarRequest) {
	return func(r *XPackMLPutCalendarRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLPutCalendar) WithErrorTrace() func(*XPackMLPutCalendarRequest) {
	return func(r *XPackMLPutCalendarRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLPutCalendar) WithFilterPath(v ...string) func(*XPackMLPutCalendarRequest) {
	return func(r *XPackMLPutCalendarRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLPutCalendar) WithHeader(h map[string]string) func(*XPackMLPutCalendarRequest) {
	return func(r *XPackMLPutCalendarRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
