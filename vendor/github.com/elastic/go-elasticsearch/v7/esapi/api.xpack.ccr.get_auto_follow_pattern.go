// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newCCRGetAutoFollowPatternFunc(t Transport) CCRGetAutoFollowPattern {
	return func(o ...func(*CCRGetAutoFollowPatternRequest)) (*Response, error) {
		var r = CCRGetAutoFollowPatternRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CCRGetAutoFollowPattern - https://www.elastic.co/guide/en/elasticsearch/reference/current/ccr-get-auto-follow-pattern.html
//
type CCRGetAutoFollowPattern func(o ...func(*CCRGetAutoFollowPatternRequest)) (*Response, error)

// CCRGetAutoFollowPatternRequest configures the CCR Get Auto Follow Pattern API request.
//
type CCRGetAutoFollowPatternRequest struct {
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
func (r CCRGetAutoFollowPatternRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_ccr") + 1 + len("auto_follow") + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString("_ccr")
	path.WriteString("/")
	path.WriteString("auto_follow")
	if r.Name != "" {
		path.WriteString("/")
		path.WriteString(r.Name)
	}

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
func (f CCRGetAutoFollowPattern) WithContext(v context.Context) func(*CCRGetAutoFollowPatternRequest) {
	return func(r *CCRGetAutoFollowPatternRequest) {
		r.ctx = v
	}
}

// WithName - the name of the auto follow pattern..
//
func (f CCRGetAutoFollowPattern) WithName(v string) func(*CCRGetAutoFollowPatternRequest) {
	return func(r *CCRGetAutoFollowPatternRequest) {
		r.Name = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CCRGetAutoFollowPattern) WithPretty() func(*CCRGetAutoFollowPatternRequest) {
	return func(r *CCRGetAutoFollowPatternRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CCRGetAutoFollowPattern) WithHuman() func(*CCRGetAutoFollowPatternRequest) {
	return func(r *CCRGetAutoFollowPatternRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CCRGetAutoFollowPattern) WithErrorTrace() func(*CCRGetAutoFollowPatternRequest) {
	return func(r *CCRGetAutoFollowPatternRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CCRGetAutoFollowPattern) WithFilterPath(v ...string) func(*CCRGetAutoFollowPatternRequest) {
	return func(r *CCRGetAutoFollowPatternRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CCRGetAutoFollowPattern) WithHeader(h map[string]string) func(*CCRGetAutoFollowPatternRequest) {
	return func(r *CCRGetAutoFollowPatternRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
