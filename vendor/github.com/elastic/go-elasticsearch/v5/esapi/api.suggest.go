// Code generated from specification version 5.6.15: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newSuggestFunc(t Transport) Suggest {
	return func(body io.Reader, o ...func(*SuggestRequest)) (*Response, error) {
		var r = SuggestRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/5.x/search-suggesters.html.
//
type Suggest func(body io.Reader, o ...func(*SuggestRequest)) (*Response, error)

// SuggestRequest configures the Suggest API request.
//
type SuggestRequest struct {
	Index []string

	Body io.Reader

	AllowNoIndices    *bool
	ExpandWildcards   string
	IgnoreUnavailable *bool
	Preference        string
	Routing           string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SuggestRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_suggest"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_suggest")

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

	if r.Preference != "" {
		params["preference"] = r.Preference
	}

	if r.Routing != "" {
		params["routing"] = r.Routing
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
func (f Suggest) WithContext(v context.Context) func(*SuggestRequest) {
	return func(r *SuggestRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names to restrict the operation; use _all to perform the operation on all indices.
//
func (f Suggest) WithIndex(v ...string) func(*SuggestRequest) {
	return func(r *SuggestRequest) {
		r.Index = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f Suggest) WithAllowNoIndices(v bool) func(*SuggestRequest) {
	return func(r *SuggestRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f Suggest) WithExpandWildcards(v string) func(*SuggestRequest) {
	return func(r *SuggestRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f Suggest) WithIgnoreUnavailable(v bool) func(*SuggestRequest) {
	return func(r *SuggestRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithPreference - specify the node or shard the operation should be performed on (default: random).
//
func (f Suggest) WithPreference(v string) func(*SuggestRequest) {
	return func(r *SuggestRequest) {
		r.Preference = v
	}
}

// WithRouting - specific routing value.
//
func (f Suggest) WithRouting(v string) func(*SuggestRequest) {
	return func(r *SuggestRequest) {
		r.Routing = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Suggest) WithPretty() func(*SuggestRequest) {
	return func(r *SuggestRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Suggest) WithHuman() func(*SuggestRequest) {
	return func(r *SuggestRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Suggest) WithErrorTrace() func(*SuggestRequest) {
	return func(r *SuggestRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Suggest) WithFilterPath(v ...string) func(*SuggestRequest) {
	return func(r *SuggestRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Suggest) WithHeader(h map[string]string) func(*SuggestRequest) {
	return func(r *SuggestRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
