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

func newClusterHealthFunc(t Transport) ClusterHealth {
	return func(o ...func(*ClusterHealthRequest)) (*Response, error) {
		var r = ClusterHealthRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ClusterHealth returns basic information about the health of the cluster.
//
//
type ClusterHealth func(o ...func(*ClusterHealthRequest)) (*Response, error)

// ClusterHealthRequest configures the Cluster Health API request.
//
type ClusterHealthRequest struct {
	Index []string

	ExpandWildcards             string
	Level                       string
	Local                       *bool
	MasterTimeout               time.Duration
	Timeout                     time.Duration
	WaitForActiveShards         string
	WaitForEvents               string
	WaitForNoInitializingShards *bool
	WaitForNoRelocatingShards   *bool
	WaitForNodes                string
	WaitForStatus               string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r ClusterHealthRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_cluster") + 1 + len("health") + 1 + len(strings.Join(r.Index, ",")))
	path.WriteString("/")
	path.WriteString("_cluster")
	path.WriteString("/")
	path.WriteString("health")
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}

	params = make(map[string]string)

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if r.Level != "" {
		params["level"] = r.Level
	}

	if r.Local != nil {
		params["local"] = strconv.FormatBool(*r.Local)
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

	if r.WaitForEvents != "" {
		params["wait_for_events"] = r.WaitForEvents
	}

	if r.WaitForNoInitializingShards != nil {
		params["wait_for_no_initializing_shards"] = strconv.FormatBool(*r.WaitForNoInitializingShards)
	}

	if r.WaitForNoRelocatingShards != nil {
		params["wait_for_no_relocating_shards"] = strconv.FormatBool(*r.WaitForNoRelocatingShards)
	}

	if r.WaitForNodes != "" {
		params["wait_for_nodes"] = r.WaitForNodes
	}

	if r.WaitForStatus != "" {
		params["wait_for_status"] = r.WaitForStatus
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
func (f ClusterHealth) WithContext(v context.Context) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.ctx = v
	}
}

// WithIndex - limit the information returned to a specific index.
//
func (f ClusterHealth) WithIndex(v ...string) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.Index = v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f ClusterHealth) WithExpandWildcards(v string) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.ExpandWildcards = v
	}
}

// WithLevel - specify the level of detail for returned information.
//
func (f ClusterHealth) WithLevel(v string) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.Level = v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f ClusterHealth) WithLocal(v bool) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f ClusterHealth) WithMasterTimeout(v time.Duration) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f ClusterHealth) WithTimeout(v time.Duration) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.Timeout = v
	}
}

// WithWaitForActiveShards - wait until the specified number of shards is active.
//
func (f ClusterHealth) WithWaitForActiveShards(v string) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.WaitForActiveShards = v
	}
}

// WithWaitForEvents - wait until all currently queued events with the given priority are processed.
//
func (f ClusterHealth) WithWaitForEvents(v string) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.WaitForEvents = v
	}
}

// WithWaitForNoInitializingShards - whether to wait until there are no initializing shards in the cluster.
//
func (f ClusterHealth) WithWaitForNoInitializingShards(v bool) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.WaitForNoInitializingShards = &v
	}
}

// WithWaitForNoRelocatingShards - whether to wait until there are no relocating shards in the cluster.
//
func (f ClusterHealth) WithWaitForNoRelocatingShards(v bool) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.WaitForNoRelocatingShards = &v
	}
}

// WithWaitForNodes - wait until the specified number of nodes is available.
//
func (f ClusterHealth) WithWaitForNodes(v string) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.WaitForNodes = v
	}
}

// WithWaitForStatus - wait until cluster is in a specific state.
//
func (f ClusterHealth) WithWaitForStatus(v string) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.WaitForStatus = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ClusterHealth) WithPretty() func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ClusterHealth) WithHuman() func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ClusterHealth) WithErrorTrace() func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ClusterHealth) WithFilterPath(v ...string) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f ClusterHealth) WithHeader(h map[string]string) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
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
func (f ClusterHealth) WithOpaqueID(s string) func(*ClusterHealthRequest) {
	return func(r *ClusterHealthRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
