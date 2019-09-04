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

func newXPackMLGetOverallBucketsFunc(t Transport) XPackMLGetOverallBuckets {
	return func(job_id string, o ...func(*XPackMLGetOverallBucketsRequest)) (*Response, error) {
		var r = XPackMLGetOverallBucketsRequest{JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLGetOverallBuckets - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-overall-buckets.html
//
type XPackMLGetOverallBuckets func(job_id string, o ...func(*XPackMLGetOverallBucketsRequest)) (*Response, error)

// XPackMLGetOverallBucketsRequest configures the X PackML Get Overall Buckets API request.
//
type XPackMLGetOverallBucketsRequest struct {
	Body io.Reader

	JobID string

	AllowNoJobs    *bool
	BucketSpan     string
	End            string
	ExcludeInterim *bool
	OverallScore   interface{}
	Start          string
	TopN           *int

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLGetOverallBucketsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("results") + 1 + len("overall_buckets"))
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
	path.WriteString("overall_buckets")

	params = make(map[string]string)

	if r.AllowNoJobs != nil {
		params["allow_no_jobs"] = strconv.FormatBool(*r.AllowNoJobs)
	}

	if r.BucketSpan != "" {
		params["bucket_span"] = r.BucketSpan
	}

	if r.End != "" {
		params["end"] = r.End
	}

	if r.ExcludeInterim != nil {
		params["exclude_interim"] = strconv.FormatBool(*r.ExcludeInterim)
	}

	if r.OverallScore != nil {
		params["overall_score"] = fmt.Sprintf("%v", r.OverallScore)
	}

	if r.Start != "" {
		params["start"] = r.Start
	}

	if r.TopN != nil {
		params["top_n"] = strconv.FormatInt(int64(*r.TopN), 10)
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
func (f XPackMLGetOverallBuckets) WithContext(v context.Context) func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		r.ctx = v
	}
}

// WithBody - Overall bucket selection details if not provided in URI.
//
func (f XPackMLGetOverallBuckets) WithBody(v io.Reader) func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		r.Body = v
	}
}

// WithAllowNoJobs - whether to ignore if a wildcard expression matches no jobs. (this includes `_all` string or when no jobs have been specified).
//
func (f XPackMLGetOverallBuckets) WithAllowNoJobs(v bool) func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		r.AllowNoJobs = &v
	}
}

// WithBucketSpan - the span of the overall buckets. defaults to the longest job bucket_span.
//
func (f XPackMLGetOverallBuckets) WithBucketSpan(v string) func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		r.BucketSpan = v
	}
}

// WithEnd - returns overall buckets with timestamps earlier than this time.
//
func (f XPackMLGetOverallBuckets) WithEnd(v string) func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		r.End = v
	}
}

// WithExcludeInterim - if true overall buckets that include interim buckets will be excluded.
//
func (f XPackMLGetOverallBuckets) WithExcludeInterim(v bool) func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		r.ExcludeInterim = &v
	}
}

// WithOverallScore - returns overall buckets with overall scores higher than this value.
//
func (f XPackMLGetOverallBuckets) WithOverallScore(v interface{}) func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		r.OverallScore = v
	}
}

// WithStart - returns overall buckets with timestamps after this time.
//
func (f XPackMLGetOverallBuckets) WithStart(v string) func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		r.Start = v
	}
}

// WithTopN - the number of top job bucket scores to be used in the overall_score calculation.
//
func (f XPackMLGetOverallBuckets) WithTopN(v int) func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		r.TopN = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLGetOverallBuckets) WithPretty() func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLGetOverallBuckets) WithHuman() func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLGetOverallBuckets) WithErrorTrace() func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLGetOverallBuckets) WithFilterPath(v ...string) func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLGetOverallBuckets) WithHeader(h map[string]string) func(*XPackMLGetOverallBucketsRequest) {
	return func(r *XPackMLGetOverallBucketsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
