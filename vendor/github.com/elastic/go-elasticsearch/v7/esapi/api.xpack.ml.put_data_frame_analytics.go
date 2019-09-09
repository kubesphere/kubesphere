// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newMLPutDataFrameAnalyticsFunc(t Transport) MLPutDataFrameAnalytics {
	return func(id string, body io.Reader, o ...func(*MLPutDataFrameAnalyticsRequest)) (*Response, error) {
		var r = MLPutDataFrameAnalyticsRequest{ID: id, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLPutDataFrameAnalytics - http://www.elastic.co/guide/en/elasticsearch/reference/current/put-dfanalytics.html
//
type MLPutDataFrameAnalytics func(id string, body io.Reader, o ...func(*MLPutDataFrameAnalyticsRequest)) (*Response, error)

// MLPutDataFrameAnalyticsRequest configures the ML Put Data Frame Analytics API request.
//
type MLPutDataFrameAnalyticsRequest struct {
	ID string

	Body io.Reader

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLPutDataFrameAnalyticsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_ml") + 1 + len("data_frame") + 1 + len("analytics") + 1 + len(r.ID))
	path.WriteString("/")
	path.WriteString("_ml")
	path.WriteString("/")
	path.WriteString("data_frame")
	path.WriteString("/")
	path.WriteString("analytics")
	path.WriteString("/")
	path.WriteString(r.ID)

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
func (f MLPutDataFrameAnalytics) WithContext(v context.Context) func(*MLPutDataFrameAnalyticsRequest) {
	return func(r *MLPutDataFrameAnalyticsRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLPutDataFrameAnalytics) WithPretty() func(*MLPutDataFrameAnalyticsRequest) {
	return func(r *MLPutDataFrameAnalyticsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLPutDataFrameAnalytics) WithHuman() func(*MLPutDataFrameAnalyticsRequest) {
	return func(r *MLPutDataFrameAnalyticsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLPutDataFrameAnalytics) WithErrorTrace() func(*MLPutDataFrameAnalyticsRequest) {
	return func(r *MLPutDataFrameAnalyticsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLPutDataFrameAnalytics) WithFilterPath(v ...string) func(*MLPutDataFrameAnalyticsRequest) {
	return func(r *MLPutDataFrameAnalyticsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLPutDataFrameAnalytics) WithHeader(h map[string]string) func(*MLPutDataFrameAnalyticsRequest) {
	return func(r *MLPutDataFrameAnalyticsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
