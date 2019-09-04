// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

func newMLStartDatafeedFunc(t Transport) MLStartDatafeed {
	return func(datafeed_id string, o ...func(*MLStartDatafeedRequest)) (*Response, error) {
		var r = MLStartDatafeedRequest{DatafeedID: datafeed_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLStartDatafeed - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-start-datafeed.html
//
type MLStartDatafeed func(datafeed_id string, o ...func(*MLStartDatafeedRequest)) (*Response, error)

// MLStartDatafeedRequest configures the ML Start Datafeed API request.
//
type MLStartDatafeedRequest struct {
	Body io.Reader

	DatafeedID string

	End     string
	Start   string
	Timeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLStartDatafeedRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_ml") + 1 + len("datafeeds") + 1 + len(r.DatafeedID) + 1 + len("_start"))
	path.WriteString("/")
	path.WriteString("_ml")
	path.WriteString("/")
	path.WriteString("datafeeds")
	path.WriteString("/")
	path.WriteString(r.DatafeedID)
	path.WriteString("/")
	path.WriteString("_start")

	params = make(map[string]string)

	if r.End != "" {
		params["end"] = r.End
	}

	if r.Start != "" {
		params["start"] = r.Start
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
func (f MLStartDatafeed) WithContext(v context.Context) func(*MLStartDatafeedRequest) {
	return func(r *MLStartDatafeedRequest) {
		r.ctx = v
	}
}

// WithBody - The start datafeed parameters.
//
func (f MLStartDatafeed) WithBody(v io.Reader) func(*MLStartDatafeedRequest) {
	return func(r *MLStartDatafeedRequest) {
		r.Body = v
	}
}

// WithEnd - the end time when the datafeed should stop. when not set, the datafeed continues in real time.
//
func (f MLStartDatafeed) WithEnd(v string) func(*MLStartDatafeedRequest) {
	return func(r *MLStartDatafeedRequest) {
		r.End = v
	}
}

// WithStart - the start time from where the datafeed should begin.
//
func (f MLStartDatafeed) WithStart(v string) func(*MLStartDatafeedRequest) {
	return func(r *MLStartDatafeedRequest) {
		r.Start = v
	}
}

// WithTimeout - controls the time to wait until a datafeed has started. default to 20 seconds.
//
func (f MLStartDatafeed) WithTimeout(v time.Duration) func(*MLStartDatafeedRequest) {
	return func(r *MLStartDatafeedRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLStartDatafeed) WithPretty() func(*MLStartDatafeedRequest) {
	return func(r *MLStartDatafeedRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLStartDatafeed) WithHuman() func(*MLStartDatafeedRequest) {
	return func(r *MLStartDatafeedRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLStartDatafeed) WithErrorTrace() func(*MLStartDatafeedRequest) {
	return func(r *MLStartDatafeedRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLStartDatafeed) WithFilterPath(v ...string) func(*MLStartDatafeedRequest) {
	return func(r *MLStartDatafeedRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLStartDatafeed) WithHeader(h map[string]string) func(*MLStartDatafeedRequest) {
	return func(r *MLStartDatafeedRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
