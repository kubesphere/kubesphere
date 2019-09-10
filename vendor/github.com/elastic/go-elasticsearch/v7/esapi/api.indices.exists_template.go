// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newIndicesExistsTemplateFunc(t Transport) IndicesExistsTemplate {
	return func(name []string, o ...func(*IndicesExistsTemplateRequest)) (*Response, error) {
		var r = IndicesExistsTemplateRequest{Name: name}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesExistsTemplate returns information about whether a particular index template exists.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/indices-templates.html.
//
type IndicesExistsTemplate func(name []string, o ...func(*IndicesExistsTemplateRequest)) (*Response, error)

// IndicesExistsTemplateRequest configures the Indices Exists Template API request.
//
type IndicesExistsTemplateRequest struct {
	Name []string

	FlatSettings  *bool
	Local         *bool
	MasterTimeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesExistsTemplateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "HEAD"

	path.Grow(1 + len("_template") + 1 + len(strings.Join(r.Name, ",")))
	path.WriteString("/")
	path.WriteString("_template")
	path.WriteString("/")
	path.WriteString(strings.Join(r.Name, ","))

	params = make(map[string]string)

	if r.FlatSettings != nil {
		params["flat_settings"] = strconv.FormatBool(*r.FlatSettings)
	}

	if r.Local != nil {
		params["local"] = strconv.FormatBool(*r.Local)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
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
func (f IndicesExistsTemplate) WithContext(v context.Context) func(*IndicesExistsTemplateRequest) {
	return func(r *IndicesExistsTemplateRequest) {
		r.ctx = v
	}
}

// WithFlatSettings - return settings in flat format (default: false).
//
func (f IndicesExistsTemplate) WithFlatSettings(v bool) func(*IndicesExistsTemplateRequest) {
	return func(r *IndicesExistsTemplateRequest) {
		r.FlatSettings = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f IndicesExistsTemplate) WithLocal(v bool) func(*IndicesExistsTemplateRequest) {
	return func(r *IndicesExistsTemplateRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f IndicesExistsTemplate) WithMasterTimeout(v time.Duration) func(*IndicesExistsTemplateRequest) {
	return func(r *IndicesExistsTemplateRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesExistsTemplate) WithPretty() func(*IndicesExistsTemplateRequest) {
	return func(r *IndicesExistsTemplateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesExistsTemplate) WithHuman() func(*IndicesExistsTemplateRequest) {
	return func(r *IndicesExistsTemplateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesExistsTemplate) WithErrorTrace() func(*IndicesExistsTemplateRequest) {
	return func(r *IndicesExistsTemplateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesExistsTemplate) WithFilterPath(v ...string) func(*IndicesExistsTemplateRequest) {
	return func(r *IndicesExistsTemplateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesExistsTemplate) WithHeader(h map[string]string) func(*IndicesExistsTemplateRequest) {
	return func(r *IndicesExistsTemplateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
