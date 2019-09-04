// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newIndicesOpenFunc(t Transport) IndicesOpen {
	return func(index []string, o ...func(*IndicesOpenRequest)) (*Response, error) {
		var r = IndicesOpenRequest{Index: index}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesOpen opens an index.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/indices-open-close.html.
//
type IndicesOpen func(index []string, o ...func(*IndicesOpenRequest)) (*Response, error)

// IndicesOpenRequest configures the Indices Open API request.
//
type IndicesOpenRequest struct {
	Index []string

	AllowNoIndices      *bool
	ExpandWildcards     string
	IgnoreUnavailable   *bool
	MasterTimeout       time.Duration
	Timeout             time.Duration
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
func (r IndicesOpenRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_open"))
	path.WriteString("/")
	path.WriteString(strings.Join(r.Index, ","))
	path.WriteString("/")
	path.WriteString("_open")

	params = make(map[string]string)

	if r.AllowNoIndices != nil {
		params["allow_no_indices"] = strconv.FormatBool(*r.AllowNoIndices)
	}

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
	}

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
func (f IndicesOpen) WithContext(v context.Context) func(*IndicesOpenRequest) {
	return func(r *IndicesOpenRequest) {
		r.ctx = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesOpen) WithAllowNoIndices(v bool) func(*IndicesOpenRequest) {
	return func(r *IndicesOpenRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesOpen) WithExpandWildcards(v string) func(*IndicesOpenRequest) {
	return func(r *IndicesOpenRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesOpen) WithIgnoreUnavailable(v bool) func(*IndicesOpenRequest) {
	return func(r *IndicesOpenRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesOpen) WithMasterTimeout(v time.Duration) func(*IndicesOpenRequest) {
	return func(r *IndicesOpenRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IndicesOpen) WithTimeout(v time.Duration) func(*IndicesOpenRequest) {
	return func(r *IndicesOpenRequest) {
		r.Timeout = v
	}
}

// WithWaitForActiveShards - sets the number of active shards to wait for before the operation returns..
//
func (f IndicesOpen) WithWaitForActiveShards(v string) func(*IndicesOpenRequest) {
	return func(r *IndicesOpenRequest) {
		r.WaitForActiveShards = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesOpen) WithPretty() func(*IndicesOpenRequest) {
	return func(r *IndicesOpenRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesOpen) WithHuman() func(*IndicesOpenRequest) {
	return func(r *IndicesOpenRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesOpen) WithErrorTrace() func(*IndicesOpenRequest) {
	return func(r *IndicesOpenRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesOpen) WithFilterPath(v ...string) func(*IndicesOpenRequest) {
	return func(r *IndicesOpenRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesOpen) WithHeader(h map[string]string) func(*IndicesOpenRequest) {
	return func(r *IndicesOpenRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
