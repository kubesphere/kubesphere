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
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func MockElasticsearchService(pattern string, fakeCode int, fakeResp string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(fakeCode)
		res.Write([]byte(fakeResp))
	})
	return httptest.NewServer(mux)
}

func TestStatisticsOnResources(t *testing.T) {
	var tests = []struct {
		description   string
		filter        events.Filter
		fakeVersion   string
		fakeCode      int
		fakeResp      string
		expected      events.Statistics
		expectedError bool
	}{{
		description: "ES index exists",
		filter:      events.Filter{},
		fakeVersion: "6",
		fakeCode:    200,
		fakeResp: `
{
  "took": 16,
  "timed_out": false,
  "_shards": {
    "total": 1,
    "successful": 1,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": 10000,
    "max_score": null,
    "hits": [

    ]
  },
  "aggregations": {
    "resources_count": {
      "value": 100
    }
  }
}
`,
		expected: events.Statistics{
			Events:    10000,
			Resources: 100,
		},
		expectedError: false,
	}, {
		description: "ES index not exists",
		filter:      events.Filter{},
		fakeVersion: "6",
		fakeCode:    404,
		fakeResp: `
{
  "error": {
    "root_cause": [
      {
        "type": "index_not_found_exception",
        "reason": "no such index [events]",
        "resource.type": "index_or_alias",
        "resource.id": "events",
        "index_uuid": "_na_",
        "index": "events"
      }
    ],
    "type": "index_not_found_exception",
    "reason": "no such index [events]",
    "resource.type": "index_or_alias",
    "resource.id": "events",
    "index_uuid": "_na_",
    "index": "events"
  },
  "status": 404
}
`,
		expectedError: true,
	}}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			mes := MockElasticsearchService("/", test.fakeCode, test.fakeResp)
			defer mes.Close()

			es, err := NewClient(&Options{Host: mes.URL, IndexPrefix: "ks-logstash-events", Version: "6"})

			if err != nil {
				t.Fatal(err)
			}

			stats, err := es.StatisticsOnResources(&test.filter)

			if test.expectedError {
				if err == nil {
					t.Fatalf("expected err like %s", test.fakeResp)
				} else if !strings.Contains(err.Error(), strconv.Itoa(test.fakeCode)) {
					t.Fatalf("err does not contain expected code: %d", test.fakeCode)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				} else if diff := cmp.Diff(stats, &test.expected); diff != "" {
					t.Fatalf("%T differ (-got, +want): %s", test.expected, diff)
				}
			}
		})
	}
}

func TestParseToQueryPart(t *testing.T) {
	q := `
{
	"bool": {
		"filter": [
			{
				"bool": {
					"should": [
						{
							"bool": {
								"filter": [
									{
										"match_phrase": {
											"involvedObject.namespace.keyword": "kubesphere-system"
										}
									},
									{
										"range": {
											"lastTimestamp": {
												"gte": "2020-01-01T01:01:01.000000001Z"
											}
										}
									}
								]
							}
						}
					],
					"minimum_should_match": 1
				}
			},
			{
				"bool": {
					"should": [
						{
							"match_phrase_prefix": {
								"involvedObject.name": "istio"
							}
						}
					],
					"minimum_should_match": 1
				}
			},
			{
				"bool": {
					"should": [
						{
							"match_phrase": {
								"reason": "unhealthy"
							}
						}
					],
					"minimum_should_match": 1
				}
			},
			{
				"range": {
					"lastTimestamp": {
						"gte": "2019-12-01T01:01:01.000000001Z"
					}
				}
			}
		]
	}
}
`
	nsCreateTime := time.Date(2020, time.Month(1), 1, 1, 1, 1, 1, time.UTC)
	startTime := nsCreateTime.AddDate(0, -1, 0)

	filter := &events.Filter{
		InvolvedObjectNamespaceMap: map[string]time.Time{
			"kubesphere-system": nsCreateTime,
		},
		InvolvedObjectNameFuzzy: []string{"istio"},
		Reasons:                 []string{"unhealthy"},
		StartTime:               &startTime,
	}

	qp := parseToQueryPart(filter)
	bs, err := json.Marshal(qp)
	if err != nil {
		panic(err)
	}

	queryPart := &map[string]interface{}{}
	if err := json.Unmarshal(bs, queryPart); err != nil {
		panic(err)
	}
	expectedQueryPart := &map[string]interface{}{}
	if err := json.Unmarshal([]byte(q), expectedQueryPart); err != nil {
		panic(err)
	}

	assert.Equal(t, expectedQueryPart, queryPart)
}
