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

func newReindexFunc(t Transport) Reindex {
	return func(body io.Reader, o ...func(*ReindexRequest)) (*Response, error) {
		var r = ReindexRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// Reindex allows to copy documents from one index to another, optionally filtering the source
// documents by a query, changing the destination index settings, or fetching the
// documents from a remote cluster.
//
//
type Reindex func(body io.Reader, o ...func(*ReindexRequest)) (*Response, error)

// ReindexRequest configures the Reindex API request.
//
type ReindexRequest struct {
	Body io.Reader

	MaxDocs             *int
	Refresh             *bool
	RequestsPerSecond   *int
	Scroll              time.Duration
	Slices              interface{}
	Timeout             time.Duration
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
func (r ReindexRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(len("/_reindex"))
	path.WriteString("/_reindex")

	params = make(map[string]string)

	if r.MaxDocs != nil {
		params["max_docs"] = strconv.FormatInt(int64(*r.MaxDocs), 10)
	}

	if r.Refresh != nil {
		params["refresh"] = strconv.FormatBool(*r.Refresh)
	}

	if r.RequestsPerSecond != nil {
		params["requests_per_second"] = strconv.FormatInt(int64(*r.RequestsPerSecond), 10)
	}

	if r.Scroll != 0 {
		params["scroll"] = formatDuration(r.Scroll)
	}

	if r.Slices != nil {
		params["slices"] = fmt.Sprintf("%v", r.Slices)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
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
func (f Reindex) WithContext(v context.Context) func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		r.ctx = v
	}
}

// WithMaxDocs - maximum number of documents to process (default: all documents).
//
func (f Reindex) WithMaxDocs(v int) func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		r.MaxDocs = &v
	}
}

// WithRefresh - should the affected indexes be refreshed?.
//
func (f Reindex) WithRefresh(v bool) func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		r.Refresh = &v
	}
}

// WithRequestsPerSecond - the throttle to set on this request in sub-requests per second. -1 means no throttle..
//
func (f Reindex) WithRequestsPerSecond(v int) func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		r.RequestsPerSecond = &v
	}
}

// WithScroll - control how long to keep the search context alive.
//
func (f Reindex) WithScroll(v time.Duration) func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		r.Scroll = v
	}
}

// WithSlices - the number of slices this task should be divided into. defaults to 1, meaning the task isn't sliced into subtasks. can be set to `auto`..
//
func (f Reindex) WithSlices(v interface{}) func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		r.Slices = v
	}
}

// WithTimeout - time each individual bulk request should wait for shards that are unavailable..
//
func (f Reindex) WithTimeout(v time.Duration) func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		r.Timeout = v
	}
}

// WithWaitForActiveShards - sets the number of shard copies that must be active before proceeding with the reindex operation. defaults to 1, meaning the primary shard only. set to `all` for all shard copies, otherwise set to any non-negative value less than or equal to the total number of copies for the shard (number of replicas + 1).
//
func (f Reindex) WithWaitForActiveShards(v string) func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		r.WaitForActiveShards = v
	}
}

// WithWaitForCompletion - should the request should block until the reindex is complete..
//
func (f Reindex) WithWaitForCompletion(v bool) func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		r.WaitForCompletion = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Reindex) WithPretty() func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Reindex) WithHuman() func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Reindex) WithErrorTrace() func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Reindex) WithFilterPath(v ...string) func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Reindex) WithHeader(h map[string]string) func(*ReindexRequest) {
	return func(r *ReindexRequest) {
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
func (f Reindex) WithOpaqueID(s string) func(*ReindexRequest) {
	return func(r *ReindexRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
