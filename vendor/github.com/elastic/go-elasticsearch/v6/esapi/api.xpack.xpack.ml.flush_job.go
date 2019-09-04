// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newXPackMLFlushJobFunc(t Transport) XPackMLFlushJob {
	return func(job_id string, o ...func(*XPackMLFlushJobRequest)) (*Response, error) {
		var r = XPackMLFlushJobRequest{JobID: job_id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLFlushJob - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-flush-job.html
//
type XPackMLFlushJob func(job_id string, o ...func(*XPackMLFlushJobRequest)) (*Response, error)

// XPackMLFlushJobRequest configures the X PackML Flush Job API request.
//
type XPackMLFlushJobRequest struct {
	Body io.Reader

	JobID string

	AdvanceTime string
	CalcInterim *bool
	End         string
	SkipTime    string
	Start       string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLFlushJobRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_xpack") + 1 + len("ml") + 1 + len("anomaly_detectors") + 1 + len(r.JobID) + 1 + len("_flush"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("ml")
	path.WriteString("/")
	path.WriteString("anomaly_detectors")
	path.WriteString("/")
	path.WriteString(r.JobID)
	path.WriteString("/")
	path.WriteString("_flush")

	params = make(map[string]string)

	if r.AdvanceTime != "" {
		params["advance_time"] = r.AdvanceTime
	}

	if r.CalcInterim != nil {
		params["calc_interim"] = strconv.FormatBool(*r.CalcInterim)
	}

	if r.End != "" {
		params["end"] = r.End
	}

	if r.SkipTime != "" {
		params["skip_time"] = r.SkipTime
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
func (f XPackMLFlushJob) WithContext(v context.Context) func(*XPackMLFlushJobRequest) {
	return func(r *XPackMLFlushJobRequest) {
		r.ctx = v
	}
}

// WithBody - Flush parameters.
//
func (f XPackMLFlushJob) WithBody(v io.Reader) func(*XPackMLFlushJobRequest) {
	return func(r *XPackMLFlushJobRequest) {
		r.Body = v
	}
}

// WithAdvanceTime - advances time to the given value generating results and updating the model for the advanced interval.
//
func (f XPackMLFlushJob) WithAdvanceTime(v string) func(*XPackMLFlushJobRequest) {
	return func(r *XPackMLFlushJobRequest) {
		r.AdvanceTime = v
	}
}

// WithCalcInterim - calculates interim results for the most recent bucket or all buckets within the latency period.
//
func (f XPackMLFlushJob) WithCalcInterim(v bool) func(*XPackMLFlushJobRequest) {
	return func(r *XPackMLFlushJobRequest) {
		r.CalcInterim = &v
	}
}

// WithEnd - when used in conjunction with calc_interim, specifies the range of buckets on which to calculate interim results.
//
func (f XPackMLFlushJob) WithEnd(v string) func(*XPackMLFlushJobRequest) {
	return func(r *XPackMLFlushJobRequest) {
		r.End = v
	}
}

// WithSkipTime - skips time to the given value without generating results or updating the model for the skipped interval.
//
func (f XPackMLFlushJob) WithSkipTime(v string) func(*XPackMLFlushJobRequest) {
	return func(r *XPackMLFlushJobRequest) {
		r.SkipTime = v
	}
}

// WithStart - when used in conjunction with calc_interim, specifies the range of buckets on which to calculate interim results.
//
func (f XPackMLFlushJob) WithStart(v string) func(*XPackMLFlushJobRequest) {
	return func(r *XPackMLFlushJobRequest) {
		r.Start = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLFlushJob) WithPretty() func(*XPackMLFlushJobRequest) {
	return func(r *XPackMLFlushJobRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLFlushJob) WithHuman() func(*XPackMLFlushJobRequest) {
	return func(r *XPackMLFlushJobRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLFlushJob) WithErrorTrace() func(*XPackMLFlushJobRequest) {
	return func(r *XPackMLFlushJobRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLFlushJob) WithFilterPath(v ...string) func(*XPackMLFlushJobRequest) {
	return func(r *XPackMLFlushJobRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLFlushJob) WithHeader(h map[string]string) func(*XPackMLFlushJobRequest) {
	return func(r *XPackMLFlushJobRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
