// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newILMPutLifecycleFunc(t Transport) ILMPutLifecycle {
	return func(o ...func(*ILMPutLifecycleRequest)) (*Response, error) {
		var r = ILMPutLifecycleRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ILMPutLifecycle - https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-put-lifecycle.html
//
type ILMPutLifecycle func(o ...func(*ILMPutLifecycleRequest)) (*Response, error)

// ILMPutLifecycleRequest configures the ILM Put Lifecycle API request.
//
type ILMPutLifecycleRequest struct {
	Body io.Reader

	Policy string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r ILMPutLifecycleRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_ilm") + 1 + len("policy") + 1 + len(r.Policy))
	path.WriteString("/")
	path.WriteString("_ilm")
	path.WriteString("/")
	path.WriteString("policy")
	if r.Policy != "" {
		path.WriteString("/")
		path.WriteString(r.Policy)
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
func (f ILMPutLifecycle) WithContext(v context.Context) func(*ILMPutLifecycleRequest) {
	return func(r *ILMPutLifecycleRequest) {
		r.ctx = v
	}
}

// WithBody - The lifecycle policy definition to register.
//
func (f ILMPutLifecycle) WithBody(v io.Reader) func(*ILMPutLifecycleRequest) {
	return func(r *ILMPutLifecycleRequest) {
		r.Body = v
	}
}

// WithPolicy - the name of the index lifecycle policy.
//
func (f ILMPutLifecycle) WithPolicy(v string) func(*ILMPutLifecycleRequest) {
	return func(r *ILMPutLifecycleRequest) {
		r.Policy = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ILMPutLifecycle) WithPretty() func(*ILMPutLifecycleRequest) {
	return func(r *ILMPutLifecycleRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ILMPutLifecycle) WithHuman() func(*ILMPutLifecycleRequest) {
	return func(r *ILMPutLifecycleRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ILMPutLifecycle) WithErrorTrace() func(*ILMPutLifecycleRequest) {
	return func(r *ILMPutLifecycleRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ILMPutLifecycle) WithFilterPath(v ...string) func(*ILMPutLifecycleRequest) {
	return func(r *ILMPutLifecycleRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f ILMPutLifecycle) WithHeader(h map[string]string) func(*ILMPutLifecycleRequest) {
	return func(r *ILMPutLifecycleRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
