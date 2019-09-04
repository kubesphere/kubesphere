// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newMLGetDatafeedStatsFunc(t Transport) MLGetDatafeedStats {
	return func(o ...func(*MLGetDatafeedStatsRequest)) (*Response, error) {
		var r = MLGetDatafeedStatsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLGetDatafeedStats - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-datafeed-stats.html
//
type MLGetDatafeedStats func(o ...func(*MLGetDatafeedStatsRequest)) (*Response, error)

// MLGetDatafeedStatsRequest configures the ML Get Datafeed Stats API request.
//
type MLGetDatafeedStatsRequest struct {
	DatafeedID string

	AllowNoDatafeeds *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLGetDatafeedStatsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_ml") + 1 + len("datafeeds") + 1 + len(r.DatafeedID) + 1 + len("_stats"))
	path.WriteString("/")
	path.WriteString("_ml")
	path.WriteString("/")
	path.WriteString("datafeeds")
	if r.DatafeedID != "" {
		path.WriteString("/")
		path.WriteString(r.DatafeedID)
	}
	path.WriteString("/")
	path.WriteString("_stats")

	params = make(map[string]string)

	if r.AllowNoDatafeeds != nil {
		params["allow_no_datafeeds"] = strconv.FormatBool(*r.AllowNoDatafeeds)
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
func (f MLGetDatafeedStats) WithContext(v context.Context) func(*MLGetDatafeedStatsRequest) {
	return func(r *MLGetDatafeedStatsRequest) {
		r.ctx = v
	}
}

// WithDatafeedID - the ID of the datafeeds stats to fetch.
//
func (f MLGetDatafeedStats) WithDatafeedID(v string) func(*MLGetDatafeedStatsRequest) {
	return func(r *MLGetDatafeedStatsRequest) {
		r.DatafeedID = v
	}
}

// WithAllowNoDatafeeds - whether to ignore if a wildcard expression matches no datafeeds. (this includes `_all` string or when no datafeeds have been specified).
//
func (f MLGetDatafeedStats) WithAllowNoDatafeeds(v bool) func(*MLGetDatafeedStatsRequest) {
	return func(r *MLGetDatafeedStatsRequest) {
		r.AllowNoDatafeeds = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLGetDatafeedStats) WithPretty() func(*MLGetDatafeedStatsRequest) {
	return func(r *MLGetDatafeedStatsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLGetDatafeedStats) WithHuman() func(*MLGetDatafeedStatsRequest) {
	return func(r *MLGetDatafeedStatsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLGetDatafeedStats) WithErrorTrace() func(*MLGetDatafeedStatsRequest) {
	return func(r *MLGetDatafeedStatsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLGetDatafeedStats) WithFilterPath(v ...string) func(*MLGetDatafeedStatsRequest) {
	return func(r *MLGetDatafeedStatsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLGetDatafeedStats) WithHeader(h map[string]string) func(*MLGetDatafeedStatsRequest) {
	return func(r *MLGetDatafeedStatsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
