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

package elasticsearch

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/json-iterator/go"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch/versions/v5"
	"kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch/versions/v6"
	"kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch/versions/v7"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInitClient(t *testing.T) {
	var tests = []struct {
		fakeResp string
		expected string
	}{
		{
			fakeResp: "es6_detect_version_major_200.json",
			expected: ElasticV6,
		},
		{
			fakeResp: "es7_detect_version_major_200.json",
			expected: ElasticV7,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			es := mockElasticsearchService("/", test.fakeResp, http.StatusOK)
			defer es.Close()

			client := &Elasticsearch{host: es.URL}
			err := client.loadClient()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(client.version, test.expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", test.expected, diff)
			}
		})
	}
}

func TestGetCurrentStats(t *testing.T) {
	var tests = []struct {
		fakeVersion string
		fakeResp    string
		fakeCode    int
		expected    logging.Statistics
		expectedErr string
	}{
		{
			fakeVersion: ElasticV6,
			fakeResp:    "es6_get_current_stats_200.json",
			fakeCode:    http.StatusOK,
			expected: logging.Statistics{
				Containers: 93,
				Logs:       241222,
			},
		},
		{
			fakeVersion: ElasticV6,
			fakeResp:    "es6_get_current_stats_404.json",
			fakeCode:    http.StatusNotFound,
			expectedErr: "type: index_not_found_exception, reason: no such index",
		},
		{
			fakeVersion: ElasticV7,
			fakeResp:    "es7_get_current_stats_200.json",
			fakeCode:    http.StatusOK,
			expected: logging.Statistics{
				Containers: 48,
				Logs:       9726,
			},
		},
		{
			fakeVersion: ElasticV7,
			fakeResp:    "es7_get_current_stats_404.json",
			fakeCode:    http.StatusNotFound,
			expectedErr: "type: index_not_found_exception, reason: no such index [ks-logstash-log-2020.05.2]",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			srv := mockElasticsearchService("/ks-logstash-log*/_search", test.fakeResp, test.fakeCode)
			defer srv.Close()

			es := newElasticsearchClient(srv, test.fakeVersion)

			result, err := es.GetCurrentStats(logging.SearchFilter{})
			if test.expectedErr != "" {
				if diff := cmp.Diff(fmt.Sprint(err), test.expectedErr); diff != "" {
					t.Fatalf("%T differ (-got, +want): %s", test.expectedErr, diff)
				}
			}
			if diff := cmp.Diff(result, test.expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", test.expected, diff)
			}
		})
	}
}

func TestCountLogsByInterval(t *testing.T) {
	var tests = []struct {
		fakeVersion string
		fakeResp    string
		fakeCode    int
		expected    logging.Histogram
		expectedErr string
	}{
		{
			fakeVersion: ElasticV7,
			fakeResp:    "es7_count_logs_by_interval_200.json",
			fakeCode:    http.StatusOK,
			expected: logging.Histogram{
				Total: 10000,
				Buckets: []logging.Bucket{
					{
						Time:  1589644800000,
						Count: 410,
					},
					{
						Time:  1589646600000,
						Count: 7465,
					},
					{
						Time:  1589648400000,
						Count: 12790,
					},
				},
			},
		},
		{
			fakeVersion: ElasticV7,
			fakeResp:    "es7_count_logs_by_interval_400.json",
			fakeCode:    http.StatusBadRequest,
			expectedErr: "type: search_phase_execution_exception, reason: all shards failed",
		},
		{
			fakeVersion: ElasticV7,
			fakeResp:    "es7_count_logs_by_interval_404.json",
			fakeCode:    http.StatusNotFound,
			expectedErr: "type: index_not_found_exception, reason: no such index [ks-logstash-log-20]",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			srv := mockElasticsearchService("/ks-logstash-log*/_search", test.fakeResp, test.fakeCode)
			defer srv.Close()

			es := newElasticsearchClient(srv, test.fakeVersion)

			result, err := es.CountLogsByInterval(logging.SearchFilter{}, "15m")
			if test.expectedErr != "" {
				if diff := cmp.Diff(fmt.Sprint(err), test.expectedErr); diff != "" {
					t.Fatalf("%T differ (-got, +want): %s", test.expectedErr, diff)
				}
			}
			if diff := cmp.Diff(result, test.expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", test.expected, diff)
			}
		})
	}
}

func TestSearchLogs(t *testing.T) {
	var tests = []struct {
		fakeVersion string
		fakeResp    string
		fakeCode    int
		expected    string
		expectedErr string
	}{
		{
			fakeVersion: ElasticV7,
			fakeResp:    "es7_search_logs_200.json",
			fakeCode:    http.StatusOK,
			expected:    "es7_search_logs_200_result.json",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var expected logging.Logs
			err := JsonFromFile(test.expected, &expected)
			if err != nil {
				t.Fatal(err)
			}

			srv := mockElasticsearchService("/ks-logstash-log*/_search", test.fakeResp, test.fakeCode)
			defer srv.Close()

			es := newElasticsearchClient(srv, test.fakeVersion)

			result, err := es.SearchLogs(logging.SearchFilter{}, 0, 10, "asc")
			if test.expectedErr != "" {
				if diff := cmp.Diff(fmt.Sprint(err), test.expectedErr); diff != "" {
					t.Fatalf("%T differ (-got, +want): %s", test.expectedErr, diff)
				}
			}
			if diff := cmp.Diff(result, expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}

func mockElasticsearchService(pattern, fakeResp string, fakeCode int) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, func(res http.ResponseWriter, req *http.Request) {
		b, _ := ioutil.ReadFile(fmt.Sprintf("./testdata/%s", fakeResp))
		res.WriteHeader(fakeCode)
		res.Write(b)
	})
	return httptest.NewServer(mux)
}

func newElasticsearchClient(srv *httptest.Server, version string) *Elasticsearch {
	es := &Elasticsearch{index: "ks-logstash-log"}
	switch version {
	case ElasticV5:
		es.c, _ = v5.New(srv.URL, "ks-logstash-log")
	case ElasticV6:
		es.c, _ = v6.New(srv.URL, "ks-logstash-log")
	case ElasticV7:
		es.c, _ = v7.New(srv.URL, "ks-logstash-log")
	}
	return es
}

func JsonFromFile(expectedFile string, expectedJsonPtr interface{}) error {
	json, err := ioutil.ReadFile(fmt.Sprintf("./testdata/%s", expectedFile))
	if err != nil {
		return err
	}
	err = jsoniter.Unmarshal(json, expectedJsonPtr)
	if err != nil {
		return err
	}

	return nil
}
