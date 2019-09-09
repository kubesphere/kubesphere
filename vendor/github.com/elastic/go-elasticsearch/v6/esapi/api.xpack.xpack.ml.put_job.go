// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newXPackMLPutJobFunc(t Transport) XPackMLPutJob {
	return func(job_id string, body io.Reader, o ...func(*XPackMLPutJobRequest)) (*Response, error) {
		var r = XPackMLPutJobRequest{JobID: job_id, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLPutJob - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-put-job.html
//
type XPackMLPutJob func(job_id string, body io.Reader, o ...func(*XPackMLPutJobRequest)) (*Response, error)

// XPackMLPutJobRequest configures the X PackML Put Job API request.
//
type XPackMLPutJobRequest struct {
	Body io.Reader

	JobID string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLPutJobRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
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
func (f XPackMLPutJob) WithContext(v context.Context) func(*XPackMLPutJobRequest) {
	return func(r *XPackMLPutJobRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLPutJob) WithPretty() func(*XPackMLPutJobRequest) {
	return func(r *XPackMLPutJobRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLPutJob) WithHuman() func(*XPackMLPutJobRequest) {
	return func(r *XPackMLPutJobRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLPutJob) WithErrorTrace() func(*XPackMLPutJobRequest) {
	return func(r *XPackMLPutJobRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLPutJob) WithFilterPath(v ...string) func(*XPackMLPutJobRequest) {
	return func(r *XPackMLPutJobRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLPutJob) WithHeader(h map[string]string) func(*XPackMLPutJobRequest) {
	return func(r *XPackMLPutJobRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
