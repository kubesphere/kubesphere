/*
Copyright 2020 The KubeSphere Authors.

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
	"kubesphere.io/kubesphere/pkg/simple/client/auditing"
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
		_, _ = res.Write([]byte(fakeResp))
	})
	return httptest.NewServer(mux)
}

func TestStatisticsOnResources(t *testing.T) {
	var tests = []struct {
		description   string
		filter        auditing.Filter
		fakeVersion   string
		fakeCode      int
		fakeResp      string
		expected      auditing.Statistics
		expectedError bool
	}{{
		description: "ES index exists",
		filter:      auditing.Filter{},
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
		expected: auditing.Statistics{
			Events:    10000,
			Resources: 100,
		},
		expectedError: false,
	}, {
		description: "ES index not exists",
		filter:      auditing.Filter{},
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
											"ObjectRef.Namespace.keyword": "kubesphere-system"
										}
									},
									{
										"range": {
											"RequestReceivedTimestamp": {
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
							"match_phrase": {
								"ObjectRef.Name.keyword": "istio"
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
							"wildcard": {
								"ObjectRef.Name.keyword": "*istio*"
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
								"Verb.keyword": "create"
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
								"Level.keyword": "Metadata"
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
							"wildcard": {
								"SourceIPs.keyword": "*192.168*"
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
								"User.Username.keyword": "system:serviceaccount:kubesphere-system:kubesphere"
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
							"wildcard": {
								"User.Username.keyword": "*system:serviceaccount*"
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
							"wildcard": {
								"User.Groups.keyword": "*system:serviceaccounts*"
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
								"ObjectRef.Resource.keyword": "devops"
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
								"ObjectRef.Subresource.keyword": "pipeline"
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
							"term": {
								"ResponseStatus.code": 404
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
								"ResponseStatus.status": "Failure"
							}
						}
					],
					"minimum_should_match": 1
				}
			},
			{
				"range": {
					"RequestReceivedTimestamp": {
						"gte": "2019-12-01T01:01:01.000000001Z",
						"lte": "2020-01-01T01:01:01.000000001Z"
					}
				}
			}
		]
	}
}
`
	nsCreateTime := time.Date(2020, time.Month(1), 1, 1, 1, 1, 1, time.UTC)
	startTime := nsCreateTime.AddDate(0, -1, 0)
	endTime := nsCreateTime.AddDate(0, 0, 0)

	filter := &auditing.Filter{
		ObjectRefNamespaceMap: map[string]time.Time{
			"kubesphere-system": nsCreateTime,
		},
		ObjectRefNames:        []string{"istio"},
		ObjectRefNameFuzzy:    []string{"istio"},
		Levels:                []string{"Metadata"},
		Verbs:                 []string{"create"},
		Users:                 []string{"system:serviceaccount:kubesphere-system:kubesphere"},
		UserFuzzy:             []string{"system:serviceaccount"},
		GroupFuzzy:            []string{"system:serviceaccounts"},
		SourceIpFuzzy:         []string{"192.168"},
		ObjectRefResources:    []string{"devops"},
		ObjectRefSubresources: []string{"pipeline"},
		ResponseCodes:         []int32{404},
		ResponseStatus:        []string{"Failure"},
		StartTime:             &startTime,
		EndTime:               &endTime,
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
