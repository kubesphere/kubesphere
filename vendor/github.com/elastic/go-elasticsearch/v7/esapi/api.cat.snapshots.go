// Code generated from specification version 7.3.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newCatSnapshotsFunc(t Transport) CatSnapshots {
	return func(o ...func(*CatSnapshotsRequest)) (*Response, error) {
		var r = CatSnapshotsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// CatSnapshots returns all snapshots in a specific repository.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/cat-snapshots.html.
//
type CatSnapshots func(o ...func(*CatSnapshotsRequest)) (*Response, error)

// CatSnapshotsRequest configures the Cat Snapshots API request.
//
type CatSnapshotsRequest struct {
	Repository []string

	Format            string
	H                 []string
	Help              *bool
	IgnoreUnavailable *bool
	MasterTimeout     time.Duration
	S                 []string
	V                 *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CatSnapshotsRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_cat") + 1 + len("snapshots") + 1 + len(strings.Join(r.Repository, ",")))
	path.WriteString("/")
	path.WriteString("_cat")
	path.WriteString("/")
	path.WriteString("snapshots")
	if len(r.Repository) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Repository, ","))
	}

	params = make(map[string]string)

	if r.Format != "" {
		params["format"] = r.Format
	}

	if len(r.H) > 0 {
		params["h"] = strings.Join(r.H, ",")
	}

	if r.Help != nil {
		params["help"] = strconv.FormatBool(*r.Help)
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.MasterTimeout != 0 {
		params["master_timeout"] = formatDuration(r.MasterTimeout)
	}

	if len(r.S) > 0 {
		params["s"] = strings.Join(r.S, ",")
	}

	if r.V != nil {
		params["v"] = strconv.FormatBool(*r.V)
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
func (f CatSnapshots) WithContext(v context.Context) func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		r.ctx = v
	}
}

// WithRepository - name of repository from which to fetch the snapshot information.
//
func (f CatSnapshots) WithRepository(v ...string) func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		r.Repository = v
	}
}

// WithFormat - a short version of the accept header, e.g. json, yaml.
//
func (f CatSnapshots) WithFormat(v string) func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		r.Format = v
	}
}

// WithH - comma-separated list of column names to display.
//
func (f CatSnapshots) WithH(v ...string) func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		r.H = v
	}
}

// WithHelp - return help information.
//
func (f CatSnapshots) WithHelp(v bool) func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		r.Help = &v
	}
}

// WithIgnoreUnavailable - set to true to ignore unavailable snapshots.
//
func (f CatSnapshots) WithIgnoreUnavailable(v bool) func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithMasterTimeout - explicit operation timeout for connection to master node.
//
func (f CatSnapshots) WithMasterTimeout(v time.Duration) func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		r.MasterTimeout = v
	}
}

// WithS - comma-separated list of column names or column aliases to sort by.
//
func (f CatSnapshots) WithS(v ...string) func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		r.S = v
	}
}

// WithV - verbose mode. display column headers.
//
func (f CatSnapshots) WithV(v bool) func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		r.V = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CatSnapshots) WithPretty() func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CatSnapshots) WithHuman() func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CatSnapshots) WithErrorTrace() func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CatSnapshots) WithFilterPath(v ...string) func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f CatSnapshots) WithHeader(h map[string]string) func(*CatSnapshotsRequest) {
	return func(r *CatSnapshotsRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
