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
	"strings"
	"time"
)

func newNodesReloadSecureSettingsFunc(t Transport) NodesReloadSecureSettings {
	return func(o ...func(*NodesReloadSecureSettingsRequest)) (*Response, error) {
		var r = NodesReloadSecureSettingsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// NodesReloadSecureSettings reloads secure settings.
//
//
type NodesReloadSecureSettings func(o ...func(*NodesReloadSecureSettingsRequest)) (*Response, error)

// NodesReloadSecureSettingsRequest configures the Nodes Reload Secure Settings API request.
//
type NodesReloadSecureSettingsRequest struct {
	Body io.Reader

	NodeID []string

	Timeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r NodesReloadSecureSettingsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_nodes") + 1 + len(strings.Join(r.NodeID, ",")) + 1 + len("reload_secure_settings"))
	path.WriteString("/")
	path.WriteString("_nodes")
	if len(r.NodeID) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.NodeID, ","))
	}
	path.WriteString("/")
	path.WriteString("reload_secure_settings")

	params = make(map[string]string)

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
func (f NodesReloadSecureSettings) WithContext(v context.Context) func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.ctx = v
	}
}

// WithBody - An object containing the password for the opensearch keystore.
//
func (f NodesReloadSecureSettings) WithBody(v io.Reader) func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.Body = v
	}
}

// WithNodeID - a list of node ids to span the reload/reinit call. should stay empty because reloading usually involves all cluster nodes..
//
func (f NodesReloadSecureSettings) WithNodeID(v ...string) func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.NodeID = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f NodesReloadSecureSettings) WithTimeout(v time.Duration) func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f NodesReloadSecureSettings) WithPretty() func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f NodesReloadSecureSettings) WithHuman() func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f NodesReloadSecureSettings) WithErrorTrace() func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f NodesReloadSecureSettings) WithFilterPath(v ...string) func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f NodesReloadSecureSettings) WithHeader(h map[string]string) func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
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
func (f NodesReloadSecureSettings) WithOpaqueID(s string) func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
