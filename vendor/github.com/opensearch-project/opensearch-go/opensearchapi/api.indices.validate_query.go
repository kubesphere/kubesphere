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

func newIndicesValidateQueryFunc(t Transport) IndicesValidateQuery {
	return func(o ...func(*IndicesValidateQueryRequest)) (*Response, error) {
		var r = IndicesValidateQueryRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesValidateQuery allows a user to validate a potentially expensive query without executing it.
//
//
type IndicesValidateQuery func(o ...func(*IndicesValidateQueryRequest)) (*Response, error)

// IndicesValidateQueryRequest configures the Indices Validate Query API request.
//
type IndicesValidateQueryRequest struct {
	Index        []string
	DocumentType []string

	Body io.Reader

	AllowNoIndices    *bool
	AllShards         *bool
	Analyzer          string
	AnalyzeWildcard   *bool
	DefaultOperator   string
	Df                string
	ExpandWildcards   string
	Explain           *bool
	IgnoreUnavailable *bool
	Lenient           *bool
	Query             string
	Rewrite           *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesValidateQueryRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len(strings.Join(r.DocumentType, ",")) + 1 + len("_validate") + 1 + len("query"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	if len(r.DocumentType) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.DocumentType, ","))
	}
	path.WriteString("/")
	path.WriteString("_validate")
	path.WriteString("/")
	path.WriteString("query")

	params = make(map[string]string)

	if r.AllowNoIndices != nil {
		params["allow_no_indices"] = strconv.FormatBool(*r.AllowNoIndices)
	}

	if r.AllShards != nil {
		params["all_shards"] = strconv.FormatBool(*r.AllShards)
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

	if r.Explain != nil {
		params["explain"] = strconv.FormatBool(*r.Explain)
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.Lenient != nil {
		params["lenient"] = strconv.FormatBool(*r.Lenient)
	}

	if r.Query != "" {
		params["q"] = r.Query
	}

	if r.Rewrite != nil {
		params["rewrite"] = strconv.FormatBool(*r.Rewrite)
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
func (f IndicesValidateQuery) WithContext(v context.Context) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.ctx = v
	}
}

// WithBody - The query definition specified with the Query DSL.
//
func (f IndicesValidateQuery) WithBody(v io.Reader) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.Body = v
	}
}

// WithIndex - a list of index names to restrict the operation; use _all to perform the operation on all indices.
//
func (f IndicesValidateQuery) WithIndex(v ...string) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.Index = v
	}
}

// WithDocumentType - a list of document types to restrict the operation; leave empty to perform the operation on all types.
//
func (f IndicesValidateQuery) WithDocumentType(v ...string) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.DocumentType = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesValidateQuery) WithAllowNoIndices(v bool) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.AllowNoIndices = &v
	}
}

// WithAllShards - execute validation on all shards instead of one random shard per index.
//
func (f IndicesValidateQuery) WithAllShards(v bool) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.AllShards = &v
	}
}

// WithAnalyzer - the analyzer to use for the query string.
//
func (f IndicesValidateQuery) WithAnalyzer(v string) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.Analyzer = v
	}
}

// WithAnalyzeWildcard - specify whether wildcard and prefix queries should be analyzed (default: false).
//
func (f IndicesValidateQuery) WithAnalyzeWildcard(v bool) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.AnalyzeWildcard = &v
	}
}

// WithDefaultOperator - the default operator for query string query (and or or).
//
func (f IndicesValidateQuery) WithDefaultOperator(v string) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.DefaultOperator = v
	}
}

// WithDf - the field to use as default where no field prefix is given in the query string.
//
func (f IndicesValidateQuery) WithDf(v string) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.Df = v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesValidateQuery) WithExpandWildcards(v string) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.ExpandWildcards = v
	}
}

// WithExplain - return detailed information about the error.
//
func (f IndicesValidateQuery) WithExplain(v bool) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.Explain = &v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesValidateQuery) WithIgnoreUnavailable(v bool) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithLenient - specify whether format-based query failures (such as providing text to a numeric field) should be ignored.
//
func (f IndicesValidateQuery) WithLenient(v bool) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.Lenient = &v
	}
}

// WithQuery - query in the lucene query string syntax.
//
func (f IndicesValidateQuery) WithQuery(v string) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.Query = v
	}
}

// WithRewrite - provide a more detailed explanation showing the actual lucene query that will be executed..
//
func (f IndicesValidateQuery) WithRewrite(v bool) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.Rewrite = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesValidateQuery) WithPretty() func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesValidateQuery) WithHuman() func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesValidateQuery) WithErrorTrace() func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesValidateQuery) WithFilterPath(v ...string) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesValidateQuery) WithHeader(h map[string]string) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
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
func (f IndicesValidateQuery) WithOpaqueID(s string) func(*IndicesValidateQueryRequest) {
	return func(r *IndicesValidateQueryRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
