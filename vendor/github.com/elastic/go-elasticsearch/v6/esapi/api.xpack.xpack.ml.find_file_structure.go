// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newXPackMLFindFileStructureFunc(t Transport) XPackMLFindFileStructure {
	return func(body io.Reader, o ...func(*XPackMLFindFileStructureRequest)) (*Response, error) {
		var r = XPackMLFindFileStructureRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMLFindFileStructure - http://www.elastic.co/guide/en/elasticsearch/reference/6.7/ml-find-file-structure.html
//
type XPackMLFindFileStructure func(body io.Reader, o ...func(*XPackMLFindFileStructureRequest)) (*Response, error)

// XPackMLFindFileStructureRequest configures the X PackML Find File Structure API request.
//
type XPackMLFindFileStructureRequest struct {
	Body io.Reader

	Charset          string
	ColumnNames      []string
	Delimiter        string
	Explain          *bool
	Format           string
	GrokPattern      string
	HasHeaderRow     *bool
	LinesToSample    *int
	Quote            string
	ShouldTrimFields *bool
	Timeout          time.Duration
	TimestampField   string
	TimestampFormat  string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMLFindFileStructureRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(len("/_xpack/ml/find_file_structure"))
	path.WriteString("/_xpack/ml/find_file_structure")

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
func (f XPackMLFindFileStructure) WithContext(v context.Context) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.ctx = v
	}
}

// WithCharset - optional parameter to specify the character set of the file.
//
func (f XPackMLFindFileStructure) WithCharset(v string) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.Charset = v
	}
}

// WithColumnNames - optional parameter containing a comma separated list of the column names for a delimited file.
//
func (f XPackMLFindFileStructure) WithColumnNames(v ...string) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.ColumnNames = v
	}
}

// WithDelimiter - optional parameter to specify the delimiter character for a delimited file - must be a single character.
//
func (f XPackMLFindFileStructure) WithDelimiter(v string) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.Delimiter = v
	}
}

// WithExplain - whether to include a commentary on how the structure was derived.
//
func (f XPackMLFindFileStructure) WithExplain(v bool) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.Explain = &v
	}
}

// WithFormat - optional parameter to specify the high level file format.
//
func (f XPackMLFindFileStructure) WithFormat(v string) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.Format = v
	}
}

// WithGrokPattern - optional parameter to specify the grok pattern that should be used to extract fields from messages in a semi-structured text file.
//
func (f XPackMLFindFileStructure) WithGrokPattern(v string) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.GrokPattern = v
	}
}

// WithHasHeaderRow - optional parameter to specify whether a delimited file includes the column names in its first row.
//
func (f XPackMLFindFileStructure) WithHasHeaderRow(v bool) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.HasHeaderRow = &v
	}
}

// WithLinesToSample - how many lines of the file should be included in the analysis.
//
func (f XPackMLFindFileStructure) WithLinesToSample(v int) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.LinesToSample = &v
	}
}

// WithQuote - optional parameter to specify the quote character for a delimited file - must be a single character.
//
func (f XPackMLFindFileStructure) WithQuote(v string) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.Quote = v
	}
}

// WithShouldTrimFields - optional parameter to specify whether the values between delimiters in a delimited file should have whitespace trimmed from them.
//
func (f XPackMLFindFileStructure) WithShouldTrimFields(v bool) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.ShouldTrimFields = &v
	}
}

// WithTimeout - timeout after which the analysis will be aborted.
//
func (f XPackMLFindFileStructure) WithTimeout(v time.Duration) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.Timeout = v
	}
}

// WithTimestampField - optional parameter to specify the timestamp field in the file.
//
func (f XPackMLFindFileStructure) WithTimestampField(v string) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.TimestampField = v
	}
}

// WithTimestampFormat - optional parameter to specify the timestamp format in the file - may be either a joda or java time format.
//
func (f XPackMLFindFileStructure) WithTimestampFormat(v string) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.TimestampFormat = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMLFindFileStructure) WithPretty() func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMLFindFileStructure) WithHuman() func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMLFindFileStructure) WithErrorTrace() func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMLFindFileStructure) WithFilterPath(v ...string) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMLFindFileStructure) WithHeader(h map[string]string) func(*XPackMLFindFileStructureRequest) {
	return func(r *XPackMLFindFileStructureRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
