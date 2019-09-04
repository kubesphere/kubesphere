// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newIndicesAnalyzeFunc(t Transport) IndicesAnalyze {
	return func(o ...func(*IndicesAnalyzeRequest)) (*Response, error) {
		var r = IndicesAnalyzeRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// IndicesAnalyze performs the analysis process on a text and return the tokens breakdown of the text.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/indices-analyze.html.
//
type IndicesAnalyze func(o ...func(*IndicesAnalyzeRequest)) (*Response, error)

// IndicesAnalyzeRequest configures the Indices Analyze API request.
//
type IndicesAnalyzeRequest struct {
	Index string

	Body io.Reader

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r IndicesAnalyzeRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(r.Index) + 1 + len("_analyze"))
	if r.Index != "" {
		path.WriteString("/")
		path.WriteString(r.Index)
	}
	path.WriteString("/")
	path.WriteString("_analyze")

	params = make(map[string]string)

	if r.Index != "" {
		params["index"] = r.Index
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
func (f IndicesAnalyze) WithContext(v context.Context) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.ctx = v
	}
}

// WithBody - Define analyzer/tokenizer parameters and the text on which the analysis should be performed.
//
func (f IndicesAnalyze) WithBody(v io.Reader) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.Body = v
	}
}

// WithIndex - the name of the index to scope the operation.
//
func (f IndicesAnalyze) WithIndex(v string) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.Index = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f IndicesAnalyze) WithPretty() func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f IndicesAnalyze) WithHuman() func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f IndicesAnalyze) WithErrorTrace() func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f IndicesAnalyze) WithFilterPath(v ...string) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f IndicesAnalyze) WithHeader(h map[string]string) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
