package elasticsearch

import (
	"github.com/google/go-cmp/cmp"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"testing"
	"time"
)

func TestMainBool(t *testing.T) {
	var tests = []struct {
		description  string
		searchFilter logging.SearchFilter
		expected     *bodyBuilder
	}{
		{
			description: "filter 2 namespaces",
			searchFilter: logging.SearchFilter{
				NamespaceFilter: map[string]time.Time{
					"kubesphere-system":         time.Unix(1582000000, 0),
					"kubesphere-logging-system": time.Unix(1582969999, 0),
				},
			},
			expected: &bodyBuilder{Body{
				Query: &Query{
					Bool: Bool{
						Filter: []Match{
							{
								Bool: &Bool{
									Should: []Match{
										{
											Bool: &Bool{
												Filter: []Match{
													{
														MatchPhrase: map[string]string{"kubernetes.namespace_name.keyword": "kubesphere-system"},
													},
													{
														Range: &Range{&Time{Gte: func() *time.Time { t := time.Unix(1582000000, 0); return &t }()}},
													},
												},
											},
										},
										{
											Bool: &Bool{
												Filter: []Match{
													{
														MatchPhrase: map[string]string{"kubernetes.namespace_name.keyword": "kubesphere-logging-system"},
													},
													{
														Range: &Range{&Time{Gte: func() *time.Time { t := time.Unix(1582969999, 0); return &t }()}},
													},
												},
											},
										},
									},
									MinimumShouldMatch: 1,
								},
							},
						},
					},
				},
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			body, err := newBodyBuilder().mainBool(test.searchFilter).bytes()
			expected, _ := test.expected.bytes()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(body, expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}

func TestCardinalityAggregation(t *testing.T) {
	var test = struct {
		description  string
		searchFilter logging.SearchFilter
		expected     *bodyBuilder
	}{
		description: "add cardinality aggregation",
		searchFilter: logging.SearchFilter{
			LogSearch: []string{"info"},
		},
		expected: &bodyBuilder{Body{
			Query: &Query{
				Bool: Bool{
					Filter: []Match{
						{
							Bool: &Bool{
								Should: []Match{
									{
										MatchPhrasePrefix: map[string]string{"log": "info"},
									},
								},
								MinimumShouldMatch: 1,
							},
						},
					},
				},
			},
			Aggs: &Aggs{
				CardinalityAggregation: &CardinalityAggregation{
					Cardinality: &Cardinality{Field: "kubernetes.docker_id.keyword"},
				},
			},
		}},
	}

	t.Run(test.description, func(t *testing.T) {
		body := newBodyBuilder().mainBool(test.searchFilter).cardinalityAggregation()
		if diff := cmp.Diff(body, test.expected); diff != "" {
			t.Fatalf("%T differ (-got, +want): %s", test.expected, diff)
		}
	})
}
