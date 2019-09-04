// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func newXPackMonitoringBulkFunc(t Transport) XPackMonitoringBulk {
	return func(body io.Reader, o ...func(*XPackMonitoringBulkRequest)) (*Response, error) {
		var r = XPackMonitoringBulkRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// XPackMonitoringBulk - https://www.elastic.co/guide/en/elasticsearch/reference/6.7/es-monitoring.html
//
type XPackMonitoringBulk func(body io.Reader, o ...func(*XPackMonitoringBulkRequest)) (*Response, error)

// XPackMonitoringBulkRequest configures the X Pack Monitoring Bulk API request.
//
type XPackMonitoringBulkRequest struct {
	DocumentType string

	Body io.Reader

	Interval         string
	SystemAPIVersion string
	SystemID         string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r XPackMonitoringBulkRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_xpack") + 1 + len("monitoring") + 1 + len(r.DocumentType) + 1 + len("_bulk"))
	path.WriteString("/")
	path.WriteString("_xpack")
	path.WriteString("/")
	path.WriteString("monitoring")
	if r.DocumentType != "" {
		path.WriteString("/")
		path.WriteString(r.DocumentType)
	}
	path.WriteString("/")
	path.WriteString("_bulk")

	params = make(map[string]string)

	if r.Interval != "" {
		params["interval"] = r.Interval
	}

	if r.SystemAPIVersion != "" {
		params["system_api_version"] = r.SystemAPIVersion
	}

	if r.SystemID != "" {
		params["system_id"] = r.SystemID
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
func (f XPackMonitoringBulk) WithContext(v context.Context) func(*XPackMonitoringBulkRequest) {
	return func(r *XPackMonitoringBulkRequest) {
		r.ctx = v
	}
}

// WithDocumentType - default document type for items which don't provide one.
//
func (f XPackMonitoringBulk) WithDocumentType(v string) func(*XPackMonitoringBulkRequest) {
	return func(r *XPackMonitoringBulkRequest) {
		r.DocumentType = v
	}
}

// WithInterval - collection interval (e.g., '10s' or '10000ms') of the payload.
//
func (f XPackMonitoringBulk) WithInterval(v string) func(*XPackMonitoringBulkRequest) {
	return func(r *XPackMonitoringBulkRequest) {
		r.Interval = v
	}
}

// WithSystemAPIVersion - api version of the monitored system.
//
func (f XPackMonitoringBulk) WithSystemAPIVersion(v string) func(*XPackMonitoringBulkRequest) {
	return func(r *XPackMonitoringBulkRequest) {
		r.SystemAPIVersion = v
	}
}

// WithSystemID - identifier of the monitored system.
//
func (f XPackMonitoringBulk) WithSystemID(v string) func(*XPackMonitoringBulkRequest) {
	return func(r *XPackMonitoringBulkRequest) {
		r.SystemID = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f XPackMonitoringBulk) WithPretty() func(*XPackMonitoringBulkRequest) {
	return func(r *XPackMonitoringBulkRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f XPackMonitoringBulk) WithHuman() func(*XPackMonitoringBulkRequest) {
	return func(r *XPackMonitoringBulkRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f XPackMonitoringBulk) WithErrorTrace() func(*XPackMonitoringBulkRequest) {
	return func(r *XPackMonitoringBulkRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f XPackMonitoringBulk) WithFilterPath(v ...string) func(*XPackMonitoringBulkRequest) {
	return func(r *XPackMonitoringBulkRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f XPackMonitoringBulk) WithHeader(h map[string]string) func(*XPackMonitoringBulkRequest) {
	return func(r *XPackMonitoringBulkRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
