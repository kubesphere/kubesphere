// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package report provides functions to report OPA's version information to an external service and process the response.
package report

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/open-policy-agent/opa/v1/keys"
	"github.com/open-policy-agent/opa/v1/logging"
	"github.com/open-policy-agent/opa/v1/version"

	"github.com/open-policy-agent/opa/v1/plugins/rest"
	"github.com/open-policy-agent/opa/v1/util"
)

// ExternalServiceURL is the base HTTP URL for a telemetry service.
// If not otherwise specified it will use the hard coded default.
//
// Override at build time via:
//
//	-ldflags "-X github.com/open-policy-agent/opa/internal/report.ExternalServiceURL=<url>"
//
// This will be overridden if the OPA_TELEMETRY_SERVICE_URL environment variable
// is provided.
var ExternalServiceURL = "https://telemetry.openpolicyagent.org"

// Reporter reports information such as the version, heap usage about the running OPA instance to an external service
type Reporter struct {
	body   map[string]any
	client rest.Client

	gatherers    map[string]Gatherer
	gatherersMtx sync.Mutex
}

// Gatherer represents a mechanism to inject additional data in the telemetry report
type Gatherer func(ctx context.Context) (any, error)

// DataResponse represents the data returned by the external service
type DataResponse struct {
	Latest ReleaseDetails `json:"latest,omitempty"`
}

// ReleaseDetails holds information about the latest OPA release
type ReleaseDetails struct {
	Download      string `json:"download,omitempty"`       // link to download the OPA release
	ReleaseNotes  string `json:"release_notes,omitempty"`  // link to the OPA release notes
	LatestRelease string `json:"latest_release,omitempty"` // latest OPA released version
	OPAUpToDate   bool   `json:"opa_up_to_date,omitempty"` // is running OPA version greater than or equal to the latest released
}

// Options supplies parameters to the reporter.
type Options struct {
	Logger logging.Logger
}

// New returns an instance of the Reporter
func New(id string, opts Options) (*Reporter, error) {
	r := Reporter{
		gatherers: map[string]Gatherer{},
	}
	r.body = map[string]any{
		"id":      id,
		"version": version.Version,
	}

	url := os.Getenv("OPA_TELEMETRY_SERVICE_URL")
	if url == "" {
		url = ExternalServiceURL
	}

	restConfig := []byte(fmt.Sprintf(`{
		"url": %q,
	}`, url))

	client, err := rest.New(restConfig, map[string]*keys.Config{}, rest.Logger(opts.Logger))
	if err != nil {
		return nil, err
	}
	r.client = client

	// heap_usage_bytes is always present, so register it unconditionally
	r.RegisterGatherer("heap_usage_bytes", readRuntimeMemStats)

	return &r, nil
}

// SendReport sends the telemetry report which includes information such as the OPA version, current memory usage to
// the external service
func (r *Reporter) SendReport(ctx context.Context) (*DataResponse, error) {
	rCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.gatherersMtx.Lock()
	defer r.gatherersMtx.Unlock()
	for key, g := range r.gatherers {
		var err error
		r.body[key], err = g(rCtx)
		if err != nil {
			return nil, fmt.Errorf("gather telemetry error for key %s: %w", key, err)
		}
	}

	resp, err := r.client.WithJSON(r.body).Do(rCtx, "POST", "/v1/version")
	if err != nil {
		return nil, err
	}

	defer util.Close(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		if resp.Body != nil {
			var result DataResponse
			err := json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				return nil, err
			}
			return &result, nil
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("server replied with HTTP %v", resp.StatusCode)
	}
}

func (r *Reporter) RegisterGatherer(key string, f Gatherer) {
	r.gatherersMtx.Lock()
	r.gatherers[key] = f
	r.gatherersMtx.Unlock()
}

// IsSet returns true if dr is populated.
func (dr *DataResponse) IsSet() bool {
	return dr != nil && dr.Latest.LatestRelease != "" && dr.Latest.Download != "" && dr.Latest.ReleaseNotes != ""
}

// Slice returns the dr as a slice of key-value string pairs. If dr is nil, this function returns an empty slice.
func (dr *DataResponse) Slice() [][2]string {

	if !dr.IsSet() {
		return nil
	}

	return [][2]string{
		{"Latest Upstream Version", strings.TrimPrefix(dr.Latest.LatestRelease, "v")},
		{"Download", dr.Latest.Download},
		{"Release Notes", dr.Latest.ReleaseNotes},
	}
}

// Pretty returns OPA release information in a human-readable format.
func (dr *DataResponse) Pretty() string {
	if !dr.IsSet() {
		return ""
	}

	pairs := dr.Slice()
	lines := make([]string, 0, len(pairs))

	for _, pair := range pairs {
		lines = append(lines, fmt.Sprintf("%v: %v", pair[0], pair[1]))
	}

	return strings.Join(lines, "\n")
}

func readRuntimeMemStats(_ context.Context) (any, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return strconv.FormatUint(m.Alloc, 10), nil
}
