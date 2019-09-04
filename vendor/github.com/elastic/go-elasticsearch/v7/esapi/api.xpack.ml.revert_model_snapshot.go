// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newMLRevertModelSnapshotFunc(t Transport) MLRevertModelSnapshot {
	return func(snapshot_id string, job_id string, o ...func(*MLRevertModelSnapshotRequest)) (*Response, error) {
		var r = MLRevertModelSnapshotRequest{SnapshotID: snapshot_id, JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLRevertModelSnapshot - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-revert-snapshot.html
//
type MLRevertModelSnapshot func(snapshot_id string, job_id string, o ...func(*MLRevertModelSnapshotRequest)) (*Response, error)

// MLRevertModelSnapshotRequest configures the ML Revert Model Snapshot API request.
//
type MLRevertModelSnapshotRequest struct {
	Body io.Reader

	JobID      string
	SnapshotID string

	DeleteInterveningResults *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLRevertModelSnapshotRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("model_snapshots") + 1 + len(r.SnapshotID) + 1 + len("_revert"))
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
	path.WriteString("/")
	path.WriteString("_revert")

	params = make(map[string]string)

	if r.DeleteInterveningResults != nil {
		params["delete_intervening_results"] = strconv.FormatBool(*r.DeleteInterveningResults)
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
func (f MLRevertModelSnapshot) WithContext(v context.Context) func(*MLRevertModelSnapshotRequest) {
	return func(r *MLRevertModelSnapshotRequest) {
		r.ctx = v
	}
}

// WithBody - Reversion options.
//
func (f MLRevertModelSnapshot) WithBody(v io.Reader) func(*MLRevertModelSnapshotRequest) {
	return func(r *MLRevertModelSnapshotRequest) {
		r.Body = v
	}
}

// WithDeleteInterveningResults - should we reset the results back to the time of the snapshot?.
//
func (f MLRevertModelSnapshot) WithDeleteInterveningResults(v bool) func(*MLRevertModelSnapshotRequest) {
	return func(r *MLRevertModelSnapshotRequest) {
		r.DeleteInterveningResults = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLRevertModelSnapshot) WithPretty() func(*MLRevertModelSnapshotRequest) {
	return func(r *MLRevertModelSnapshotRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLRevertModelSnapshot) WithHuman() func(*MLRevertModelSnapshotRequest) {
	return func(r *MLRevertModelSnapshotRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLRevertModelSnapshot) WithErrorTrace() func(*MLRevertModelSnapshotRequest) {
	return func(r *MLRevertModelSnapshotRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLRevertModelSnapshot) WithFilterPath(v ...string) func(*MLRevertModelSnapshotRequest) {
	return func(r *MLRevertModelSnapshotRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLRevertModelSnapshot) WithHeader(h map[string]string) func(*MLRevertModelSnapshotRequest) {
	return func(r *MLRevertModelSnapshotRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
