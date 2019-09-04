// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newIndicesUnfreezeFunc(t Transport) IndicesUnfreeze {
	return func(index string, o ...func(*IndicesUnfreezeRequest)) (*Response, error) {
		var r = IndicesUnfreezeRequest{Index: index}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesUnfreeze - https://www.elastic.co/guide/en/elasticsearch/reference/current/frozen.html
//
type IndicesUnfreeze func(index string, o ...func(*IndicesUnfreezeRequest)) (*Response, error)

// IndicesUnfreezeRequest configures the Indices Unfreeze API request.
//
type IndicesUnfreezeRequest struct {
	Index string

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
func (r IndicesUnfreezeRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(r.Index) + 1 + len("_unfreeze"))
	path.WriteString("/")
	path.WriteString(r.Index)
	path.WriteString("/")
	path.WriteString("_unfreeze")

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
func (f IndicesUnfreeze) WithContext(v context.Context) func(*IndicesUnfreezeRequest) {
	return func(r *IndicesUnfreezeRequest) {
		r.ctx = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesUnfreeze) WithAllowNoIndices(v bool) func(*IndicesUnfreezeRequest) {
	return func(r *IndicesUnfreezeRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesUnfreeze) WithExpandWildcards(v string) func(*IndicesUnfreezeRequest) {
	return func(r *IndicesUnfreezeRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesUnfreeze) WithIgnoreUnavailable(v bool) func(*IndicesUnfreezeRequest) {
	return func(r *IndicesUnfreezeRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesUnfreeze) WithMasterTimeout(v time.Duration) func(*IndicesUnfreezeRequest) {
	return func(r *IndicesUnfreezeRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IndicesUnfreeze) WithTimeout(v time.Duration) func(*IndicesUnfreezeRequest) {
	return func(r *IndicesUnfreezeRequest) {
		r.Timeout = v
	}
}

// WithWaitForActiveShards - sets the number of active shards to wait for before the operation returns..
//
func (f IndicesUnfreeze) WithWaitForActiveShards(v string) func(*IndicesUnfreezeRequest) {
	return func(r *IndicesUnfreezeRequest) {
		r.WaitForActiveShards = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesUnfreeze) WithPretty() func(*IndicesUnfreezeRequest) {
	return func(r *IndicesUnfreezeRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesUnfreeze) WithHuman() func(*IndicesUnfreezeRequest) {
	return func(r *IndicesUnfreezeRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesUnfreeze) WithErrorTrace() func(*IndicesUnfreezeRequest) {
	return func(r *IndicesUnfreezeRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesUnfreeze) WithFilterPath(v ...string) func(*IndicesUnfreezeRequest) {
	return func(r *IndicesUnfreezeRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesUnfreeze) WithHeader(h map[string]string) func(*IndicesUnfreezeRequest) {
	return func(r *IndicesUnfreezeRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
