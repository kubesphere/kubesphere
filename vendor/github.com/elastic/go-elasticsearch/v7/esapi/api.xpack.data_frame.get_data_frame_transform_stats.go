// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newDataFrameGetDataFrameTransformStatsFunc(t Transport) DataFrameGetDataFrameTransformStats {
	return func(o ...func(*DataFrameGetDataFrameTransformStatsRequest)) (*Response, error) {
		var r = DataFrameGetDataFrameTransformStatsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// DataFrameGetDataFrameTransformStats - https://www.elastic.co/guide/en/elasticsearch/reference/current/get-data-frame-transform-stats.html
//
type DataFrameGetDataFrameTransformStats func(o ...func(*DataFrameGetDataFrameTransformStatsRequest)) (*Response, error)

// DataFrameGetDataFrameTransformStatsRequest configures the Data Frame Get Data Frame Transform Stats API request.
//
type DataFrameGetDataFrameTransformStatsRequest struct {
	TransformID string

	AllowNoMatch *bool
	From         *int
	Size         *int

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r DataFrameGetDataFrameTransformStatsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_data_frame") + 1 + len("transforms") + 1 + len(r.TransformID) + 1 + len("_stats"))
	path.WriteString("/")
	path.WriteString("_data_frame")
	path.WriteString("/")
	path.WriteString("transforms")
	if r.TransformID != "" {
		path.WriteString("/")
		path.WriteString(r.TransformID)
	}
	path.WriteString("/")
	path.WriteString("_stats")

	params = make(map[string]string)

	if r.AllowNoMatch != nil {
		params["allow_no_match"] = strconv.FormatBool(*r.AllowNoMatch)
	}

	if r.From != nil {
		params["from"] = strconv.FormatInt(int64(*r.From), 10)
	}

	if r.Size != nil {
		params["size"] = strconv.FormatInt(int64(*r.Size), 10)
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
func (f DataFrameGetDataFrameTransformStats) WithContext(v context.Context) func(*DataFrameGetDataFrameTransformStatsRequest) {
	return func(r *DataFrameGetDataFrameTransformStatsRequest) {
		r.ctx = v
	}
}

// WithTransformID - the ID of the transform for which to get stats. '_all' or '*' implies all transforms.
//
func (f DataFrameGetDataFrameTransformStats) WithTransformID(v string) func(*DataFrameGetDataFrameTransformStatsRequest) {
	return func(r *DataFrameGetDataFrameTransformStatsRequest) {
		r.TransformID = v
	}
}

// WithAllowNoMatch - whether to ignore if a wildcard expression matches no data frame transforms. (this includes `_all` string or when no data frame transforms have been specified).
//
func (f DataFrameGetDataFrameTransformStats) WithAllowNoMatch(v bool) func(*DataFrameGetDataFrameTransformStatsRequest) {
	return func(r *DataFrameGetDataFrameTransformStatsRequest) {
		r.AllowNoMatch = &v
	}
}

// WithFrom - skips a number of transform stats, defaults to 0.
//
func (f DataFrameGetDataFrameTransformStats) WithFrom(v int) func(*DataFrameGetDataFrameTransformStatsRequest) {
	return func(r *DataFrameGetDataFrameTransformStatsRequest) {
		r.From = &v
	}
}

// WithSize - specifies a max number of transform stats to get, defaults to 100.
//
func (f DataFrameGetDataFrameTransformStats) WithSize(v int) func(*DataFrameGetDataFrameTransformStatsRequest) {
	return func(r *DataFrameGetDataFrameTransformStatsRequest) {
		r.Size = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f DataFrameGetDataFrameTransformStats) WithPretty() func(*DataFrameGetDataFrameTransformStatsRequest) {
	return func(r *DataFrameGetDataFrameTransformStatsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f DataFrameGetDataFrameTransformStats) WithHuman() func(*DataFrameGetDataFrameTransformStatsRequest) {
	return func(r *DataFrameGetDataFrameTransformStatsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f DataFrameGetDataFrameTransformStats) WithErrorTrace() func(*DataFrameGetDataFrameTransformStatsRequest) {
	return func(r *DataFrameGetDataFrameTransformStatsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f DataFrameGetDataFrameTransformStats) WithFilterPath(v ...string) func(*DataFrameGetDataFrameTransformStatsRequest) {
	return func(r *DataFrameGetDataFrameTransformStatsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f DataFrameGetDataFrameTransformStats) WithHeader(h map[string]string) func(*DataFrameGetDataFrameTransformStatsRequest) {
	return func(r *DataFrameGetDataFrameTransformStatsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
