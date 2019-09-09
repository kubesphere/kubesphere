// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newMLFindFileStructureFunc(t Transport) MLFindFileStructure {
	return func(body io.Reader, o ...func(*MLFindFileStructureRequest)) (*Response, error) {
		var r = MLFindFileStructureRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// MLFindFileStructure - http://www.elastic.co/guide/en/elasticsearch/reference/current/ml-find-file-structure.html
//
type MLFindFileStructure func(body io.Reader, o ...func(*MLFindFileStructureRequest)) (*Response, error)

// MLFindFileStructureRequest configures the ML Find File Structure API request.
//
type MLFindFileStructureRequest struct {
	Body io.Reader

	Charset            string
	ColumnNames        []string
	Delimiter          string
	Explain            *bool
	Format             string
	GrokPattern        string
	HasHeaderRow       *bool
	LineMergeSizeLimit *int
	LinesToSample      *int
	Quote              string
	ShouldTrimFields   *bool
	Timeout            time.Duration
	TimestampField     string
	TimestampFormat    string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r MLFindFileStructureRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(len("/_ml/find_file_structure"))
	path.WriteString("/_ml/find_file_structure")

	params = make(map[string]string)

	if r.Charset != "" {
		params["charset"] = r.Charset
	}

	if len(r.ColumnNames) > 0 {
		params["column_names"] = strings.Join(r.ColumnNames, ",")
	}

	if r.Delimiter != "" {
		params["delimiter"] = r.Delimiter
	}

	if r.Explain != nil {
		params["explain"] = strconv.FormatBool(*r.Explain)
	}

	if r.Format != "" {
		params["format"] = r.Format
	}

	if r.GrokPattern != "" {
		params["grok_pattern"] = r.GrokPattern
	}

	if r.HasHeaderRow != nil {
		params["has_header_row"] = strconv.FormatBool(*r.HasHeaderRow)
	}

	if r.LineMergeSizeLimit != nil {
		params["line_merge_size_limit"] = strconv.FormatInt(int64(*r.LineMergeSizeLimit), 10)
	}

	if r.LinesToSample != nil {
		params["lines_to_sample"] = strconv.FormatInt(int64(*r.LinesToSample), 10)
	}

	if r.Quote != "" {
		params["quote"] = r.Quote
	}

	if r.ShouldTrimFields != nil {
		params["should_trim_fields"] = strconv.FormatBool(*r.ShouldTrimFields)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
	}

	if r.TimestampField != "" {
		params["timestamp_field"] = r.TimestampField
	}

	if r.TimestampFormat != "" {
		params["timestamp_format"] = r.TimestampFormat
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
func (f MLFindFileStructure) WithContext(v context.Context) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.ctx = v
	}
}

// WithCharset - optional parameter to specify the character set of the file.
//
func (f MLFindFileStructure) WithCharset(v string) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.Charset = v
	}
}

// WithColumnNames - optional parameter containing a comma separated list of the column names for a delimited file.
//
func (f MLFindFileStructure) WithColumnNames(v ...string) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.ColumnNames = v
	}
}

// WithDelimiter - optional parameter to specify the delimiter character for a delimited file - must be a single character.
//
func (f MLFindFileStructure) WithDelimiter(v string) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.Delimiter = v
	}
}

// WithExplain - whether to include a commentary on how the structure was derived.
//
func (f MLFindFileStructure) WithExplain(v bool) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.Explain = &v
	}
}

// WithFormat - optional parameter to specify the high level file format.
//
func (f MLFindFileStructure) WithFormat(v string) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.Format = v
	}
}

// WithGrokPattern - optional parameter to specify the grok pattern that should be used to extract fields from messages in a semi-structured text file.
//
func (f MLFindFileStructure) WithGrokPattern(v string) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.GrokPattern = v
	}
}

// WithHasHeaderRow - optional parameter to specify whether a delimited file includes the column names in its first row.
//
func (f MLFindFileStructure) WithHasHeaderRow(v bool) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.HasHeaderRow = &v
	}
}

// WithLineMergeSizeLimit - maximum number of characters permitted in a single message when lines are merged to create messages..
//
func (f MLFindFileStructure) WithLineMergeSizeLimit(v int) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.LineMergeSizeLimit = &v
	}
}

// WithLinesToSample - how many lines of the file should be included in the analysis.
//
func (f MLFindFileStructure) WithLinesToSample(v int) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.LinesToSample = &v
	}
}

// WithQuote - optional parameter to specify the quote character for a delimited file - must be a single character.
//
func (f MLFindFileStructure) WithQuote(v string) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.Quote = v
	}
}

// WithShouldTrimFields - optional parameter to specify whether the values between delimiters in a delimited file should have whitespace trimmed from them.
//
func (f MLFindFileStructure) WithShouldTrimFields(v bool) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.ShouldTrimFields = &v
	}
}

// WithTimeout - timeout after which the analysis will be aborted.
//
func (f MLFindFileStructure) WithTimeout(v time.Duration) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.Timeout = v
	}
}

// WithTimestampField - optional parameter to specify the timestamp field in the file.
//
func (f MLFindFileStructure) WithTimestampField(v string) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.TimestampField = v
	}
}

// WithTimestampFormat - optional parameter to specify the timestamp format in the file - may be either a joda or java time format.
//
func (f MLFindFileStructure) WithTimestampFormat(v string) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.TimestampFormat = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f MLFindFileStructure) WithPretty() func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f MLFindFileStructure) WithHuman() func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f MLFindFileStructure) WithErrorTrace() func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f MLFindFileStructure) WithFilterPath(v ...string) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f MLFindFileStructure) WithHeader(h map[string]string) func(*MLFindFileStructureRequest) {
	return func(r *MLFindFileStructureRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
