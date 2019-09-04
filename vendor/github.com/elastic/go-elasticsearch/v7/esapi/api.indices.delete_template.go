// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newIndicesDeleteTemplateFunc(t Transport) IndicesDeleteTemplate {
	return func(name string, o ...func(*IndicesDeleteTemplateRequest)) (*Response, error) {
		var r = IndicesDeleteTemplateRequest{Name: name}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesDeleteTemplate deletes an index template.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/indices-templates.html.
//
type IndicesDeleteTemplate func(name string, o ...func(*IndicesDeleteTemplateRequest)) (*Response, error)

// IndicesDeleteTemplateRequest configures the Indices Delete Template API request.
//
type IndicesDeleteTemplateRequest struct {
	Name string

	MasterTimeout time.Duration
	Timeout       time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesDeleteTemplateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_template") + 1 + len(r.Name))
	path.WriteString("/")
	path.WriteString("_template")
	path.WriteString("/")
	path.WriteString(r.Name)

	params = make(map[string]string)

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
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
func (f IndicesDeleteTemplate) WithContext(v context.Context) func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesDeleteTemplate) WithMasterTimeout(v time.Duration) func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IndicesDeleteTemplate) WithTimeout(v time.Duration) func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesDeleteTemplate) WithPretty() func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesDeleteTemplate) WithHuman() func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesDeleteTemplate) WithErrorTrace() func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesDeleteTemplate) WithFilterPath(v ...string) func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesDeleteTemplate) WithHeader(h map[string]string) func(*IndicesDeleteTemplateRequest) {
	return func(r *IndicesDeleteTemplateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
