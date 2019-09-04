// Code generated from specification version 5.6.15: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newMpercolateFunc(t Transport) Mpercolate {
	return func(body io.Reader, o ...func(*MpercolateRequest)) (*Response, error) {
		var r = MpercolateRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/5.x/search-percolate.html.
//
type Mpercolate func(body io.Reader, o ...func(*MpercolateRequest)) (*Response, error)

// MpercolateRequest configures the Mpercolate API request.
//
type MpercolateRequest struct {
	Index        string
	DocumentType string

	Body io.Reader

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
func (r MpercolateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(r.Index) + 1 + len(r.DocumentType) + 1 + len("_mpercolate"))
	if r.Index != "" {
		path.WriteString("/")
		path.WriteString(r.Index)
	}
	if r.DocumentType != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentType)
	}
	path.WriteString("/")
	path.WriteString("_mpercolate")

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
func (f Mpercolate) WithContext(v context.Context) func(*MpercolateRequest) {
	return func(r *MpercolateRequest) {
		r.ctx = v
	}
}

// WithIndex - the index of the document being count percolated to use as default.
//
func (f Mpercolate) WithIndex(v string) func(*MpercolateRequest) {
	return func(r *MpercolateRequest) {
		r.Index = v
	}
}

// WithDocumentType - the type of the document being percolated to use as default..
//
func (f Mpercolate) WithDocumentType(v string) func(*MpercolateRequest) {
	return func(r *MpercolateRequest) {
		r.DocumentType = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f Mpercolate) WithAllowNoIndices(v bool) func(*MpercolateRequest) {
	return func(r *MpercolateRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f Mpercolate) WithExpandWildcards(v string) func(*MpercolateRequest) {
	return func(r *MpercolateRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f Mpercolate) WithIgnoreUnavailable(v bool) func(*MpercolateRequest) {
	return func(r *MpercolateRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Mpercolate) WithPretty() func(*MpercolateRequest) {
	return func(r *MpercolateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Mpercolate) WithHuman() func(*MpercolateRequest) {
	return func(r *MpercolateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Mpercolate) WithErrorTrace() func(*MpercolateRequest) {
	return func(r *MpercolateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Mpercolate) WithFilterPath(v ...string) func(*MpercolateRequest) {
	return func(r *MpercolateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Mpercolate) WithHeader(h map[string]string) func(*MpercolateRequest) {
	return func(r *MpercolateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
