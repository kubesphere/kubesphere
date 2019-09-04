// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newMLDeleteDatafeedFunc(t Transport) MLDeleteDatafeed {
	return func(datafeed_id string, o ...func(*MLDeleteDatafeedRequest)) (*Response, error) {
		var r = MLDeleteDatafeedRequest{DatafeedID: datafeed_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLDeleteDatafeed - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-delete-datafeed.html
//
type MLDeleteDatafeed func(datafeed_id string, o ...func(*MLDeleteDatafeedRequest)) (*Response, error)

// MLDeleteDatafeedRequest configures the ML Delete Datafeed API request.
//
type MLDeleteDatafeedRequest struct {
	DatafeedID string

	Force *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLDeleteDatafeedRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_ml") + 1 + len("datafeeds") + 1 + len(r.DatafeedID))
	path.WriteString("/")
	path.WriteString("_ml")
	path.WriteString("/")
	path.WriteString("datafeeds")
	path.WriteString("/")
	path.WriteString(r.DatafeedID)

	params = make(map[string]string)

	if r.Force != nil {
		params["force"] = strconv.FormatBool(*r.Force)
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
func (f MLDeleteDatafeed) WithContext(v context.Context) func(*MLDeleteDatafeedRequest) {
	return func(r *MLDeleteDatafeedRequest) {
		r.ctx = v
	}
}

// WithForce - true if the datafeed should be forcefully deleted.
//
func (f MLDeleteDatafeed) WithForce(v bool) func(*MLDeleteDatafeedRequest) {
	return func(r *MLDeleteDatafeedRequest) {
		r.Force = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLDeleteDatafeed) WithPretty() func(*MLDeleteDatafeedRequest) {
	return func(r *MLDeleteDatafeedRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLDeleteDatafeed) WithHuman() func(*MLDeleteDatafeedRequest) {
	return func(r *MLDeleteDatafeedRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLDeleteDatafeed) WithErrorTrace() func(*MLDeleteDatafeedRequest) {
	return func(r *MLDeleteDatafeedRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLDeleteDatafeed) WithFilterPath(v ...string) func(*MLDeleteDatafeedRequest) {
	return func(r *MLDeleteDatafeedRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLDeleteDatafeed) WithHeader(h map[string]string) func(*MLDeleteDatafeedRequest) {
	return func(r *MLDeleteDatafeedRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
