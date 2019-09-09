// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newNodesStatsFunc(t Transport) NodesStats {
	return func(o ...func(*NodesStatsRequest)) (*Response, error) {
		var r = NodesStatsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// NodesStats returns statistical information about nodes in the cluster.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/cluster-nodes-stats.html.
//
type NodesStats func(o ...func(*NodesStatsRequest)) (*Response, error)

// NodesStatsRequest configures the Nodes Stats API request.
//
type NodesStatsRequest struct {
	IndexMetric []string
	Metric      []string
	NodeID      []string

	CompletionFields        []string
	FielddataFields         []string
	Fields                  []string
	Groups                  *bool
	IncludeSegmentFileSizes *bool
	Level                   string
	Timeout                 time.Duration
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
func (r NodesStatsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_nodes") + 1 + len(strings.Join(r.NodeID, ",")) + 1 + len("stats") + 1 + len(strings.Join(r.Metric, ",")) + 1 + len(strings.Join(r.IndexMetric, ",")))
	path.WriteString("/")
	path.WriteString("_nodes")
	if len(r.NodeID) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.NodeID, ","))
	}
	path.WriteString("/")
	path.WriteString("stats")
	if len(r.Metric) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Metric, ","))
	}
	if len(r.IndexMetric) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.IndexMetric, ","))
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

	if r.Groups != nil {
		params["groups"] = strconv.FormatBool(*r.Groups)
	}

	if r.IncludeSegmentFileSizes != nil {
		params["include_segment_file_sizes"] = strconv.FormatBool(*r.IncludeSegmentFileSizes)
	}

	if r.Level != "" {
		params["level"] = r.Level
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
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
func (f NodesStats) WithContext(v context.Context) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.ctx = v
	}
}

// WithIndexMetric - limit the information returned for `indices` metric to the specific index metrics. isn't used if `indices` (or `all`) metric isn't specified..
//
func (f NodesStats) WithIndexMetric(v ...string) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.IndexMetric = v
	}
}

// WithMetric - limit the information returned to the specified metrics.
//
func (f NodesStats) WithMetric(v ...string) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.Metric = v
	}
}

// WithNodeID - a list of node ids or names to limit the returned information; use `_local` to return information from the node you're connecting to, leave empty to get information from all nodes.
//
func (f NodesStats) WithNodeID(v ...string) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.NodeID = v
	}
}

// WithCompletionFields - a list of fields for `fielddata` and `suggest` index metric (supports wildcards).
//
func (f NodesStats) WithCompletionFields(v ...string) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.CompletionFields = v
	}
}

// WithFielddataFields - a list of fields for `fielddata` index metric (supports wildcards).
//
func (f NodesStats) WithFielddataFields(v ...string) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.FielddataFields = v
	}
}

// WithFields - a list of fields for `fielddata` and `completion` index metric (supports wildcards).
//
func (f NodesStats) WithFields(v ...string) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.Fields = v
	}
}

// WithGroups - a list of search groups for `search` index metric.
//
func (f NodesStats) WithGroups(v bool) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.Groups = &v
	}
}

// WithIncludeSegmentFileSizes - whether to report the aggregated disk usage of each one of the lucene index files (only applies if segment stats are requested).
//
func (f NodesStats) WithIncludeSegmentFileSizes(v bool) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.IncludeSegmentFileSizes = &v
	}
}

// WithLevel - return indices stats aggregated at index, node or shard level.
//
func (f NodesStats) WithLevel(v string) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.Level = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f NodesStats) WithTimeout(v time.Duration) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.Timeout = v
	}
}

// WithTypes - a list of document types for the `indexing` index metric.
//
func (f NodesStats) WithTypes(v ...string) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.Types = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f NodesStats) WithPretty() func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f NodesStats) WithHuman() func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f NodesStats) WithErrorTrace() func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f NodesStats) WithFilterPath(v ...string) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f NodesStats) WithHeader(h map[string]string) func(*NodesStatsRequest) {
	return func(r *NodesStatsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
