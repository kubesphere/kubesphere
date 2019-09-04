// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newSearchShardsFunc(t Transport) SearchShards {
	return func(o ...func(*SearchShardsRequest)) (*Response, error) {
		var r = SearchShardsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SearchShards returns information about the indices and shards that a search request would be executed against.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/search-shards.html.
//
type SearchShards func(o ...func(*SearchShardsRequest)) (*Response, error)

// SearchShardsRequest configures the Search Shards API request.
//
type SearchShardsRequest struct {
	Index []string

	AllowNoIndices    *bool
	ExpandWildcards   string
	IgnoreUnavailable *bool
	Local             *bool
	Preference        string
	Routing           string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SearchShardsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_search_shards"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_search_shards")

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

	if r.Local != nil {
		params["local"] = strconv.FormatBool(*r.Local)
	}

	if r.Preference != "" {
		params["preference"] = r.Preference
	}

	if r.Routing != "" {
		params["routing"] = r.Routing
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
func (f SearchShards) WithContext(v context.Context) func(*SearchShardsRequest) {
	return func(r *SearchShardsRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names to search; use _all to perform the operation on all indices.
//
func (f SearchShards) WithIndex(v ...string) func(*SearchShardsRequest) {
	return func(r *SearchShardsRequest) {
		r.Index = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f SearchShards) WithAllowNoIndices(v bool) func(*SearchShardsRequest) {
	return func(r *SearchShardsRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f SearchShards) WithExpandWildcards(v string) func(*SearchShardsRequest) {
	return func(r *SearchShardsRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f SearchShards) WithIgnoreUnavailable(v bool) func(*SearchShardsRequest) {
	return func(r *SearchShardsRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f SearchShards) WithLocal(v bool) func(*SearchShardsRequest) {
	return func(r *SearchShardsRequest) {
		r.Local = &v
	}
}

// WithPreference - specify the node or shard the operation should be performed on (default: random).
//
func (f SearchShards) WithPreference(v string) func(*SearchShardsRequest) {
	return func(r *SearchShardsRequest) {
		r.Preference = v
	}
}

// WithRouting - specific routing value.
//
func (f SearchShards) WithRouting(v string) func(*SearchShardsRequest) {
	return func(r *SearchShardsRequest) {
		r.Routing = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SearchShards) WithPretty() func(*SearchShardsRequest) {
	return func(r *SearchShardsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SearchShards) WithHuman() func(*SearchShardsRequest) {
	return func(r *SearchShardsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SearchShards) WithErrorTrace() func(*SearchShardsRequest) {
	return func(r *SearchShardsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SearchShards) WithFilterPath(v ...string) func(*SearchShardsRequest) {
	return func(r *SearchShardsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SearchShards) WithHeader(h map[string]string) func(*SearchShardsRequest) {
	return func(r *SearchShardsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
