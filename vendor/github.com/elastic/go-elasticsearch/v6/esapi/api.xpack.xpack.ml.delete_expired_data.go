// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackMLDeleteExpiredDataFunc(t Transport) XPackMLDeleteExpiredData {
	return func(o ...func(*XPackMLDeleteExpiredDataRequest)) (*Response, error) {
		var r = XPackMLDeleteExpiredDataRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLDeleteExpiredData -
//
type XPackMLDeleteExpiredData func(o ...func(*XPackMLDeleteExpiredDataRequest)) (*Response, error)

// XPackMLDeleteExpiredDataRequest configures the X PackML Delete Expired Data API request.
//
type XPackMLDeleteExpiredDataRequest struct {
	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLDeleteExpiredDataRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(len("/_xpack/ml/_delete_expired_data"))
	path.WriteString("/_xpack/ml/_delete_expired_data")

	params = make(map[string]string)

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
func (f XPackMLDeleteExpiredData) WithContext(v context.Context) func(*XPackMLDeleteExpiredDataRequest) {
	return func(r *XPackMLDeleteExpiredDataRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLDeleteExpiredData) WithPretty() func(*XPackMLDeleteExpiredDataRequest) {
	return func(r *XPackMLDeleteExpiredDataRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLDeleteExpiredData) WithHuman() func(*XPackMLDeleteExpiredDataRequest) {
	return func(r *XPackMLDeleteExpiredDataRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLDeleteExpiredData) WithErrorTrace() func(*XPackMLDeleteExpiredDataRequest) {
	return func(r *XPackMLDeleteExpiredDataRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLDeleteExpiredData) WithFilterPath(v ...string) func(*XPackMLDeleteExpiredDataRequest) {
	return func(r *XPackMLDeleteExpiredDataRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLDeleteExpiredData) WithHeader(h map[string]string) func(*XPackMLDeleteExpiredDataRequest) {
	return func(r *XPackMLDeleteExpiredDataRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
