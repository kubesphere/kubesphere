// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

func newXPackGraphExploreFunc(t Transport) XPackGraphExplore {
	return func(o ...func(*XPackGraphExploreRequest)) (*Response, error) {
		var r = XPackGraphExploreRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackGraphExplore - https://www.elastic.co/guide/en/elasticsearch/reference/current/graph-explore-api.html
//
type XPackGraphExplore func(o ...func(*XPackGraphExploreRequest)) (*Response, error)

// XPackGraphExploreRequest configures the X Pack Graph Explore API request.
//
type XPackGraphExploreRequest struct {
	Index        []string
	DocumentType []string

	Body io.Reader

	Routing string
	Timeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackGraphExploreRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len(strings.Join(r.DocumentType, ",")) + 1 + len("_xpack") + 1 + len("graph") + 1 + len("_explore"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	if len(r.DocumentType) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.DocumentType, ","))
	}
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("graph")
	path.WriteString("/")
	path.WriteString("_explore")

	params = make(map[string]string)

	if r.Routing != "" {
		params["routing"] = r.Routing
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
func (f XPackGraphExplore) WithContext(v context.Context) func(*XPackGraphExploreRequest) {
	return func(r *XPackGraphExploreRequest) {
		r.ctx = v
	}
}

// WithBody - Graph Query DSL.
//
func (f XPackGraphExplore) WithBody(v io.Reader) func(*XPackGraphExploreRequest) {
	return func(r *XPackGraphExploreRequest) {
		r.Body = v
	}
}

// WithIndex - a list of index names to search; use _all to perform the operation on all indices.
//
func (f XPackGraphExplore) WithIndex(v ...string) func(*XPackGraphExploreRequest) {
	return func(r *XPackGraphExploreRequest) {
		r.Index = v
	}
}

// WithDocumentType - a list of document types to search; leave empty to perform the operation on all types.
//
func (f XPackGraphExplore) WithDocumentType(v ...string) func(*XPackGraphExploreRequest) {
	return func(r *XPackGraphExploreRequest) {
		r.DocumentType = v
	}
}

// WithRouting - specific routing value.
//
func (f XPackGraphExplore) WithRouting(v string) func(*XPackGraphExploreRequest) {
	return func(r *XPackGraphExploreRequest) {
		r.Routing = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f XPackGraphExplore) WithTimeout(v time.Duration) func(*XPackGraphExploreRequest) {
	return func(r *XPackGraphExploreRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackGraphExplore) WithPretty() func(*XPackGraphExploreRequest) {
	return func(r *XPackGraphExploreRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackGraphExplore) WithHuman() func(*XPackGraphExploreRequest) {
	return func(r *XPackGraphExploreRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackGraphExplore) WithErrorTrace() func(*XPackGraphExploreRequest) {
	return func(r *XPackGraphExploreRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackGraphExplore) WithFilterPath(v ...string) func(*XPackGraphExploreRequest) {
	return func(r *XPackGraphExploreRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackGraphExplore) WithHeader(h map[string]string) func(*XPackGraphExploreRequest) {
	return func(r *XPackGraphExploreRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
