// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newCCRFollowFunc(t Transport) CCRFollow {
	return func(index string, body io.Reader, o ...func(*CCRFollowRequest)) (*Response, error) {
		var r = CCRFollowRequest{Index: index, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CCRFollow - https://www.elastic.co/guide/en/elasticsearch/reference/current/ccr-put-follow.html
//
type CCRFollow func(index string, body io.Reader, o ...func(*CCRFollowRequest)) (*Response, error)

// CCRFollowRequest configures the CCR Follow API request.
//
type CCRFollowRequest struct {
	Index string

	Body io.Reader

	WaitForActiveShards string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CCRFollowRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len(r.Index) + 1 + len("_ccr") + 1 + len("follow"))
	path.WriteString("/")
	path.WriteString(r.Index)
	path.WriteString("/")
	path.WriteString("_ccr")
	path.WriteString("/")
	path.WriteString("follow")

	params = make(map[string]string)

	if r.WaitForActiveShards != "" {
		params["wait_for_active_shards"] = r.WaitForActiveShards
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
func (f CCRFollow) WithContext(v context.Context) func(*CCRFollowRequest) {
	return func(r *CCRFollowRequest) {
		r.ctx = v
	}
}

// WithWaitForActiveShards - sets the number of shard copies that must be active before returning. defaults to 0. set to `all` for all shard copies, otherwise set to any non-negative value less than or equal to the total number of copies for the shard (number of replicas + 1).
//
func (f CCRFollow) WithWaitForActiveShards(v string) func(*CCRFollowRequest) {
	return func(r *CCRFollowRequest) {
		r.WaitForActiveShards = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CCRFollow) WithPretty() func(*CCRFollowRequest) {
	return func(r *CCRFollowRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CCRFollow) WithHuman() func(*CCRFollowRequest) {
	return func(r *CCRFollowRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CCRFollow) WithErrorTrace() func(*CCRFollowRequest) {
	return func(r *CCRFollowRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CCRFollow) WithFilterPath(v ...string) func(*CCRFollowRequest) {
	return func(r *CCRFollowRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CCRFollow) WithHeader(h map[string]string) func(*CCRFollowRequest) {
	return func(r *CCRFollowRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
