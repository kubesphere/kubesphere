// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newMgetFunc(t Transport) Mget {
	return func(body io.Reader, o ...func(*MgetRequest)) (*Response, error) {
		var r = MgetRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// Mget allows to get multiple documents in one request.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/docs-multi-get.html.
//
type Mget func(body io.Reader, o ...func(*MgetRequest)) (*Response, error)

// MgetRequest configures the Mget API request.
//
type MgetRequest struct {
	Index        string
	DocumentType string

	Body io.Reader

	Preference     string
	Realtime       *bool
	Refresh        *bool
	Routing        string
	Source         []string
	SourceExcludes []string
	SourceIncludes []string
	StoredFields   []string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MgetRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(r.Index) + 1 + len(r.DocumentType) + 1 + len("_mget"))
	if r.Index != "" {
		path.WriteString("/")
		path.WriteString(r.Index)
	}
	if r.DocumentType != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentType)
	}
	path.WriteString("/")
	path.WriteString("_mget")

	params = make(map[string]string)

	if r.Preference != "" {
		params["preference"] = r.Preference
	}

	if r.Realtime != nil {
		params["realtime"] = strconv.FormatBool(*r.Realtime)
	}

	if r.Refresh != nil {
		params["refresh"] = strconv.FormatBool(*r.Refresh)
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

	if len(r.StoredFields) > 0 {
		params["stored_fields"] = strings.Join(r.StoredFields, ",")
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
func (f Mget) WithContext(v context.Context) func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.ctx = v
	}
}

// WithIndex - the name of the index.
//
func (f Mget) WithIndex(v string) func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.Index = v
	}
}

// WithDocumentType - the type of the document.
//
func (f Mget) WithDocumentType(v string) func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.DocumentType = v
	}
}

// WithPreference - specify the node or shard the operation should be performed on (default: random).
//
func (f Mget) WithPreference(v string) func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.Preference = v
	}
}

// WithRealtime - specify whether to perform the operation in realtime or search mode.
//
func (f Mget) WithRealtime(v bool) func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.Realtime = &v
	}
}

// WithRefresh - refresh the shard containing the document before performing the operation.
//
func (f Mget) WithRefresh(v bool) func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.Refresh = &v
	}
}

// WithRouting - specific routing value.
//
func (f Mget) WithRouting(v string) func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.Routing = v
	}
}

// WithSource - true or false to return the _source field or not, or a list of fields to return.
//
func (f Mget) WithSource(v ...string) func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.Source = v
	}
}

// WithSourceExcludes - a list of fields to exclude from the returned _source field.
//
func (f Mget) WithSourceExcludes(v ...string) func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.SourceExcludes = v
	}
}

// WithSourceIncludes - a list of fields to extract and return from the _source field.
//
func (f Mget) WithSourceIncludes(v ...string) func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.SourceIncludes = v
	}
}

// WithStoredFields - a list of stored fields to return in the response.
//
func (f Mget) WithStoredFields(v ...string) func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.StoredFields = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Mget) WithPretty() func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Mget) WithHuman() func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Mget) WithErrorTrace() func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Mget) WithFilterPath(v ...string) func(*MgetRequest) {
	return func(r *MgetRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Mget) WithHeader(h map[string]string) func(*MgetRequest) {
	return func(r *MgetRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
