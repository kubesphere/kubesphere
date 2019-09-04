// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newIngestGetPipelineFunc(t Transport) IngestGetPipeline {
	return func(o ...func(*IngestGetPipelineRequest)) (*Response, error) {
		var r = IngestGetPipelineRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IngestGetPipeline returns a pipeline.
//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/master/get-pipeline-api.html.
//
type IngestGetPipeline func(o ...func(*IngestGetPipelineRequest)) (*Response, error)

// IngestGetPipelineRequest configures the Ingest Get Pipeline API request.
//
type IngestGetPipelineRequest struct {
	PipelineID string

	MasterTimeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IngestGetPipelineRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_ingest") + 1 + len("pipeline") + 1 + len(r.PipelineID))
	path.WriteString("/")
	path.WriteString("_ingest")
	path.WriteString("/")
	path.WriteString("pipeline")
	if r.PipelineID != "" {
		path.WriteString("/")
		path.WriteString(r.PipelineID)
	}

	params = make(map[string]string)

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
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
func (f IngestGetPipeline) WithContext(v context.Context) func(*IngestGetPipelineRequest) {
	return func(r *IngestGetPipelineRequest) {
		r.ctx = v
	}
}

// WithPipelineID - comma separated list of pipeline ids. wildcards supported.
//
func (f IngestGetPipeline) WithPipelineID(v string) func(*IngestGetPipelineRequest) {
	return func(r *IngestGetPipelineRequest) {
		r.PipelineID = v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f IngestGetPipeline) WithMasterTimeout(v time.Duration) func(*IngestGetPipelineRequest) {
	return func(r *IngestGetPipelineRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IngestGetPipeline) WithPretty() func(*IngestGetPipelineRequest) {
	return func(r *IngestGetPipelineRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IngestGetPipeline) WithHuman() func(*IngestGetPipelineRequest) {
	return func(r *IngestGetPipelineRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IngestGetPipeline) WithErrorTrace() func(*IngestGetPipelineRequest) {
	return func(r *IngestGetPipelineRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IngestGetPipeline) WithFilterPath(v ...string) func(*IngestGetPipelineRequest) {
	return func(r *IngestGetPipelineRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IngestGetPipeline) WithHeader(h map[string]string) func(*IngestGetPipelineRequest) {
	return func(r *IngestGetPipelineRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
