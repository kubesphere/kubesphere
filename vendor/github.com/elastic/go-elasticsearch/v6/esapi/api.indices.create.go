// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newIndicesCreateFunc(t Transport) IndicesCreate {
	return func(index string, o ...func(*IndicesCreateRequest)) (*Response, error) {
		var r = IndicesCreateRequest{Index: index}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesCreate creates an index with optional settings and mappings.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/indices-create-index.html.
//
type IndicesCreate func(index string, o ...func(*IndicesCreateRequest)) (*Response, error)

// IndicesCreateRequest configures the Indices Create API request.
//
type IndicesCreateRequest struct {
	Index string

	Body io.Reader

	IncludeTypeName     *bool
	MasterTimeout       time.Duration
	Timeout             time.Duration
	UpdateAllTypes      *bool
	WaitForActiveShards string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesCreateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len(r.Index))
	path.WriteString("/")
	path.WriteString(r.Index)

	params = make(map[string]string)

	if r.IncludeTypeName != nil {
		params["include_type_name"] = strconv.FormatBool(*r.IncludeTypeName)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
	}

	if r.UpdateAllTypes != nil {
		params["update_all_types"] = strconv.FormatBool(*r.UpdateAllTypes)
	}

	if r.WaitForActiveShards != "" {
		params["wait_for_active_shards"] = r.WaitForActiveShards
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
func (f IndicesCreate) WithContext(v context.Context) func(*IndicesCreateRequest) {
	return func(r *IndicesCreateRequest) {
		r.ctx = v
	}
}

// WithBody - The configuration for the index (`settings` and `mappings`).
//
func (f IndicesCreate) WithBody(v io.Reader) func(*IndicesCreateRequest) {
	return func(r *IndicesCreateRequest) {
		r.Body = v
	}
}

// WithIncludeTypeName - whether a type should be expected in the body of the mappings..
//
func (f IndicesCreate) WithIncludeTypeName(v bool) func(*IndicesCreateRequest) {
	return func(r *IndicesCreateRequest) {
		r.IncludeTypeName = &v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f IndicesCreate) WithMasterTimeout(v time.Duration) func(*IndicesCreateRequest) {
	return func(r *IndicesCreateRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IndicesCreate) WithTimeout(v time.Duration) func(*IndicesCreateRequest) {
	return func(r *IndicesCreateRequest) {
		r.Timeout = v
	}
}

// WithUpdateAllTypes - whether to update the mapping for all fields with the same name across all types or not.
//
func (f IndicesCreate) WithUpdateAllTypes(v bool) func(*IndicesCreateRequest) {
	return func(r *IndicesCreateRequest) {
		r.UpdateAllTypes = &v
	}
}

// WithWaitForActiveShards - set the number of active shards to wait for before the operation returns..
//
func (f IndicesCreate) WithWaitForActiveShards(v string) func(*IndicesCreateRequest) {
	return func(r *IndicesCreateRequest) {
		r.WaitForActiveShards = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesCreate) WithPretty() func(*IndicesCreateRequest) {
	return func(r *IndicesCreateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesCreate) WithHuman() func(*IndicesCreateRequest) {
	return func(r *IndicesCreateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesCreate) WithErrorTrace() func(*IndicesCreateRequest) {
	return func(r *IndicesCreateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesCreate) WithFilterPath(v ...string) func(*IndicesCreateRequest) {
	return func(r *IndicesCreateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesCreate) WithHeader(h map[string]string) func(*IndicesCreateRequest) {
	return func(r *IndicesCreateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
