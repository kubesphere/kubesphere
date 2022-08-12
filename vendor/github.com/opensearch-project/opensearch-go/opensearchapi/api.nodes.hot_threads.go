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

func newNodesHotThreadsFunc(t Transport) NodesHotThreads {
	return func(o ...func(*NodesHotThreadsRequest)) (*Response, error) {
		var r = NodesHotThreadsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// NodesHotThreads returns information about hot threads on each node in the cluster.
//
//
type NodesHotThreads func(o ...func(*NodesHotThreadsRequest)) (*Response, error)

// NodesHotThreadsRequest configures the Nodes Hot Threads API request.
//
type NodesHotThreadsRequest struct {
	NodeID []string

	IgnoreIdleThreads *bool
	Interval          time.Duration
	Snapshots         *int
	Threads           *int
	Timeout           time.Duration
	DocumentType      string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r NodesHotThreadsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_cluster") + 1 + len("nodes") + 1 + len(strings.Join(r.NodeID, ",")) + 1 + len("hot_threads"))
	path.WriteString("/")
	path.WriteString("_cluster")
	path.WriteString("/")
	path.WriteString("nodes")
	if len(r.NodeID) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.NodeID, ","))
	}
	path.WriteString("/")
	path.WriteString("hot_threads")

	params = make(map[string]string)

	if r.IgnoreIdleThreads != nil {
		params["ignore_idle_threads"] = strconv.FormatBool(*r.IgnoreIdleThreads)
	}

	if r.Interval != 0 {
		params["interval"] = formatDuration(r.Interval)
	}

	if r.Snapshots != nil {
		params["snapshots"] = strconv.FormatInt(int64(*r.Snapshots), 10)
	}

	if r.Threads != nil {
		params["threads"] = strconv.FormatInt(int64(*r.Threads), 10)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
	}

	if r.DocumentType != "" {
		params["type"] = r.DocumentType
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
func (f NodesHotThreads) WithContext(v context.Context) func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
		r.ctx = v
	}
}

// WithNodeID - a list of node ids or names to limit the returned information; use `_local` to return information from the node you're connecting to, leave empty to get information from all nodes.
//
func (f NodesHotThreads) WithNodeID(v ...string) func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
		r.NodeID = v
	}
}

// WithIgnoreIdleThreads - don't show threads that are in known-idle places, such as waiting on a socket select or pulling from an empty task queue (default: true).
//
func (f NodesHotThreads) WithIgnoreIdleThreads(v bool) func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
		r.IgnoreIdleThreads = &v
	}
}

// WithInterval - the interval for the second sampling of threads.
//
func (f NodesHotThreads) WithInterval(v time.Duration) func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
		r.Interval = v
	}
}

// WithSnapshots - number of samples of thread stacktrace (default: 10).
//
func (f NodesHotThreads) WithSnapshots(v int) func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
		r.Snapshots = &v
	}
}

// WithThreads - specify the number of threads to provide information for (default: 3).
//
func (f NodesHotThreads) WithThreads(v int) func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
		r.Threads = &v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f NodesHotThreads) WithTimeout(v time.Duration) func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
		r.Timeout = v
	}
}

// WithDocumentType - the type to sample (default: cpu).
//
func (f NodesHotThreads) WithDocumentType(v string) func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
		r.DocumentType = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f NodesHotThreads) WithPretty() func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f NodesHotThreads) WithHuman() func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f NodesHotThreads) WithErrorTrace() func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f NodesHotThreads) WithFilterPath(v ...string) func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f NodesHotThreads) WithHeader(h map[string]string) func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
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
func (f NodesHotThreads) WithOpaqueID(s string) func(*NodesHotThreadsRequest) {
	return func(r *NodesHotThreadsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
