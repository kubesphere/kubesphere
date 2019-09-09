// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newCCRPauseFollowFunc(t Transport) CCRPauseFollow {
	return func(index string, o ...func(*CCRPauseFollowRequest)) (*Response, error) {
		var r = CCRPauseFollowRequest{Index: index}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CCRPauseFollow - https://www.elastic.co/guide/en/elasticsearch/reference/current/ccr-post-pause-follow.html
//
type CCRPauseFollow func(index string, o ...func(*CCRPauseFollowRequest)) (*Response, error)

// CCRPauseFollowRequest configures the CCR Pause Follow API request.
//
type CCRPauseFollowRequest struct {
	Index string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CCRPauseFollowRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(r.Index) + 1 + len("_ccr") + 1 + len("pause_follow"))
	path.WriteString("/")
	path.WriteString(r.Index)
	path.WriteString("/")
	path.WriteString("_ccr")
	path.WriteString("/")
	path.WriteString("pause_follow")

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
func (f CCRPauseFollow) WithContext(v context.Context) func(*CCRPauseFollowRequest) {
	return func(r *CCRPauseFollowRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CCRPauseFollow) WithPretty() func(*CCRPauseFollowRequest) {
	return func(r *CCRPauseFollowRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CCRPauseFollow) WithHuman() func(*CCRPauseFollowRequest) {
	return func(r *CCRPauseFollowRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CCRPauseFollow) WithErrorTrace() func(*CCRPauseFollowRequest) {
	return func(r *CCRPauseFollowRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CCRPauseFollow) WithFilterPath(v ...string) func(*CCRPauseFollowRequest) {
	return func(r *CCRPauseFollowRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CCRPauseFollow) WithHeader(h map[string]string) func(*CCRPauseFollowRequest) {
	return func(r *CCRPauseFollowRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
