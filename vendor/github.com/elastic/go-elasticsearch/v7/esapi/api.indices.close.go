// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newIndicesCloseFunc(t Transport) IndicesClose {
	return func(index []string, o ...func(*IndicesCloseRequest)) (*Response, error) {
		var r = IndicesCloseRequest{Index: index}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesClose closes an index.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/indices-open-close.html.
//
type IndicesClose func(index []string, o ...func(*IndicesCloseRequest)) (*Response, error)

// IndicesCloseRequest configures the Indices Close API request.
//
type IndicesCloseRequest struct {
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
func (r IndicesCloseRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_close"))
	path.WriteString("/")
	path.WriteString(strings.Join(r.Index, ","))
	path.WriteString("/")
	path.WriteString("_close")

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
func (f IndicesClose) WithContext(v context.Context) func(*IndicesCloseRequest) {
	return func(r *IndicesCloseRequest) {
		r.ctx = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesClose) WithAllowNoIndices(v bool) func(*IndicesCloseRequest) {
	return func(r *IndicesCloseRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesClose) WithExpandWildcards(v string) func(*IndicesCloseRequest) {
	return func(r *IndicesCloseRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesClose) WithIgnoreUnavailable(v bool) func(*IndicesCloseRequest) {
	return func(r *IndicesCloseRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesClose) WithMasterTimeout(v time.Duration) func(*IndicesCloseRequest) {
	return func(r *IndicesCloseRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IndicesClose) WithTimeout(v time.Duration) func(*IndicesCloseRequest) {
	return func(r *IndicesCloseRequest) {
		r.Timeout = v
	}
}

// WithWaitForActiveShards - sets the number of active shards to wait for before the operation returns..
//
func (f IndicesClose) WithWaitForActiveShards(v string) func(*IndicesCloseRequest) {
	return func(r *IndicesCloseRequest) {
		r.WaitForActiveShards = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesClose) WithPretty() func(*IndicesCloseRequest) {
	return func(r *IndicesCloseRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesClose) WithHuman() func(*IndicesCloseRequest) {
	return func(r *IndicesCloseRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesClose) WithErrorTrace() func(*IndicesCloseRequest) {
	return func(r *IndicesCloseRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesClose) WithFilterPath(v ...string) func(*IndicesCloseRequest) {
	return func(r *IndicesCloseRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesClose) WithHeader(h map[string]string) func(*IndicesCloseRequest) {
	return func(r *IndicesCloseRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
