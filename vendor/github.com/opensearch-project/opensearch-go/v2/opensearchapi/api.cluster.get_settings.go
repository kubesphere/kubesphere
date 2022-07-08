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

func newClusterGetSettingsFunc(t Transport) ClusterGetSettings {
	return func(o ...func(*ClusterGetSettingsRequest)) (*Response, error) {
		var r = ClusterGetSettingsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ClusterGetSettings returns cluster settings.
//
//
type ClusterGetSettings func(o ...func(*ClusterGetSettingsRequest)) (*Response, error)

// ClusterGetSettingsRequest configures the Cluster Get Settings API request.
//
type ClusterGetSettingsRequest struct {
	FlatSettings    *bool
	IncludeDefaults *bool
	MasterTimeout   time.Duration
	Timeout         time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r ClusterGetSettingsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_cluster/settings"))
	path.WriteString("/_cluster/settings")

	params = make(map[string]string)

	if r.FlatSettings != nil {
		params["flat_settings"] = strconv.FormatBool(*r.FlatSettings)
	}

	if r.IncludeDefaults != nil {
		params["include_defaults"] = strconv.FormatBool(*r.IncludeDefaults)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
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
func (f ClusterGetSettings) WithContext(v context.Context) func(*ClusterGetSettingsRequest) {
	return func(r *ClusterGetSettingsRequest) {
		r.ctx = v
	}
}

// WithFlatSettings - return settings in flat format (default: false).
//
func (f ClusterGetSettings) WithFlatSettings(v bool) func(*ClusterGetSettingsRequest) {
	return func(r *ClusterGetSettingsRequest) {
		r.FlatSettings = &v
	}
}

// WithIncludeDefaults - whether to return all default clusters setting..
//
func (f ClusterGetSettings) WithIncludeDefaults(v bool) func(*ClusterGetSettingsRequest) {
	return func(r *ClusterGetSettingsRequest) {
		r.IncludeDefaults = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f ClusterGetSettings) WithMasterTimeout(v time.Duration) func(*ClusterGetSettingsRequest) {
	return func(r *ClusterGetSettingsRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f ClusterGetSettings) WithTimeout(v time.Duration) func(*ClusterGetSettingsRequest) {
	return func(r *ClusterGetSettingsRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ClusterGetSettings) WithPretty() func(*ClusterGetSettingsRequest) {
	return func(r *ClusterGetSettingsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ClusterGetSettings) WithHuman() func(*ClusterGetSettingsRequest) {
	return func(r *ClusterGetSettingsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ClusterGetSettings) WithErrorTrace() func(*ClusterGetSettingsRequest) {
	return func(r *ClusterGetSettingsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ClusterGetSettings) WithFilterPath(v ...string) func(*ClusterGetSettingsRequest) {
	return func(r *ClusterGetSettingsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f ClusterGetSettings) WithHeader(h map[string]string) func(*ClusterGetSettingsRequest) {
	return func(r *ClusterGetSettingsRequest) {
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
func (f ClusterGetSettings) WithOpaqueID(s string) func(*ClusterGetSettingsRequest) {
	return func(r *ClusterGetSettingsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
