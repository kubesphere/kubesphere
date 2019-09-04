// Code generated from specification version 7.0.0 (5e798c1): DO NOT EDIT

package esapi

import (
	"context"
	"strings"
	"time"
)

func newNodesUsageFunc(t Transport) NodesUsage {
	return func(o ...func(*NodesUsageRequest)) (*Response, error) {
		var r = NodesUsageRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// NodesUsage returns low-level information about REST actions usage on nodes.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/cluster-nodes-usage.html.
//
type NodesUsage func(o ...func(*NodesUsageRequest)) (*Response, error)

// NodesUsageRequest configures the Nodes Usage API request.
//
type NodesUsageRequest struct {
	Metric  []string
	NodeID  []string
	Timeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r NodesUsageRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_nodes") + 1 + len(strings.Join(r.NodeID, ",")) + 1 + len("usage") + 1 + len(strings.Join(r.Metric, ",")))
	path.WriteString("/")
	path.WriteString("_nodes")
	if len(r.NodeID) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.NodeID, ","))
	}
	path.WriteString("/")
	path.WriteString("usage")
	if len(r.Metric) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Metric, ","))
	}

	params = make(map[string]string)

	if r.Timeout != 0 {
		params["timeout"] = time.Duration(r.Timeout * time.Millisecond).String()
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
func (f NodesUsage) WithContext(v context.Context) func(*NodesUsageRequest) {
	return func(r *NodesUsageRequest) {
		r.ctx = v
	}
}

// WithMetric - limit the information returned to the specified metrics.
//
func (f NodesUsage) WithMetric(v ...string) func(*NodesUsageRequest) {
	return func(r *NodesUsageRequest) {
		r.Metric = v
	}
}

// WithNodeID - a list of node ids or names to limit the returned information; use `_local` to return information from the node you're connecting to, leave empty to get information from all nodes.
//
func (f NodesUsage) WithNodeID(v ...string) func(*NodesUsageRequest) {
	return func(r *NodesUsageRequest) {
		r.NodeID = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f NodesUsage) WithTimeout(v time.Duration) func(*NodesUsageRequest) {
	return func(r *NodesUsageRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f NodesUsage) WithPretty() func(*NodesUsageRequest) {
	return func(r *NodesUsageRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f NodesUsage) WithHuman() func(*NodesUsageRequest) {
	return func(r *NodesUsageRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f NodesUsage) WithErrorTrace() func(*NodesUsageRequest) {
	return func(r *NodesUsageRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f NodesUsage) WithFilterPath(v ...string) func(*NodesUsageRequest) {
	return func(r *NodesUsageRequest) {
		r.FilterPath = v
	}
}
