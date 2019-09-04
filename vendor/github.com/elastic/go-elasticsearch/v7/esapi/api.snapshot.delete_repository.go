// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newSnapshotDeleteRepositoryFunc(t Transport) SnapshotDeleteRepository {
	return func(repository []string, o ...func(*SnapshotDeleteRepositoryRequest)) (*Response, error) {
		var r = SnapshotDeleteRepositoryRequest{Repository: repository}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SnapshotDeleteRepository deletes a repository.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/modules-snapshots.html.
//
type SnapshotDeleteRepository func(repository []string, o ...func(*SnapshotDeleteRepositoryRequest)) (*Response, error)

// SnapshotDeleteRepositoryRequest configures the Snapshot Delete Repository API request.
//
type SnapshotDeleteRepositoryRequest struct {
	Repository []string

	MasterTimeout time.Duration
	Timeout       time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SnapshotDeleteRepositoryRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_snapshot") + 1 + len(strings.Join(r.Repository, ",")))
	path.WriteString("/")
	path.WriteString("_snapshot")
	path.WriteString("/")
	path.WriteString(strings.Join(r.Repository, ","))

	params = make(map[string]string)

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
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
func (f SnapshotDeleteRepository) WithContext(v context.Context) func(*SnapshotDeleteRepositoryRequest) {
	return func(r *SnapshotDeleteRepositoryRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f SnapshotDeleteRepository) WithMasterTimeout(v time.Duration) func(*SnapshotDeleteRepositoryRequest) {
	return func(r *SnapshotDeleteRepositoryRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f SnapshotDeleteRepository) WithTimeout(v time.Duration) func(*SnapshotDeleteRepositoryRequest) {
	return func(r *SnapshotDeleteRepositoryRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SnapshotDeleteRepository) WithPretty() func(*SnapshotDeleteRepositoryRequest) {
	return func(r *SnapshotDeleteRepositoryRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SnapshotDeleteRepository) WithHuman() func(*SnapshotDeleteRepositoryRequest) {
	return func(r *SnapshotDeleteRepositoryRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SnapshotDeleteRepository) WithErrorTrace() func(*SnapshotDeleteRepositoryRequest) {
	return func(r *SnapshotDeleteRepositoryRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SnapshotDeleteRepository) WithFilterPath(v ...string) func(*SnapshotDeleteRepositoryRequest) {
	return func(r *SnapshotDeleteRepositoryRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SnapshotDeleteRepository) WithHeader(h map[string]string) func(*SnapshotDeleteRepositoryRequest) {
	return func(r *SnapshotDeleteRepositoryRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
