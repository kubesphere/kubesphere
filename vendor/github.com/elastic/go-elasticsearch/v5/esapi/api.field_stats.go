// Code generated from specification version 5.6.15: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newFieldStatsFunc(t Transport) FieldStats {
	return func(o ...func(*FieldStatsRequest)) (*Response, error) {
		var r = FieldStatsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/5.x/search-field-stats.html.
//
type FieldStats func(o ...func(*FieldStatsRequest)) (*Response, error)

// FieldStatsRequest configures the Field Stats API request.
//
type FieldStatsRequest struct {
	Index []string

	Body io.Reader

	AllowNoIndices    *bool
	ExpandWildcards   string
	Fields            []string
	IgnoreUnavailable *bool
	Level             string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r FieldStatsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_field_stats"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_field_stats")

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

	if r.Level != "" {
		params["level"] = r.Level
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
func (f FieldStats) WithContext(v context.Context) func(*FieldStatsRequest) {
	return func(r *FieldStatsRequest) {
		r.ctx = v
	}
}

// WithBody - Field json objects containing the name and optionally a range to filter out indices result, that have results outside the defined bounds.
//
func (f FieldStats) WithBody(v io.Reader) func(*FieldStatsRequest) {
	return func(r *FieldStatsRequest) {
		r.Body = v
	}
}

// WithIndex - a list of index names; use _all to perform the operation on all indices.
//
func (f FieldStats) WithIndex(v ...string) func(*FieldStatsRequest) {
	return func(r *FieldStatsRequest) {
		r.Index = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f FieldStats) WithAllowNoIndices(v bool) func(*FieldStatsRequest) {
	return func(r *FieldStatsRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f FieldStats) WithExpandWildcards(v string) func(*FieldStatsRequest) {
	return func(r *FieldStatsRequest) {
		r.ExpandWildcards = v
	}
}

// WithFields - a list of fields for to get field statistics for (min value, max value, and more).
//
func (f FieldStats) WithFields(v ...string) func(*FieldStatsRequest) {
	return func(r *FieldStatsRequest) {
		r.Fields = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f FieldStats) WithIgnoreUnavailable(v bool) func(*FieldStatsRequest) {
	return func(r *FieldStatsRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithLevel - defines if field stats should be returned on a per index level or on a cluster wide level.
//
func (f FieldStats) WithLevel(v string) func(*FieldStatsRequest) {
	return func(r *FieldStatsRequest) {
		r.Level = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f FieldStats) WithPretty() func(*FieldStatsRequest) {
	return func(r *FieldStatsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f FieldStats) WithHuman() func(*FieldStatsRequest) {
	return func(r *FieldStatsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f FieldStats) WithErrorTrace() func(*FieldStatsRequest) {
	return func(r *FieldStatsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f FieldStats) WithFilterPath(v ...string) func(*FieldStatsRequest) {
	return func(r *FieldStatsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f FieldStats) WithHeader(h map[string]string) func(*FieldStatsRequest) {
	return func(r *FieldStatsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
