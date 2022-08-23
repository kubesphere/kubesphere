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

func newMsearchFunc(t Transport) Msearch {
	return func(body io.Reader, o ...func(*MsearchRequest)) (*Response, error) {
		var r = MsearchRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// Msearch allows to execute several search operations in one request.
//
//
type Msearch func(body io.Reader, o ...func(*MsearchRequest)) (*Response, error)

// MsearchRequest configures the Msearch API request.
//
type MsearchRequest struct {
	Index        []string
	DocumentType []string

	Body io.Reader

	CcsMinimizeRoundtrips      *bool
	MaxConcurrentSearches      *int
	MaxConcurrentShardRequests *int
	PreFilterShardSize         *int
	RestTotalHitsAsInt         *bool
	SearchType                 string
	TypedKeys                  *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MsearchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len(strings.Join(r.DocumentType, ",")) + 1 + len("_msearch"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	if len(r.DocumentType) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.DocumentType, ","))
	}
	path.WriteString("/")
	path.WriteString("_msearch")

	params = make(map[string]string)

	if r.CcsMinimizeRoundtrips != nil {
		params["ccs_minimize_roundtrips"] = strconv.FormatBool(*r.CcsMinimizeRoundtrips)
	}

	if r.MaxConcurrentSearches != nil {
		params["max_concurrent_searches"] = strconv.FormatInt(int64(*r.MaxConcurrentSearches), 10)
	}

	if r.MaxConcurrentShardRequests != nil {
		params["max_concurrent_shard_requests"] = strconv.FormatInt(int64(*r.MaxConcurrentShardRequests), 10)
	}

	if r.PreFilterShardSize != nil {
		params["pre_filter_shard_size"] = strconv.FormatInt(int64(*r.PreFilterShardSize), 10)
	}

	if r.RestTotalHitsAsInt != nil {
		params["rest_total_hits_as_int"] = strconv.FormatBool(*r.RestTotalHitsAsInt)
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
func (f Msearch) WithContext(v context.Context) func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names to use as default.
//
func (f Msearch) WithIndex(v ...string) func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.Index = v
	}
}

// WithDocumentType - a list of document types to use as default.
//
func (f Msearch) WithDocumentType(v ...string) func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.DocumentType = v
	}
}

// WithCcsMinimizeRoundtrips - indicates whether network round-trips should be minimized as part of cross-cluster search requests execution.
//
func (f Msearch) WithCcsMinimizeRoundtrips(v bool) func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.CcsMinimizeRoundtrips = &v
	}
}

// WithMaxConcurrentSearches - controls the maximum number of concurrent searches the multi search api will execute.
//
func (f Msearch) WithMaxConcurrentSearches(v int) func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.MaxConcurrentSearches = &v
	}
}

// WithMaxConcurrentShardRequests - the number of concurrent shard requests each sub search executes concurrently per node. this value should be used to limit the impact of the search on the cluster in order to limit the number of concurrent shard requests.
//
func (f Msearch) WithMaxConcurrentShardRequests(v int) func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.MaxConcurrentShardRequests = &v
	}
}

// WithPreFilterShardSize - a threshold that enforces a pre-filter roundtrip to prefilter search shards based on query rewriting if the number of shards the search request expands to exceeds the threshold. this filter roundtrip can limit the number of shards significantly if for instance a shard can not match any documents based on its rewrite method ie. if date filters are mandatory to match but the shard bounds and the query are disjoint..
//
func (f Msearch) WithPreFilterShardSize(v int) func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.PreFilterShardSize = &v
	}
}

// WithRestTotalHitsAsInt - indicates whether hits.total should be rendered as an integer or an object in the rest search response.
//
func (f Msearch) WithRestTotalHitsAsInt(v bool) func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.RestTotalHitsAsInt = &v
	}
}

// WithSearchType - search operation type.
//
func (f Msearch) WithSearchType(v string) func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.SearchType = v
	}
}

// WithTypedKeys - specify whether aggregation and suggester names should be prefixed by their respective types in the response.
//
func (f Msearch) WithTypedKeys(v bool) func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.TypedKeys = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Msearch) WithPretty() func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Msearch) WithHuman() func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Msearch) WithErrorTrace() func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Msearch) WithFilterPath(v ...string) func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Msearch) WithHeader(h map[string]string) func(*MsearchRequest) {
	return func(r *MsearchRequest) {
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
func (f Msearch) WithOpaqueID(s string) func(*MsearchRequest) {
	return func(r *MsearchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
