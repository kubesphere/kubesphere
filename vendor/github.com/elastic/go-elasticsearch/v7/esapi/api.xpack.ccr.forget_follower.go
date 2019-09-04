// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newCCRForgetFollowerFunc(t Transport) CCRForgetFollower {
	return func(index string, body io.Reader, o ...func(*CCRForgetFollowerRequest)) (*Response, error) {
		var r = CCRForgetFollowerRequest{Index: index, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CCRForgetFollower - http://www.elastic.co/guide/en/elasticsearch/reference/current
//
type CCRForgetFollower func(index string, body io.Reader, o ...func(*CCRForgetFollowerRequest)) (*Response, error)

// CCRForgetFollowerRequest configures the CCR Forget Follower API request.
//
type CCRForgetFollowerRequest struct {
	Index string

	Body io.Reader

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CCRForgetFollowerRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(r.Index) + 1 + len("_ccr") + 1 + len("forget_follower"))
	path.WriteString("/")
	path.WriteString(r.Index)
	path.WriteString("/")
	path.WriteString("_ccr")
	path.WriteString("/")
	path.WriteString("forget_follower")

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
func (f CCRForgetFollower) WithContext(v context.Context) func(*CCRForgetFollowerRequest) {
	return func(r *CCRForgetFollowerRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CCRForgetFollower) WithPretty() func(*CCRForgetFollowerRequest) {
	return func(r *CCRForgetFollowerRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CCRForgetFollower) WithHuman() func(*CCRForgetFollowerRequest) {
	return func(r *CCRForgetFollowerRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CCRForgetFollower) WithErrorTrace() func(*CCRForgetFollowerRequest) {
	return func(r *CCRForgetFollowerRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CCRForgetFollower) WithFilterPath(v ...string) func(*CCRForgetFollowerRequest) {
	return func(r *CCRForgetFollowerRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CCRForgetFollower) WithHeader(h map[string]string) func(*CCRForgetFollowerRequest) {
	return func(r *CCRForgetFollowerRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
