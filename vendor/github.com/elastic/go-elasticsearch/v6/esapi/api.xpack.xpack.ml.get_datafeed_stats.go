// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newXPackMLGetDatafeedStatsFunc(t Transport) XPackMLGetDatafeedStats {
	return func(o ...func(*XPackMLGetDatafeedStatsRequest)) (*Response, error) {
		var r = XPackMLGetDatafeedStatsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLGetDatafeedStats - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-datafeed-stats.html
//
type XPackMLGetDatafeedStats func(o ...func(*XPackMLGetDatafeedStatsRequest)) (*Response, error)

// XPackMLGetDatafeedStatsRequest configures the X PackML Get Datafeed Stats API request.
//
type XPackMLGetDatafeedStatsRequest struct {
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
func (r XPackMLGetDatafeedStatsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("datafeeds") + 1 + len(r.DatafeedID) + 1 + len("_stats"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
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
func (f XPackMLGetDatafeedStats) WithContext(v context.Context) func(*XPackMLGetDatafeedStatsRequest) {
	return func(r *XPackMLGetDatafeedStatsRequest) {
		r.ctx = v
	}
}

// WithDatafeedID - the ID of the datafeeds stats to fetch.
//
func (f XPackMLGetDatafeedStats) WithDatafeedID(v string) func(*XPackMLGetDatafeedStatsRequest) {
	return func(r *XPackMLGetDatafeedStatsRequest) {
		r.DatafeedID = v
	}
}

// WithAllowNoDatafeeds - whether to ignore if a wildcard expression matches no datafeeds. (this includes `_all` string or when no datafeeds have been specified).
//
func (f XPackMLGetDatafeedStats) WithAllowNoDatafeeds(v bool) func(*XPackMLGetDatafeedStatsRequest) {
	return func(r *XPackMLGetDatafeedStatsRequest) {
		r.AllowNoDatafeeds = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLGetDatafeedStats) WithPretty() func(*XPackMLGetDatafeedStatsRequest) {
	return func(r *XPackMLGetDatafeedStatsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLGetDatafeedStats) WithHuman() func(*XPackMLGetDatafeedStatsRequest) {
	return func(r *XPackMLGetDatafeedStatsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLGetDatafeedStats) WithErrorTrace() func(*XPackMLGetDatafeedStatsRequest) {
	return func(r *XPackMLGetDatafeedStatsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLGetDatafeedStats) WithFilterPath(v ...string) func(*XPackMLGetDatafeedStatsRequest) {
	return func(r *XPackMLGetDatafeedStatsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLGetDatafeedStats) WithHeader(h map[string]string) func(*XPackMLGetDatafeedStatsRequest) {
	return func(r *XPackMLGetDatafeedStatsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
