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
	"time"
)

func newIndicesRolloverFunc(t Transport) IndicesRollover {
	return func(alias string, o ...func(*IndicesRolloverRequest)) (*Response, error) {
		var r = IndicesRolloverRequest{Alias: alias}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesRollover updates an alias to point to a new index when the existing index
// is considered to be too large or too old.
//
//
type IndicesRollover func(alias string, o ...func(*IndicesRolloverRequest)) (*Response, error)

// IndicesRolloverRequest configures the Indices Rollover API request.
//
type IndicesRolloverRequest struct {
	Body io.Reader

	Alias    string
	NewIndex string

	DryRun              *bool
	IncludeTypeName     *bool
	MasterTimeout       time.Duration
	Timeout             time.Duration
	WaitForActiveShards string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesRolloverRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(r.Alias) + 1 + len("_rollover") + 1 + len(r.NewIndex))
	path.WriteString("/")
	path.WriteString(r.Alias)
	path.WriteString("/")
	path.WriteString("_rollover")
	if r.NewIndex != "" {
		path.WriteString("/")
		path.WriteString(r.NewIndex)
	}

	params = make(map[string]string)

	if r.DryRun != nil {
		params["dry_run"] = strconv.FormatBool(*r.DryRun)
	}

	if r.IncludeTypeName != nil {
		params["include_type_name"] = strconv.FormatBool(*r.IncludeTypeName)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
	}

	if r.WaitForActiveShards != "" {
		params["wait_for_active_shards"] = r.WaitForActiveShards
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
func (f IndicesRollover) WithContext(v context.Context) func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
		r.ctx = v
	}
}

// WithBody - The conditions that needs to be met for executing rollover.
//
func (f IndicesRollover) WithBody(v io.Reader) func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
		r.Body = v
	}
}

// WithNewIndex - the name of the rollover index.
//
func (f IndicesRollover) WithNewIndex(v string) func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
		r.NewIndex = v
	}
}

// WithDryRun - if set to true the rollover action will only be validated but not actually performed even if a condition matches. the default is false.
//
func (f IndicesRollover) WithDryRun(v bool) func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
		r.DryRun = &v
	}
}

// WithIncludeTypeName - whether a type should be included in the body of the mappings..
//
func (f IndicesRollover) WithIncludeTypeName(v bool) func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
		r.IncludeTypeName = &v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesRollover) WithMasterTimeout(v time.Duration) func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IndicesRollover) WithTimeout(v time.Duration) func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
		r.Timeout = v
	}
}

// WithWaitForActiveShards - set the number of active shards to wait for on the newly created rollover index before the operation returns..
//
func (f IndicesRollover) WithWaitForActiveShards(v string) func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
		r.WaitForActiveShards = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesRollover) WithPretty() func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesRollover) WithHuman() func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesRollover) WithErrorTrace() func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesRollover) WithFilterPath(v ...string) func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesRollover) WithHeader(h map[string]string) func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
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
func (f IndicesRollover) WithOpaqueID(s string) func(*IndicesRolloverRequest) {
	return func(r *IndicesRolloverRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
