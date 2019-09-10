// Code generated from specification version 5.6.15: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newCreateFunc(t Transport) Create {
	return func(index string, id string, body io.Reader, o ...func(*CreateRequest)) (*Response, error) {
		var r = CreateRequest{Index: index, DocumentID: id, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// Create creates a new document in the index.
//
// Returns a 409 response when a document with a same ID already exists in the index.
//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/5.x/docs-index_.html.
//
type Create func(index string, id string, body io.Reader, o ...func(*CreateRequest)) (*Response, error)

// CreateRequest configures the Create API request.
//
type CreateRequest struct {
	Index        string
	DocumentType string
	DocumentID   string

	Body io.Reader

	Parent              string
	Pipeline            string
	Refresh             string
	Routing             string
	Timeout             time.Duration
	Timestamp           time.Duration
	TTL                 time.Duration
	Version             *int
	VersionType         string
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
func (r CreateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	if r.DocumentType == "" {
		r.DocumentType = "_doc"
	}

	path.Grow(1 + len(r.Index) + 1 + len(r.DocumentType) + 1 + len(r.DocumentID) + 1 + len("_create"))
	path.WriteString("/")
	path.WriteString(r.Index)
	path.WriteString("/")
	path.WriteString(r.DocumentType)
	path.WriteString("/")
	path.WriteString(r.DocumentID)
	path.WriteString("/")
	path.WriteString("_create")

	params = make(map[string]string)

	if r.Parent != "" {
		params["parent"] = r.Parent
	}

	if r.Pipeline != "" {
		params["pipeline"] = r.Pipeline
	}

	if r.Refresh != "" {
		params["refresh"] = r.Refresh
	}

	if r.Routing != "" {
		params["routing"] = r.Routing
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
	}

	if r.Timestamp != 0 {
		params["timestamp"] = formatDuration(r.Timestamp)
	}

	if r.TTL != 0 {
		params["ttl"] = formatDuration(r.TTL)
	}

	if r.Version != nil {
		params["version"] = strconv.FormatInt(int64(*r.Version), 10)
	}

	if r.VersionType != "" {
		params["version_type"] = r.VersionType
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
func (f Create) WithContext(v context.Context) func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.ctx = v
	}
}

// WithDocumentType - the type of the document.
//
func (f Create) WithDocumentType(v string) func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.DocumentType = v
	}
}

// WithParent - ID of the parent document.
//
func (f Create) WithParent(v string) func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.Parent = v
	}
}

// WithPipeline - the pipeline ID to preprocess incoming documents with.
//
func (f Create) WithPipeline(v string) func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.Pipeline = v
	}
}

// WithRefresh - if `true` then refresh the affected shards to make this operation visible to search, if `wait_for` then wait for a refresh to make this operation visible to search, if `false` (the default) then do nothing with refreshes..
//
func (f Create) WithRefresh(v string) func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.Refresh = v
	}
}

// WithRouting - specific routing value.
//
func (f Create) WithRouting(v string) func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.Routing = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f Create) WithTimeout(v time.Duration) func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.Timeout = v
	}
}

// WithTimestamp - explicit timestamp for the document.
//
func (f Create) WithTimestamp(v time.Duration) func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.Timestamp = v
	}
}

// WithTTL - expiration time for the document.
//
func (f Create) WithTTL(v time.Duration) func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.TTL = v
	}
}

// WithVersion - explicit version number for concurrency control.
//
func (f Create) WithVersion(v int) func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.Version = &v
	}
}

// WithVersionType - specific version type.
//
func (f Create) WithVersionType(v string) func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.VersionType = v
	}
}

// WithWaitForActiveShards - sets the number of shard copies that must be active before proceeding with the index operation. defaults to 1, meaning the primary shard only. set to `all` for all shard copies, otherwise set to any non-negative value less than or equal to the total number of copies for the shard (number of replicas + 1).
//
func (f Create) WithWaitForActiveShards(v string) func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.WaitForActiveShards = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Create) WithPretty() func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Create) WithHuman() func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Create) WithErrorTrace() func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Create) WithFilterPath(v ...string) func(*CreateRequest) {
	return func(r *CreateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Create) WithHeader(h map[string]string) func(*CreateRequest) {
	return func(r *CreateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
