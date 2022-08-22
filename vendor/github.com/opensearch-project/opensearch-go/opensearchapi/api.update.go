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
	"strconv"
	"strings"
	"time"
)

func newUpdateFunc(t Transport) Update {
	return func(index string, id string, body io.Reader, o ...func(*UpdateRequest)) (*Response, error) {
		var r = UpdateRequest{Index: index, DocumentID: id, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// Update updates a document with a script or partial document.
//
//
type Update func(index string, id string, body io.Reader, o ...func(*UpdateRequest)) (*Response, error)

// UpdateRequest configures the Update API request.
//
type UpdateRequest struct {
	Index        string
	DocumentType string
	DocumentID   string

	Body io.Reader

	IfPrimaryTerm       *int
	IfSeqNo             *int
	Lang                string
	Refresh             string
	RequireAlias        *bool
	RetryOnConflict     *int
	Routing             string
	Source              []string
	SourceExcludes      []string
	SourceIncludes      []string
	Timeout             time.Duration
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
func (r UpdateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	if r.DocumentType == "" {
		r.DocumentType = "_doc"
	}

	path.Grow(1 + len(r.Index) + 1 + len(r.DocumentType) + 1 + len(r.DocumentID) + 1 + len("_update"))
	path.WriteString("/")
	path.WriteString(r.Index)
	if r.DocumentType != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentType)
	}
	path.WriteString("/")
	path.WriteString(r.DocumentID)
	path.WriteString("/")
	path.WriteString("_update")

	params = make(map[string]string)

	if r.IfPrimaryTerm != nil {
		params["if_primary_term"] = strconv.FormatInt(int64(*r.IfPrimaryTerm), 10)
	}

	if r.IfSeqNo != nil {
		params["if_seq_no"] = strconv.FormatInt(int64(*r.IfSeqNo), 10)
	}

	if r.Lang != "" {
		params["lang"] = r.Lang
	}

	if r.Refresh != "" {
		params["refresh"] = r.Refresh
	}

	if r.RequireAlias != nil {
		params["require_alias"] = strconv.FormatBool(*r.RequireAlias)
	}

	if r.RetryOnConflict != nil {
		params["retry_on_conflict"] = strconv.FormatInt(int64(*r.RetryOnConflict), 10)
	}

	if r.Routing != "" {
		params["routing"] = r.Routing
	}

	if len(r.Source) > 0 {
		params["_source"] = strings.Join(r.Source, ",")
	}

	if len(r.SourceExcludes) > 0 {
		params["_source_excludes"] = strings.Join(r.SourceExcludes, ",")
	}

	if len(r.SourceIncludes) > 0 {
		params["_source_includes"] = strings.Join(r.SourceIncludes, ",")
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
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
func (f Update) WithContext(v context.Context) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.ctx = v
	}
}

// WithDocumentType - the type of the document.
//
func (f Update) WithDocumentType(v string) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.DocumentType = v
	}
}

// WithIfPrimaryTerm - only perform the update operation if the last operation that has changed the document has the specified primary term.
//
func (f Update) WithIfPrimaryTerm(v int) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.IfPrimaryTerm = &v
	}
}

// WithIfSeqNo - only perform the update operation if the last operation that has changed the document has the specified sequence number.
//
func (f Update) WithIfSeqNo(v int) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.IfSeqNo = &v
	}
}

// WithLang - the script language (default: painless).
//
func (f Update) WithLang(v string) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.Lang = v
	}
}

// WithRefresh - if `true` then refresh the affected shards to make this operation visible to search, if `wait_for` then wait for a refresh to make this operation visible to search, if `false` (the default) then do nothing with refreshes..
//
func (f Update) WithRefresh(v string) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.Refresh = v
	}
}

// WithRequireAlias - when true, requires destination is an alias. default is false.
//
func (f Update) WithRequireAlias(v bool) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.RequireAlias = &v
	}
}

// WithRetryOnConflict - specify how many times should the operation be retried when a conflict occurs (default: 0).
//
func (f Update) WithRetryOnConflict(v int) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.RetryOnConflict = &v
	}
}

// WithRouting - specific routing value.
//
func (f Update) WithRouting(v string) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.Routing = v
	}
}

// WithSource - true or false to return the _source field or not, or a list of fields to return.
//
func (f Update) WithSource(v ...string) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.Source = v
	}
}

// WithSourceExcludes - a list of fields to exclude from the returned _source field.
//
func (f Update) WithSourceExcludes(v ...string) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.SourceExcludes = v
	}
}

// WithSourceIncludes - a list of fields to extract and return from the _source field.
//
func (f Update) WithSourceIncludes(v ...string) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.SourceIncludes = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f Update) WithTimeout(v time.Duration) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.Timeout = v
	}
}

// WithWaitForActiveShards - sets the number of shard copies that must be active before proceeding with the update operation. defaults to 1, meaning the primary shard only. set to `all` for all shard copies, otherwise set to any non-negative value less than or equal to the total number of copies for the shard (number of replicas + 1).
//
func (f Update) WithWaitForActiveShards(v string) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.WaitForActiveShards = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Update) WithPretty() func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Update) WithHuman() func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Update) WithErrorTrace() func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Update) WithFilterPath(v ...string) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Update) WithHeader(h map[string]string) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
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
func (f Update) WithOpaqueID(s string) func(*UpdateRequest) {
	return func(r *UpdateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
