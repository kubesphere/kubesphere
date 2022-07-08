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

func newSnapshotGetRepositoryFunc(t Transport) SnapshotGetRepository {
	return func(o ...func(*SnapshotGetRepositoryRequest)) (*Response, error) {
		var r = SnapshotGetRepositoryRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SnapshotGetRepository returns information about a repository.
//
//
type SnapshotGetRepository func(o ...func(*SnapshotGetRepositoryRequest)) (*Response, error)

// SnapshotGetRepositoryRequest configures the Snapshot Get Repository API request.
//
type SnapshotGetRepositoryRequest struct {
	Repository []string

	Local         *bool
	MasterTimeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SnapshotGetRepositoryRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_snapshot") + 1 + len(strings.Join(r.Repository, ",")))
	path.WriteString("/")
	path.WriteString("_snapshot")
	if len(r.Repository) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Repository, ","))
	}

	params = make(map[string]string)

	if r.Local != nil {
		params["local"] = strconv.FormatBool(*r.Local)
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
func (f SnapshotGetRepository) WithContext(v context.Context) func(*SnapshotGetRepositoryRequest) {
	return func(r *SnapshotGetRepositoryRequest) {
		r.ctx = v
	}
}

// WithRepository - a list of repository names.
//
func (f SnapshotGetRepository) WithRepository(v ...string) func(*SnapshotGetRepositoryRequest) {
	return func(r *SnapshotGetRepositoryRequest) {
		r.Repository = v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f SnapshotGetRepository) WithLocal(v bool) func(*SnapshotGetRepositoryRequest) {
	return func(r *SnapshotGetRepositoryRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f SnapshotGetRepository) WithMasterTimeout(v time.Duration) func(*SnapshotGetRepositoryRequest) {
	return func(r *SnapshotGetRepositoryRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SnapshotGetRepository) WithPretty() func(*SnapshotGetRepositoryRequest) {
	return func(r *SnapshotGetRepositoryRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SnapshotGetRepository) WithHuman() func(*SnapshotGetRepositoryRequest) {
	return func(r *SnapshotGetRepositoryRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SnapshotGetRepository) WithErrorTrace() func(*SnapshotGetRepositoryRequest) {
	return func(r *SnapshotGetRepositoryRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SnapshotGetRepository) WithFilterPath(v ...string) func(*SnapshotGetRepositoryRequest) {
	return func(r *SnapshotGetRepositoryRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SnapshotGetRepository) WithHeader(h map[string]string) func(*SnapshotGetRepositoryRequest) {
	return func(r *SnapshotGetRepositoryRequest) {
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
func (f SnapshotGetRepository) WithOpaqueID(s string) func(*SnapshotGetRepositoryRequest) {
	return func(r *SnapshotGetRepositoryRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
