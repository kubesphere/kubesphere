// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newXPackMLGetModelSnapshotsFunc(t Transport) XPackMLGetModelSnapshots {
	return func(job_id string, o ...func(*XPackMLGetModelSnapshotsRequest)) (*Response, error) {
		var r = XPackMLGetModelSnapshotsRequest{JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLGetModelSnapshots - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-snapshot.html
//
type XPackMLGetModelSnapshots func(job_id string, o ...func(*XPackMLGetModelSnapshotsRequest)) (*Response, error)

// XPackMLGetModelSnapshotsRequest configures the X PackML Get Model Snapshots API request.
//
type XPackMLGetModelSnapshotsRequest struct {
	Body io.Reader

	JobID      string
	SnapshotID string

	Desc  *bool
	End   interface{}
	From  *int
	Size  *int
	Sort  string
	Start interface{}

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLGetModelSnapshotsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

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
	if r.SnapshotID != "" {
		path.WriteString("/")
		path.WriteString(r.SnapshotID)
	}

	params = make(map[string]string)

	if r.Desc != nil {
		params["desc"] = strconv.FormatBool(*r.Desc)
	}

	if r.End != nil {
		params["end"] = fmt.Sprintf("%v", r.End)
	}

	if r.From != nil {
		params["from"] = strconv.FormatInt(int64(*r.From), 10)
	}

	if r.Size != nil {
		params["size"] = strconv.FormatInt(int64(*r.Size), 10)
	}

	if r.Sort != "" {
		params["sort"] = r.Sort
	}

	if r.Start != nil {
		params["start"] = fmt.Sprintf("%v", r.Start)
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
func (f XPackMLGetModelSnapshots) WithContext(v context.Context) func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		r.ctx = v
	}
}

// WithBody - Model snapshot selection criteria.
//
func (f XPackMLGetModelSnapshots) WithBody(v io.Reader) func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		r.Body = v
	}
}

// WithSnapshotID - the ID of the snapshot to fetch.
//
func (f XPackMLGetModelSnapshots) WithSnapshotID(v string) func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		r.SnapshotID = v
	}
}

// WithDesc - true if the results should be sorted in descending order.
//
func (f XPackMLGetModelSnapshots) WithDesc(v bool) func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		r.Desc = &v
	}
}

// WithEnd - the filter 'end' query parameter.
//
func (f XPackMLGetModelSnapshots) WithEnd(v interface{}) func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		r.End = v
	}
}

// WithFrom - skips a number of documents.
//
func (f XPackMLGetModelSnapshots) WithFrom(v int) func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		r.From = &v
	}
}

// WithSize - the default number of documents returned in queries as a string..
//
func (f XPackMLGetModelSnapshots) WithSize(v int) func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		r.Size = &v
	}
}

// WithSort - name of the field to sort on.
//
func (f XPackMLGetModelSnapshots) WithSort(v string) func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		r.Sort = v
	}
}

// WithStart - the filter 'start' query parameter.
//
func (f XPackMLGetModelSnapshots) WithStart(v interface{}) func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		r.Start = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLGetModelSnapshots) WithPretty() func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLGetModelSnapshots) WithHuman() func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLGetModelSnapshots) WithErrorTrace() func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLGetModelSnapshots) WithFilterPath(v ...string) func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLGetModelSnapshots) WithHeader(h map[string]string) func(*XPackMLGetModelSnapshotsRequest) {
	return func(r *XPackMLGetModelSnapshotsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
