// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newXPackMLStopDatafeedFunc(t Transport) XPackMLStopDatafeed {
	return func(datafeed_id string, o ...func(*XPackMLStopDatafeedRequest)) (*Response, error) {
		var r = XPackMLStopDatafeedRequest{DatafeedID: datafeed_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLStopDatafeed - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-stop-datafeed.html
//
type XPackMLStopDatafeed func(datafeed_id string, o ...func(*XPackMLStopDatafeedRequest)) (*Response, error)

// XPackMLStopDatafeedRequest configures the X PackML Stop Datafeed API request.
//
type XPackMLStopDatafeedRequest struct {
	DatafeedID string

	AllowNoDatafeeds *bool
	Force            *bool
	Timeout          time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLStopDatafeedRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("datafeeds") + 1 + len(r.DatafeedID) + 1 + len("_stop"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("datafeeds")
	path.WriteString("/")
	path.WriteString(r.DatafeedID)
	path.WriteString("/")
	path.WriteString("_stop")

	params = make(map[string]string)

	if r.AllowNoDatafeeds != nil {
		params["allow_no_datafeeds"] = strconv.FormatBool(*r.AllowNoDatafeeds)
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
func (f XPackMLStopDatafeed) WithContext(v context.Context) func(*XPackMLStopDatafeedRequest) {
	return func(r *XPackMLStopDatafeedRequest) {
		r.ctx = v
	}
}

// WithAllowNoDatafeeds - whether to ignore if a wildcard expression matches no datafeeds. (this includes `_all` string or when no datafeeds have been specified).
//
func (f XPackMLStopDatafeed) WithAllowNoDatafeeds(v bool) func(*XPackMLStopDatafeedRequest) {
	return func(r *XPackMLStopDatafeedRequest) {
		r.AllowNoDatafeeds = &v
	}
}

// WithForce - true if the datafeed should be forcefully stopped..
//
func (f XPackMLStopDatafeed) WithForce(v bool) func(*XPackMLStopDatafeedRequest) {
	return func(r *XPackMLStopDatafeedRequest) {
		r.Force = &v
	}
}

// WithTimeout - controls the time to wait until a datafeed has stopped. default to 20 seconds.
//
func (f XPackMLStopDatafeed) WithTimeout(v time.Duration) func(*XPackMLStopDatafeedRequest) {
	return func(r *XPackMLStopDatafeedRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLStopDatafeed) WithPretty() func(*XPackMLStopDatafeedRequest) {
	return func(r *XPackMLStopDatafeedRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLStopDatafeed) WithHuman() func(*XPackMLStopDatafeedRequest) {
	return func(r *XPackMLStopDatafeedRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLStopDatafeed) WithErrorTrace() func(*XPackMLStopDatafeedRequest) {
	return func(r *XPackMLStopDatafeedRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLStopDatafeed) WithFilterPath(v ...string) func(*XPackMLStopDatafeedRequest) {
	return func(r *XPackMLStopDatafeedRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLStopDatafeed) WithHeader(h map[string]string) func(*XPackMLStopDatafeedRequest) {
	return func(r *XPackMLStopDatafeedRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
