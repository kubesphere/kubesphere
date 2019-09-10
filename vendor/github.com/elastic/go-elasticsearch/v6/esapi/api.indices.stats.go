// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newIndicesStatsFunc(t Transport) IndicesStats {
	return func(o ...func(*IndicesStatsRequest)) (*Response, error) {
		var r = IndicesStatsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesStats provides statistics on operations happening in an index.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/indices-stats.html.
//
type IndicesStats func(o ...func(*IndicesStatsRequest)) (*Response, error)

// IndicesStatsRequest configures the Indices Stats API request.
//
type IndicesStatsRequest struct {
	Index []string

	Metric []string

	CompletionFields        []string
	FielddataFields         []string
	Fields                  []string
	Groups                  []string
	IncludeSegmentFileSizes *bool
	Level                   string
	Types                   []string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesStatsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_stats") + 1 + len(strings.Join(r.Metric, ",")))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_stats")
	if len(r.Metric) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Metric, ","))
	}

	params = make(map[string]string)

	if len(r.CompletionFields) > 0 {
		params["completion_fields"] = strings.Join(r.CompletionFields, ",")
	}

	if len(r.FielddataFields) > 0 {
		params["fielddata_fields"] = strings.Join(r.FielddataFields, ",")
	}

	if len(r.Fields) > 0 {
		params["fields"] = strings.Join(r.Fields, ",")
	}

	if len(r.Groups) > 0 {
		params["groups"] = strings.Join(r.Groups, ",")
	}

	if r.IncludeSegmentFileSizes != nil {
		params["include_segment_file_sizes"] = strconv.FormatBool(*r.IncludeSegmentFileSizes)
	}

	if r.Level != "" {
		params["level"] = r.Level
	}

	if len(r.Types) > 0 {
		params["types"] = strings.Join(r.Types, ",")
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
func (f IndicesStats) WithContext(v context.Context) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index names; use _all to perform the operation on all indices.
//
func (f IndicesStats) WithIndex(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Index = v
	}
}

// WithMetric - limit the information returned the specific metrics..
//
func (f IndicesStats) WithMetric(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Metric = v
	}
}

// WithCompletionFields - a list of fields for `fielddata` and `suggest` index metric (supports wildcards).
//
func (f IndicesStats) WithCompletionFields(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.CompletionFields = v
	}
}

// WithFielddataFields - a list of fields for `fielddata` index metric (supports wildcards).
//
func (f IndicesStats) WithFielddataFields(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.FielddataFields = v
	}
}

// WithFields - a list of fields for `fielddata` and `completion` index metric (supports wildcards).
//
func (f IndicesStats) WithFields(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Fields = v
	}
}

// WithGroups - a list of search groups for `search` index metric.
//
func (f IndicesStats) WithGroups(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Groups = v
	}
}

// WithIncludeSegmentFileSizes - whether to report the aggregated disk usage of each one of the lucene index files (only applies if segment stats are requested).
//
func (f IndicesStats) WithIncludeSegmentFileSizes(v bool) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.IncludeSegmentFileSizes = &v
	}
}

// WithLevel - return stats aggregated at cluster, index or shard level.
//
func (f IndicesStats) WithLevel(v string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Level = v
	}
}

// WithTypes - a list of document types for the `indexing` index metric.
//
func (f IndicesStats) WithTypes(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Types = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesStats) WithPretty() func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesStats) WithHuman() func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesStats) WithErrorTrace() func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesStats) WithFilterPath(v ...string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesStats) WithHeader(h map[string]string) func(*IndicesStatsRequest) {
	return func(r *IndicesStatsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
