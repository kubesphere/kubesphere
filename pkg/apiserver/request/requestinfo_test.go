/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package request

import (
	"net/http"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
)

func newTestRequestInfoResolver() RequestInfoResolver {
	requestInfoResolver := &RequestInfoFactory{
		APIPrefixes:          sets.New("api", "apis", "kapis", "kapi"),
		GrouplessAPIPrefixes: sets.New("api", "kapi"),
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
			name:                      "list clusterRoles of cluster gondor",
			url:                       "/clusters/gondor/apis/rbac.authorization.k8s.io/v1/clusterroles",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedResource:          "clusterroles",
			expectedIsResourceRequest: true,
			expectedCluster:           "gondor",
			expectedKubernetesRequest: true,
		},
		{
			name:                      "list nodes",
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
			name:                      "list nodes of cluster gondor",
			url:                       "/clusters/gondor/api/v1/nodes",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedResource:          "nodes",
			expectedIsResourceRequest: true,
			expectedCluster:           "gondor",
			expectedKubernetesRequest: true,
		},
		{
			name:                      "list roles of cluster gondor",
			url:                       "/clusters/gondor/apis/rbac.authorization.k8s.io/v1/namespaces/namespace1/roles",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedResource:          "roles",
			expectedIsResourceRequest: true,
			expectedNamespace:         "namespace1",
			expectedCluster:           "gondor",
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
			name:                      "list namespaces of cluster gondor",
			url:                       "/clusters/gondor/kapis/resources.kubesphere.io/v1alpha3/workspaces/workspace1/namespaces",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedResource:          "namespaces",
			expectedIsResourceRequest: true,
			expectedWorkspace:         "workspace1",
			expectedCluster:           "gondor",
			expectedKubernetesRequest: false,
		},
		{
			name:                      "list clusters",
			url:                       "/apis/cluster.kubesphere.io/v1alpha1/clusters",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedResource:          "clusters",
			expectedIsResourceRequest: true,
			expectedWorkspace:         "",
			expectedCluster:           "",
			expectedKubernetesRequest: true,
		},
		{
			name:                      "get cluster gondor",
			url:                       "/apis/cluster.kubesphere.io/v1alpha1/clusters/gondor",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "get",
			expectedResource:          "clusters",
			expectedIsResourceRequest: true,
			expectedWorkspace:         "",
			expectedCluster:           "",
			expectedKubernetesRequest: true,
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
			name:                      "",
			url:                       "/kapis/tenant.kubesphere.io/v1alpha2/workspaces",
			method:                    http.MethodGet,
			expectedErr:               nil,
			expectedVerb:              "list",
			expectedNamespace:         "",
			expectedCluster:           "",
			expectedWorkspace:         "",
			expectedKubernetesRequest: false,
			expectedIsResourceRequest: true,
			expectedResource:          "workspaces",
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
		t.Run(test.url, func(t *testing.T) {
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
		})

	}
}
