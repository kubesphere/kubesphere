// Code generated from specification version 7.0.0 (5e798c1): DO NOT EDIT

package esapi

import (
	"context"
	"strings"
	"time"
)

func newNodesReloadSecureSettingsFunc(t Transport) NodesReloadSecureSettings {
	return func(o ...func(*NodesReloadSecureSettingsRequest)) (*Response, error) {
		var r = NodesReloadSecureSettingsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// NodesReloadSecureSettings reloads secure settings.
//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/master/secure-settings.html#reloadable-secure-settings.
//
type NodesReloadSecureSettings func(o ...func(*NodesReloadSecureSettingsRequest)) (*Response, error)

// NodesReloadSecureSettingsRequest configures the Nodes   Reload Secure Settings API request.
//
type NodesReloadSecureSettingsRequest struct {
	NodeID  []string
	Timeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r NodesReloadSecureSettingsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(1 + len("_nodes") + 1 + len(strings.Join(r.NodeID, ",")) + 1 + len("reload_secure_settings"))
	path.WriteString("/")
	path.WriteString("_nodes")
	if len(r.NodeID) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.NodeID, ","))
	}
	path.WriteString("/")
	path.WriteString("reload_secure_settings")

	params = make(map[string]string)

	if r.Timeout != 0 {
		params["timeout"] = time.Duration(r.Timeout * time.Millisecond).String()
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
func (f NodesReloadSecureSettings) WithContext(v context.Context) func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.ctx = v
	}
}

// WithNodeID - a list of node ids to span the reload/reinit call. should stay empty because reloading usually involves all cluster nodes..
//
func (f NodesReloadSecureSettings) WithNodeID(v ...string) func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.NodeID = v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f NodesReloadSecureSettings) WithTimeout(v time.Duration) func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.Timeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f NodesReloadSecureSettings) WithPretty() func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f NodesReloadSecureSettings) WithHuman() func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f NodesReloadSecureSettings) WithErrorTrace() func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f NodesReloadSecureSettings) WithFilterPath(v ...string) func(*NodesReloadSecureSettingsRequest) {
	return func(r *NodesReloadSecureSettingsRequest) {
		r.FilterPath = v
	}
}
