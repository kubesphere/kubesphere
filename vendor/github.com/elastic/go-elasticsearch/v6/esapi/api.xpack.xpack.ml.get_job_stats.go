// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newXPackMLGetJobStatsFunc(t Transport) XPackMLGetJobStats {
	return func(o ...func(*XPackMLGetJobStatsRequest)) (*Response, error) {
		var r = XPackMLGetJobStatsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLGetJobStats - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-job-stats.html
//
type XPackMLGetJobStats func(o ...func(*XPackMLGetJobStatsRequest)) (*Response, error)

// XPackMLGetJobStatsRequest configures the X PackML Get Job Stats API request.
//
type XPackMLGetJobStatsRequest struct {
	JobID string

	AllowNoJobs *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLGetJobStatsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("_stats"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
	if r.JobID != "" {
		path.WriteString("/")
		path.WriteString(r.JobID)
	}
	path.WriteString("/")
	path.WriteString("_stats")

	params = make(map[string]string)

	if r.AllowNoJobs != nil {
		params["allow_no_jobs"] = strconv.FormatBool(*r.AllowNoJobs)
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
func (f XPackMLGetJobStats) WithContext(v context.Context) func(*XPackMLGetJobStatsRequest) {
	return func(r *XPackMLGetJobStatsRequest) {
		r.ctx = v
	}
}

// WithJobID - the ID of the jobs stats to fetch.
//
func (f XPackMLGetJobStats) WithJobID(v string) func(*XPackMLGetJobStatsRequest) {
	return func(r *XPackMLGetJobStatsRequest) {
		r.JobID = v
	}
}

// WithAllowNoJobs - whether to ignore if a wildcard expression matches no jobs. (this includes `_all` string or when no jobs have been specified).
//
func (f XPackMLGetJobStats) WithAllowNoJobs(v bool) func(*XPackMLGetJobStatsRequest) {
	return func(r *XPackMLGetJobStatsRequest) {
		r.AllowNoJobs = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLGetJobStats) WithPretty() func(*XPackMLGetJobStatsRequest) {
	return func(r *XPackMLGetJobStatsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLGetJobStats) WithHuman() func(*XPackMLGetJobStatsRequest) {
	return func(r *XPackMLGetJobStatsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLGetJobStats) WithErrorTrace() func(*XPackMLGetJobStatsRequest) {
	return func(r *XPackMLGetJobStatsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLGetJobStats) WithFilterPath(v ...string) func(*XPackMLGetJobStatsRequest) {
	return func(r *XPackMLGetJobStatsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLGetJobStats) WithHeader(h map[string]string) func(*XPackMLGetJobStatsRequest) {
	return func(r *XPackMLGetJobStatsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
