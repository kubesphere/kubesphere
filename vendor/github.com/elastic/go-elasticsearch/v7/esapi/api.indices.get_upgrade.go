// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newIndicesGetUpgradeFunc(t Transport) IndicesGetUpgrade {
	return func(o ...func(*IndicesGetUpgradeRequest)) (*Response, error) {
		var r = IndicesGetUpgradeRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesGetUpgrade the _upgrade API is no longer useful and will be removed.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/indices-upgrade.html.
//
type IndicesGetUpgrade func(o ...func(*IndicesGetUpgradeRequest)) (*Response, error)

// IndicesGetUpgradeRequest configures the Indices Get Upgrade API request.
//
type IndicesGetUpgradeRequest struct {
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
func (r IndicesGetUpgradeRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_upgrade"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_upgrade")

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
func (f IndicesGetUpgrade) WithContext(v context.Context) func(*IndicesGetUpgradeRequest) {
	return func(r *IndicesGetUpgradeRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names; use _all to perform the operation on all indices.
//
func (f IndicesGetUpgrade) WithIndex(v ...string) func(*IndicesGetUpgradeRequest) {
	return func(r *IndicesGetUpgradeRequest) {
		r.Index = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f IndicesGetUpgrade) WithAllowNoIndices(v bool) func(*IndicesGetUpgradeRequest) {
	return func(r *IndicesGetUpgradeRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f IndicesGetUpgrade) WithExpandWildcards(v string) func(*IndicesGetUpgradeRequest) {
	return func(r *IndicesGetUpgradeRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f IndicesGetUpgrade) WithIgnoreUnavailable(v bool) func(*IndicesGetUpgradeRequest) {
	return func(r *IndicesGetUpgradeRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesGetUpgrade) WithPretty() func(*IndicesGetUpgradeRequest) {
	return func(r *IndicesGetUpgradeRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesGetUpgrade) WithHuman() func(*IndicesGetUpgradeRequest) {
	return func(r *IndicesGetUpgradeRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesGetUpgrade) WithErrorTrace() func(*IndicesGetUpgradeRequest) {
	return func(r *IndicesGetUpgradeRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesGetUpgrade) WithFilterPath(v ...string) func(*IndicesGetUpgradeRequest) {
	return func(r *IndicesGetUpgradeRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesGetUpgrade) WithHeader(h map[string]string) func(*IndicesGetUpgradeRequest) {
	return func(r *IndicesGetUpgradeRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
