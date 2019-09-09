// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func newTermvectorsFunc(t Transport) Termvectors {
	return func(index string, o ...func(*TermvectorsRequest)) (*Response, error) {
		var r = TermvectorsRequest{Index: index}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// Termvectors returns information and statistics about terms in the fields of a particular document.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/docs-termvectors.html.
//
type Termvectors func(index string, o ...func(*TermvectorsRequest)) (*Response, error)

// TermvectorsRequest configures the Termvectors API request.
//
type TermvectorsRequest struct {
	Index        string
	DocumentType string
	DocumentID   string

	Body io.Reader

	Fields          []string
	FieldStatistics *bool
	Offsets         *bool
	Payloads        *bool
	Positions       *bool
	Preference      string
	Realtime        *bool
	Routing         string
	TermStatistics  *bool
	Version         *int
	VersionType     string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r TermvectorsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	if r.DocumentType == "" {
		r.DocumentType = "_doc"
	}

	path.Grow(1 + len(r.Index) + 1 + len(r.DocumentType) + 1 + len(r.DocumentID) + 1 + len("_termvectors"))
	path.WriteString("/")
	path.WriteString(r.Index)
	if r.DocumentType != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentType)
	}
	if r.DocumentID != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentID)
	}
	path.WriteString("/")
	path.WriteString("_termvectors")

	params = make(map[string]string)

	if len(r.Fields) > 0 {
		params["fields"] = strings.Join(r.Fields, ",")
	}

	if r.FieldStatistics != nil {
		params["field_statistics"] = strconv.FormatBool(*r.FieldStatistics)
	}

	if r.Offsets != nil {
		params["offsets"] = strconv.FormatBool(*r.Offsets)
	}

	if r.Payloads != nil {
		params["payloads"] = strconv.FormatBool(*r.Payloads)
	}

	if r.Positions != nil {
		params["positions"] = strconv.FormatBool(*r.Positions)
	}

	if r.Preference != "" {
		params["preference"] = r.Preference
	}

	if r.Realtime != nil {
		params["realtime"] = strconv.FormatBool(*r.Realtime)
	}

	if r.Routing != "" {
		params["routing"] = r.Routing
	}

	if r.TermStatistics != nil {
		params["term_statistics"] = strconv.FormatBool(*r.TermStatistics)
	}

	if r.Version != nil {
		params["version"] = strconv.FormatInt(int64(*r.Version), 10)
	}

	if r.VersionType != "" {
		params["version_type"] = r.VersionType
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
func (f Termvectors) WithContext(v context.Context) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.ctx = v
	}
}

// WithBody - Define parameters and or supply a document to get termvectors for. See documentation..
//
func (f Termvectors) WithBody(v io.Reader) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.Body = v
	}
}

// WithDocumentID - the ID of the document, when not specified a doc param should be supplied..
//
func (f Termvectors) WithDocumentID(v string) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.DocumentID = v
	}
}

// WithDocumentType - the type of the document..
//
func (f Termvectors) WithDocumentType(v string) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.DocumentType = v
	}
}

// WithFields - a list of fields to return..
//
func (f Termvectors) WithFields(v ...string) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.Fields = v
	}
}

// WithFieldStatistics - specifies if document count, sum of document frequencies and sum of total term frequencies should be returned..
//
func (f Termvectors) WithFieldStatistics(v bool) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.FieldStatistics = &v
	}
}

// WithOffsets - specifies if term offsets should be returned..
//
func (f Termvectors) WithOffsets(v bool) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.Offsets = &v
	}
}

// WithPayloads - specifies if term payloads should be returned..
//
func (f Termvectors) WithPayloads(v bool) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.Payloads = &v
	}
}

// WithPositions - specifies if term positions should be returned..
//
func (f Termvectors) WithPositions(v bool) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.Positions = &v
	}
}

// WithPreference - specify the node or shard the operation should be performed on (default: random)..
//
func (f Termvectors) WithPreference(v string) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.Preference = v
	}
}

// WithRealtime - specifies if request is real-time as opposed to near-real-time (default: true)..
//
func (f Termvectors) WithRealtime(v bool) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.Realtime = &v
	}
}

// WithRouting - specific routing value..
//
func (f Termvectors) WithRouting(v string) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.Routing = v
	}
}

// WithTermStatistics - specifies if total term frequency and document frequency should be returned..
//
func (f Termvectors) WithTermStatistics(v bool) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.TermStatistics = &v
	}
}

// WithVersion - explicit version number for concurrency control.
//
func (f Termvectors) WithVersion(v int) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.Version = &v
	}
}

// WithVersionType - specific version type.
//
func (f Termvectors) WithVersionType(v string) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.VersionType = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Termvectors) WithPretty() func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Termvectors) WithHuman() func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Termvectors) WithErrorTrace() func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Termvectors) WithFilterPath(v ...string) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Termvectors) WithHeader(h map[string]string) func(*TermvectorsRequest) {
	return func(r *TermvectorsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
