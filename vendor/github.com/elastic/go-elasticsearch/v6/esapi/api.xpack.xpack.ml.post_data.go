// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newXPackMLPostDataFunc(t Transport) XPackMLPostData {
	return func(job_id string, body io.Reader, o ...func(*XPackMLPostDataRequest)) (*Response, error) {
		var r = XPackMLPostDataRequest{JobID: job_id, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLPostData - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-post-data.html
//
type XPackMLPostData func(job_id string, body io.Reader, o ...func(*XPackMLPostDataRequest)) (*Response, error)

// XPackMLPostDataRequest configures the X PackML Post Data API request.
//
type XPackMLPostDataRequest struct {
	Body io.Reader

	JobID string

	ResetEnd   string
	ResetStart string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLPostDataRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("_data"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
	path.WriteString("/")
	path.WriteString(r.JobID)
	path.WriteString("/")
	path.WriteString("_data")

	params = make(map[string]string)

	if r.ResetEnd != "" {
		params["reset_end"] = r.ResetEnd
	}

	if r.ResetStart != "" {
		params["reset_start"] = r.ResetStart
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
func (f XPackMLPostData) WithContext(v context.Context) func(*XPackMLPostDataRequest) {
	return func(r *XPackMLPostDataRequest) {
		r.ctx = v
	}
}

// WithResetEnd - optional parameter to specify the end of the bucket resetting range.
//
func (f XPackMLPostData) WithResetEnd(v string) func(*XPackMLPostDataRequest) {
	return func(r *XPackMLPostDataRequest) {
		r.ResetEnd = v
	}
}

// WithResetStart - optional parameter to specify the start of the bucket resetting range.
//
func (f XPackMLPostData) WithResetStart(v string) func(*XPackMLPostDataRequest) {
	return func(r *XPackMLPostDataRequest) {
		r.ResetStart = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLPostData) WithPretty() func(*XPackMLPostDataRequest) {
	return func(r *XPackMLPostDataRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLPostData) WithHuman() func(*XPackMLPostDataRequest) {
	return func(r *XPackMLPostDataRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLPostData) WithErrorTrace() func(*XPackMLPostDataRequest) {
	return func(r *XPackMLPostDataRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLPostData) WithFilterPath(v ...string) func(*XPackMLPostDataRequest) {
	return func(r *XPackMLPostDataRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLPostData) WithHeader(h map[string]string) func(*XPackMLPostDataRequest) {
	return func(r *XPackMLPostDataRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
