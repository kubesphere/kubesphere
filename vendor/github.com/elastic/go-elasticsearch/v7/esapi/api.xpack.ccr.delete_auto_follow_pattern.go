// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newCCRDeleteAutoFollowPatternFunc(t Transport) CCRDeleteAutoFollowPattern {
	return func(name string, o ...func(*CCRDeleteAutoFollowPatternRequest)) (*Response, error) {
		var r = CCRDeleteAutoFollowPatternRequest{Name: name}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CCRDeleteAutoFollowPattern - https://www.elastic.co/guide/en/elasticsearch/reference/current/ccr-delete-auto-follow-pattern.html
//
type CCRDeleteAutoFollowPattern func(name string, o ...func(*CCRDeleteAutoFollowPatternRequest)) (*Response, error)

// CCRDeleteAutoFollowPatternRequest configures the CCR Delete Auto Follow Pattern API request.
//
type CCRDeleteAutoFollowPatternRequest struct {
	Name string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CCRDeleteAutoFollowPatternRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_ccr") + 1 + len("auto_follow") + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString("_ccr")
	path.WriteString("/")
	path.WriteString("auto_follow")
	path.WriteString("/")
	path.WriteString(r.Name)

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
func (f CCRDeleteAutoFollowPattern) WithContext(v context.Context) func(*CCRDeleteAutoFollowPatternRequest) {
	return func(r *CCRDeleteAutoFollowPatternRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CCRDeleteAutoFollowPattern) WithPretty() func(*CCRDeleteAutoFollowPatternRequest) {
	return func(r *CCRDeleteAutoFollowPatternRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CCRDeleteAutoFollowPattern) WithHuman() func(*CCRDeleteAutoFollowPatternRequest) {
	return func(r *CCRDeleteAutoFollowPatternRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CCRDeleteAutoFollowPattern) WithErrorTrace() func(*CCRDeleteAutoFollowPatternRequest) {
	return func(r *CCRDeleteAutoFollowPatternRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CCRDeleteAutoFollowPattern) WithFilterPath(v ...string) func(*CCRDeleteAutoFollowPatternRequest) {
	return func(r *CCRDeleteAutoFollowPatternRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CCRDeleteAutoFollowPattern) WithHeader(h map[string]string) func(*CCRDeleteAutoFollowPatternRequest) {
	return func(r *CCRDeleteAutoFollowPatternRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
