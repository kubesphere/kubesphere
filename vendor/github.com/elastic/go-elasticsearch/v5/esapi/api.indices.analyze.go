// Code generated from specification version 5.6.15: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
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
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/5.x/indices-analyze.html.
//
type IndicesAnalyze func(o ...func(*IndicesAnalyzeRequest)) (*Response, error)

// IndicesAnalyzeRequest configures the Indices Analyze API request.
//
type IndicesAnalyzeRequest struct {
	Index string

	Body io.Reader

	Analyzer    string
	Attributes  []string
	CharFilter  []string
	Explain     *bool
	Field       string
	Filter      []string
	Format      string
	PreferLocal *bool
	Text        []string
	Tokenizer   string

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

	if r.Analyzer != "" {
		params["analyzer"] = r.Analyzer
	}

	if len(r.Attributes) > 0 {
		params["attributes"] = strings.Join(r.Attributes, ",")
	}

	if len(r.CharFilter) > 0 {
		params["char_filter"] = strings.Join(r.CharFilter, ",")
	}

	if r.Explain != nil {
		params["explain"] = strconv.FormatBool(*r.Explain)
	}

	if r.Field != "" {
		params["field"] = r.Field
	}

	if len(r.Filter) > 0 {
		params["filter"] = strings.Join(r.Filter, ",")
	}

	if r.Format != "" {
		params["format"] = r.Format
	}

	if r.Index != "" {
		params["index"] = r.Index
	}

	if r.PreferLocal != nil {
		params["prefer_local"] = strconv.FormatBool(*r.PreferLocal)
	}

	if len(r.Text) > 0 {
		params["text"] = strings.Join(r.Text, ",")
	}

	if r.Tokenizer != "" {
		params["tokenizer"] = r.Tokenizer
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

// WithBody - The text on which the analysis should be performed.
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

// WithAnalyzer - the name of the analyzer to use.
//
func (f IndicesAnalyze) WithAnalyzer(v string) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.Analyzer = v
	}
}

// WithAttributes - a list of token attributes to output, this parameter works only with `explain=true`.
//
func (f IndicesAnalyze) WithAttributes(v ...string) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.Attributes = v
	}
}

// WithCharFilter - a list of character filters to use for the analysis.
//
func (f IndicesAnalyze) WithCharFilter(v ...string) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.CharFilter = v
	}
}

// WithExplain - with `true`, outputs more advanced details. (default: false).
//
func (f IndicesAnalyze) WithExplain(v bool) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.Explain = &v
	}
}

// WithField - use the analyzer configured for this field (instead of passing the analyzer name).
//
func (f IndicesAnalyze) WithField(v string) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.Field = v
	}
}

// WithFilter - a list of filters to use for the analysis.
//
func (f IndicesAnalyze) WithFilter(v ...string) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.Filter = v
	}
}

// WithFormat - format of the output.
//
func (f IndicesAnalyze) WithFormat(v string) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.Format = v
	}
}

// WithPreferLocal - with `true`, specify that a local shard should be used if available, with `false`, use a random shard (default: true).
//
func (f IndicesAnalyze) WithPreferLocal(v bool) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.PreferLocal = &v
	}
}

// WithText - the text on which the analysis should be performed (when request body is not used).
//
func (f IndicesAnalyze) WithText(v ...string) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.Text = v
	}
}

// WithTokenizer - the name of the tokenizer to use for the analysis.
//
func (f IndicesAnalyze) WithTokenizer(v string) func(*IndicesAnalyzeRequest) {
	return func(r *IndicesAnalyzeRequest) {
		r.Tokenizer = v
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
