// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newSnapshotGetFunc(t Transport) SnapshotGet {
	return func(repository string, snapshot []string, o ...func(*SnapshotGetRequest)) (*Response, error) {
		var r = SnapshotGetRequest{Repository: repository, Snapshot: snapshot}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// SnapshotGet returns information about a snapshot.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/modules-snapshots.html.
//
type SnapshotGet func(repository string, snapshot []string, o ...func(*SnapshotGetRequest)) (*Response, error)

// SnapshotGetRequest configures the Snapshot Get API request.
//
type SnapshotGetRequest struct {
	Repository string
	Snapshot   []string

	IgnoreUnavailable *bool
	MasterTimeout     time.Duration
	Verbose           *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SnapshotGetRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_snapshot") + 1 + len(r.Repository) + 1 + len(strings.Join(r.Snapshot, ",")))
	path.WriteString("/")
	path.WriteString("_snapshot")
	path.WriteString("/")
	path.WriteString(r.Repository)
	path.WriteString("/")
	path.WriteString(strings.Join(r.Snapshot, ","))

	params = make(map[string]string)

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if r.Verbose != nil {
		params["verbose"] = strconv.FormatBool(*r.Verbose)
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
func (f SnapshotGet) WithContext(v context.Context) func(*SnapshotGetRequest) {
	return func(r *SnapshotGetRequest) {
		r.ctx = v
	}
}

// WithIgnoreUnavailable - whether to ignore unavailable snapshots, defaults to false which means a snapshotmissingexception is thrown.
//
func (f SnapshotGet) WithIgnoreUnavailable(v bool) func(*SnapshotGetRequest) {
	return func(r *SnapshotGetRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f SnapshotGet) WithMasterTimeout(v time.Duration) func(*SnapshotGetRequest) {
	return func(r *SnapshotGetRequest) {
		r.MasterTimeout = v
	}
}

// WithVerbose - whether to show verbose snapshot info or only show the basic info found in the repository index blob.
//
func (f SnapshotGet) WithVerbose(v bool) func(*SnapshotGetRequest) {
	return func(r *SnapshotGetRequest) {
		r.Verbose = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f SnapshotGet) WithPretty() func(*SnapshotGetRequest) {
	return func(r *SnapshotGetRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f SnapshotGet) WithHuman() func(*SnapshotGetRequest) {
	return func(r *SnapshotGetRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f SnapshotGet) WithErrorTrace() func(*SnapshotGetRequest) {
	return func(r *SnapshotGetRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f SnapshotGet) WithFilterPath(v ...string) func(*SnapshotGetRequest) {
	return func(r *SnapshotGetRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f SnapshotGet) WithHeader(h map[string]string) func(*SnapshotGetRequest) {
	return func(r *SnapshotGetRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
