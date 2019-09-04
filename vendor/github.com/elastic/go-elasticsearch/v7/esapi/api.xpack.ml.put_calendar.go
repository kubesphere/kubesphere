// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newMLPutCalendarFunc(t Transport) MLPutCalendar {
	return func(calendar_id string, o ...func(*MLPutCalendarRequest)) (*Response, error) {
		var r = MLPutCalendarRequest{CalendarID: calendar_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLPutCalendar -
//
type MLPutCalendar func(calendar_id string, o ...func(*MLPutCalendarRequest)) (*Response, error)

// MLPutCalendarRequest configures the ML Put Calendar API request.
//
type MLPutCalendarRequest struct {
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
func (r MLPutCalendarRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_ml") + 1 + len("calendars") + 1 + len(r.CalendarID))
	path.WriteString("/")
	path.WriteString("_ml")
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
func (f MLPutCalendar) WithContext(v context.Context) func(*MLPutCalendarRequest) {
	return func(r *MLPutCalendarRequest) {
		r.ctx = v
	}
}

// WithBody - The calendar details.
//
func (f MLPutCalendar) WithBody(v io.Reader) func(*MLPutCalendarRequest) {
	return func(r *MLPutCalendarRequest) {
		r.Body = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLPutCalendar) WithPretty() func(*MLPutCalendarRequest) {
	return func(r *MLPutCalendarRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLPutCalendar) WithHuman() func(*MLPutCalendarRequest) {
	return func(r *MLPutCalendarRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLPutCalendar) WithErrorTrace() func(*MLPutCalendarRequest) {
	return func(r *MLPutCalendarRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLPutCalendar) WithFilterPath(v ...string) func(*MLPutCalendarRequest) {
	return func(r *MLPutCalendarRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLPutCalendar) WithHeader(h map[string]string) func(*MLPutCalendarRequest) {
	return func(r *MLPutCalendarRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
