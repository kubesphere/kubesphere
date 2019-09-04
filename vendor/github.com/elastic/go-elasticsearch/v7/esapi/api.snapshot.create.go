// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newSnapshotCreateFunc(t Transport) SnapshotCreate {
	return func(repository string, snapshot string, o ...func(*SnapshotCreateRequest)) (*Response, error) {
		var r = SnapshotCreateRequest{Repository: repository, Snapshot: snapshot}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SnapshotCreate creates a snapshot in a repository.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/modules-snapshots.html.
//
type SnapshotCreate func(repository string, snapshot string, o ...func(*SnapshotCreateRequest)) (*Response, error)

// SnapshotCreateRequest configures the Snapshot Create API request.
//
type SnapshotCreateRequest struct {
	Body io.Reader

	Repository string
	Snapshot   string

	MasterTimeout     time.Duration
	WaitForCompletion *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SnapshotCreateRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(1 + len("_snapshot") + 1 + len(r.Repository) + 1 + len(r.Snapshot))
	path.WriteString("/")
	path.WriteString("_snapshot")
	path.WriteString("/")
	path.WriteString(r.Repository)
	path.WriteString("/")
	path.WriteString(r.Snapshot)

	params = make(map[string]string)

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.WaitForCompletion != nil {
		params["wait_for_completion"] = strconv.FormatBool(*r.WaitForCompletion)
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
func (f SnapshotCreate) WithContext(v context.Context) func(*SnapshotCreateRequest) {
	return func(r *SnapshotCreateRequest) {
		r.ctx = v
	}
}

// WithBody - The snapshot definition.
//
func (f SnapshotCreate) WithBody(v io.Reader) func(*SnapshotCreateRequest) {
	return func(r *SnapshotCreateRequest) {
		r.Body = v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f SnapshotCreate) WithMasterTimeout(v time.Duration) func(*SnapshotCreateRequest) {
	return func(r *SnapshotCreateRequest) {
		r.MasterTimeout = v
	}
}

// WithWaitForCompletion - should this request wait until the operation has completed before returning.
//
func (f SnapshotCreate) WithWaitForCompletion(v bool) func(*SnapshotCreateRequest) {
	return func(r *SnapshotCreateRequest) {
		r.WaitForCompletion = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SnapshotCreate) WithPretty() func(*SnapshotCreateRequest) {
	return func(r *SnapshotCreateRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SnapshotCreate) WithHuman() func(*SnapshotCreateRequest) {
	return func(r *SnapshotCreateRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SnapshotCreate) WithErrorTrace() func(*SnapshotCreateRequest) {
	return func(r *SnapshotCreateRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SnapshotCreate) WithFilterPath(v ...string) func(*SnapshotCreateRequest) {
	return func(r *SnapshotCreateRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SnapshotCreate) WithHeader(h map[string]string) func(*SnapshotCreateRequest) {
	return func(r *SnapshotCreateRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
