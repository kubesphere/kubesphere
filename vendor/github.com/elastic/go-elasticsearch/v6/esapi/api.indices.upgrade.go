// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newIndicesUpgradeFunc(t Transport) IndicesUpgrade {
	return func(o ...func(*IndicesUpgradeRequest)) (*Response, error) {
		var r = IndicesUpgradeRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesUpgrade the _upgrade API is no longer useful and will be removed.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/indices-upgrade.html.
//
type IndicesUpgrade func(o ...func(*IndicesUpgradeRequest)) (*Response, error)

// IndicesUpgradeRequest configures the Indices Upgrade API request.
//
type IndicesUpgradeRequest struct {
	Index []string

	AllowNoIndices      *bool
	ExpandWildcards     string
	IgnoreUnavailable   *bool
	OnlyAncientSegments *bool
	WaitForCompletion   *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesUpgradeRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_upgrade"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_upgrade")

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

	if r.OnlyAncientSegments != nil {
		params["only_ancient_segments"] = strconv.FormatBool(*r.OnlyAncientSegments)
	}

	if r.WaitForCompletion != nil {
		params["wait_for_completion"] = strconv.FormatBool(*r.WaitForCompletion)
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
func (f IndicesUpgrade) WithContext(v context.Context) func(*IndicesUpgradeRequest) {
	return func(r *IndicesUpgradeRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names; use _all to perform the operation on all indices.
//
func (f IndicesUpgrade) WithIndex(v ...string) func(*IndicesUpgradeRequest) {
	return func(r *IndicesUpgradeRequest) {
		r.Index = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesUpgrade) WithAllowNoIndices(v bool) func(*IndicesUpgradeRequest) {
	return func(r *IndicesUpgradeRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesUpgrade) WithExpandWildcards(v string) func(*IndicesUpgradeRequest) {
	return func(r *IndicesUpgradeRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesUpgrade) WithIgnoreUnavailable(v bool) func(*IndicesUpgradeRequest) {
	return func(r *IndicesUpgradeRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithOnlyAncientSegments - if true, only ancient (an older lucene major release) segments will be upgraded.
//
func (f IndicesUpgrade) WithOnlyAncientSegments(v bool) func(*IndicesUpgradeRequest) {
	return func(r *IndicesUpgradeRequest) {
		r.OnlyAncientSegments = &v
	}
}

// WithWaitForCompletion - specify whether the request should block until the all segments are upgraded (default: false).
//
func (f IndicesUpgrade) WithWaitForCompletion(v bool) func(*IndicesUpgradeRequest) {
	return func(r *IndicesUpgradeRequest) {
		r.WaitForCompletion = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesUpgrade) WithPretty() func(*IndicesUpgradeRequest) {
	return func(r *IndicesUpgradeRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesUpgrade) WithHuman() func(*IndicesUpgradeRequest) {
	return func(r *IndicesUpgradeRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesUpgrade) WithErrorTrace() func(*IndicesUpgradeRequest) {
	return func(r *IndicesUpgradeRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesUpgrade) WithFilterPath(v ...string) func(*IndicesUpgradeRequest) {
	return func(r *IndicesUpgradeRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesUpgrade) WithHeader(h map[string]string) func(*IndicesUpgradeRequest) {
	return func(r *IndicesUpgradeRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
