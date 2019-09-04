// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newXPackMLDeleteDatafeedFunc(t Transport) XPackMLDeleteDatafeed {
	return func(datafeed_id string, o ...func(*XPackMLDeleteDatafeedRequest)) (*Response, error) {
		var r = XPackMLDeleteDatafeedRequest{DatafeedID: datafeed_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLDeleteDatafeed - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-delete-datafeed.html
//
type XPackMLDeleteDatafeed func(datafeed_id string, o ...func(*XPackMLDeleteDatafeedRequest)) (*Response, error)

// XPackMLDeleteDatafeedRequest configures the X PackML Delete Datafeed API request.
//
type XPackMLDeleteDatafeedRequest struct {
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
func (r XPackMLDeleteDatafeedRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("datafeeds") + 1 + len(r.DatafeedID))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
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
func (f XPackMLDeleteDatafeed) WithContext(v context.Context) func(*XPackMLDeleteDatafeedRequest) {
	return func(r *XPackMLDeleteDatafeedRequest) {
		r.ctx = v
	}
}

// WithForce - true if the datafeed should be forcefully deleted.
//
func (f XPackMLDeleteDatafeed) WithForce(v bool) func(*XPackMLDeleteDatafeedRequest) {
	return func(r *XPackMLDeleteDatafeedRequest) {
		r.Force = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLDeleteDatafeed) WithPretty() func(*XPackMLDeleteDatafeedRequest) {
	return func(r *XPackMLDeleteDatafeedRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLDeleteDatafeed) WithHuman() func(*XPackMLDeleteDatafeedRequest) {
	return func(r *XPackMLDeleteDatafeedRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLDeleteDatafeed) WithErrorTrace() func(*XPackMLDeleteDatafeedRequest) {
	return func(r *XPackMLDeleteDatafeedRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLDeleteDatafeed) WithFilterPath(v ...string) func(*XPackMLDeleteDatafeedRequest) {
	return func(r *XPackMLDeleteDatafeedRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLDeleteDatafeed) WithHeader(h map[string]string) func(*XPackMLDeleteDatafeedRequest) {
	return func(r *XPackMLDeleteDatafeedRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
