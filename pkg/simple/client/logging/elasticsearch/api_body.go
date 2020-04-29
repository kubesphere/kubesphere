package elasticsearch

import (
	"fmt"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"time"
)

const (
	podNameMaxLength          = 63
	podNameSuffixLength       = 6  // 5 characters + 1 hyphen
	replicaSetSuffixMaxLength = 11 // max 10 characters + 1 hyphen
)

type bodyBuilder struct {
	Body
}

func newBodyBuilder() *bodyBuilder {
	return &bodyBuilder{}
}

func (bb *bodyBuilder) bytes() ([]byte, error) {
	return json.Marshal(bb.Body)
}

// The mainBody func builds api body for query.
// TODO: Should use an elegant pakcage for building query body, but `elastic/go-elasticsearch` doesn't provide it currently.
//
// Example:
// GET kapis/logging.kubesphere.io/v1alpha2/cluster?start_time=0&end_time=156576063993&namespaces=kubesphere-system&pod_query=ks-apiserver
// -----
//{
//    "from":0,
//    "size":10,
//    "sort":[
//        {
//            "time": "desc"
//        }
//    ],
//    "query":{
//        "bool":{
//            "filter":[
//                {
//                    "bool":{
//                        "should":[
//                            {
//                                "bool":{
//                                    "filter":[
//                                        {
//                                            "match_phrase":{
//                                                "kubernetes.namespace_name.keyword":"kubesphere-system"
//                                            }
//                                        },
//                                        {
//                                            "range":{
//                                                "time":{
//                                                    "gte":"1572315987000"
//                                                }
//                                            }
//                                        }
//                                    ]
//                                }
//                            }
//                        ],
//                        "minimum_should_match":1
//                    }
//                },
//                {
//                    "bool":{
//                        "should":[
//                            {
//                                "match_phrase_prefix":{
//                                    "kubernetes.pod_name":"ks-apiserver"
//                                }
//                            }
//                        ],
//                        "minimum_should_match":1
//                    }
//                },
//                {
//                    "range":{
//                        "time":{
//                            "gte":"0",
//                            "lte":"156576063993"
//                        }
//                    }
//                }
//            ]
//        }
//    }
//}
func (bb *bodyBuilder) mainBool(sf logging.SearchFilter) *bodyBuilder {
	var ms []Match

	// literal matching
	if len(sf.NamespaceFilter) != 0 {
		var b Bool
		for ns := range sf.NamespaceFilter {
			match := Match{
				Bool: &Bool{
					Filter: []Match{
						{
							MatchPhrase: map[string]string{
								"kubernetes.namespace_name.keyword": ns,
							},
						},
						{
							Range: &Range{
								Time: &Time{
									Gte: func() *time.Time { t := sf.NamespaceFilter[ns]; return &t }(),
								},
							},
						},
					},
				},
			}
			b.Should = append(b.Should, match)
		}
		b.MinimumShouldMatch = 1
		ms = append(ms, Match{Bool: &b})
	}
	if sf.WorkloadFilter != nil {
		var b Bool
		for _, wk := range sf.WorkloadFilter {
			b.Should = append(b.Should, Match{Regexp: map[string]string{"kubernetes.pod_name.keyword": podNameRegexp(wk)}})
		}
		b.MinimumShouldMatch = 1
		ms = append(ms, Match{Bool: &b})
	}
	if sf.PodFilter != nil {
		var b Bool
		for _, po := range sf.PodFilter {
			b.Should = append(b.Should, Match{MatchPhrase: map[string]string{"kubernetes.pod_name.keyword": po}})
		}
		b.MinimumShouldMatch = 1
		ms = append(ms, Match{Bool: &b})
	}
	if sf.ContainerFilter != nil {
		var b Bool
		for _, c := range sf.ContainerFilter {
			b.Should = append(b.Should, Match{MatchPhrase: map[string]string{"kubernetes.container_name.keyword": c}})
		}
		b.MinimumShouldMatch = 1
		ms = append(ms, Match{Bool: &b})
	}

	// fuzzy matching
	if sf.WorkloadSearch != nil {
		var b Bool
		for _, wk := range sf.WorkloadSearch {
			b.Should = append(b.Should, Match{MatchPhrasePrefix: map[string]string{"kubernetes.pod_name": wk}})
		}

		b.MinimumShouldMatch = 1
		ms = append(ms, Match{Bool: &b})
	}
	if sf.PodSearch != nil {
		var b Bool
		for _, po := range sf.PodSearch {
			b.Should = append(b.Should, Match{MatchPhrasePrefix: map[string]string{"kubernetes.pod_name": po}})
		}
		b.MinimumShouldMatch = 1
		ms = append(ms, Match{Bool: &b})
	}
	if sf.ContainerSearch != nil {
		var b Bool
		for _, c := range sf.ContainerSearch {
			b.Should = append(b.Should, Match{MatchPhrasePrefix: map[string]string{"kubernetes.container_name": c}})
		}
		b.MinimumShouldMatch = 1
		ms = append(ms, Match{Bool: &b})
	}
	if sf.LogSearch != nil {
		var b Bool
		for _, l := range sf.LogSearch {
			b.Should = append(b.Should, Match{MatchPhrasePrefix: map[string]string{"log": l}})
		}
		b.MinimumShouldMatch = 1
		ms = append(ms, Match{Bool: &b})
	}

	if !sf.Starttime.IsZero() || !sf.Endtime.IsZero() {
		fromTo := Match{
			Range: &Range{&Time{
				Gte: &sf.Starttime,
				Lte: &sf.Endtime,
			}},
		}
		ms = append(ms, fromTo)
	}

	bb.Body.Query = &Query{Bool{Filter: ms}}
	return bb
}

func (bb *bodyBuilder) cardinalityAggregation() *bodyBuilder {
	bb.Body.Aggs = &Aggs{
		CardinalityAggregation: &CardinalityAggregation{
			&Cardinality{
				Field: "kubernetes.docker_id.keyword",
			},
		},
	}
	return bb
}

func (bb *bodyBuilder) dateHistogramAggregation(interval string) *bodyBuilder {
	if interval == "" {
		interval = "15m"
	}

	bb.Body.Aggs = &Aggs{
		DateHistogramAggregation: &DateHistogramAggregation{
			&DateHistogram{
				Field:    "time",
				Interval: interval,
			},
		},
	}
	return bb
}

func (bb *bodyBuilder) from(n int64) *bodyBuilder {
	bb.From = n
	return bb
}

func (bb *bodyBuilder) size(n int64) *bodyBuilder {
	bb.Size = n
	return bb
}

func (bb *bodyBuilder) sort(o string) *bodyBuilder {
	if o != "asc" {
		o = "desc"
	}

	bb.Sorts = []map[string]string{{"time": o}}
	return bb
}

func podNameRegexp(workloadName string) string {
	var regexp string
	if len(workloadName) <= podNameMaxLength-replicaSetSuffixMaxLength-podNameSuffixLength {
		// match deployment pods, eg. <deploy>-579dfbcddd-24znw
		// replicaset rand string is limited to vowels
		// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/util/rand/rand.go#L83
		regexp += workloadName + "-[bcdfghjklmnpqrstvwxz2456789]{1,10}-[a-z0-9]{5}|"
		// match statefulset pods, eg. <sts>-0
		regexp += workloadName + "-[0-9]+|"
		// match pods of daemonset or job, eg. <ds>-29tdk, <job>-5xqvl
		regexp += workloadName + "-[a-z0-9]{5}"
	} else if len(workloadName) <= podNameMaxLength-podNameSuffixLength {
		replicaSetSuffixLength := podNameMaxLength - podNameSuffixLength - len(workloadName)
		regexp += fmt.Sprintf("%s%d%s", workloadName+"-[bcdfghjklmnpqrstvwxz2456789]{", replicaSetSuffixLength, "}[a-z0-9]{5}|")
		regexp += workloadName + "-[0-9]+|"
		regexp += workloadName + "-[a-z0-9]{5}"
	} else {
		// Rand suffix may overwrites the workload name if the name is too long
		// This won't happen for StatefulSet because long name will cause ReplicaSet fails during StatefulSet creation.
		regexp += workloadName[:podNameMaxLength-podNameSuffixLength+1] + "[a-z0-9]{5}|"
		regexp += workloadName + "-[0-9]+"
	}
	return regexp
}

func parseResponse(body []byte) (Response, error) {
	var res Response
	err := json.Unmarshal(body, &res)
	if err != nil {
		klog.Error(err)
		return Response{}, err
	}
	return res, nil
}
