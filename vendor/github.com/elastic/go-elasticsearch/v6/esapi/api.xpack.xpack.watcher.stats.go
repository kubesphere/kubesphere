// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newXPackWatcherStatsFunc(t Transport) XPackWatcherStats {
	return func(o ...func(*XPackWatcherStatsRequest)) (*Response, error) {
		var r = XPackWatcherStatsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackWatcherStats - http://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-stats.html
//
type XPackWatcherStats func(o ...func(*XPackWatcherStatsRequest)) (*Response, error)

// XPackWatcherStatsRequest configures the X Pack Watcher Stats API request.
//
type XPackWatcherStatsRequest struct {
	Metric string

	EmitStacktraces *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackWatcherStatsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("watcher") + 1 + len("stats") + 1 + len(r.Metric))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("watcher")
	path.WriteString("/")
	path.WriteString("stats")
	if r.Metric != "" {
		path.WriteString("/")
		path.WriteString(r.Metric)
	}

	params = make(map[string]string)

	if r.EmitStacktraces != nil {
		params["emit_stacktraces"] = strconv.FormatBool(*r.EmitStacktraces)
	}

	if r.Metric != "" {
		params["metric"] = r.Metric
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
func (f XPackWatcherStats) WithContext(v context.Context) func(*XPackWatcherStatsRequest) {
	return func(r *XPackWatcherStatsRequest) {
		r.ctx = v
	}
}

// WithMetric - controls what additional stat metrics should be include in the response.
//
func (f XPackWatcherStats) WithMetric(v string) func(*XPackWatcherStatsRequest) {
	return func(r *XPackWatcherStatsRequest) {
		r.Metric = v
	}
}

// WithEmitStacktraces - emits stack traces of currently running watches.
//
func (f XPackWatcherStats) WithEmitStacktraces(v bool) func(*XPackWatcherStatsRequest) {
	return func(r *XPackWatcherStatsRequest) {
		r.EmitStacktraces = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackWatcherStats) WithPretty() func(*XPackWatcherStatsRequest) {
	return func(r *XPackWatcherStatsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackWatcherStats) WithHuman() func(*XPackWatcherStatsRequest) {
	return func(r *XPackWatcherStatsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackWatcherStats) WithErrorTrace() func(*XPackWatcherStatsRequest) {
	return func(r *XPackWatcherStatsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackWatcherStats) WithFilterPath(v ...string) func(*XPackWatcherStatsRequest) {
	return func(r *XPackWatcherStatsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackWatcherStats) WithHeader(h map[string]string) func(*XPackWatcherStatsRequest) {
	return func(r *XPackWatcherStatsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
