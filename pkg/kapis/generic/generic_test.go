/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package generic

import (
	"github.com/emicklei/go-restful"
	"github.com/google/go-cmp/cmp"
	"net/http/httptest"
	"net/url"
	"testing"
)

var group = "test.kubesphere.io"
var version = "v1"
var scheme = "http"

func TestNewGenericProxy(t *testing.T) {
	var testCases = []struct {
		description string
		endpoint    string
		query       string
		expected    *url.URL
	}{
		{
			description: "Endpoint with path",
			endpoint:    "http://awesome.kubesphere-system.svc:8080/api",
			query:       "/kapis/test.kubesphere.io/v1/foo/bar?id=1&time=whatever",
			expected: &url.URL{
				Scheme:   scheme,
				Host:     "awesome.kubesphere-system.svc:8080",
				Path:     "/api/v1/foo/bar",
				RawQuery: "id=1&time=whatever",
			},
		},
		{
			description: "Endpoint without path",
			endpoint:    "http://awesome.kubesphere-system.svc:8080",
			query:       "/kapis/test.kubesphere.io/v1/foo/bar?id=1&time=whatever",
			expected: &url.URL{
				Scheme:   scheme,
				Host:     "awesome.kubesphere-system.svc:8080",
				Path:     "/v1/foo/bar",
				RawQuery: "id=1&time=whatever",
			},
		},
	}

	for _, testCase := range testCases {
		proxy, err := NewGenericProxy(testCase.endpoint, group, version)
		if err != nil {
			t.Error(err)
		}

		t.Run(testCase.description, func(t *testing.T) {
			request := httptest.NewRequest("GET", testCase.query, nil)
			u := proxy.makeURL(restful.NewRequest(request))
			if diff := cmp.Diff(u, testCase.expected); len(diff) != 0 {
				t.Errorf("%T differ (-got, +want): %s", testCase.expected, diff)
			}
		})
	}
}
