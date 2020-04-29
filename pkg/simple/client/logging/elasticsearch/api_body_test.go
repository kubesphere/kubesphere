package elasticsearch

import (
	"github.com/google/go-cmp/cmp"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"testing"
)

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
