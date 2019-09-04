// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newMLGetRecordsFunc(t Transport) MLGetRecords {
	return func(job_id string, o ...func(*MLGetRecordsRequest)) (*Response, error) {
		var r = MLGetRecordsRequest{JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLGetRecords - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-record.html
//
type MLGetRecords func(job_id string, o ...func(*MLGetRecordsRequest)) (*Response, error)

// MLGetRecordsRequest configures the ML Get Records API request.
//
type MLGetRecordsRequest struct {
	Body io.Reader

	JobID string

	Desc           *bool
	End            string
	ExcludeInterim *bool
	From           *int
	RecordScore    interface{}
	Size           *int
	Sort           string
	Start          string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLGetRecordsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("results") + 1 + len("records"))
	path.WriteString("/")
	path.WriteString("_ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
	path.WriteString("/")
	path.WriteString(r.JobID)
	path.WriteString("/")
	path.WriteString("results")
	path.WriteString("/")
	path.WriteString("records")

	params = make(map[string]string)

	if r.Desc != nil {
		params["desc"] = strconv.FormatBool(*r.Desc)
	}

	if r.End != "" {
		params["end"] = r.End
	}

	if r.ExcludeInterim != nil {
		params["exclude_interim"] = strconv.FormatBool(*r.ExcludeInterim)
	}

	if r.From != nil {
		params["from"] = strconv.FormatInt(int64(*r.From), 10)
	}

	if r.RecordScore != nil {
		params["record_score"] = fmt.Sprintf("%v", r.RecordScore)
	}

	if r.Size != nil {
		params["size"] = strconv.FormatInt(int64(*r.Size), 10)
	}

	if r.Sort != "" {
		params["sort"] = r.Sort
	}

	if r.Start != "" {
		params["start"] = r.Start
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
func (f MLGetRecords) WithContext(v context.Context) func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.ctx = v
	}
}

// WithBody - Record selection criteria.
//
func (f MLGetRecords) WithBody(v io.Reader) func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.Body = v
	}
}

// WithDesc - set the sort direction.
//
func (f MLGetRecords) WithDesc(v bool) func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.Desc = &v
	}
}

// WithEnd - end time filter for records.
//
func (f MLGetRecords) WithEnd(v string) func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.End = v
	}
}

// WithExcludeInterim - exclude interim results.
//
func (f MLGetRecords) WithExcludeInterim(v bool) func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.ExcludeInterim = &v
	}
}

// WithFrom - skips a number of records.
//
func (f MLGetRecords) WithFrom(v int) func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.From = &v
	}
}

// WithRecordScore - .
//
func (f MLGetRecords) WithRecordScore(v interface{}) func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.RecordScore = v
	}
}

// WithSize - specifies a max number of records to get.
//
func (f MLGetRecords) WithSize(v int) func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.Size = &v
	}
}

// WithSort - sort records by a particular field.
//
func (f MLGetRecords) WithSort(v string) func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.Sort = v
	}
}

// WithStart - start time filter for records.
//
func (f MLGetRecords) WithStart(v string) func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.Start = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLGetRecords) WithPretty() func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLGetRecords) WithHuman() func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLGetRecords) WithErrorTrace() func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLGetRecords) WithFilterPath(v ...string) func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLGetRecords) WithHeader(h map[string]string) func(*MLGetRecordsRequest) {
	return func(r *MLGetRecordsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
