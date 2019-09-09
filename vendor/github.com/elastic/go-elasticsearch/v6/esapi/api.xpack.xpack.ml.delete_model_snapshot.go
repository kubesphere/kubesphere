// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
)

func newXPackMLDeleteModelSnapshotFunc(t Transport) XPackMLDeleteModelSnapshot {
	return func(snapshot_id string, job_id string, o ...func(*XPackMLDeleteModelSnapshotRequest)) (*Response, error) {
		var r = XPackMLDeleteModelSnapshotRequest{SnapshotID: snapshot_id, JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLDeleteModelSnapshot - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-delete-snapshot.html
//
type XPackMLDeleteModelSnapshot func(snapshot_id string, job_id string, o ...func(*XPackMLDeleteModelSnapshotRequest)) (*Response, error)

// XPackMLDeleteModelSnapshotRequest configures the X PackML Delete Model Snapshot API request.
//
type XPackMLDeleteModelSnapshotRequest struct {
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
func (r XPackMLDeleteModelSnapshotRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("model_snapshots") + 1 + len(r.SnapshotID))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
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
func (f XPackMLDeleteModelSnapshot) WithContext(v context.Context) func(*XPackMLDeleteModelSnapshotRequest) {
	return func(r *XPackMLDeleteModelSnapshotRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLDeleteModelSnapshot) WithPretty() func(*XPackMLDeleteModelSnapshotRequest) {
	return func(r *XPackMLDeleteModelSnapshotRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLDeleteModelSnapshot) WithHuman() func(*XPackMLDeleteModelSnapshotRequest) {
	return func(r *XPackMLDeleteModelSnapshotRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLDeleteModelSnapshot) WithErrorTrace() func(*XPackMLDeleteModelSnapshotRequest) {
	return func(r *XPackMLDeleteModelSnapshotRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLDeleteModelSnapshot) WithFilterPath(v ...string) func(*XPackMLDeleteModelSnapshotRequest) {
	return func(r *XPackMLDeleteModelSnapshotRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLDeleteModelSnapshot) WithHeader(h map[string]string) func(*XPackMLDeleteModelSnapshotRequest) {
	return func(r *XPackMLDeleteModelSnapshotRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
