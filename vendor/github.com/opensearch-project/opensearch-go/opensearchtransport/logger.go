// SPDX-License-Identifier: Apache-2.0
//
// The OpenSearch Contributors require contributions made to
// this file be licensed under the Apache-2.0 license or a
// compatible open source license.
//
// Modifications Copyright OpenSearch Contributors. See
// GitHub history for details.

// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package opensearchtransport

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var debugLogger DebuggingLogger

// Logger defines an interface for logging request and response.
//
type Logger interface {
	// LogRoundTrip should not modify the request or response, except for consuming and closing the body.
	// Implementations have to check for nil values in request and response.
	LogRoundTrip(*http.Request, *http.Response, error, time.Time, time.Duration) error
	// RequestBodyEnabled makes the client pass a copy of request body to the logger.
	RequestBodyEnabled() bool
	// ResponseBodyEnabled makes the client pass a copy of response body to the logger.
	ResponseBodyEnabled() bool
}

// DebuggingLogger defines the interface for a debugging logger.
//
type DebuggingLogger interface {
	Log(a ...interface{}) error
	Logf(format string, a ...interface{}) error
}

// TextLogger prints the log message in plain text.
//
type TextLogger struct {
	Output             io.Writer
	EnableRequestBody  bool
	EnableResponseBody bool
}

// ColorLogger prints the log message in a terminal-optimized plain text.
//
type ColorLogger struct {
	Output             io.Writer
	EnableRequestBody  bool
	EnableResponseBody bool
}

// CurlLogger prints the log message as a runnable curl command.
//
type CurlLogger struct {
	Output             io.Writer
	EnableRequestBody  bool
	EnableResponseBody bool
}

// JSONLogger prints the log message as JSON.
//
type JSONLogger struct {
	Output             io.Writer
	EnableRequestBody  bool
	EnableResponseBody bool
}

// debuggingLogger prints debug messages as plain text.
//
type debuggingLogger struct {
	Output io.Writer
}

// LogRoundTrip prints the information about request and response.
//
func (l *TextLogger) LogRoundTrip(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) error {
	fmt.Fprintf(l.Output, "%s %s %s [status:%d request:%s]\n",
		start.Format(time.RFC3339),
		req.Method,
		req.URL.String(),
		resStatusCode(res),
		dur.Truncate(time.Millisecond),
	)
	if l.RequestBodyEnabled() && req != nil && req.Body != nil && req.Body != http.NoBody {
		var buf bytes.Buffer
		if req.GetBody != nil {
			b, _ := req.GetBody()
			buf.ReadFrom(b)
		} else {
			buf.ReadFrom(req.Body)
		}
		logBodyAsText(l.Output, &buf, ">")
	}
	if l.ResponseBodyEnabled() && res != nil && res.Body != nil && res.Body != http.NoBody {
		defer res.Body.Close()
		var buf bytes.Buffer
		buf.ReadFrom(res.Body)
		logBodyAsText(l.Output, &buf, "<")
	}
	if err != nil {
		fmt.Fprintf(l.Output, "! ERROR: %v\n", err)
	}
	return nil
}

// RequestBodyEnabled returns true when the request body should be logged.
func (l *TextLogger) RequestBodyEnabled() bool { return l.EnableRequestBody }

// ResponseBodyEnabled returns true when the response body should be logged.
func (l *TextLogger) ResponseBodyEnabled() bool { return l.EnableResponseBody }

// LogRoundTrip prints the information about request and response.
//
func (l *ColorLogger) LogRoundTrip(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) error {
	query, _ := url.QueryUnescape(req.URL.RawQuery)
	if query != "" {
		query = "?" + query
	}

	var (
		status string
		color  string
	)

	status = res.Status
	switch {
	case res.StatusCode > 0 && res.StatusCode < 300:
		color = "\x1b[32m"
	case res.StatusCode > 299 && res.StatusCode < 500:
		color = "\x1b[33m"
	case res.StatusCode > 499:
		color = "\x1b[31m"
	default:
		status = "ERROR"
		color = "\x1b[31;4m"
	}

	fmt.Fprintf(l.Output, "%6s \x1b[1;4m%s://%s%s\x1b[0m%s %s%s\x1b[0m \x1b[2m%s\x1b[0m\n",
		req.Method,
		req.URL.Scheme,
		req.URL.Host,
		req.URL.Path,
		query,
		color,
		status,
		dur.Truncate(time.Millisecond),
	)

	if l.RequestBodyEnabled() && req != nil && req.Body != nil && req.Body != http.NoBody {
		var buf bytes.Buffer
		if req.GetBody != nil {
			b, _ := req.GetBody()
			buf.ReadFrom(b)
		} else {
			buf.ReadFrom(req.Body)
		}
		fmt.Fprint(l.Output, "\x1b[2m")
		logBodyAsText(l.Output, &buf, "       »")
		fmt.Fprint(l.Output, "\x1b[0m")
	}

	if l.ResponseBodyEnabled() && res != nil && res.Body != nil && res.Body != http.NoBody {
		defer res.Body.Close()
		var buf bytes.Buffer
		buf.ReadFrom(res.Body)
		fmt.Fprint(l.Output, "\x1b[2m")
		logBodyAsText(l.Output, &buf, "       «")
		fmt.Fprint(l.Output, "\x1b[0m")
	}

	if err != nil {
		fmt.Fprintf(l.Output, "\x1b[31;1m» ERROR \x1b[31m%v\x1b[0m\n", err)
	}

	if l.RequestBodyEnabled() || l.ResponseBodyEnabled() {
		fmt.Fprintf(l.Output, "\x1b[2m%s\x1b[0m\n", strings.Repeat("─", 80))
	}
	return nil
}

// RequestBodyEnabled returns true when the request body should be logged.
func (l *ColorLogger) RequestBodyEnabled() bool { return l.EnableRequestBody }

// ResponseBodyEnabled returns true when the response body should be logged.
func (l *ColorLogger) ResponseBodyEnabled() bool { return l.EnableResponseBody }

// LogRoundTrip prints the information about request and response.
//
func (l *CurlLogger) LogRoundTrip(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) error {
	var b bytes.Buffer

	var query string
	qvalues := url.Values{}
	for k, v := range req.URL.Query() {
		if k == "pretty" {
			continue
		}
		for _, qv := range v {
			qvalues.Add(k, qv)
		}
	}
	if len(qvalues) > 0 {
		query = qvalues.Encode()
	}

	b.WriteString(`curl`)
	if req.Method == "HEAD" {
		b.WriteString(" --head")
	} else {
		fmt.Fprintf(&b, " -X %s", req.Method)
	}

	if len(req.Header) > 0 {
		for k, vv := range req.Header {
			if k == "Authorization" || k == "User-Agent" {
				continue
			}
			v := strings.Join(vv, ",")
			b.WriteString(fmt.Sprintf(" -H '%s: %s'", k, v))
		}
	}

	b.WriteString(" 'http://localhost:9200")
	b.WriteString(req.URL.Path)
	b.WriteString("?pretty")
	if query != "" {
		fmt.Fprintf(&b, "&%s", query)
	}
	b.WriteString("'")

	if req != nil && req.Body != nil && req.Body != http.NoBody {
		var buf bytes.Buffer
		if req.GetBody != nil {
			b, _ := req.GetBody()
			buf.ReadFrom(b)
		} else {
			buf.ReadFrom(req.Body)
		}

		b.Grow(buf.Len())
		b.WriteString(" -d \\\n'")
		json.Indent(&b, buf.Bytes(), "", " ")
		b.WriteString("'")
	}

	b.WriteRune('\n')

	var status string
	status = res.Status

	fmt.Fprintf(&b, "# => %s [%s] %s\n", start.UTC().Format(time.RFC3339), status, dur.Truncate(time.Millisecond))
	if l.ResponseBodyEnabled() && res != nil && res.Body != nil && res.Body != http.NoBody {
		var buf bytes.Buffer
		buf.ReadFrom(res.Body)

		b.Grow(buf.Len())
		b.WriteString("# ")
		json.Indent(&b, buf.Bytes(), "# ", " ")
	}

	b.WriteString("\n")
	if l.ResponseBodyEnabled() && res != nil && res.Body != nil && res.Body != http.NoBody {
		b.WriteString("\n")
	}

	b.WriteTo(l.Output)

	return nil
}

// RequestBodyEnabled returns true when the request body should be logged.
func (l *CurlLogger) RequestBodyEnabled() bool { return l.EnableRequestBody }

// ResponseBodyEnabled returns true when the response body should be logged.
func (l *CurlLogger) ResponseBodyEnabled() bool { return l.EnableResponseBody }

// LogRoundTrip prints the information about request and response.
//
func (l *JSONLogger) LogRoundTrip(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) error {
	// TODO: Research performance optimization of using sync.Pool

	bsize := 200
	var b = bytes.NewBuffer(make([]byte, 0, bsize))
	var v = make([]byte, 0, bsize)

	appendTime := func(t time.Time) {
		v = v[:0]
		v = t.AppendFormat(v, time.RFC3339)
		b.Write(v)
	}

	appendQuote := func(s string) {
		v = v[:0]
		v = strconv.AppendQuote(v, s)
		b.Write(v)
	}

	appendInt := func(i int64) {
		v = v[:0]
		v = strconv.AppendInt(v, i, 10)
		b.Write(v)
	}

	port := req.URL.Port()

	b.WriteRune('{')
	// -- Timestamp
	b.WriteString(`"@timestamp":"`)
	appendTime(start.UTC())
	b.WriteRune('"')
	// -- Event
	b.WriteString(`,"event":{`)
	b.WriteString(`"duration":`)
	appendInt(dur.Nanoseconds())
	b.WriteRune('}')
	// -- URL
	b.WriteString(`,"url":{`)
	b.WriteString(`"scheme":`)
	appendQuote(req.URL.Scheme)
	b.WriteString(`,"domain":`)
	appendQuote(req.URL.Hostname())
	if port != "" {
		b.WriteString(`,"port":`)
		b.WriteString(port)
	}
	b.WriteString(`,"path":`)
	appendQuote(req.URL.Path)
	b.WriteString(`,"query":`)
	appendQuote(req.URL.RawQuery)
	b.WriteRune('}') // Close "url"
	// -- HTTP
	b.WriteString(`,"http":`)
	// ---- Request
	b.WriteString(`{"request":{`)
	b.WriteString(`"method":`)
	appendQuote(req.Method)
	if l.RequestBodyEnabled() && req != nil && req.Body != nil && req.Body != http.NoBody {
		var buf bytes.Buffer
		if req.GetBody != nil {
			b, _ := req.GetBody()
			buf.ReadFrom(b)
		} else {
			buf.ReadFrom(req.Body)
		}

		b.Grow(buf.Len() + 8)
		b.WriteString(`,"body":`)
		appendQuote(buf.String())
	}
	b.WriteRune('}') // Close "http.request"
	// ---- Response
	b.WriteString(`,"response":{`)
	b.WriteString(`"status_code":`)
	appendInt(int64(resStatusCode(res)))
	if l.ResponseBodyEnabled() && res != nil && res.Body != nil && res.Body != http.NoBody {
		defer res.Body.Close()
		var buf bytes.Buffer
		buf.ReadFrom(res.Body)

		b.Grow(buf.Len() + 8)
		b.WriteString(`,"body":`)
		appendQuote(buf.String())
	}
	b.WriteRune('}') // Close "http.response"
	b.WriteRune('}') // Close "http"
	// -- Error
	if err != nil {
		b.WriteString(`,"error":{"message":`)
		appendQuote(err.Error())
		b.WriteRune('}') // Close "error"
	}
	b.WriteRune('}')
	b.WriteRune('\n')
	b.WriteTo(l.Output)

	return nil
}

// RequestBodyEnabled returns true when the request body should be logged.
func (l *JSONLogger) RequestBodyEnabled() bool { return l.EnableRequestBody }

// ResponseBodyEnabled returns true when the response body should be logged.
func (l *JSONLogger) ResponseBodyEnabled() bool { return l.EnableResponseBody }

// Log prints the arguments to output in default format.
//
func (l *debuggingLogger) Log(a ...interface{}) error {
	_, err := fmt.Fprint(l.Output, a...)
	return err
}

// Logf prints formats the arguments and prints them to output.
//
func (l *debuggingLogger) Logf(format string, a ...interface{}) error {
	_, err := fmt.Fprintf(l.Output, format, a...)
	return err
}

func logBodyAsText(dst io.Writer, body io.Reader, prefix string) {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		s := scanner.Text()
		if s != "" {
			fmt.Fprintf(dst, "%s %s\n", prefix, s)
		}
	}
}

func duplicateBody(body io.ReadCloser) (io.ReadCloser, io.ReadCloser, error) {
	var (
		b1 bytes.Buffer
		b2 bytes.Buffer
		tr = io.TeeReader(body, &b2)
	)
	_, err := b1.ReadFrom(tr)
	if err != nil {
		return ioutil.NopCloser(io.MultiReader(&b1, errorReader{err: err})), ioutil.NopCloser(io.MultiReader(&b2, errorReader{err: err})), err
	}
	defer func() { body.Close() }()

	return ioutil.NopCloser(&b1), ioutil.NopCloser(&b2), nil
}

func resStatusCode(res *http.Response) int {
	if res == nil {
		return -1
	}
	return res.StatusCode
}

type errorReader struct{ err error }

func (r errorReader) Read(p []byte) (int, error) { return 0, r.err }
