// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newMLDeleteModelSnapshotFunc(t Transport) MLDeleteModelSnapshot {
	return func(snapshot_id string, job_id string, o ...func(*MLDeleteModelSnapshotRequest)) (*Response, error) {
		var r = MLDeleteModelSnapshotRequest{SnapshotID: snapshot_id, JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLDeleteModelSnapshot - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-delete-snapshot.html
//
type MLDeleteModelSnapshot func(snapshot_id string, job_id string, o ...func(*MLDeleteModelSnapshotRequest)) (*Response, error)

// MLDeleteModelSnapshotRequest configures the ML Delete Model Snapshot API request.
//
type MLDeleteModelSnapshotRequest struct {
	JobID      string
	SnapshotID string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLDeleteModelSnapshotRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("model_snapshots") + 1 + len(r.SnapshotID))
	path.WriteString("/")
	path.WriteString("_ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
	path.WriteString("/")
	path.WriteString(r.JobID)
	path.WriteString("/")
	path.WriteString("model_snapshots")
	path.WriteString("/")
	path.WriteString(r.SnapshotID)

	params = make(map[string]string)

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
func (f MLDeleteModelSnapshot) WithContext(v context.Context) func(*MLDeleteModelSnapshotRequest) {
	return func(r *MLDeleteModelSnapshotRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLDeleteModelSnapshot) WithPretty() func(*MLDeleteModelSnapshotRequest) {
	return func(r *MLDeleteModelSnapshotRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLDeleteModelSnapshot) WithHuman() func(*MLDeleteModelSnapshotRequest) {
	return func(r *MLDeleteModelSnapshotRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLDeleteModelSnapshot) WithErrorTrace() func(*MLDeleteModelSnapshotRequest) {
	return func(r *MLDeleteModelSnapshotRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLDeleteModelSnapshot) WithFilterPath(v ...string) func(*MLDeleteModelSnapshotRequest) {
	return func(r *MLDeleteModelSnapshotRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLDeleteModelSnapshot) WithHeader(h map[string]string) func(*MLDeleteModelSnapshotRequest) {
	return func(r *MLDeleteModelSnapshotRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
