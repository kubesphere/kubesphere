// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newMLStopDataFrameAnalyticsFunc(t Transport) MLStopDataFrameAnalytics {
	return func(id string, o ...func(*MLStopDataFrameAnalyticsRequest)) (*Response, error) {
		var r = MLStopDataFrameAnalyticsRequest{ID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLStopDataFrameAnalytics - http://www.elastic.co/guide/en/elasticsearch/reference/current/stop-dfanalytics.html
//
type MLStopDataFrameAnalytics func(id string, o ...func(*MLStopDataFrameAnalyticsRequest)) (*Response, error)

// MLStopDataFrameAnalyticsRequest configures the ML Stop Data Frame Analytics API request.
//
type MLStopDataFrameAnalyticsRequest struct {
	ID string

	Body io.Reader

	AllowNoMatch *bool
	Force        *bool
	Timeout      time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLStopDataFrameAnalyticsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_ml") + 1 + len("data_frame") + 1 + len("analytics") + 1 + len(r.ID) + 1 + len("_stop"))
	path.WriteString("/")
	path.WriteString("_ml")
	path.WriteString("/")
	path.WriteString("data_frame")
	path.WriteString("/")
	path.WriteString("analytics")
	path.WriteString("/")
	path.WriteString(r.ID)
	path.WriteString("/")
	path.WriteString("_stop")

	params = make(map[string]string)

	if r.AllowNoMatch != nil {
		params["allow_no_match"] = strconv.FormatBool(*r.AllowNoMatch)
	}

	if r.Force != nil {
		params["force"] = strconv.FormatBool(*r.Force)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
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
func (f MLStopDataFrameAnalytics) WithContext(v context.Context) func(*MLStopDataFrameAnalyticsRequest) {
	return func(r *MLStopDataFrameAnalyticsRequest) {
		r.ctx = v
	}
}

// WithBody - The stop data frame analytics parameters.
//
func (f MLStopDataFrameAnalytics) WithBody(v io.Reader) func(*MLStopDataFrameAnalyticsRequest) {
	return func(r *MLStopDataFrameAnalyticsRequest) {
		r.Body = v
	}
}

// WithAllowNoMatch - whether to ignore if a wildcard expression matches no data frame analytics. (this includes `_all` string or when no data frame analytics have been specified).
//
func (f MLStopDataFrameAnalytics) WithAllowNoMatch(v bool) func(*MLStopDataFrameAnalyticsRequest) {
	return func(r *MLStopDataFrameAnalyticsRequest) {
		r.AllowNoMatch = &v
	}
}

// WithForce - true if the data frame analytics should be forcefully stopped.
//
func (f MLStopDataFrameAnalytics) WithForce(v bool) func(*MLStopDataFrameAnalyticsRequest) {
	return func(r *MLStopDataFrameAnalyticsRequest) {
		r.Force = &v
	}
}

// WithTimeout - controls the time to wait until the task has stopped. defaults to 20 seconds.
//
func (f MLStopDataFrameAnalytics) WithTimeout(v time.Duration) func(*MLStopDataFrameAnalyticsRequest) {
	return func(r *MLStopDataFrameAnalyticsRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLStopDataFrameAnalytics) WithPretty() func(*MLStopDataFrameAnalyticsRequest) {
	return func(r *MLStopDataFrameAnalyticsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLStopDataFrameAnalytics) WithHuman() func(*MLStopDataFrameAnalyticsRequest) {
	return func(r *MLStopDataFrameAnalyticsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLStopDataFrameAnalytics) WithErrorTrace() func(*MLStopDataFrameAnalyticsRequest) {
	return func(r *MLStopDataFrameAnalyticsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLStopDataFrameAnalytics) WithFilterPath(v ...string) func(*MLStopDataFrameAnalyticsRequest) {
	return func(r *MLStopDataFrameAnalyticsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLStopDataFrameAnalytics) WithHeader(h map[string]string) func(*MLStopDataFrameAnalyticsRequest) {
	return func(r *MLStopDataFrameAnalyticsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
