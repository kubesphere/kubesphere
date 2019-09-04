// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newCCRResumeFollowFunc(t Transport) CCRResumeFollow {
	return func(index string, o ...func(*CCRResumeFollowRequest)) (*Response, error) {
		var r = CCRResumeFollowRequest{Index: index}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CCRResumeFollow - https://www.elastic.co/guide/en/elasticsearch/reference/current/ccr-post-resume-follow.html
//
type CCRResumeFollow func(index string, o ...func(*CCRResumeFollowRequest)) (*Response, error)

// CCRResumeFollowRequest configures the CCR Resume Follow API request.
//
type CCRResumeFollowRequest struct {
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
func (r CCRResumeFollowRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(r.Index) + 1 + len("_ccr") + 1 + len("resume_follow"))
	path.WriteString("/")
	path.WriteString(r.Index)
	path.WriteString("/")
	path.WriteString("_ccr")
	path.WriteString("/")
	path.WriteString("resume_follow")

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
func (f CCRResumeFollow) WithContext(v context.Context) func(*CCRResumeFollowRequest) {
	return func(r *CCRResumeFollowRequest) {
		r.ctx = v
	}
}

// WithBody - The name of the leader index and other optional ccr related parameters.
//
func (f CCRResumeFollow) WithBody(v io.Reader) func(*CCRResumeFollowRequest) {
	return func(r *CCRResumeFollowRequest) {
		r.Body = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CCRResumeFollow) WithPretty() func(*CCRResumeFollowRequest) {
	return func(r *CCRResumeFollowRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CCRResumeFollow) WithHuman() func(*CCRResumeFollowRequest) {
	return func(r *CCRResumeFollowRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CCRResumeFollow) WithErrorTrace() func(*CCRResumeFollowRequest) {
	return func(r *CCRResumeFollowRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CCRResumeFollow) WithFilterPath(v ...string) func(*CCRResumeFollowRequest) {
	return func(r *CCRResumeFollowRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CCRResumeFollow) WithHeader(h map[string]string) func(*CCRResumeFollowRequest) {
	return func(r *CCRResumeFollowRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
