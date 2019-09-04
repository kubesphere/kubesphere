// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newILMRetryFunc(t Transport) ILMRetry {
	return func(o ...func(*ILMRetryRequest)) (*Response, error) {
		var r = ILMRetryRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ILMRetry - https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-retry-policy.html
//
type ILMRetry func(o ...func(*ILMRetryRequest)) (*Response, error)

// ILMRetryRequest configures the ILM Retry API request.
//
type ILMRetryRequest struct {
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
func (r ILMRetryRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(r.Index) + 1 + len("_ilm") + 1 + len("retry"))
	if r.Index != "" {
		path.WriteString("/")
		path.WriteString(r.Index)
	}
	path.WriteString("/")
	path.WriteString("_ilm")
	path.WriteString("/")
	path.WriteString("retry")

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
func (f ILMRetry) WithContext(v context.Context) func(*ILMRetryRequest) {
	return func(r *ILMRetryRequest) {
		r.ctx = v
	}
}

// WithIndex - the name of the indices (comma-separated) whose failed lifecycle step is to be retry.
//
func (f ILMRetry) WithIndex(v string) func(*ILMRetryRequest) {
	return func(r *ILMRetryRequest) {
		r.Index = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ILMRetry) WithPretty() func(*ILMRetryRequest) {
	return func(r *ILMRetryRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ILMRetry) WithHuman() func(*ILMRetryRequest) {
	return func(r *ILMRetryRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ILMRetry) WithErrorTrace() func(*ILMRetryRequest) {
	return func(r *ILMRetryRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ILMRetry) WithFilterPath(v ...string) func(*ILMRetryRequest) {
	return func(r *ILMRetryRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f ILMRetry) WithHeader(h map[string]string) func(*ILMRetryRequest) {
	return func(r *ILMRetryRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
