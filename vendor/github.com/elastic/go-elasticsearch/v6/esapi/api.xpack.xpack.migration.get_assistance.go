// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newXPackMigrationGetAssistanceFunc(t Transport) XPackMigrationGetAssistance {
	return func(o ...func(*XPackMigrationGetAssistanceRequest)) (*Response, error) {
		var r = XPackMigrationGetAssistanceRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMigrationGetAssistance - https://www.elastic.co/guide/en/elasticsearch/reference/current/migration-api-assistance.html
//
type XPackMigrationGetAssistance func(o ...func(*XPackMigrationGetAssistanceRequest)) (*Response, error)

// XPackMigrationGetAssistanceRequest configures the X Pack Migration Get Assistance API request.
//
type XPackMigrationGetAssistanceRequest struct {
	Index []string

	AllowNoIndices    *bool
	ExpandWildcards   string
	IgnoreUnavailable *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMigrationGetAssistanceRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("migration") + 1 + len("assistance") + 1 + len(strings.Join(r.Index, ",")))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("migration")
	path.WriteString("/")
	path.WriteString("assistance")
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}

	params = make(map[string]string)

	if r.AllowNoIndices != nil {
		params["allow_no_indices"] = strconv.FormatBool(*r.AllowNoIndices)
	}

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
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
func (f XPackMigrationGetAssistance) WithContext(v context.Context) func(*XPackMigrationGetAssistanceRequest) {
	return func(r *XPackMigrationGetAssistanceRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names; use _all to perform the operation on all indices.
//
func (f XPackMigrationGetAssistance) WithIndex(v ...string) func(*XPackMigrationGetAssistanceRequest) {
	return func(r *XPackMigrationGetAssistanceRequest) {
		r.Index = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f XPackMigrationGetAssistance) WithAllowNoIndices(v bool) func(*XPackMigrationGetAssistanceRequest) {
	return func(r *XPackMigrationGetAssistanceRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f XPackMigrationGetAssistance) WithExpandWildcards(v string) func(*XPackMigrationGetAssistanceRequest) {
	return func(r *XPackMigrationGetAssistanceRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f XPackMigrationGetAssistance) WithIgnoreUnavailable(v bool) func(*XPackMigrationGetAssistanceRequest) {
	return func(r *XPackMigrationGetAssistanceRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMigrationGetAssistance) WithPretty() func(*XPackMigrationGetAssistanceRequest) {
	return func(r *XPackMigrationGetAssistanceRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMigrationGetAssistance) WithHuman() func(*XPackMigrationGetAssistanceRequest) {
	return func(r *XPackMigrationGetAssistanceRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMigrationGetAssistance) WithErrorTrace() func(*XPackMigrationGetAssistanceRequest) {
	return func(r *XPackMigrationGetAssistanceRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMigrationGetAssistance) WithFilterPath(v ...string) func(*XPackMigrationGetAssistanceRequest) {
	return func(r *XPackMigrationGetAssistanceRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMigrationGetAssistance) WithHeader(h map[string]string) func(*XPackMigrationGetAssistanceRequest) {
	return func(r *XPackMigrationGetAssistanceRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
