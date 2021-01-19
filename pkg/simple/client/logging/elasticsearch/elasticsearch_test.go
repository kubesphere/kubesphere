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
	"kubesphere.io/kubesphere/pkg/simple/client/es"
	"kubesphere.io/kubesphere/pkg/simple/client/es/query"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetCurrentStats(t *testing.T) {
	var tests = []struct {
		fakeVersion string
		fakeResp    string
		fakeCode    int
		expected    logging.Statistics
		expectedErr string
	}{
		{
			fakeVersion: es.ElasticV6,
			fakeResp:    "es6_get_current_stats_200.json",
			fakeCode:    http.StatusOK,
			expected: logging.Statistics{
				Containers: 93,
				Logs:       241222,
			},
		},
		{
			fakeVersion: es.ElasticV6,
			fakeResp:    "es6_get_current_stats_404.json",
			fakeCode:    http.StatusNotFound,
			expectedErr: "type: index_not_found_exception, reason: no such index",
		},
		{
			fakeVersion: es.ElasticV7,
			fakeResp:    "es7_get_current_stats_200.json",
			fakeCode:    http.StatusOK,
			expected: logging.Statistics{
				Containers: 48,
				Logs:       9726,
			},
		},
		{
			fakeVersion: es.ElasticV7,
			fakeResp:    "es7_get_current_stats_404.json",
			fakeCode:    http.StatusNotFound,
			expectedErr: "type: index_not_found_exception, reason: no such index [ks-logstash-log-2020.05.2]",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			srv := mockElasticsearchService("/ks-logstash-log*/_search", test.fakeResp, test.fakeCode)
			defer srv.Close()

			client, err := NewClient(&logging.Options{
				Host:        srv.URL,
				IndexPrefix: "ks-logstash-log",
				Version:     test.fakeVersion,
			})
			if err != nil {
				t.Fatalf("create client error, %s", err)
			}

			result, err := client.GetCurrentStats(logging.SearchFilter{})
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
			fakeVersion: es.ElasticV7,
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
			fakeVersion: es.ElasticV7,
			fakeResp:    "es7_count_logs_by_interval_400.json",
			fakeCode:    http.StatusBadRequest,
			expectedErr: "type: search_phase_execution_exception, reason: all shards failed",
		},
		{
			fakeVersion: es.ElasticV7,
			fakeResp:    "es7_count_logs_by_interval_404.json",
			fakeCode:    http.StatusNotFound,
			expectedErr: "type: index_not_found_exception, reason: no such index [ks-logstash-log-20]",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			srv := mockElasticsearchService("/ks-logstash-log*/_search", test.fakeResp, test.fakeCode)
			defer srv.Close()

			client, err := NewClient(&logging.Options{
				Host:        srv.URL,
				IndexPrefix: "ks-logstash-log",
				Version:     test.fakeVersion,
			})
			if err != nil {
				t.Fatalf("create client error, %s", err)
			}

			result, err := client.CountLogsByInterval(logging.SearchFilter{}, "15m")
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
			fakeVersion: es.ElasticV7,
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

			client, err := NewClient(&logging.Options{
				Host:        srv.URL,
				IndexPrefix: "ks-logstash-log",
				Version:     test.fakeVersion,
			})
			if err != nil {
				t.Fatalf("create client error, %s", err)
			}

			result, err := client.SearchLogs(logging.SearchFilter{}, 0, 10, "asc")
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

func TestParseToQueryPart(t *testing.T) {
	var tests = []struct {
		filter   logging.SearchFilter
		expected string
	}{
		{
			filter: logging.SearchFilter{
				NamespaceFilter: map[string]*time.Time{
					"default": func() *time.Time { t := time.Unix(1589981934, 0); return &t }(),
				},
			},
			expected: "api_body_1.json",
		},
		{
			filter: logging.SearchFilter{
				WorkloadFilter: []string{"mysql"},
				Starttime:      time.Unix(1589980934, 0),
				Endtime:        time.Unix(1589981934, 0),
			},
			expected: "api_body_2.json",
		},
		{
			filter: logging.SearchFilter{
				PodFilter: []string{"mysql"},
				PodSearch: []string{"mysql-a8w3s-10945j"},
				LogSearch: []string{"info"},
			},
			expected: "api_body_3.json",
		},
		{
			filter: logging.SearchFilter{
				ContainerFilter: []string{"mysql-1"},
				ContainerSearch: []string{"mysql-3"},
			},
			expected: "api_body_4.json",
		},
		{
			filter: logging.SearchFilter{
				Starttime: time.Unix(1590744676, 0),
			},
			expected: "api_body_7.json",
		},
		{
			filter: logging.SearchFilter{
				NamespaceFilter: map[string]*time.Time{
					"default": nil,
				},
			},
			expected: "api_body_8.json",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {

			expected, err := ioutil.ReadFile(fmt.Sprintf("./testdata/%s", test.expected))
			if err != nil {
				t.Fatalf("read expected error, %s", err.Error())
			}

			result, _ := query.NewBuilder().WithQuery(parseToQueryPart(test.filter)).Bytes()
			if diff := cmp.Diff(string(result), string(result)); diff != "" {
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
		_, _ = res.Write(b)
	})
	return httptest.NewServer(mux)
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
