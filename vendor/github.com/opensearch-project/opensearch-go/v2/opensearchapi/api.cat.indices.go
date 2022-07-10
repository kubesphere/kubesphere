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
	"time"
)

func newCatIndicesFunc(t Transport) CatIndices {
	return func(o ...func(*CatIndicesRequest)) (*Response, error) {
		var r = CatIndicesRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CatIndices returns information about indices: number of primaries and replicas, document counts, disk size, ...
//
//
type CatIndices func(o ...func(*CatIndicesRequest)) (*Response, error)

// CatIndicesRequest configures the Cat Indices API request.
//
type CatIndicesRequest struct {
	Index []string

	Bytes                   string
	ExpandWildcards         string
	Format                  string
	H                       []string
	Health                  string
	Help                    *bool
	IncludeUnloadedSegments *bool
	Local                   *bool
	MasterTimeout           time.Duration
	Pri                     *bool
	S                       []string
	Time                    string
	V                       *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CatIndicesRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_cat") + 1 + len("indices") + 1 + len(strings.Join(r.Index, ",")))
	path.WriteString("/")
	path.WriteString("_cat")
	path.WriteString("/")
	path.WriteString("indices")
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}

	params = make(map[string]string)

	if r.Bytes != "" {
		params["bytes"] = r.Bytes
	}

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if r.Format != "" {
		params["format"] = r.Format
	}

	if len(r.H) > 0 {
		params["h"] = strings.Join(r.H, ",")
	}

	if r.Health != "" {
		params["health"] = r.Health
	}

	if r.Help != nil {
		params["help"] = strconv.FormatBool(*r.Help)
	}

	if r.IncludeUnloadedSegments != nil {
		params["include_unloaded_segments"] = strconv.FormatBool(*r.IncludeUnloadedSegments)
	}

	if r.Local != nil {
		params["local"] = strconv.FormatBool(*r.Local)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.Pri != nil {
		params["pri"] = strconv.FormatBool(*r.Pri)
	}

	if len(r.S) > 0 {
		params["s"] = strings.Join(r.S, ",")
	}

	if r.Time != "" {
		params["time"] = r.Time
	}

	if r.V != nil {
		params["v"] = strconv.FormatBool(*r.V)
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
func (f CatIndices) WithContext(v context.Context) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names to limit the returned information.
//
func (f CatIndices) WithIndex(v ...string) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.Index = v
	}
}

// WithBytes - the unit in which to display byte values.
//
func (f CatIndices) WithBytes(v string) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.Bytes = v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f CatIndices) WithExpandWildcards(v string) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.ExpandWildcards = v
	}
}

// WithFormat - a short version of the accept header, e.g. json, yaml.
//
func (f CatIndices) WithFormat(v string) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.Format = v
	}
}

// WithH - comma-separated list of column names to display.
//
func (f CatIndices) WithH(v ...string) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.H = v
	}
}

// WithHealth - a health status ("green", "yellow", or "red" to filter only indices matching the specified health status.
//
func (f CatIndices) WithHealth(v string) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.Health = v
	}
}

// WithHelp - return help information.
//
func (f CatIndices) WithHelp(v bool) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.Help = &v
	}
}

// WithIncludeUnloadedSegments - if set to true segment stats will include stats for segments that are not currently loaded into memory.
//
func (f CatIndices) WithIncludeUnloadedSegments(v bool) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.IncludeUnloadedSegments = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f CatIndices) WithLocal(v bool) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f CatIndices) WithMasterTimeout(v time.Duration) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.MasterTimeout = v
	}
}

// WithPri - set to true to return stats only for primary shards.
//
func (f CatIndices) WithPri(v bool) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.Pri = &v
	}
}

// WithS - comma-separated list of column names or column aliases to sort by.
//
func (f CatIndices) WithS(v ...string) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.S = v
	}
}

// WithTime - the unit in which to display time values.
//
func (f CatIndices) WithTime(v string) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.Time = v
	}
}

// WithV - verbose mode. display column headers.
//
func (f CatIndices) WithV(v bool) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.V = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CatIndices) WithPretty() func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CatIndices) WithHuman() func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CatIndices) WithErrorTrace() func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CatIndices) WithFilterPath(v ...string) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CatIndices) WithHeader(h map[string]string) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
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
func (f CatIndices) WithOpaqueID(s string) func(*CatIndicesRequest) {
	return func(r *CatIndicesRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
