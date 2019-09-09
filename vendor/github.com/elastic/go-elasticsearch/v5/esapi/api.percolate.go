// Code generated from specification version 5.6.15: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newPercolateFunc(t Transport) Percolate {
	return func(index string, o ...func(*PercolateRequest)) (*Response, error) {
		var r = PercolateRequest{Index: index}
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
type Percolate func(index string, o ...func(*PercolateRequest)) (*Response, error)

// PercolateRequest configures the Percolate API request.
//
type PercolateRequest struct {
	Index        string
	DocumentType string
	DocumentID   string

	Body io.Reader

	AllowNoIndices      *bool
	ExpandWildcards     string
	IgnoreUnavailable   *bool
	PercolateFormat     string
	PercolateIndex      string
	PercolatePreference string
	PercolateRouting    string
	PercolateType       string
	Preference          string
	Routing             []string
	Version             *int
	VersionType         string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r PercolateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(r.Index) + 1 + len(r.DocumentType) + 1 + len(r.DocumentID) + 1 + len("_percolate"))
	path.WriteString("/")
	path.WriteString(r.Index)
	path.WriteString("/")
	path.WriteString(r.DocumentType)
	if r.DocumentID != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentID)
	}
	path.WriteString("/")
	path.WriteString("_percolate")

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

	if r.PercolateFormat != "" {
		params["percolate_format"] = r.PercolateFormat
	}

	if r.PercolateIndex != "" {
		params["percolate_index"] = r.PercolateIndex
	}

	if r.PercolatePreference != "" {
		params["percolate_preference"] = r.PercolatePreference
	}

	if r.PercolateRouting != "" {
		params["percolate_routing"] = r.PercolateRouting
	}

	if r.PercolateType != "" {
		params["percolate_type"] = r.PercolateType
	}

	if r.Preference != "" {
		params["preference"] = r.Preference
	}

	if len(r.Routing) > 0 {
		params["routing"] = strings.Join(r.Routing, ",")
	}

	if r.Version != nil {
		params["version"] = strconv.FormatInt(int64(*r.Version), 10)
	}

	if r.VersionType != "" {
		params["version_type"] = r.VersionType
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
func (f Percolate) WithContext(v context.Context) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.ctx = v
	}
}

// WithBody - The percolator request definition using the percolate DSL.
//
func (f Percolate) WithBody(v io.Reader) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.Body = v
	}
}

// WithDocumentID - substitute the document in the request body with a document that is known by the specified ID. on top of the ID, the index and type parameter will be used to retrieve the document from within the cluster..
//
func (f Percolate) WithDocumentID(v string) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.DocumentID = v
	}
}

// WithDocumentType - the type of the document being percolated..
//
func (f Percolate) WithDocumentType(v string) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.DocumentType = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f Percolate) WithAllowNoIndices(v bool) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.AllowNoIndices = &v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f Percolate) WithExpandWildcards(v string) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.ExpandWildcards = v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f Percolate) WithIgnoreUnavailable(v bool) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithPercolateFormat - return an array of matching query ids instead of objects.
//
func (f Percolate) WithPercolateFormat(v string) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.PercolateFormat = v
	}
}

// WithPercolateIndex - the index to percolate the document into. defaults to index..
//
func (f Percolate) WithPercolateIndex(v string) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.PercolateIndex = v
	}
}

// WithPercolatePreference - which shard to prefer when executing the percolate request..
//
func (f Percolate) WithPercolatePreference(v string) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.PercolatePreference = v
	}
}

// WithPercolateRouting - the routing value to use when percolating the existing document..
//
func (f Percolate) WithPercolateRouting(v string) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.PercolateRouting = v
	}
}

// WithPercolateType - the type to percolate document into. defaults to type..
//
func (f Percolate) WithPercolateType(v string) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.PercolateType = v
	}
}

// WithPreference - specify the node or shard the operation should be performed on (default: random).
//
func (f Percolate) WithPreference(v string) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.Preference = v
	}
}

// WithRouting - a list of specific routing values.
//
func (f Percolate) WithRouting(v ...string) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.Routing = v
	}
}

// WithVersion - explicit version number for concurrency control.
//
func (f Percolate) WithVersion(v int) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.Version = &v
	}
}

// WithVersionType - specific version type.
//
func (f Percolate) WithVersionType(v string) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.VersionType = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Percolate) WithPretty() func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Percolate) WithHuman() func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Percolate) WithErrorTrace() func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Percolate) WithFilterPath(v ...string) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Percolate) WithHeader(h map[string]string) func(*PercolateRequest) {
	return func(r *PercolateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
