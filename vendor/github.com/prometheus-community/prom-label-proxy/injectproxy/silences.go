// Copyright 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package injectproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"

	runtimeclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/prometheus/alertmanager/pkg/labels"
)

func (r *routes) silences(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		r.listSilences(w, req)
	case "POST":
		r.postSilence(w, req)
	default:
		http.NotFound(w, req)
	}
}

func (r *routes) listSilences(w http.ResponseWriter, req *http.Request) {
	var (
		q               = req.URL.Query()
		proxyLabelMatch = labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  r.label,
			Value: mustLabelValue(req.Context()),
		}
		modified = []string{proxyLabelMatch.String()}
	)
	for _, filter := range q["filter"] {
		m, err := labels.ParseMatcher(filter)
		if err != nil {
			http.Error(w, fmt.Sprintf("bad request: can't parse filter %q: %v", filter, err), http.StatusBadRequest)
			return
		}
		if m.Name == r.label {
			continue
		}
		modified = append(modified, filter)
	}

	q["filter"] = modified
	q.Del(r.label)
	req.URL.RawQuery = q.Encode()

	r.handler.ServeHTTP(w, req)
}

func (r *routes) postSilence(w http.ResponseWriter, req *http.Request) {
	var (
		sil    models.PostableSilence
		lvalue = mustLabelValue(req.Context())
	)
	if err := json.NewDecoder(req.Body).Decode(&sil); err != nil {
		http.Error(w, fmt.Sprintf("bad request: can't decode: %v", err), http.StatusBadRequest)
		return
	}

	if sil.ID != "" {
		// This is an update for an existing silence.
		existing, err := r.getSilenceByID(req.Context(), sil.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("proxy error: can't get silence: %v", err), http.StatusBadGateway)
			return
		}

		if !hasMatcherForLabel(existing.Matchers, r.label, lvalue) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	var falsy bool
	modified := models.Matchers{
		&models.Matcher{Name: &(r.label), Value: &lvalue, IsRegex: &falsy},
	}
	for _, m := range sil.Matchers {
		if m.Name != nil && *m.Name == r.label {
			continue
		}
		modified = append(modified, m)
	}
	// At least one matcher in addition to the enforced label is required,
	// otherwise all alerts would be silenced
	if len(modified) < 2 {
		http.Error(w, "need at least one matcher, got none", http.StatusBadRequest)
		return
	}
	sil.Matchers = modified

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&sil); err != nil {
		http.Error(w, fmt.Sprintf("can't encode: %v", err), http.StatusInternalServerError)
		return
	}

	req = req.Clone(req.Context())
	req.Body = ioutil.NopCloser(&buf)
	req.URL.RawQuery = ""
	req.Header["Content-Length"] = []string{strconv.Itoa(buf.Len())}
	req.ContentLength = int64(buf.Len())

	r.handler.ServeHTTP(w, req)
}

func (r *routes) deleteSilence(w http.ResponseWriter, req *http.Request) {
	silID := strings.TrimPrefix(req.URL.Path, "/api/v2/silence/")
	if silID == "" || silID == req.URL.Path {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Get the silence by ID and verify that it has the expected label.
	sil, err := r.getSilenceByID(req.Context(), silID)
	if err != nil {
		http.Error(w, fmt.Sprintf("proxy error: %v", err), http.StatusBadGateway)
		return
	}

	if !hasMatcherForLabel(sil.Matchers, r.label, mustLabelValue(req.Context())) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	req.URL.RawQuery = ""
	r.handler.ServeHTTP(w, req)
}

func (r *routes) getSilenceByID(ctx context.Context, id string) (*models.GettableSilence, error) {
	amc := client.New(
		runtimeclient.New(r.upstream.Host, path.Join(r.upstream.Path, "/api/v2"), []string{r.upstream.Scheme}),
		strfmt.Default,
	)
	params := silence.NewGetSilenceParams().WithContext(ctx)
	params.SetSilenceID(strfmt.UUID(id))
	sil, err := amc.Silence.GetSilence(params)
	if err != nil {
		return nil, err
	}
	return sil.Payload, nil
}

func hasMatcherForLabel(matchers models.Matchers, name, value string) bool {
	for _, m := range matchers {
		if *m.Name == name && !*m.IsRegex && *m.Value == value {
			return true
		}
	}
	return false
}
