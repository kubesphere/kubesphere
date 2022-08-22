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

func newDeleteFunc(t Transport) Delete {
	return func(index string, id string, o ...func(*DeleteRequest)) (*Response, error) {
		var r = DeleteRequest{Index: index, DocumentID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// Delete removes a document from the index.
//
//
type Delete func(index string, id string, o ...func(*DeleteRequest)) (*Response, error)

// DeleteRequest configures the Delete API request.
//
type DeleteRequest struct {
	Index        string
	DocumentType string
	DocumentID   string

	IfPrimaryTerm       *int
	IfSeqNo             *int
	Refresh             string
	Routing             string
	Timeout             time.Duration
	Version             *int
	VersionType         string
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
func (r DeleteRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	if r.DocumentType == "" {
		r.DocumentType = "_doc"
	}

	path.Grow(1 + len(r.Index) + 1 + len(r.DocumentType) + 1 + len(r.DocumentID))
	path.WriteString("/")
	path.WriteString(r.Index)
	if r.DocumentType != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentType)
	}
	path.WriteString("/")
	path.WriteString(r.DocumentID)

	params = make(map[string]string)

	if r.IfPrimaryTerm != nil {
		params["if_primary_term"] = strconv.FormatInt(int64(*r.IfPrimaryTerm), 10)
	}

	if r.IfSeqNo != nil {
		params["if_seq_no"] = strconv.FormatInt(int64(*r.IfSeqNo), 10)
	}

	if r.Refresh != "" {
		params["refresh"] = r.Refresh
	}

	if r.Routing != "" {
		params["routing"] = r.Routing
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
	}

	if r.Version != nil {
		params["version"] = strconv.FormatInt(int64(*r.Version), 10)
	}

	if r.VersionType != "" {
		params["version_type"] = r.VersionType
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
func (f Delete) WithContext(v context.Context) func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.ctx = v
	}
}

// WithDocumentType - the type of the document.
//
func (f Delete) WithDocumentType(v string) func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.DocumentType = v
	}
}

// WithIfPrimaryTerm - only perform the delete operation if the last operation that has changed the document has the specified primary term.
//
func (f Delete) WithIfPrimaryTerm(v int) func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.IfPrimaryTerm = &v
	}
}

// WithIfSeqNo - only perform the delete operation if the last operation that has changed the document has the specified sequence number.
//
func (f Delete) WithIfSeqNo(v int) func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.IfSeqNo = &v
	}
}

// WithRefresh - if `true` then refresh the affected shards to make this operation visible to search, if `wait_for` then wait for a refresh to make this operation visible to search, if `false` (the default) then do nothing with refreshes..
//
func (f Delete) WithRefresh(v string) func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.Refresh = v
	}
}

// WithRouting - specific routing value.
//
func (f Delete) WithRouting(v string) func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.Routing = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f Delete) WithTimeout(v time.Duration) func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.Timeout = v
	}
}

// WithVersion - explicit version number for concurrency control.
//
func (f Delete) WithVersion(v int) func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.Version = &v
	}
}

// WithVersionType - specific version type.
//
func (f Delete) WithVersionType(v string) func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.VersionType = v
	}
}

// WithWaitForActiveShards - sets the number of shard copies that must be active before proceeding with the delete operation. defaults to 1, meaning the primary shard only. set to `all` for all shard copies, otherwise set to any non-negative value less than or equal to the total number of copies for the shard (number of replicas + 1).
//
func (f Delete) WithWaitForActiveShards(v string) func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.WaitForActiveShards = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Delete) WithPretty() func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Delete) WithHuman() func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Delete) WithErrorTrace() func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Delete) WithFilterPath(v ...string) func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Delete) WithHeader(h map[string]string) func(*DeleteRequest) {
	return func(r *DeleteRequest) {
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
func (f Delete) WithOpaqueID(s string) func(*DeleteRequest) {
	return func(r *DeleteRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
