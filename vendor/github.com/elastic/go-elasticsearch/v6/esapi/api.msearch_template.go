// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newMsearchTemplateFunc(t Transport) MsearchTemplate {
	return func(body io.Reader, o ...func(*MsearchTemplateRequest)) (*Response, error) {
		var r = MsearchTemplateRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MsearchTemplate allows to execute several search template operations in one request.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/current/search-multi-search.html.
//
type MsearchTemplate func(body io.Reader, o ...func(*MsearchTemplateRequest)) (*Response, error)

// MsearchTemplateRequest configures the Msearch Template API request.
//
type MsearchTemplateRequest struct {
	Index        []string
	DocumentType []string

	Body io.Reader

	MaxConcurrentSearches *int
	RestTotalHitsAsInt    *bool
	SearchType            string
	TypedKeys             *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MsearchTemplateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len(strings.Join(r.DocumentType, ",")) + 1 + len("_msearch") + 1 + len("template"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	if len(r.DocumentType) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.DocumentType, ","))
	}
	path.WriteString("/")
	path.WriteString("_msearch")
	path.WriteString("/")
	path.WriteString("template")

	params = make(map[string]string)

	if r.MaxConcurrentSearches != nil {
		params["max_concurrent_searches"] = strconv.FormatInt(int64(*r.MaxConcurrentSearches), 10)
	}

	if r.RestTotalHitsAsInt != nil {
		params["rest_total_hits_as_int"] = strconv.FormatBool(*r.RestTotalHitsAsInt)
	}

	if r.SearchType != "" {
		params["search_type"] = r.SearchType
	}

	if r.TypedKeys != nil {
		params["typed_keys"] = strconv.FormatBool(*r.TypedKeys)
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

	req, _ := newRequest(method, path.String(), r.Body)

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
func (f MsearchTemplate) WithContext(v context.Context) func(*MsearchTemplateRequest) {
	return func(r *MsearchTemplateRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names to use as default.
//
func (f MsearchTemplate) WithIndex(v ...string) func(*MsearchTemplateRequest) {
	return func(r *MsearchTemplateRequest) {
		r.Index = v
	}
}

// WithDocumentType - a list of document types to use as default.
//
func (f MsearchTemplate) WithDocumentType(v ...string) func(*MsearchTemplateRequest) {
	return func(r *MsearchTemplateRequest) {
		r.DocumentType = v
	}
}

// WithMaxConcurrentSearches - controls the maximum number of concurrent searches the multi search api will execute.
//
func (f MsearchTemplate) WithMaxConcurrentSearches(v int) func(*MsearchTemplateRequest) {
	return func(r *MsearchTemplateRequest) {
		r.MaxConcurrentSearches = &v
	}
}

// WithRestTotalHitsAsInt - this parameter is ignored in this version. it is used in the next major version to control whether the rest response should render the total.hits as an object or a number.
//
func (f MsearchTemplate) WithRestTotalHitsAsInt(v bool) func(*MsearchTemplateRequest) {
	return func(r *MsearchTemplateRequest) {
		r.RestTotalHitsAsInt = &v
	}
}

// WithSearchType - search operation type.
//
func (f MsearchTemplate) WithSearchType(v string) func(*MsearchTemplateRequest) {
	return func(r *MsearchTemplateRequest) {
		r.SearchType = v
	}
}

// WithTypedKeys - specify whether aggregation and suggester names should be prefixed by their respective types in the response.
//
func (f MsearchTemplate) WithTypedKeys(v bool) func(*MsearchTemplateRequest) {
	return func(r *MsearchTemplateRequest) {
		r.TypedKeys = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MsearchTemplate) WithPretty() func(*MsearchTemplateRequest) {
	return func(r *MsearchTemplateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MsearchTemplate) WithHuman() func(*MsearchTemplateRequest) {
	return func(r *MsearchTemplateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MsearchTemplate) WithErrorTrace() func(*MsearchTemplateRequest) {
	return func(r *MsearchTemplateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MsearchTemplate) WithFilterPath(v ...string) func(*MsearchTemplateRequest) {
	return func(r *MsearchTemplateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MsearchTemplate) WithHeader(h map[string]string) func(*MsearchTemplateRequest) {
	return func(r *MsearchTemplateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
