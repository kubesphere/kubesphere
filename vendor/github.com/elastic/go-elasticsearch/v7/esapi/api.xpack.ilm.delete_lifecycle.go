// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newILMDeleteLifecycleFunc(t Transport) ILMDeleteLifecycle {
	return func(o ...func(*ILMDeleteLifecycleRequest)) (*Response, error) {
		var r = ILMDeleteLifecycleRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ILMDeleteLifecycle - https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-delete-lifecycle.html
//
type ILMDeleteLifecycle func(o ...func(*ILMDeleteLifecycleRequest)) (*Response, error)

// ILMDeleteLifecycleRequest configures the ILM Delete Lifecycle API request.
//
type ILMDeleteLifecycleRequest struct {
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
func (r ILMDeleteLifecycleRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

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
func (f ILMDeleteLifecycle) WithContext(v context.Context) func(*ILMDeleteLifecycleRequest) {
	return func(r *ILMDeleteLifecycleRequest) {
		r.ctx = v
	}
}

// WithPolicy - the name of the index lifecycle policy.
//
func (f ILMDeleteLifecycle) WithPolicy(v string) func(*ILMDeleteLifecycleRequest) {
	return func(r *ILMDeleteLifecycleRequest) {
		r.Policy = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ILMDeleteLifecycle) WithPretty() func(*ILMDeleteLifecycleRequest) {
	return func(r *ILMDeleteLifecycleRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ILMDeleteLifecycle) WithHuman() func(*ILMDeleteLifecycleRequest) {
	return func(r *ILMDeleteLifecycleRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ILMDeleteLifecycle) WithErrorTrace() func(*ILMDeleteLifecycleRequest) {
	return func(r *ILMDeleteLifecycleRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ILMDeleteLifecycle) WithFilterPath(v ...string) func(*ILMDeleteLifecycleRequest) {
	return func(r *ILMDeleteLifecycleRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f ILMDeleteLifecycle) WithHeader(h map[string]string) func(*ILMDeleteLifecycleRequest) {
	return func(r *ILMDeleteLifecycleRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
