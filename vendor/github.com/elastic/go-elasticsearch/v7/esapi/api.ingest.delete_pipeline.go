// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func newIngestDeletePipelineFunc(t Transport) IngestDeletePipeline {
	return func(id string, o ...func(*IngestDeletePipelineRequest)) (*Response, error) {
		var r = IngestDeletePipelineRequest{PipelineID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IngestDeletePipeline deletes a pipeline.
//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/master/delete-pipeline-api.html.
//
type IngestDeletePipeline func(id string, o ...func(*IngestDeletePipelineRequest)) (*Response, error)

// IngestDeletePipelineRequest configures the Ingest Delete Pipeline API request.
//
type IngestDeletePipelineRequest struct {
	PipelineID string

	MasterTimeout time.Duration
	Timeout       time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IngestDeletePipelineRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "DELETE"

	path.Grow(1 + len("_ingest") + 1 + len("pipeline") + 1 + len(r.PipelineID))
	path.WriteString("/")
	path.WriteString("_ingest")
	path.WriteString("/")
	path.WriteString("pipeline")
	path.WriteString("/")
	path.WriteString(r.PipelineID)

	params = make(map[string]string)

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
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
func (f IngestDeletePipeline) WithContext(v context.Context) func(*IngestDeletePipelineRequest) {
	return func(r *IngestDeletePipelineRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f IngestDeletePipeline) WithMasterTimeout(v time.Duration) func(*IngestDeletePipelineRequest) {
	return func(r *IngestDeletePipelineRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IngestDeletePipeline) WithTimeout(v time.Duration) func(*IngestDeletePipelineRequest) {
	return func(r *IngestDeletePipelineRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IngestDeletePipeline) WithPretty() func(*IngestDeletePipelineRequest) {
	return func(r *IngestDeletePipelineRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IngestDeletePipeline) WithHuman() func(*IngestDeletePipelineRequest) {
	return func(r *IngestDeletePipelineRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IngestDeletePipeline) WithErrorTrace() func(*IngestDeletePipelineRequest) {
	return func(r *IngestDeletePipelineRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IngestDeletePipeline) WithFilterPath(v ...string) func(*IngestDeletePipelineRequest) {
	return func(r *IngestDeletePipelineRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IngestDeletePipeline) WithHeader(h map[string]string) func(*IngestDeletePipelineRequest) {
	return func(r *IngestDeletePipelineRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
