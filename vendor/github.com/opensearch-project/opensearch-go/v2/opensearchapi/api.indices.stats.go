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
	"net/http"
	"strconv"
	"strings"
)

func newIndicesStatsFunc(t Transport) IndicesStats {
	return func(o ...func(*IndicesStatsRequest)) (*Response, error) {
		var r = IndicesStatsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesStats provides statistics on operations happening in an index.
//
//
type IndicesStats func(o ...func(*IndicesStatsRequest)) (*Response, error)

// IndicesStatsRequest configures the Indices Stats API request.
//
type IndicesStatsRequest struct {
	Index []string

	Metric []string

	CompletionFields        []string
	ExpandWildcards         string
	FielddataFields         []string
	Fields                  []string
	ForbidClosedIndices     *bool
	Groups                  []string
	IncludeSegmentFileSizes *bool
	IncludeUnloadedSegments *bool
	Level                   string
	Types                   []string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesStatsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_stats") + 1 + len(strings.Join(r.Metric, ",")))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_stats")
	if len(r.Metric) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Metric, ","))
	}

	params = make(map[string]string)

	if len(r.CompletionFields) > 0 {
		params["completion_fields"] = strings.Join(r.CompletionFields, ",")
	}

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if len(r.FielddataFields) > 0 {
		params["fielddata_fields"] = strings.Join(r.FielddataFields, ",")
	}

	if len(r.Fields) > 0 {
		params["fields"] = strings.Join(r.Fields, ",")
	}

	if r.ForbidClosedIndices != nil {
		params["forbid_closed_indices"] = strconv.FormatBool(*r.ForbidClosedIndices)
	}

	if len(r.Groups) > 0 {
		params["groups"] = strings.Join(r.Groups, ",")
	}

	if r.IncludeSegmentFileSizes != nil {
		params["include_segment_file_sizes"] = strconv.FormatBool(*r.IncludeSegmentFileSizes)
	}

	if r.IncludeUnloadedSegments != nil {
		params["include_unloaded_segments"] = strconv.FormatBool(*r.IncludeUnloadedSegments)
	}

	if r.Level != "" {
		params["level"] = r.Level
	}

	if len(r.Types) > 0 {
		params["types"] = strings.Join(r.Types, ",")
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

	req, err := newRequest(method, path.String(), nil)
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
func (f IndicesStats) WithContext(v context.Context) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names; use _all to perform the operation on all indices.
//
func (f IndicesStats) WithIndex(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Index = v
	}
}

// WithMetric - limit the information returned the specific metrics..
//
func (f IndicesStats) WithMetric(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Metric = v
	}
}

// WithCompletionFields - a list of fields for `fielddata` and `suggest` index metric (supports wildcards).
//
func (f IndicesStats) WithCompletionFields(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.CompletionFields = v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesStats) WithExpandWildcards(v string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.ExpandWildcards = v
	}
}

// WithFielddataFields - a list of fields for `fielddata` index metric (supports wildcards).
//
func (f IndicesStats) WithFielddataFields(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.FielddataFields = v
	}
}

// WithFields - a list of fields for `fielddata` and `completion` index metric (supports wildcards).
//
func (f IndicesStats) WithFields(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Fields = v
	}
}

// WithForbidClosedIndices - if set to false stats will also collected from closed indices if explicitly specified or if expand_wildcards expands to closed indices.
//
func (f IndicesStats) WithForbidClosedIndices(v bool) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.ForbidClosedIndices = &v
	}
}

// WithGroups - a list of search groups for `search` index metric.
//
func (f IndicesStats) WithGroups(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Groups = v
	}
}

// WithIncludeSegmentFileSizes - whether to report the aggregated disk usage of each one of the lucene index files (only applies if segment stats are requested).
//
func (f IndicesStats) WithIncludeSegmentFileSizes(v bool) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.IncludeSegmentFileSizes = &v
	}
}

// WithIncludeUnloadedSegments - if set to true segment stats will include stats for segments that are not currently loaded into memory.
//
func (f IndicesStats) WithIncludeUnloadedSegments(v bool) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.IncludeUnloadedSegments = &v
	}
}

// WithLevel - return stats aggregated at cluster, index or shard level.
//
func (f IndicesStats) WithLevel(v string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Level = v
	}
}

// WithTypes - a list of document types for the `indexing` index metric.
//
func (f IndicesStats) WithTypes(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Types = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesStats) WithPretty() func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesStats) WithHuman() func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesStats) WithErrorTrace() func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesStats) WithFilterPath(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesStats) WithHeader(h map[string]string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
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
func (f IndicesStats) WithOpaqueID(s string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
