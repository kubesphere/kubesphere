// SPDX-License-Identifier: Apache-2.0
//
// The OpenSearch Contributors require contributions made to
// this file be licensed under the Apache-2.0 license or a
// compatible open source license.
//
// Modifications Copyright OpenSearch Contributors. See
// GitHub history for details.

// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package opensearchapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newDeleteByQueryFunc(t Transport) DeleteByQuery {
	return func(index []string, body io.Reader, o ...func(*DeleteByQueryRequest)) (*Response, error) {
		var r = DeleteByQueryRequest{Index: index, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// DeleteByQuery deletes documents matching the provided query.
//
//
type DeleteByQuery func(index []string, body io.Reader, o ...func(*DeleteByQueryRequest)) (*Response, error)

// DeleteByQueryRequest configures the Delete By Query API request.
//
type DeleteByQueryRequest struct {
	Index        []string
	DocumentType []string

	Body io.Reader

	AllowNoIndices      *bool
	Analyzer            string
	AnalyzeWildcard     *bool
	Conflicts           string
	DefaultOperator     string
	Df                  string
	ExpandWildcards     string
	From                *int
	IgnoreUnavailable   *bool
	Lenient             *bool
	MaxDocs             *int
	Preference          string
	Query               string
	Refresh             *bool
	RequestCache        *bool
	RequestsPerSecond   *int
	Routing             []string
	Scroll              time.Duration
	ScrollSize          *int
	SearchTimeout       time.Duration
	SearchType          string
	Size                *int
	Slices              interface{}
	Sort                []string
	Source              []string
	SourceExcludes      []string
	SourceIncludes      []string
	Stats               []string
	TerminateAfter      *int
	Timeout             time.Duration
	Version             *bool
	WaitForActiveShards string
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
func (r DeleteByQueryRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len(strings.Join(r.DocumentType, ",")) + 1 + len("_delete_by_query"))
	path.WriteString("/")
	path.WriteString(strings.Join(r.Index, ","))
	if len(r.DocumentType) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.DocumentType, ","))
	}
	path.WriteString("/")
	path.WriteString("_delete_by_query")

	params = make(map[string]string)

	if r.AllowNoIndices != nil {
		params["allow_no_indices"] = strconv.FormatBool(*r.AllowNoIndices)
	}

	if r.Analyzer != "" {
		params["analyzer"] = r.Analyzer
	}

	if r.AnalyzeWildcard != nil {
		params["analyze_wildcard"] = strconv.FormatBool(*r.AnalyzeWildcard)
	}

	if r.Conflicts != "" {
		params["conflicts"] = r.Conflicts
	}

	if r.DefaultOperator != "" {
		params["default_operator"] = r.DefaultOperator
	}

	if r.Df != "" {
		params["df"] = r.Df
	}

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if r.From != nil {
		params["from"] = strconv.FormatInt(int64(*r.From), 10)
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.Lenient != nil {
		params["lenient"] = strconv.FormatBool(*r.Lenient)
	}

	if r.MaxDocs != nil {
		params["max_docs"] = strconv.FormatInt(int64(*r.MaxDocs), 10)
	}

	if r.Preference != "" {
		params["preference"] = r.Preference
	}

	if r.Query != "" {
		params["q"] = r.Query
	}

	if r.Refresh != nil {
		params["refresh"] = strconv.FormatBool(*r.Refresh)
	}

	if r.RequestCache != nil {
		params["request_cache"] = strconv.FormatBool(*r.RequestCache)
	}

	if r.RequestsPerSecond != nil {
		params["requests_per_second"] = strconv.FormatInt(int64(*r.RequestsPerSecond), 10)
	}

	if len(r.Routing) > 0 {
		params["routing"] = strings.Join(r.Routing, ",")
	}

	if r.Scroll != 0 {
		params["scroll"] = formatDuration(r.Scroll)
	}

	if r.ScrollSize != nil {
		params["scroll_size"] = strconv.FormatInt(int64(*r.ScrollSize), 10)
	}

	if r.SearchTimeout != 0 {
		params["search_timeout"] = formatDuration(r.SearchTimeout)
	}

	if r.SearchType != "" {
		params["search_type"] = r.SearchType
	}

	if r.Size != nil {
		params["size"] = strconv.FormatInt(int64(*r.Size), 10)
	}

	if r.Slices != nil {
		params["slices"] = fmt.Sprintf("%v", r.Slices)
	}

	if len(r.Sort) > 0 {
		params["sort"] = strings.Join(r.Sort, ",")
	}

	if len(r.Source) > 0 {
		params["_source"] = strings.Join(r.Source, ",")
	}

	if len(r.SourceExcludes) > 0 {
		params["_source_excludes"] = strings.Join(r.SourceExcludes, ",")
	}

	if len(r.SourceIncludes) > 0 {
		params["_source_includes"] = strings.Join(r.SourceIncludes, ",")
	}

	if len(r.Stats) > 0 {
		params["stats"] = strings.Join(r.Stats, ",")
	}

	if r.TerminateAfter != nil {
		params["terminate_after"] = strconv.FormatInt(int64(*r.TerminateAfter), 10)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
	}

	if r.Version != nil {
		params["version"] = strconv.FormatBool(*r.Version)
	}

	if r.WaitForActiveShards != "" {
		params["wait_for_active_shards"] = r.WaitForActiveShards
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

	req, err := newRequest(method, path.String(), r.Body)
	if err != nil {
		return nil, err
	}

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
func (f DeleteByQuery) WithContext(v context.Context) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.ctx = v
	}
}

// WithDocumentType - a list of document types to search; leave empty to perform the operation on all types.
//
func (f DeleteByQuery) WithDocumentType(v ...string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.DocumentType = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f DeleteByQuery) WithAllowNoIndices(v bool) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.AllowNoIndices = &v
	}
}

// WithAnalyzer - the analyzer to use for the query string.
//
func (f DeleteByQuery) WithAnalyzer(v string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Analyzer = v
	}
}

// WithAnalyzeWildcard - specify whether wildcard and prefix queries should be analyzed (default: false).
//
func (f DeleteByQuery) WithAnalyzeWildcard(v bool) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.AnalyzeWildcard = &v
	}
}

// WithConflicts - what to do when the delete by query hits version conflicts?.
//
func (f DeleteByQuery) WithConflicts(v string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Conflicts = v
	}
}

// WithDefaultOperator - the default operator for query string query (and or or).
//
func (f DeleteByQuery) WithDefaultOperator(v string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.DefaultOperator = v
	}
}

// WithDf - the field to use as default where no field prefix is given in the query string.
//
func (f DeleteByQuery) WithDf(v string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Df = v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f DeleteByQuery) WithExpandWildcards(v string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.ExpandWildcards = v
	}
}

// WithFrom - starting offset (default: 0).
//
func (f DeleteByQuery) WithFrom(v int) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.From = &v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f DeleteByQuery) WithIgnoreUnavailable(v bool) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithLenient - specify whether format-based query failures (such as providing text to a numeric field) should be ignored.
//
func (f DeleteByQuery) WithLenient(v bool) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Lenient = &v
	}
}

// WithMaxDocs - maximum number of documents to process (default: all documents).
//
func (f DeleteByQuery) WithMaxDocs(v int) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.MaxDocs = &v
	}
}

// WithPreference - specify the node or shard the operation should be performed on (default: random).
//
func (f DeleteByQuery) WithPreference(v string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Preference = v
	}
}

// WithQuery - query in the lucene query string syntax.
//
func (f DeleteByQuery) WithQuery(v string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Query = v
	}
}

// WithRefresh - should the effected indexes be refreshed?.
//
func (f DeleteByQuery) WithRefresh(v bool) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Refresh = &v
	}
}

// WithRequestCache - specify if request cache should be used for this request or not, defaults to index level setting.
//
func (f DeleteByQuery) WithRequestCache(v bool) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.RequestCache = &v
	}
}

// WithRequestsPerSecond - the throttle for this request in sub-requests per second. -1 means no throttle..
//
func (f DeleteByQuery) WithRequestsPerSecond(v int) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.RequestsPerSecond = &v
	}
}

// WithRouting - a list of specific routing values.
//
func (f DeleteByQuery) WithRouting(v ...string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Routing = v
	}
}

// WithScroll - specify how long a consistent view of the index should be maintained for scrolled search.
//
func (f DeleteByQuery) WithScroll(v time.Duration) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Scroll = v
	}
}

// WithScrollSize - size on the scroll request powering the delete by query.
//
func (f DeleteByQuery) WithScrollSize(v int) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.ScrollSize = &v
	}
}

// WithSearchTimeout - explicit timeout for each search request. defaults to no timeout..
//
func (f DeleteByQuery) WithSearchTimeout(v time.Duration) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.SearchTimeout = v
	}
}

// WithSearchType - search operation type.
//
func (f DeleteByQuery) WithSearchType(v string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.SearchType = v
	}
}

// WithSize - deprecated, please use `max_docs` instead.
//
func (f DeleteByQuery) WithSize(v int) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Size = &v
	}
}

// WithSlices - the number of slices this task should be divided into. defaults to 1, meaning the task isn't sliced into subtasks. can be set to `auto`..
//
func (f DeleteByQuery) WithSlices(v interface{}) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Slices = v
	}
}

// WithSort - a list of <field>:<direction> pairs.
//
func (f DeleteByQuery) WithSort(v ...string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Sort = v
	}
}

// WithSource - true or false to return the _source field or not, or a list of fields to return.
//
func (f DeleteByQuery) WithSource(v ...string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Source = v
	}
}

// WithSourceExcludes - a list of fields to exclude from the returned _source field.
//
func (f DeleteByQuery) WithSourceExcludes(v ...string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.SourceExcludes = v
	}
}

// WithSourceIncludes - a list of fields to extract and return from the _source field.
//
func (f DeleteByQuery) WithSourceIncludes(v ...string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.SourceIncludes = v
	}
}

// WithStats - specific 'tag' of the request for logging and statistical purposes.
//
func (f DeleteByQuery) WithStats(v ...string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Stats = v
	}
}

// WithTerminateAfter - the maximum number of documents to collect for each shard, upon reaching which the query execution will terminate early..
//
func (f DeleteByQuery) WithTerminateAfter(v int) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.TerminateAfter = &v
	}
}

// WithTimeout - time each individual bulk request should wait for shards that are unavailable..
//
func (f DeleteByQuery) WithTimeout(v time.Duration) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Timeout = v
	}
}

// WithVersion - specify whether to return document version as part of a hit.
//
func (f DeleteByQuery) WithVersion(v bool) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Version = &v
	}
}

// WithWaitForActiveShards - sets the number of shard copies that must be active before proceeding with the delete by query operation. defaults to 1, meaning the primary shard only. set to `all` for all shard copies, otherwise set to any non-negative value less than or equal to the total number of copies for the shard (number of replicas + 1).
//
func (f DeleteByQuery) WithWaitForActiveShards(v string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.WaitForActiveShards = v
	}
}

// WithWaitForCompletion - should the request should block until the delete by query is complete..
//
func (f DeleteByQuery) WithWaitForCompletion(v bool) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.WaitForCompletion = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f DeleteByQuery) WithPretty() func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f DeleteByQuery) WithHuman() func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f DeleteByQuery) WithErrorTrace() func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f DeleteByQuery) WithFilterPath(v ...string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f DeleteByQuery) WithHeader(h map[string]string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}

// WithOpaqueID adds the X-Opaque-Id header to the HTTP request.
//
func (f DeleteByQuery) WithOpaqueID(s string) func(*DeleteByQueryRequest) {
	return func(r *DeleteByQueryRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
