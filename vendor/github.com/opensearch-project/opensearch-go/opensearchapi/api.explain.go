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

func newExplainFunc(t Transport) Explain {
	return func(index string, id string, o ...func(*ExplainRequest)) (*Response, error) {
		var r = ExplainRequest{Index: index, DocumentID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// Explain returns information about why a specific matches (or doesn't match) a query.
//
//
type Explain func(index string, id string, o ...func(*ExplainRequest)) (*Response, error)

// ExplainRequest configures the Explain API request.
//
type ExplainRequest struct {
	Index        string
	DocumentType string
	DocumentID   string

	Body io.Reader

	Analyzer        string
	AnalyzeWildcard *bool
	DefaultOperator string
	Df              string
	Lenient         *bool
	Preference      string
	Query           string
	Routing         string
	Source          []string
	SourceExcludes  []string
	SourceIncludes  []string
	StoredFields    []string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r ExplainRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	if r.DocumentType == "" {
		r.DocumentType = "_doc"
	}

	path.Grow(1 + len(r.Index) + 1 + len(r.DocumentType) + 1 + len(r.DocumentID) + 1 + len("_explain"))
	path.WriteString("/")
	path.WriteString(r.Index)
	if r.DocumentType != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentType)
	}
	path.WriteString("/")
	path.WriteString(r.DocumentID)
	path.WriteString("/")
	path.WriteString("_explain")

	params = make(map[string]string)

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

	if r.Lenient != nil {
		params["lenient"] = strconv.FormatBool(*r.Lenient)
	}

	if r.Preference != "" {
		params["preference"] = r.Preference
	}

	if r.Query != "" {
		params["q"] = r.Query
	}

	if r.Routing != "" {
		params["routing"] = r.Routing
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

	if len(r.StoredFields) > 0 {
		params["stored_fields"] = strings.Join(r.StoredFields, ",")
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
func (f Explain) WithContext(v context.Context) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.ctx = v
	}
}

// WithBody - The query definition using the Query DSL.
//
func (f Explain) WithBody(v io.Reader) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.Body = v
	}
}

// WithDocumentType - the type of the document.
//
func (f Explain) WithDocumentType(v string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.DocumentType = v
	}
}

// WithAnalyzer - the analyzer for the query string query.
//
func (f Explain) WithAnalyzer(v string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.Analyzer = v
	}
}

// WithAnalyzeWildcard - specify whether wildcards and prefix queries in the query string query should be analyzed (default: false).
//
func (f Explain) WithAnalyzeWildcard(v bool) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.AnalyzeWildcard = &v
	}
}

// WithDefaultOperator - the default operator for query string query (and or or).
//
func (f Explain) WithDefaultOperator(v string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.DefaultOperator = v
	}
}

// WithDf - the default field for query string query (default: _all).
//
func (f Explain) WithDf(v string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.Df = v
	}
}

// WithLenient - specify whether format-based query failures (such as providing text to a numeric field) should be ignored.
//
func (f Explain) WithLenient(v bool) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.Lenient = &v
	}
}

// WithPreference - specify the node or shard the operation should be performed on (default: random).
//
func (f Explain) WithPreference(v string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.Preference = v
	}
}

// WithQuery - query in the lucene query string syntax.
//
func (f Explain) WithQuery(v string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.Query = v
	}
}

// WithRouting - specific routing value.
//
func (f Explain) WithRouting(v string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.Routing = v
	}
}

// WithSource - true or false to return the _source field or not, or a list of fields to return.
//
func (f Explain) WithSource(v ...string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.Source = v
	}
}

// WithSourceExcludes - a list of fields to exclude from the returned _source field.
//
func (f Explain) WithSourceExcludes(v ...string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.SourceExcludes = v
	}
}

// WithSourceIncludes - a list of fields to extract and return from the _source field.
//
func (f Explain) WithSourceIncludes(v ...string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.SourceIncludes = v
	}
}

// WithStoredFields - a list of stored fields to return in the response.
//
func (f Explain) WithStoredFields(v ...string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.StoredFields = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Explain) WithPretty() func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Explain) WithHuman() func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Explain) WithErrorTrace() func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Explain) WithFilterPath(v ...string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Explain) WithHeader(h map[string]string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
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
func (f Explain) WithOpaqueID(s string) func(*ExplainRequest) {
	return func(r *ExplainRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
