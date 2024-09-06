/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package generic

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/google/go-cmp/cmp"
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
