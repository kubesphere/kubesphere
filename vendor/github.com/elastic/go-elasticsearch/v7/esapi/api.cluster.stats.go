// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newClusterStatsFunc(t Transport) ClusterStats {
	return func(o ...func(*ClusterStatsRequest)) (*Response, error) {
		var r = ClusterStatsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ClusterStats returns high-level overview of cluster statistics.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/cluster-stats.html.
//
type ClusterStats func(o ...func(*ClusterStatsRequest)) (*Response, error)

// ClusterStatsRequest configures the Cluster Stats API request.
//
type ClusterStatsRequest struct {
	NodeID []string

	FlatSettings *bool
	Timeout      time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r ClusterStatsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/nodes/_cluster/stats/nodes/") + len(strings.Join(r.NodeID, ",")))
	path.WriteString("/")
	path.WriteString("_cluster")
	path.WriteString("/")
	path.WriteString("stats")
	if len(r.NodeID) > 0 {
		path.WriteString("/")
		path.WriteString("nodes")
		path.WriteString("/")
		path.WriteString(strings.Join(r.NodeID, ","))
	}

	params = make(map[string]string)

	if r.FlatSettings != nil {
		params["flat_settings"] = strconv.FormatBool(*r.FlatSettings)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
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
func (f ClusterStats) WithContext(v context.Context) func(*ClusterStatsRequest) {
	return func(r *ClusterStatsRequest) {
		r.ctx = v
	}
}

// WithNodeID - a list of node ids or names to limit the returned information; use `_local` to return information from the node you're connecting to, leave empty to get information from all nodes.
//
func (f ClusterStats) WithNodeID(v ...string) func(*ClusterStatsRequest) {
	return func(r *ClusterStatsRequest) {
		r.NodeID = v
	}
}

// WithFlatSettings - return settings in flat format (default: false).
//
func (f ClusterStats) WithFlatSettings(v bool) func(*ClusterStatsRequest) {
	return func(r *ClusterStatsRequest) {
		r.FlatSettings = &v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f ClusterStats) WithTimeout(v time.Duration) func(*ClusterStatsRequest) {
	return func(r *ClusterStatsRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ClusterStats) WithPretty() func(*ClusterStatsRequest) {
	return func(r *ClusterStatsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ClusterStats) WithHuman() func(*ClusterStatsRequest) {
	return func(r *ClusterStatsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ClusterStats) WithErrorTrace() func(*ClusterStatsRequest) {
	return func(r *ClusterStatsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ClusterStats) WithFilterPath(v ...string) func(*ClusterStatsRequest) {
	return func(r *ClusterStatsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f ClusterStats) WithHeader(h map[string]string) func(*ClusterStatsRequest) {
	return func(r *ClusterStatsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
