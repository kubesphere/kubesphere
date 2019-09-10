// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newRenderSearchTemplateFunc(t Transport) RenderSearchTemplate {
	return func(o ...func(*RenderSearchTemplateRequest)) (*Response, error) {
		var r = RenderSearchTemplateRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// RenderSearchTemplate allows to use the Mustache language to pre-render a search definition.
//
// See full documentation at http://www.elasticsearch.org/guide/en/elasticsearch/reference/master/search-template.html.
//
type RenderSearchTemplate func(o ...func(*RenderSearchTemplateRequest)) (*Response, error)

// RenderSearchTemplateRequest configures the Render Search Template API request.
//
type RenderSearchTemplateRequest struct {
	TemplateID string

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
func (r RenderSearchTemplateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_render") + 1 + len("template") + 1 + len(r.TemplateID))
	path.WriteString("/")
	path.WriteString("_render")
	path.WriteString("/")
	path.WriteString("template")
	if r.TemplateID != "" {
		path.WriteString("/")
		path.WriteString(r.TemplateID)
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
func (f RenderSearchTemplate) WithContext(v context.Context) func(*RenderSearchTemplateRequest) {
	return func(r *RenderSearchTemplateRequest) {
		r.ctx = v
	}
}

// WithBody - The search definition template and its params.
//
func (f RenderSearchTemplate) WithBody(v io.Reader) func(*RenderSearchTemplateRequest) {
	return func(r *RenderSearchTemplateRequest) {
		r.Body = v
	}
}

// WithTemplateID - the ID of the stored search template.
//
func (f RenderSearchTemplate) WithTemplateID(v string) func(*RenderSearchTemplateRequest) {
	return func(r *RenderSearchTemplateRequest) {
		r.TemplateID = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f RenderSearchTemplate) WithPretty() func(*RenderSearchTemplateRequest) {
	return func(r *RenderSearchTemplateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f RenderSearchTemplate) WithHuman() func(*RenderSearchTemplateRequest) {
	return func(r *RenderSearchTemplateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f RenderSearchTemplate) WithErrorTrace() func(*RenderSearchTemplateRequest) {
	return func(r *RenderSearchTemplateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f RenderSearchTemplate) WithFilterPath(v ...string) func(*RenderSearchTemplateRequest) {
	return func(r *RenderSearchTemplateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f RenderSearchTemplate) WithHeader(h map[string]string) func(*RenderSearchTemplateRequest) {
	return func(r *RenderSearchTemplateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
