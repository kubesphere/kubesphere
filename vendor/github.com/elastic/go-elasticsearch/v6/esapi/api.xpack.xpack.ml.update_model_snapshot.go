// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newXPackMLUpdateModelSnapshotFunc(t Transport) XPackMLUpdateModelSnapshot {
	return func(snapshot_id string, job_id string, body io.Reader, o ...func(*XPackMLUpdateModelSnapshotRequest)) (*Response, error) {
		var r = XPackMLUpdateModelSnapshotRequest{SnapshotID: snapshot_id, JobID: job_id, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLUpdateModelSnapshot - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-update-snapshot.html
//
type XPackMLUpdateModelSnapshot func(snapshot_id string, job_id string, body io.Reader, o ...func(*XPackMLUpdateModelSnapshotRequest)) (*Response, error)

// XPackMLUpdateModelSnapshotRequest configures the X PackML Update Model Snapshot API request.
//
type XPackMLUpdateModelSnapshotRequest struct {
	Body io.Reader

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
func (r XPackMLUpdateModelSnapshotRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("model_snapshots") + 1 + len(r.SnapshotID) + 1 + len("_update"))
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
	path.WriteString("/")
	path.WriteString("_update")

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
func (f XPackMLUpdateModelSnapshot) WithContext(v context.Context) func(*XPackMLUpdateModelSnapshotRequest) {
	return func(r *XPackMLUpdateModelSnapshotRequest) {
		r.ctx = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLUpdateModelSnapshot) WithPretty() func(*XPackMLUpdateModelSnapshotRequest) {
	return func(r *XPackMLUpdateModelSnapshotRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLUpdateModelSnapshot) WithHuman() func(*XPackMLUpdateModelSnapshotRequest) {
	return func(r *XPackMLUpdateModelSnapshotRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLUpdateModelSnapshot) WithErrorTrace() func(*XPackMLUpdateModelSnapshotRequest) {
	return func(r *XPackMLUpdateModelSnapshotRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLUpdateModelSnapshot) WithFilterPath(v ...string) func(*XPackMLUpdateModelSnapshotRequest) {
	return func(r *XPackMLUpdateModelSnapshotRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLUpdateModelSnapshot) WithHeader(h map[string]string) func(*XPackMLUpdateModelSnapshotRequest) {
	return func(r *XPackMLUpdateModelSnapshotRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
