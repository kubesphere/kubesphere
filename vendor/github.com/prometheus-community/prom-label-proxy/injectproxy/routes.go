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
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

type routes struct {
	upstream  *url.URL
	handler   http.Handler
	label     string
	mux       *http.ServeMux
	modifiers map[string]func(*http.Response) error
}

func NewRoutes(upstream *url.URL, label string) *routes {
	proxy := httputil.NewSingleHostReverseProxy(upstream)

	r := &routes{
		upstream: upstream,
		handler:  proxy,
		label:    label,
	}
	mux := http.NewServeMux()
	mux.Handle("/federate", enforceMethods(r.federate, "GET"))
	mux.Handle("/api/v1/query", enforceMethods(r.query, "GET", "POST"))
	mux.Handle("/api/v1/query_range", enforceMethods(r.query, "GET", "POST"))
	mux.Handle("/api/v1/alerts", enforceMethods(r.noop, "GET"))
	mux.Handle("/api/v1/rules", enforceMethods(r.noop, "GET"))
	mux.Handle("/api/v2/silences", enforceMethods(r.silences, "GET", "POST"))
	mux.Handle("/api/v2/silences/", enforceMethods(r.silences, "GET", "POST"))
	mux.Handle("/api/v2/silence/", enforceMethods(r.deleteSilence, "DELETE"))
	r.mux = mux
	r.modifiers = map[string]func(*http.Response) error{
		"/api/v1/rules":  modifyAPIResponse(r.filterRules),
		"/api/v1/alerts": modifyAPIResponse(r.filterAlerts),
	}
	proxy.ModifyResponse = r.ModifyResponse
	return r
}

func (r *routes) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	lvalue := req.URL.Query().Get(r.label)
	if lvalue == "" {
		http.Error(w, fmt.Sprintf("Bad request. The %q query parameter must be provided.", r.label), http.StatusBadRequest)
		return
	}
	req = req.WithContext(withLabelValue(req.Context(), lvalue))
	// Remove the proxy label from the query parameters.
	q := req.URL.Query()
	q.Del(r.label)
	req.URL.RawQuery = q.Encode()

	r.mux.ServeHTTP(w, req)
}

func (r *routes) ModifyResponse(resp *http.Response) error {
	m, found := r.modifiers[resp.Request.URL.Path]
	if !found {
		// Return the server's response unmodified.
		return nil
	}
	return m(resp)
}

func enforceMethods(h http.HandlerFunc, methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, m := range methods {
			if m == req.Method {
				h(w, req)
				return
			}
		}
		http.NotFound(w, req)
	})
}

type ctxKey int

const keyLabel ctxKey = iota

func mustLabelValue(ctx context.Context) string {
	label, ok := ctx.Value(keyLabel).(string)
	if !ok {
		panic(fmt.Sprintf("can't find the %q value in the context", keyLabel))
	}
	if label == "" {
		panic(fmt.Sprintf("empty %q value in the context", keyLabel))
	}
	return label
}

func withLabelValue(ctx context.Context, label string) context.Context {
	return context.WithValue(ctx, keyLabel, label)
}

func (r *routes) noop(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

func (r *routes) query(w http.ResponseWriter, req *http.Request) {
	expr, err := parser.ParseExpr(req.FormValue("query"))
	if err != nil {
		return
	}

	e := NewEnforcer([]*labels.Matcher{
		{
			Name:  r.label,
			Type:  labels.MatchEqual,
			Value: mustLabelValue(req.Context()),
		},
	}...)
	if err := e.EnforceNode(expr); err != nil {
		return
	}

	q := req.URL.Query()
	q.Set("query", expr.String())
	req.URL.RawQuery = q.Encode()

	r.handler.ServeHTTP(w, req)
}

func (r *routes) federate(w http.ResponseWriter, req *http.Request) {
	matcher := &labels.Matcher{
		Name:  r.label,
		Type:  labels.MatchEqual,
		Value: mustLabelValue(req.Context()),
	}

	q := req.URL.Query()
	q.Set("match[]", "{"+matcher.String()+"}")
	req.URL.RawQuery = q.Encode()

	r.handler.ServeHTTP(w, req)
}
