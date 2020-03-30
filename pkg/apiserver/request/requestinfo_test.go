/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package request

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"net/http"
	"testing"
)

func newTestRequestInfoResolver() RequestInfoResolver {
	requestInfoResolver := &RequestInfoFactory{
		APIPrefixes:          sets.NewString("api", "apis", "kapis", "kapi"),
		GrouplessAPIPrefixes: sets.NewString("api", "kapi"),
	}

	return requestInfoResolver
}

func TestRequestInfoFactory_NewRequestInfo(t *testing.T) {
	tests := []struct {
		name                      string
		url                       string
		method                    string
		expectedErr               error
		expectedVerb              string
		expectedResource          string
		expectedIsResourceRequest bool
		expectedCluster           string
		expectedWorkspace         string
		expectedNamespace         string
		expectedKubernetesRequest bool
	}{
		{
			name:                      "login",
			url:                       "/oauth/authorize?client_id=ks-console&response_type=token",
			method:                    http.MethodPost,
			expectedErr:               nil,
			expectedVerb:              "POST",
			expectedResource:          "",
			expectedIsResourceRequest: false,
			expectedCluster:           "",
			expectedKubernetesRequest: false,
		},
		{
			name:                      "list cluster roles",
			url:                       "/apis/rbac.authorization.k8s.io/v1/clusters/cluster1/clusterroles",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedResource:          "clusterroles",
			expectedIsResourceRequest: true,
			expectedCluster:           "cluster1",
			expectedKubernetesRequest: true,
		},
		{
			name:                      "list cluster nodes",
			url:                       "/api/v1/clusters/cluster1/nodes",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedResource:          "nodes",
			expectedIsResourceRequest: true,
			expectedCluster:           "cluster1",
			expectedKubernetesRequest: true,
		},
		{
			name:                      "list cluster nodes",
			url:                       "/api/v1/clusters/cluster1/nodes",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedResource:          "nodes",
			expectedIsResourceRequest: true,
			expectedCluster:           "cluster1",
			expectedKubernetesRequest: true,
		},
		{
			name:                      "list cluster nodes",
			url:                       "/api/v1/nodes",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedResource:          "nodes",
			expectedIsResourceRequest: true,
			expectedCluster:           "",
			expectedKubernetesRequest: true,
		},
		{
			name:                      "list roles",
			url:                       "/apis/rbac.authorization.k8s.io/v1/clusters/cluster1/namespaces/namespace1/roles",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedResource:          "roles",
			expectedIsResourceRequest: true,
			expectedNamespace:         "namespace1",
			expectedCluster:           "cluster1",
			expectedKubernetesRequest: true,
		},
		{
			name:                      "list roles",
			url:                       "/apis/rbac.authorization.k8s.io/v1/namespaces/namespace1/roles",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedResource:          "roles",
			expectedIsResourceRequest: true,
			expectedCluster:           "",
			expectedNamespace:         "namespace1",
			expectedKubernetesRequest: true,
		},
		{
			name:                      "list namespaces",
			url:                       "/kapis/resources.kubesphere.io/v1alpha3/workspaces/workspace1/namespaces",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedResource:          "namespaces",
			expectedIsResourceRequest: true,
			expectedWorkspace:         "workspace1",
			expectedCluster:           "",
			expectedKubernetesRequest: false,
		},
		{
			name:                      "list namespaces",
			url:                       "/kapis/resources.kubesphere.io/v1alpha3/clusters/cluster1/workspaces/workspace1/namespaces",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedResource:          "namespaces",
			expectedIsResourceRequest: true,
			expectedWorkspace:         "workspace1",
			expectedCluster:           "cluster1",
			expectedKubernetesRequest: false,
		},
		{
			name:                      "random query",
			url:                       "/foo/bar",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "GET",
			expectedResource:          "",
			expectedIsResourceRequest: false,
			expectedWorkspace:         "",
			expectedCluster:           "",
			expectedKubernetesRequest: false,
		},
		{
			name:                      "kubesphere api without clusters",
			url:                       "/kapis/foo/bar/",
			method:                    http.MethodPost,
			expectedErr:               nil,
			expectedVerb:              "POST",
			expectedResource:          "",
			expectedNamespace:         "",
			expectedWorkspace:         "",
			expectedCluster:           "",
			expectedIsResourceRequest: false,
			expectedKubernetesRequest: false,
		},
	}

	requestInfoResolver := newTestRequestInfoResolver()

	for _, test := range tests {
		req, err := http.NewRequest(test.method, test.url, nil)
		if err != nil {
			t.Fatal(err)
		}
		requestInfo, err := requestInfoResolver.NewRequestInfo(req)

		if err != nil {
			if test.expectedErr != err {
				t.Errorf("%s: expected error %v, actual %v", test.name, test.expectedErr, err)
			}
		} else {
			if test.expectedVerb != requestInfo.Verb {
				t.Errorf("%s: expected verb %v, actual %+v", test.name, test.expectedVerb, requestInfo.Verb)
			}
			if test.expectedResource != requestInfo.Resource {
				t.Errorf("%s: expected resource %v, actual %+v", test.name, test.expectedResource, requestInfo.Resource)
			}
			if test.expectedIsResourceRequest != requestInfo.IsResourceRequest {
				t.Errorf("%s: expected is resource request %v, actual %+v", test.name, test.expectedIsResourceRequest, requestInfo.IsResourceRequest)
			}
			if test.expectedCluster != requestInfo.Cluster {
				t.Errorf("%s: expected cluster %v, actual %+v", test.name, test.expectedCluster, requestInfo.Cluster)
			}
			if test.expectedWorkspace != requestInfo.Workspace {
				t.Errorf("%s: expected workspace %v, actual %+v", test.name, test.expectedWorkspace, requestInfo.Workspace)
			}
			if test.expectedNamespace != requestInfo.Namespace {
				t.Errorf("%s: expected namespace %v, actual %+v", test.name, test.expectedNamespace, requestInfo.Namespace)
			}

			if test.expectedKubernetesRequest != requestInfo.IsKubernetesRequest {
				t.Errorf("%s: expected kubernetes request %v, actual %+v", test.name, test.expectedKubernetesRequest, requestInfo.IsKubernetesRequest)
			}
		}
	}
}
