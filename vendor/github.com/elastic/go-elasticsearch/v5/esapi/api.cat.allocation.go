// Code generated from specification version 5.6.15: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newCatAllocationFunc(t Transport) CatAllocation {
	return func(o ...func(*CatAllocationRequest)) (*Response, error) {
		var r = CatAllocationRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CatAllocation provides a snapshot of how many shards are allocated to each data node and how much disk space they are using.
//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/5.x/cat-allocation.html.
//
type CatAllocation func(o ...func(*CatAllocationRequest)) (*Response, error)

// CatAllocationRequest configures the Cat Allocation API request.
//
type CatAllocationRequest struct {
	NodeID []string

	Bytes         string
	Format        string
	H             []string
	Help          *bool
	Local         *bool
	MasterTimeout time.Duration
	S             []string
	V             *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CatAllocationRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_cat") + 1 + len("allocation") + 1 + len(strings.Join(r.NodeID, ",")))
	path.WriteString("/")
	path.WriteString("_cat")
	path.WriteString("/")
	path.WriteString("allocation")
	if len(r.NodeID) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.NodeID, ","))
	}

	params = make(map[string]string)

	if r.Bytes != "" {
		params["bytes"] = r.Bytes
	}

	if r.Format != "" {
		params["format"] = r.Format
	}

	if len(r.H) > 0 {
		params["h"] = strings.Join(r.H, ",")
	}

	if r.Help != nil {
		params["help"] = strconv.FormatBool(*r.Help)
	}

	if r.Local != nil {
		params["local"] = strconv.FormatBool(*r.Local)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if len(r.S) > 0 {
		params["s"] = strings.Join(r.S, ",")
	}

	if r.V != nil {
		params["v"] = strconv.FormatBool(*r.V)
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
func (f CatAllocation) WithContext(v context.Context) func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.ctx = v
	}
}

// WithNodeID - a list of node ids or names to limit the returned information.
//
func (f CatAllocation) WithNodeID(v ...string) func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.NodeID = v
	}
}

// WithBytes - the unit in which to display byte values.
//
func (f CatAllocation) WithBytes(v string) func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.Bytes = v
	}
}

// WithFormat - a short version of the accept header, e.g. json, yaml.
//
func (f CatAllocation) WithFormat(v string) func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.Format = v
	}
}

// WithH - comma-separated list of column names to display.
//
func (f CatAllocation) WithH(v ...string) func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.H = v
	}
}

// WithHelp - return help information.
//
func (f CatAllocation) WithHelp(v bool) func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.Help = &v
	}
}

// WithLocal - return local information, do not retrieve the state from master node (default: false).
//
func (f CatAllocation) WithLocal(v bool) func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.Local = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f CatAllocation) WithMasterTimeout(v time.Duration) func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.MasterTimeout = v
	}
}

// WithS - comma-separated list of column names or column aliases to sort by.
//
func (f CatAllocation) WithS(v ...string) func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.S = v
	}
}

// WithV - verbose mode. display column headers.
//
func (f CatAllocation) WithV(v bool) func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.V = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CatAllocation) WithPretty() func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CatAllocation) WithHuman() func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CatAllocation) WithErrorTrace() func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CatAllocation) WithFilterPath(v ...string) func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CatAllocation) WithHeader(h map[string]string) func(*CatAllocationRequest) {
	return func(r *CatAllocationRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
