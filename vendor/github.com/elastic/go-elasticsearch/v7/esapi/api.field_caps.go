// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newFieldCapsFunc(t Transport) FieldCaps {
	return func(o ...func(*FieldCapsRequest)) (*Response, error) {
		var r = FieldCapsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// FieldCaps returns the information about the capabilities of fields among multiple indices.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/search-field-caps.html.
//
type FieldCaps func(o ...func(*FieldCapsRequest)) (*Response, error)

// FieldCapsRequest configures the Field Caps API request.
//
type FieldCapsRequest struct {
	Index []string

	AllowNoIndices    *bool
	ExpandWildcards   string
	Fields            []string
	IgnoreUnavailable *bool
	IncludeUnmapped   *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r FieldCapsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_field_caps"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_field_caps")

	params = make(map[string]string)

	if r.AllowNoIndices != nil {
		params["allow_no_indices"] = strconv.FormatBool(*r.AllowNoIndices)
	}

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if len(r.Fields) > 0 {
		params["fields"] = strings.Join(r.Fields, ",")
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.IncludeUnmapped != nil {
		params["include_unmapped"] = strconv.FormatBool(*r.IncludeUnmapped)
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

	req, _ := newRequest(method, path.String(), nil)

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
func (f FieldCaps) WithContext(v context.Context) func(*FieldCapsRequest) {
	return func(r *FieldCapsRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names; use _all to perform the operation on all indices.
//
func (f FieldCaps) WithIndex(v ...string) func(*FieldCapsRequest) {
	return func(r *FieldCapsRequest) {
		r.Index = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f FieldCaps) WithAllowNoIndices(v bool) func(*FieldCapsRequest) {
	return func(r *FieldCapsRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f FieldCaps) WithExpandWildcards(v string) func(*FieldCapsRequest) {
	return func(r *FieldCapsRequest) {
		r.ExpandWildcards = v
	}
}

// WithFields - a list of field names.
//
func (f FieldCaps) WithFields(v ...string) func(*FieldCapsRequest) {
	return func(r *FieldCapsRequest) {
		r.Fields = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f FieldCaps) WithIgnoreUnavailable(v bool) func(*FieldCapsRequest) {
	return func(r *FieldCapsRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithIncludeUnmapped - indicates whether unmapped fields should be included in the response..
//
func (f FieldCaps) WithIncludeUnmapped(v bool) func(*FieldCapsRequest) {
	return func(r *FieldCapsRequest) {
		r.IncludeUnmapped = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f FieldCaps) WithPretty() func(*FieldCapsRequest) {
	return func(r *FieldCapsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f FieldCaps) WithHuman() func(*FieldCapsRequest) {
	return func(r *FieldCapsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f FieldCaps) WithErrorTrace() func(*FieldCapsRequest) {
	return func(r *FieldCapsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f FieldCaps) WithFilterPath(v ...string) func(*FieldCapsRequest) {
	return func(r *FieldCapsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f FieldCaps) WithHeader(h map[string]string) func(*FieldCapsRequest) {
	return func(r *FieldCapsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
