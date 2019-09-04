// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

func newIngestPutPipelineFunc(t Transport) IngestPutPipeline {
	return func(id string, body io.Reader, o ...func(*IngestPutPipelineRequest)) (*Response, error) {
		var r = IngestPutPipelineRequest{PipelineID: id, Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IngestPutPipeline creates or updates a pipeline.
//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/master/put-pipeline-api.html.
//
type IngestPutPipeline func(id string, body io.Reader, o ...func(*IngestPutPipelineRequest)) (*Response, error)

// IngestPutPipelineRequest configures the Ingest Put Pipeline API request.
//
type IngestPutPipelineRequest struct {
	PipelineID string

	Body io.Reader

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
func (r IngestPutPipelineRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

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
func (f IngestPutPipeline) WithContext(v context.Context) func(*IngestPutPipelineRequest) {
	return func(r *IngestPutPipelineRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f IngestPutPipeline) WithMasterTimeout(v time.Duration) func(*IngestPutPipelineRequest) {
	return func(r *IngestPutPipelineRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f IngestPutPipeline) WithTimeout(v time.Duration) func(*IngestPutPipelineRequest) {
	return func(r *IngestPutPipelineRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IngestPutPipeline) WithPretty() func(*IngestPutPipelineRequest) {
	return func(r *IngestPutPipelineRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IngestPutPipeline) WithHuman() func(*IngestPutPipelineRequest) {
	return func(r *IngestPutPipelineRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IngestPutPipeline) WithErrorTrace() func(*IngestPutPipelineRequest) {
	return func(r *IngestPutPipelineRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IngestPutPipeline) WithFilterPath(v ...string) func(*IngestPutPipelineRequest) {
	return func(r *IngestPutPipelineRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IngestPutPipeline) WithHeader(h map[string]string) func(*IngestPutPipelineRequest) {
	return func(r *IngestPutPipelineRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
