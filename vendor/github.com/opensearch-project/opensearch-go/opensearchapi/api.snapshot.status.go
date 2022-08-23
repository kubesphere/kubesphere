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

func newSnapshotStatusFunc(t Transport) SnapshotStatus {
	return func(o ...func(*SnapshotStatusRequest)) (*Response, error) {
		var r = SnapshotStatusRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SnapshotStatus returns information about the status of a snapshot.
//
//
type SnapshotStatus func(o ...func(*SnapshotStatusRequest)) (*Response, error)

// SnapshotStatusRequest configures the Snapshot Status API request.
//
type SnapshotStatusRequest struct {
	Repository string
	Snapshot   []string

	IgnoreUnavailable *bool
	MasterTimeout     time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SnapshotStatusRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_snapshot") + 1 + len(r.Repository) + 1 + len(strings.Join(r.Snapshot, ",")) + 1 + len("_status"))
	path.WriteString("/")
	path.WriteString("_snapshot")
	if r.Repository != "" {
		path.WriteString("/")
		path.WriteString(r.Repository)
	}
	if len(r.Snapshot) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Snapshot, ","))
	}
	path.WriteString("/")
	path.WriteString("_status")

	params = make(map[string]string)

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
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
func (f SnapshotStatus) WithContext(v context.Context) func(*SnapshotStatusRequest) {
	return func(r *SnapshotStatusRequest) {
		r.ctx = v
	}
}

// WithRepository - a repository name.
//
func (f SnapshotStatus) WithRepository(v string) func(*SnapshotStatusRequest) {
	return func(r *SnapshotStatusRequest) {
		r.Repository = v
	}
}

// WithSnapshot - a list of snapshot names.
//
func (f SnapshotStatus) WithSnapshot(v ...string) func(*SnapshotStatusRequest) {
	return func(r *SnapshotStatusRequest) {
		r.Snapshot = v
	}
}

// WithIgnoreUnavailable - whether to ignore unavailable snapshots, defaults to false which means a snapshotmissingexception is thrown.
//
func (f SnapshotStatus) WithIgnoreUnavailable(v bool) func(*SnapshotStatusRequest) {
	return func(r *SnapshotStatusRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f SnapshotStatus) WithMasterTimeout(v time.Duration) func(*SnapshotStatusRequest) {
	return func(r *SnapshotStatusRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SnapshotStatus) WithPretty() func(*SnapshotStatusRequest) {
	return func(r *SnapshotStatusRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SnapshotStatus) WithHuman() func(*SnapshotStatusRequest) {
	return func(r *SnapshotStatusRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SnapshotStatus) WithErrorTrace() func(*SnapshotStatusRequest) {
	return func(r *SnapshotStatusRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SnapshotStatus) WithFilterPath(v ...string) func(*SnapshotStatusRequest) {
	return func(r *SnapshotStatusRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SnapshotStatus) WithHeader(h map[string]string) func(*SnapshotStatusRequest) {
	return func(r *SnapshotStatusRequest) {
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
func (f SnapshotStatus) WithOpaqueID(s string) func(*SnapshotStatusRequest) {
	return func(r *SnapshotStatusRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
