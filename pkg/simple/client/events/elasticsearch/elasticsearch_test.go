package elasticsearch

import (
	"github.com/stretchr/testify/assert"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
	"testing"
	"time"
)

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
