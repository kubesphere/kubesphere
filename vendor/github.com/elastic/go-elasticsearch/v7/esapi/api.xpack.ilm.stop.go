// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newILMStopFunc(t Transport) ILMStop {
	return func(o ...func(*ILMStopRequest)) (*Response, error) {
		var r = ILMStopRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ILMStop - https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-stop.html
//
type ILMStop func(o ...func(*ILMStopRequest)) (*Response, error)

// ILMStopRequest configures the ILM Stop API request.
//
type ILMStopRequest struct {
	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r ILMStopRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(len("/_ilm/stop"))
	path.WriteString("/_ilm/stop")

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
func (f ILMStop) WithContext(v context.Context) func(*ILMStopRequest) {
	return func(r *ILMStopRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ILMStop) WithPretty() func(*ILMStopRequest) {
	return func(r *ILMStopRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ILMStop) WithHuman() func(*ILMStopRequest) {
	return func(r *ILMStopRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ILMStop) WithErrorTrace() func(*ILMStopRequest) {
	return func(r *ILMStopRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ILMStop) WithFilterPath(v ...string) func(*ILMStopRequest) {
	return func(r *ILMStopRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f ILMStop) WithHeader(h map[string]string) func(*ILMStopRequest) {
	return func(r *ILMStopRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
