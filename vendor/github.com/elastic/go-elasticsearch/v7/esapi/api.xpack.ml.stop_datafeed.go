// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newMLStopDatafeedFunc(t Transport) MLStopDatafeed {
	return func(datafeed_id string, o ...func(*MLStopDatafeedRequest)) (*Response, error) {
		var r = MLStopDatafeedRequest{DatafeedID: datafeed_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLStopDatafeed - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-stop-datafeed.html
//
type MLStopDatafeed func(datafeed_id string, o ...func(*MLStopDatafeedRequest)) (*Response, error)

// MLStopDatafeedRequest configures the ML Stop Datafeed API request.
//
type MLStopDatafeedRequest struct {
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
func (r MLStopDatafeedRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_ml") + 1 + len("datafeeds") + 1 + len(r.DatafeedID) + 1 + len("_stop"))
	path.WriteString("/")
	path.WriteString("_ml")
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
func (f MLStopDatafeed) WithContext(v context.Context) func(*MLStopDatafeedRequest) {
	return func(r *MLStopDatafeedRequest) {
		r.ctx = v
	}
}

// WithAllowNoDatafeeds - whether to ignore if a wildcard expression matches no datafeeds. (this includes `_all` string or when no datafeeds have been specified).
//
func (f MLStopDatafeed) WithAllowNoDatafeeds(v bool) func(*MLStopDatafeedRequest) {
	return func(r *MLStopDatafeedRequest) {
		r.AllowNoDatafeeds = &v
	}
}

// WithForce - true if the datafeed should be forcefully stopped..
//
func (f MLStopDatafeed) WithForce(v bool) func(*MLStopDatafeedRequest) {
	return func(r *MLStopDatafeedRequest) {
		r.Force = &v
	}
}

// WithTimeout - controls the time to wait until a datafeed has stopped. default to 20 seconds.
//
func (f MLStopDatafeed) WithTimeout(v time.Duration) func(*MLStopDatafeedRequest) {
	return func(r *MLStopDatafeedRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLStopDatafeed) WithPretty() func(*MLStopDatafeedRequest) {
	return func(r *MLStopDatafeedRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLStopDatafeed) WithHuman() func(*MLStopDatafeedRequest) {
	return func(r *MLStopDatafeedRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLStopDatafeed) WithErrorTrace() func(*MLStopDatafeedRequest) {
	return func(r *MLStopDatafeedRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLStopDatafeed) WithFilterPath(v ...string) func(*MLStopDatafeedRequest) {
	return func(r *MLStopDatafeedRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLStopDatafeed) WithHeader(h map[string]string) func(*MLStopDatafeedRequest) {
	return func(r *MLStopDatafeedRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
