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
	"strings"
	"time"
)

func newClusterPostVotingConfigExclusionsFunc(t Transport) ClusterPostVotingConfigExclusions {
	return func(o ...func(*ClusterPostVotingConfigExclusionsRequest)) (*Response, error) {
		var r = ClusterPostVotingConfigExclusionsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ClusterPostVotingConfigExclusions updates the cluster voting config exclusions by node ids or node names.
//
//
type ClusterPostVotingConfigExclusions func(o ...func(*ClusterPostVotingConfigExclusionsRequest)) (*Response, error)

// ClusterPostVotingConfigExclusionsRequest configures the Cluster Post Voting Config Exclusions API request.
//
type ClusterPostVotingConfigExclusionsRequest struct {
	NodeIds   string
	NodeNames string
	Timeout   time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r ClusterPostVotingConfigExclusionsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(len("/_cluster/voting_config_exclusions"))
	path.WriteString("/_cluster/voting_config_exclusions")

	params = make(map[string]string)

	if r.NodeIds != "" {
		params["node_ids"] = r.NodeIds
	}

	if r.NodeNames != "" {
		params["node_names"] = r.NodeNames
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
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
func (f ClusterPostVotingConfigExclusions) WithContext(v context.Context) func(*ClusterPostVotingConfigExclusionsRequest) {
	return func(r *ClusterPostVotingConfigExclusionsRequest) {
		r.ctx = v
	}
}

// WithNodeIds - a list of the persistent ids of the nodes to exclude from the voting configuration. if specified, you may not also specify ?node_names..
//
func (f ClusterPostVotingConfigExclusions) WithNodeIds(v string) func(*ClusterPostVotingConfigExclusionsRequest) {
	return func(r *ClusterPostVotingConfigExclusionsRequest) {
		r.NodeIds = v
	}
}

// WithNodeNames - a list of the names of the nodes to exclude from the voting configuration. if specified, you may not also specify ?node_ids..
//
func (f ClusterPostVotingConfigExclusions) WithNodeNames(v string) func(*ClusterPostVotingConfigExclusionsRequest) {
	return func(r *ClusterPostVotingConfigExclusionsRequest) {
		r.NodeNames = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f ClusterPostVotingConfigExclusions) WithTimeout(v time.Duration) func(*ClusterPostVotingConfigExclusionsRequest) {
	return func(r *ClusterPostVotingConfigExclusionsRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ClusterPostVotingConfigExclusions) WithPretty() func(*ClusterPostVotingConfigExclusionsRequest) {
	return func(r *ClusterPostVotingConfigExclusionsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ClusterPostVotingConfigExclusions) WithHuman() func(*ClusterPostVotingConfigExclusionsRequest) {
	return func(r *ClusterPostVotingConfigExclusionsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ClusterPostVotingConfigExclusions) WithErrorTrace() func(*ClusterPostVotingConfigExclusionsRequest) {
	return func(r *ClusterPostVotingConfigExclusionsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ClusterPostVotingConfigExclusions) WithFilterPath(v ...string) func(*ClusterPostVotingConfigExclusionsRequest) {
	return func(r *ClusterPostVotingConfigExclusionsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f ClusterPostVotingConfigExclusions) WithHeader(h map[string]string) func(*ClusterPostVotingConfigExclusionsRequest) {
	return func(r *ClusterPostVotingConfigExclusionsRequest) {
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
func (f ClusterPostVotingConfigExclusions) WithOpaqueID(s string) func(*ClusterPostVotingConfigExclusionsRequest) {
	return func(r *ClusterPostVotingConfigExclusionsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
