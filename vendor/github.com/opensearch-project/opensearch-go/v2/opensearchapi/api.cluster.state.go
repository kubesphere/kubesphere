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

func newClusterStateFunc(t Transport) ClusterState {
	return func(o ...func(*ClusterStateRequest)) (*Response, error) {
		var r = ClusterStateRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ClusterState returns a comprehensive information about the state of the cluster.
//
//
type ClusterState func(o ...func(*ClusterStateRequest)) (*Response, error)

// ClusterStateRequest configures the Cluster State API request.
//
type ClusterStateRequest struct {
	Index []string

	Metric []string

	AllowNoIndices         *bool
	ExpandWildcards        string
	FlatSettings           *bool
	IgnoreUnavailable      *bool
	Local                  *bool
	MasterTimeout          time.Duration
	WaitForMetadataVersion *int
	WaitForTimeout         time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r ClusterStateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_cluster") + 1 + len("state") + 1 + len(strings.Join(r.Metric, ",")) + 1 + len(strings.Join(r.Index, ",")))
	path.WriteString("/")
	path.WriteString("_cluster")
	path.WriteString("/")
	path.WriteString("state")
	if len(r.Metric) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Metric, ","))
	}
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}

	params = make(map[string]string)

	if r.AllowNoIndices != nil {
		params["allow_no_indices"] = strconv.FormatBool(*r.AllowNoIndices)
	}

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if r.FlatSettings != nil {
		params["flat_settings"] = strconv.FormatBool(*r.FlatSettings)
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.Local != nil {
		params["local"] = strconv.FormatBool(*r.Local)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.WaitForMetadataVersion != nil {
		params["wait_for_metadata_version"] = strconv.FormatInt(int64(*r.WaitForMetadataVersion), 10)
	}

	if r.WaitForTimeout != 0 {
		params["wait_for_timeout"] = formatDuration(r.WaitForTimeout)
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
func (f ClusterState) WithContext(v context.Context) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names; use _all to perform the operation on all indices.
//
func (f ClusterState) WithIndex(v ...string) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.Index = v
	}
}

// WithMetric - limit the information returned to the specified metrics.
//
func (f ClusterState) WithMetric(v ...string) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.Metric = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f ClusterState) WithAllowNoIndices(v bool) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f ClusterState) WithExpandWildcards(v string) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.ExpandWildcards = v
	}
}

// WithFlatSettings - return settings in flat format (default: false).
//
func (f ClusterState) WithFlatSettings(v bool) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.FlatSettings = &v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f ClusterState) WithIgnoreUnavailable(v bool) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f ClusterState) WithLocal(v bool) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f ClusterState) WithMasterTimeout(v time.Duration) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.MasterTimeout = v
	}
}

// WithWaitForMetadataVersion - wait for the metadata version to be equal or greater than the specified metadata version.
//
func (f ClusterState) WithWaitForMetadataVersion(v int) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.WaitForMetadataVersion = &v
	}
}

// WithWaitForTimeout - the maximum time to wait for wait_for_metadata_version before timing out.
//
func (f ClusterState) WithWaitForTimeout(v time.Duration) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.WaitForTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ClusterState) WithPretty() func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ClusterState) WithHuman() func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ClusterState) WithErrorTrace() func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ClusterState) WithFilterPath(v ...string) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f ClusterState) WithHeader(h map[string]string) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
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
func (f ClusterState) WithOpaqueID(s string) func(*ClusterStateRequest) {
	return func(r *ClusterStateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
