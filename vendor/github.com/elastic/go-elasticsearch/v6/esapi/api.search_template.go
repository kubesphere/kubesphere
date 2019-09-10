// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newSearchTemplateFunc(t Transport) SearchTemplate {
	return func(body io.Reader, o ...func(*SearchTemplateRequest)) (*Response, error) {
		var r = SearchTemplateRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SearchTemplate allows to use the Mustache language to pre-render a search definition.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/current/search-template.html.
//
type SearchTemplate func(body io.Reader, o ...func(*SearchTemplateRequest)) (*Response, error)

// SearchTemplateRequest configures the Search Template API request.
//
type SearchTemplateRequest struct {
	Index        []string
	DocumentType []string

	Body io.Reader

	AllowNoIndices     *bool
	ExpandWildcards    string
	Explain            *bool
	IgnoreThrottled    *bool
	IgnoreUnavailable  *bool
	Preference         string
	Profile            *bool
	RestTotalHitsAsInt *bool
	Routing            []string
	Scroll             time.Duration
	SearchType         string
	TypedKeys          *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SearchTemplateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len(strings.Join(r.DocumentType, ",")) + 1 + len("_search") + 1 + len("template"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	if len(r.DocumentType) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.DocumentType, ","))
	}
	path.WriteString("/")
	path.WriteString("_search")
	path.WriteString("/")
	path.WriteString("template")

	params = make(map[string]string)

	if r.AllowNoIndices != nil {
		params["allow_no_indices"] = strconv.FormatBool(*r.AllowNoIndices)
	}

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if r.Explain != nil {
		params["explain"] = strconv.FormatBool(*r.Explain)
	}

	if r.IgnoreThrottled != nil {
		params["ignore_throttled"] = strconv.FormatBool(*r.IgnoreThrottled)
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.Preference != "" {
		params["preference"] = r.Preference
	}

	if r.Profile != nil {
		params["profile"] = strconv.FormatBool(*r.Profile)
	}

	if r.RestTotalHitsAsInt != nil {
		params["rest_total_hits_as_int"] = strconv.FormatBool(*r.RestTotalHitsAsInt)
	}

	if len(r.Routing) > 0 {
		params["routing"] = strings.Join(r.Routing, ",")
	}

	if r.Scroll != 0 {
		params["scroll"] = formatDuration(r.Scroll)
	}

	if r.SearchType != "" {
		params["search_type"] = r.SearchType
	}

	if r.TypedKeys != nil {
		params["typed_keys"] = strconv.FormatBool(*r.TypedKeys)
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
func (f SearchTemplate) WithContext(v context.Context) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names to search; use _all to perform the operation on all indices.
//
func (f SearchTemplate) WithIndex(v ...string) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.Index = v
	}
}

// WithDocumentType - a list of document types to search; leave empty to perform the operation on all types.
//
func (f SearchTemplate) WithDocumentType(v ...string) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.DocumentType = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f SearchTemplate) WithAllowNoIndices(v bool) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f SearchTemplate) WithExpandWildcards(v string) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.ExpandWildcards = v
	}
}

// WithExplain - specify whether to return detailed information about score computation as part of a hit.
//
func (f SearchTemplate) WithExplain(v bool) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.Explain = &v
	}
}

// WithIgnoreThrottled - whether specified concrete, expanded or aliased indices should be ignored when throttled.
//
func (f SearchTemplate) WithIgnoreThrottled(v bool) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.IgnoreThrottled = &v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f SearchTemplate) WithIgnoreUnavailable(v bool) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithPreference - specify the node or shard the operation should be performed on (default: random).
//
func (f SearchTemplate) WithPreference(v string) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.Preference = v
	}
}

// WithProfile - specify whether to profile the query execution.
//
func (f SearchTemplate) WithProfile(v bool) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.Profile = &v
	}
}

// WithRestTotalHitsAsInt - this parameter is ignored in this version. it is used in the next major version to control whether the rest response should render the total.hits as an object or a number.
//
func (f SearchTemplate) WithRestTotalHitsAsInt(v bool) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.RestTotalHitsAsInt = &v
	}
}

// WithRouting - a list of specific routing values.
//
func (f SearchTemplate) WithRouting(v ...string) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.Routing = v
	}
}

// WithScroll - specify how long a consistent view of the index should be maintained for scrolled search.
//
func (f SearchTemplate) WithScroll(v time.Duration) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.Scroll = v
	}
}

// WithSearchType - search operation type.
//
func (f SearchTemplate) WithSearchType(v string) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.SearchType = v
	}
}

// WithTypedKeys - specify whether aggregation and suggester names should be prefixed by their respective types in the response.
//
func (f SearchTemplate) WithTypedKeys(v bool) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.TypedKeys = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SearchTemplate) WithPretty() func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SearchTemplate) WithHuman() func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SearchTemplate) WithErrorTrace() func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SearchTemplate) WithFilterPath(v ...string) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SearchTemplate) WithHeader(h map[string]string) func(*SearchTemplateRequest) {
	return func(r *SearchTemplateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
