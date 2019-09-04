// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newILMMoveToStepFunc(t Transport) ILMMoveToStep {
	return func(o ...func(*ILMMoveToStepRequest)) (*Response, error) {
		var r = ILMMoveToStepRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ILMMoveToStep - https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-move-to-step.html
//
type ILMMoveToStep func(o ...func(*ILMMoveToStepRequest)) (*Response, error)

// ILMMoveToStepRequest configures the ILM Move To Step API request.
//
type ILMMoveToStepRequest struct {
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
func (r ILMMoveToStepRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_ilm") + 1 + len("move") + 1 + len(r.Index))
	path.WriteString("/")
	path.WriteString("_ilm")
	path.WriteString("/")
	path.WriteString("move")
	if r.Index != "" {
		path.WriteString("/")
		path.WriteString(r.Index)
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
func (f ILMMoveToStep) WithContext(v context.Context) func(*ILMMoveToStepRequest) {
	return func(r *ILMMoveToStepRequest) {
		r.ctx = v
	}
}

// WithBody - The new lifecycle step to move to.
//
func (f ILMMoveToStep) WithBody(v io.Reader) func(*ILMMoveToStepRequest) {
	return func(r *ILMMoveToStepRequest) {
		r.Body = v
	}
}

// WithIndex - the name of the index whose lifecycle step is to change.
//
func (f ILMMoveToStep) WithIndex(v string) func(*ILMMoveToStepRequest) {
	return func(r *ILMMoveToStepRequest) {
		r.Index = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ILMMoveToStep) WithPretty() func(*ILMMoveToStepRequest) {
	return func(r *ILMMoveToStepRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ILMMoveToStep) WithHuman() func(*ILMMoveToStepRequest) {
	return func(r *ILMMoveToStepRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ILMMoveToStep) WithErrorTrace() func(*ILMMoveToStepRequest) {
	return func(r *ILMMoveToStepRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ILMMoveToStep) WithFilterPath(v ...string) func(*ILMMoveToStepRequest) {
	return func(r *ILMMoveToStepRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f ILMMoveToStep) WithHeader(h map[string]string) func(*ILMMoveToStepRequest) {
	return func(r *ILMMoveToStepRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
