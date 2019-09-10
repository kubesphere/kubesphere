// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newSnapshotDeleteFunc(t Transport) SnapshotDelete {
	return func(repository string, snapshot string, o ...func(*SnapshotDeleteRequest)) (*Response, error) {
		var r = SnapshotDeleteRequest{Repository: repository, Snapshot: snapshot}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SnapshotDelete deletes a snapshot.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/modules-snapshots.html.
//
type SnapshotDelete func(repository string, snapshot string, o ...func(*SnapshotDeleteRequest)) (*Response, error)

// SnapshotDeleteRequest configures the Snapshot Delete API request.
//
type SnapshotDeleteRequest struct {
	Repository string
	Snapshot   string

	MasterTimeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SnapshotDeleteRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

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
func (f SnapshotDelete) WithContext(v context.Context) func(*SnapshotDeleteRequest) {
	return func(r *SnapshotDeleteRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f SnapshotDelete) WithMasterTimeout(v time.Duration) func(*SnapshotDeleteRequest) {
	return func(r *SnapshotDeleteRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SnapshotDelete) WithPretty() func(*SnapshotDeleteRequest) {
	return func(r *SnapshotDeleteRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SnapshotDelete) WithHuman() func(*SnapshotDeleteRequest) {
	return func(r *SnapshotDeleteRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SnapshotDelete) WithErrorTrace() func(*SnapshotDeleteRequest) {
	return func(r *SnapshotDeleteRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SnapshotDelete) WithFilterPath(v ...string) func(*SnapshotDeleteRequest) {
	return func(r *SnapshotDeleteRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SnapshotDelete) WithHeader(h map[string]string) func(*SnapshotDeleteRequest) {
	return func(r *SnapshotDeleteRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
