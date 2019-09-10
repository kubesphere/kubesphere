// Code generated from specification version 5.6.15: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newIndicesFlushSyncedFunc(t Transport) IndicesFlushSynced {
	return func(o ...func(*IndicesFlushSyncedRequest)) (*Response, error) {
		var r = IndicesFlushSyncedRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesFlushSynced performs a synced flush operation on one or more indices.
//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/5.x/indices-synced-flush.html.
//
type IndicesFlushSynced func(o ...func(*IndicesFlushSyncedRequest)) (*Response, error)

// IndicesFlushSyncedRequest configures the Indices Flush Synced API request.
//
type IndicesFlushSyncedRequest struct {
	Index []string

	AllowNoIndices    *bool
	ExpandWildcards   string
	IgnoreUnavailable *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesFlushSyncedRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_flush") + 1 + len("synced"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_flush")
	path.WriteString("/")
	path.WriteString("synced")

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
func (f IndicesFlushSynced) WithContext(v context.Context) func(*IndicesFlushSyncedRequest) {
	return func(r *IndicesFlushSyncedRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names; use _all for all indices.
//
func (f IndicesFlushSynced) WithIndex(v ...string) func(*IndicesFlushSyncedRequest) {
	return func(r *IndicesFlushSyncedRequest) {
		r.Index = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesFlushSynced) WithAllowNoIndices(v bool) func(*IndicesFlushSyncedRequest) {
	return func(r *IndicesFlushSyncedRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesFlushSynced) WithExpandWildcards(v string) func(*IndicesFlushSyncedRequest) {
	return func(r *IndicesFlushSyncedRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesFlushSynced) WithIgnoreUnavailable(v bool) func(*IndicesFlushSyncedRequest) {
	return func(r *IndicesFlushSyncedRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesFlushSynced) WithPretty() func(*IndicesFlushSyncedRequest) {
	return func(r *IndicesFlushSyncedRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesFlushSynced) WithHuman() func(*IndicesFlushSyncedRequest) {
	return func(r *IndicesFlushSyncedRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesFlushSynced) WithErrorTrace() func(*IndicesFlushSyncedRequest) {
	return func(r *IndicesFlushSyncedRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesFlushSynced) WithFilterPath(v ...string) func(*IndicesFlushSyncedRequest) {
	return func(r *IndicesFlushSyncedRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesFlushSynced) WithHeader(h map[string]string) func(*IndicesFlushSyncedRequest) {
	return func(r *IndicesFlushSyncedRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
