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

func newClusterPutSettingsFunc(t Transport) ClusterPutSettings {
	return func(body io.Reader, o ...func(*ClusterPutSettingsRequest)) (*Response, error) {
		var r = ClusterPutSettingsRequest{Body: body}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// ClusterPutSettings updates the cluster settings.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/cluster-update-settings.html.
//
type ClusterPutSettings func(body io.Reader, o ...func(*ClusterPutSettingsRequest)) (*Response, error)

// ClusterPutSettingsRequest configures the Cluster Put Settings API request.
//
type ClusterPutSettingsRequest struct {
	Body io.Reader

	FlatSettings  *bool
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
func (r ClusterPutSettingsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "PUT"

	path.Grow(len("/_cluster/settings"))
	path.WriteString("/_cluster/settings")

	params = make(map[string]string)

	if r.FlatSettings != nil {
		params["flat_settings"] = strconv.FormatBool(*r.FlatSettings)
	}

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
func (f ClusterPutSettings) WithContext(v context.Context) func(*ClusterPutSettingsRequest) {
	return func(r *ClusterPutSettingsRequest) {
		r.ctx = v
	}
}

// WithFlatSettings - return settings in flat format (default: false).
//
func (f ClusterPutSettings) WithFlatSettings(v bool) func(*ClusterPutSettingsRequest) {
	return func(r *ClusterPutSettingsRequest) {
		r.FlatSettings = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f ClusterPutSettings) WithMasterTimeout(v time.Duration) func(*ClusterPutSettingsRequest) {
	return func(r *ClusterPutSettingsRequest) {
		r.MasterTimeout = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f ClusterPutSettings) WithTimeout(v time.Duration) func(*ClusterPutSettingsRequest) {
	return func(r *ClusterPutSettingsRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f ClusterPutSettings) WithPretty() func(*ClusterPutSettingsRequest) {
	return func(r *ClusterPutSettingsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f ClusterPutSettings) WithHuman() func(*ClusterPutSettingsRequest) {
	return func(r *ClusterPutSettingsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f ClusterPutSettings) WithErrorTrace() func(*ClusterPutSettingsRequest) {
	return func(r *ClusterPutSettingsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f ClusterPutSettings) WithFilterPath(v ...string) func(*ClusterPutSettingsRequest) {
	return func(r *ClusterPutSettingsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f ClusterPutSettings) WithHeader(h map[string]string) func(*ClusterPutSettingsRequest) {
	return func(r *ClusterPutSettingsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
