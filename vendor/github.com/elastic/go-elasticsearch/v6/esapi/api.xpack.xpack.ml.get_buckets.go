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

func newXPackMLGetBucketsFunc(t Transport) XPackMLGetBuckets {
	return func(job_id string, o ...func(*XPackMLGetBucketsRequest)) (*Response, error) {
		var r = XPackMLGetBucketsRequest{JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLGetBuckets - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-bucket.html
//
type XPackMLGetBuckets func(job_id string, o ...func(*XPackMLGetBucketsRequest)) (*Response, error)

// XPackMLGetBucketsRequest configures the X PackML Get Buckets API request.
//
type XPackMLGetBucketsRequest struct {
	Body io.Reader

	JobID     string
	Timestamp string

	AnomalyScore   interface{}
	Desc           *bool
	End            string
	ExcludeInterim *bool
	Expand         *bool
	From           *int
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
func (r XPackMLGetBucketsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("results") + 1 + len("buckets") + 1 + len(r.Timestamp))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
	path.WriteString("/")
	path.WriteString(r.JobID)
	path.WriteString("/")
	path.WriteString("results")
	path.WriteString("/")
	path.WriteString("buckets")
	if r.Timestamp != "" {
		path.WriteString("/")
		path.WriteString(r.Timestamp)
	}

	params = make(map[string]string)

	if r.AnomalyScore != nil {
		params["anomaly_score"] = fmt.Sprintf("%v", r.AnomalyScore)
	}

	if r.Desc != nil {
		params["desc"] = strconv.FormatBool(*r.Desc)
	}

	if r.End != "" {
		params["end"] = r.End
	}

	if r.ExcludeInterim != nil {
		params["exclude_interim"] = strconv.FormatBool(*r.ExcludeInterim)
	}

	if r.Expand != nil {
		params["expand"] = strconv.FormatBool(*r.Expand)
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
func (f XPackMLGetBuckets) WithContext(v context.Context) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.ctx = v
	}
}

// WithBody - Bucket selection details if not provided in URI.
//
func (f XPackMLGetBuckets) WithBody(v io.Reader) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.Body = v
	}
}

// WithTimestamp - the timestamp of the desired single bucket result.
//
func (f XPackMLGetBuckets) WithTimestamp(v string) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.Timestamp = v
	}
}

// WithAnomalyScore - filter for the most anomalous buckets.
//
func (f XPackMLGetBuckets) WithAnomalyScore(v interface{}) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.AnomalyScore = v
	}
}

// WithDesc - set the sort direction.
//
func (f XPackMLGetBuckets) WithDesc(v bool) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.Desc = &v
	}
}

// WithEnd - end time filter for buckets.
//
func (f XPackMLGetBuckets) WithEnd(v string) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.End = v
	}
}

// WithExcludeInterim - exclude interim results.
//
func (f XPackMLGetBuckets) WithExcludeInterim(v bool) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.ExcludeInterim = &v
	}
}

// WithExpand - include anomaly records.
//
func (f XPackMLGetBuckets) WithExpand(v bool) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.Expand = &v
	}
}

// WithFrom - skips a number of buckets.
//
func (f XPackMLGetBuckets) WithFrom(v int) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.From = &v
	}
}

// WithSize - specifies a max number of buckets to get.
//
func (f XPackMLGetBuckets) WithSize(v int) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.Size = &v
	}
}

// WithSort - sort buckets by a particular field.
//
func (f XPackMLGetBuckets) WithSort(v string) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.Sort = v
	}
}

// WithStart - start time filter for buckets.
//
func (f XPackMLGetBuckets) WithStart(v string) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.Start = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLGetBuckets) WithPretty() func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLGetBuckets) WithHuman() func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLGetBuckets) WithErrorTrace() func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLGetBuckets) WithFilterPath(v ...string) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLGetBuckets) WithHeader(h map[string]string) func(*XPackMLGetBucketsRequest) {
	return func(r *XPackMLGetBucketsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
