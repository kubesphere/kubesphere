// Code generated from specification version 7.0.0 (5e798c1): DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"strconv"
	"strings"
)

func newRankEvalFunc(t Transport) RankEval {
	return func(body io.Reader, o ...func(*RankEvalRequest)) (*Response, error) {
		var r = RankEvalRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// RankEval allows to evaluate the quality of ranked search results over a set of typical search queries
//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/master/search-rank-eval.html.
//
type RankEval func(body io.Reader, o ...func(*RankEvalRequest)) (*Response, error)

// RankEvalRequest configures the Rank Eval API request.
//
type RankEvalRequest struct {
	Index []string
	Body  io.Reader

	AllowNoIndices    *bool
	ExpandWildcards   string
	IgnoreUnavailable *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r RankEvalRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_rank_eval"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_rank_eval")

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
func (f RankEval) WithContext(v context.Context) func(*RankEvalRequest) {
	return func(r *RankEvalRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names to search; use _all to perform the operation on all indices.
//
func (f RankEval) WithIndex(v ...string) func(*RankEvalRequest) {
	return func(r *RankEvalRequest) {
		r.Index = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f RankEval) WithAllowNoIndices(v bool) func(*RankEvalRequest) {
	return func(r *RankEvalRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f RankEval) WithExpandWildcards(v string) func(*RankEvalRequest) {
	return func(r *RankEvalRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f RankEval) WithIgnoreUnavailable(v bool) func(*RankEvalRequest) {
	return func(r *RankEvalRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f RankEval) WithPretty() func(*RankEvalRequest) {
	return func(r *RankEvalRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f RankEval) WithHuman() func(*RankEvalRequest) {
	return func(r *RankEvalRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f RankEval) WithErrorTrace() func(*RankEvalRequest) {
	return func(r *RankEvalRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f RankEval) WithFilterPath(v ...string) func(*RankEvalRequest) {
	return func(r *RankEvalRequest) {
		r.FilterPath = v
	}
}
