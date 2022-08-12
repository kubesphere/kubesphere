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
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newCountFunc(t Transport) Count {
	return func(o ...func(*CountRequest)) (*Response, error) {
		var r = CountRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// Count returns number of documents matching a query.
//
//
type Count func(o ...func(*CountRequest)) (*Response, error)

// CountRequest configures the Count API request.
//
type CountRequest struct {
	Index        []string
	DocumentType []string

	Body io.Reader

	AllowNoIndices    *bool
	Analyzer          string
	AnalyzeWildcard   *bool
	DefaultOperator   string
	Df                string
	ExpandWildcards   string
	IgnoreThrottled   *bool
	IgnoreUnavailable *bool
	Lenient           *bool
	MinScore          *int
	Preference        string
	Query             string
	Routing           []string
	TerminateAfter    *int

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CountRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len(strings.Join(r.DocumentType, ",")) + 1 + len("_count"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	if len(r.DocumentType) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.DocumentType, ","))
	}
	path.WriteString("/")
	path.WriteString("_count")

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

	if r.DefaultOperator != "" {
		params["default_operator"] = r.DefaultOperator
	}

	if r.Df != "" {
		params["df"] = r.Df
	}

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if r.IgnoreThrottled != nil {
		params["ignore_throttled"] = strconv.FormatBool(*r.IgnoreThrottled)
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.Lenient != nil {
		params["lenient"] = strconv.FormatBool(*r.Lenient)
	}

	if r.MinScore != nil {
		params["min_score"] = strconv.FormatInt(int64(*r.MinScore), 10)
	}

	if r.Preference != "" {
		params["preference"] = r.Preference
	}

	if r.Query != "" {
		params["q"] = r.Query
	}

	if len(r.Routing) > 0 {
		params["routing"] = strings.Join(r.Routing, ",")
	}

	if r.TerminateAfter != nil {
		params["terminate_after"] = strconv.FormatInt(int64(*r.TerminateAfter), 10)
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
func (f Count) WithContext(v context.Context) func(*CountRequest) {
	return func(r *CountRequest) {
		r.ctx = v
	}
}

// WithBody - A query to restrict the results specified with the Query DSL (optional).
//
func (f Count) WithBody(v io.Reader) func(*CountRequest) {
	return func(r *CountRequest) {
		r.Body = v
	}
}

// WithIndex - a list of indices to restrict the results.
//
func (f Count) WithIndex(v ...string) func(*CountRequest) {
	return func(r *CountRequest) {
		r.Index = v
	}
}

// WithDocumentType - a list of types to restrict the results.
//
func (f Count) WithDocumentType(v ...string) func(*CountRequest) {
	return func(r *CountRequest) {
		r.DocumentType = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f Count) WithAllowNoIndices(v bool) func(*CountRequest) {
	return func(r *CountRequest) {
		r.AllowNoIndices = &v
	}
}

// WithAnalyzer - the analyzer to use for the query string.
//
func (f Count) WithAnalyzer(v string) func(*CountRequest) {
	return func(r *CountRequest) {
		r.Analyzer = v
	}
}

// WithAnalyzeWildcard - specify whether wildcard and prefix queries should be analyzed (default: false).
//
func (f Count) WithAnalyzeWildcard(v bool) func(*CountRequest) {
	return func(r *CountRequest) {
		r.AnalyzeWildcard = &v
	}
}

// WithDefaultOperator - the default operator for query string query (and or or).
//
func (f Count) WithDefaultOperator(v string) func(*CountRequest) {
	return func(r *CountRequest) {
		r.DefaultOperator = v
	}
}

// WithDf - the field to use as default where no field prefix is given in the query string.
//
func (f Count) WithDf(v string) func(*CountRequest) {
	return func(r *CountRequest) {
		r.Df = v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f Count) WithExpandWildcards(v string) func(*CountRequest) {
	return func(r *CountRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreThrottled - whether specified concrete, expanded or aliased indices should be ignored when throttled.
//
func (f Count) WithIgnoreThrottled(v bool) func(*CountRequest) {
	return func(r *CountRequest) {
		r.IgnoreThrottled = &v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f Count) WithIgnoreUnavailable(v bool) func(*CountRequest) {
	return func(r *CountRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithLenient - specify whether format-based query failures (such as providing text to a numeric field) should be ignored.
//
func (f Count) WithLenient(v bool) func(*CountRequest) {
	return func(r *CountRequest) {
		r.Lenient = &v
	}
}

// WithMinScore - include only documents with a specific `_score` value in the result.
//
func (f Count) WithMinScore(v int) func(*CountRequest) {
	return func(r *CountRequest) {
		r.MinScore = &v
	}
}

// WithPreference - specify the node or shard the operation should be performed on (default: random).
//
func (f Count) WithPreference(v string) func(*CountRequest) {
	return func(r *CountRequest) {
		r.Preference = v
	}
}

// WithQuery - query in the lucene query string syntax.
//
func (f Count) WithQuery(v string) func(*CountRequest) {
	return func(r *CountRequest) {
		r.Query = v
	}
}

// WithRouting - a list of specific routing values.
//
func (f Count) WithRouting(v ...string) func(*CountRequest) {
	return func(r *CountRequest) {
		r.Routing = v
	}
}

// WithTerminateAfter - the maximum count for each shard, upon reaching which the query execution will terminate early.
//
func (f Count) WithTerminateAfter(v int) func(*CountRequest) {
	return func(r *CountRequest) {
		r.TerminateAfter = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Count) WithPretty() func(*CountRequest) {
	return func(r *CountRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Count) WithHuman() func(*CountRequest) {
	return func(r *CountRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Count) WithErrorTrace() func(*CountRequest) {
	return func(r *CountRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Count) WithFilterPath(v ...string) func(*CountRequest) {
	return func(r *CountRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Count) WithHeader(h map[string]string) func(*CountRequest) {
	return func(r *CountRequest) {
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
func (f Count) WithOpaqueID(s string) func(*CountRequest) {
	return func(r *CountRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
