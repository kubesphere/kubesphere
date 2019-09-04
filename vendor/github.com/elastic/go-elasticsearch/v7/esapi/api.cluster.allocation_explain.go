// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newClusterAllocationExplainFunc(t Transport) ClusterAllocationExplain {
	return func(o ...func(*ClusterAllocationExplainRequest)) (*Response, error) {
		var r = ClusterAllocationExplainRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ClusterAllocationExplain provides explanations for shard allocations in the cluster.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/cluster-allocation-explain.html.
//
type ClusterAllocationExplain func(o ...func(*ClusterAllocationExplainRequest)) (*Response, error)

// ClusterAllocationExplainRequest configures the Cluster Allocation Explain API request.
//
type ClusterAllocationExplainRequest struct {
	Body io.Reader

	IncludeDiskInfo     *bool
	IncludeYesDecisions *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r ClusterAllocationExplainRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(len("/_cluster/allocation/explain"))
	path.WriteString("/_cluster/allocation/explain")

	params = make(map[string]string)

	if r.IncludeDiskInfo != nil {
		params["include_disk_info"] = strconv.FormatBool(*r.IncludeDiskInfo)
	}

	if r.IncludeYesDecisions != nil {
		params["include_yes_decisions"] = strconv.FormatBool(*r.IncludeYesDecisions)
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
func (f ClusterAllocationExplain) WithContext(v context.Context) func(*ClusterAllocationExplainRequest) {
	return func(r *ClusterAllocationExplainRequest) {
		r.ctx = v
	}
}

// WithBody - The index, shard, and primary flag to explain. Empty means 'explain the first unassigned shard'.
//
func (f ClusterAllocationExplain) WithBody(v io.Reader) func(*ClusterAllocationExplainRequest) {
	return func(r *ClusterAllocationExplainRequest) {
		r.Body = v
	}
}

// WithIncludeDiskInfo - return information about disk usage and shard sizes (default: false).
//
func (f ClusterAllocationExplain) WithIncludeDiskInfo(v bool) func(*ClusterAllocationExplainRequest) {
	return func(r *ClusterAllocationExplainRequest) {
		r.IncludeDiskInfo = &v
	}
}

// WithIncludeYesDecisions - return 'yes' decisions in explanation (default: false).
//
func (f ClusterAllocationExplain) WithIncludeYesDecisions(v bool) func(*ClusterAllocationExplainRequest) {
	return func(r *ClusterAllocationExplainRequest) {
		r.IncludeYesDecisions = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ClusterAllocationExplain) WithPretty() func(*ClusterAllocationExplainRequest) {
	return func(r *ClusterAllocationExplainRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ClusterAllocationExplain) WithHuman() func(*ClusterAllocationExplainRequest) {
	return func(r *ClusterAllocationExplainRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ClusterAllocationExplain) WithErrorTrace() func(*ClusterAllocationExplainRequest) {
	return func(r *ClusterAllocationExplainRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ClusterAllocationExplain) WithFilterPath(v ...string) func(*ClusterAllocationExplainRequest) {
	return func(r *ClusterAllocationExplainRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f ClusterAllocationExplain) WithHeader(h map[string]string) func(*ClusterAllocationExplainRequest) {
	return func(r *ClusterAllocationExplainRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
