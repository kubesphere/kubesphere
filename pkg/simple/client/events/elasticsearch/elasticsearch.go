package elasticsearch

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	es5 "github.com/elastic/go-elasticsearch/v5"
	es6 "github.com/elastic/go-elasticsearch/v6"
	es7 "github.com/elastic/go-elasticsearch/v7"
	jsoniter "github.com/json-iterator/go"
	corev1 "k8s.io/api/core/v1"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Elasticsearch struct {
	c    client
	opts struct {
		index string
	}
}

func (es *Elasticsearch) SearchEvents(filter *events.Filter, from, size int64,
	sort string) (*events.Events, error) {
	queryPart := parseToQueryPart(filter)
	if sort == "" {
		sort = "desc"
	}
	sortPart := []map[string]interface{}{{
		"lastTimestamp": map[string]string{"order": sort},
	}}
	b := map[string]interface{}{
		"from":  from,
		"size":  size,
		"query": queryPart,
		"sort":  sortPart,
	}

	body, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	resp, err := es.c.ExSearch(&Request{
		Index: es.opts.index,
		Body:  bytes.NewBuffer(body),
	})
	if err != nil || resp == nil {
		return nil, err
	}

	var innerHits []struct {
		*corev1.Event `json:"_source"`
	}
	if err := json.Unmarshal(resp.Hits.Hits, &innerHits); err != nil {
		return nil, err
	}
	evts := events.Events{Total: resp.Hits.Total}
	for _, hit := range innerHits {
		evts.Records = append(evts.Records, hit.Event)
	}
	return &evts, nil
}

func (es *Elasticsearch) CountOverTime(filter *events.Filter, interval string) (*events.Histogram, error) {
	if interval == "" {
		interval = "15m"
	}

	queryPart := parseToQueryPart(filter)
	aggName := "events_count_over_lasttimestamp"
	aggsPart := map[string]interface{}{
		aggName: map[string]interface{}{
			"date_histogram": map[string]string{
				"field":    "lastTimestamp",
				"interval": interval,
			},
		},
	}
	b := map[string]interface{}{
		"query": queryPart,
		"aggs":  aggsPart,
		"size":  0, // do not get docs
	}

	body, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	resp, err := es.c.ExSearch(&Request{
		Index: es.opts.index,
		Body:  bytes.NewBuffer(body),
	})
	if err != nil || resp == nil {
		return nil, err
	}

	raw := resp.Aggregations[aggName]
	var agg struct {
		Buckets []struct {
			KeyAsString string `json:"key_as_string"`
			Key         int64  `json:"key"`
			DocCount    int64  `json:"doc_count"`
		} `json:"buckets"`
	}
	if err := json.Unmarshal(raw, &agg); err != nil {
		return nil, err
	}
	histo := events.Histogram{Total: resp.Hits.Total}
	for _, b := range agg.Buckets {
		histo.Buckets = append(histo.Buckets,
			events.Bucket{Time: b.Key, Count: b.DocCount})
	}
	return &histo, nil
}

func (es *Elasticsearch) StatisticsOnResources(filter *events.Filter) (*events.Statistics, error) {
	queryPart := parseToQueryPart(filter)
	aggName := "resources_count"
	aggsPart := map[string]interface{}{
		aggName: map[string]interface{}{
			"cardinality": map[string]string{
				"field": "involvedObject.uid.keyword",
			},
		},
	}
	b := map[string]interface{}{
		"query": queryPart,
		"aggs":  aggsPart,
		"size":  0, // do not get docs
	}

	body, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	resp, err := es.c.ExSearch(&Request{
		Index: es.opts.index,
		Body:  bytes.NewBuffer(body),
	})
	if err != nil || resp == nil {
		return nil, err
	}

	raw := resp.Aggregations[aggName]
	var agg struct {
		Value int64 `json:"value"`
	}
	if err := json.Unmarshal(raw, &agg); err != nil {
		return nil, err
	}

	return &events.Statistics{
		Resources: agg.Value,
		Events:    resp.Hits.Total,
	}, nil
}

func NewClient(options *Options) (*Elasticsearch, error) {
	clientV5 := func() (*ClientV5, error) {
		c, err := es5.NewClient(es5.Config{Addresses: []string{options.Host}})
		if err != nil {
			return nil, err
		}
		return (*ClientV5)(c), nil
	}
	clientV6 := func() (*ClientV6, error) {
		c, err := es6.NewClient(es6.Config{Addresses: []string{options.Host}})
		if err != nil {
			return nil, err
		}
		return (*ClientV6)(c), nil
	}
	clientV7 := func() (*ClientV7, error) {
		c, err := es7.NewClient(es7.Config{Addresses: []string{options.Host}})
		if err != nil {
			return nil, err
		}
		return (*ClientV7)(c), nil
	}

	var (
		version = options.Version
		es      = Elasticsearch{}
		err     error
	)
	es.opts.index = fmt.Sprintf("%s*", options.IndexPrefix)

	if options.Version == "" {
		var c5 *ClientV5
		if c5, err = clientV5(); err == nil {
			if version, err = c5.Version(); err == nil {
				es.c = c5
			}
		}
	}
	if err != nil {
		return nil, err
	}

	switch strings.Split(version, ".")[0] {
	case "5":
		if es.c == nil {
			es.c, err = clientV5()
		}
	case "6":
		es.c, err = clientV6()
	case "7":
		es.c, err = clientV7()
	default:
		err = fmt.Errorf("unsupported elasticsearch version %s", version)
	}
	if err != nil {
		return nil, err
	}
	return &es, nil
}

func parseToQueryPart(f *events.Filter) interface{} {
	if f == nil {
		return nil
	}
	type BoolBody struct {
		Filter             []map[string]interface{} `json:"filter,omitempty"`
		Should             []map[string]interface{} `json:"should,omitempty"`
		MinimumShouldMatch *int                     `json:"minimum_should_match,omitempty"`
		MustNot            []map[string]interface{} `json:"must_not,omitempty"`
	}
	var mini = 1
	b := BoolBody{}
	queryBody := map[string]interface{}{
		"bool": &b,
	}

	if len(f.InvolvedObjectNamespaceMap) > 0 {
		bi := BoolBody{MinimumShouldMatch: &mini}
		for k, v := range f.InvolvedObjectNamespaceMap {
			if k == "" {
				bi.Should = append(bi.Should, map[string]interface{}{
					"bool": &BoolBody{
						MustNot: []map[string]interface{}{{
							"exists": map[string]string{"field": "involvedObject.namespace"},
						}},
					},
				})
			} else {
				bi.Should = append(bi.Should, map[string]interface{}{
					"bool": &BoolBody{
						Filter: []map[string]interface{}{{
							"match_phrase": map[string]string{"involvedObject.namespace.keyword": k},
						}, {
							"range": map[string]interface{}{
								"lastTimestamp": map[string]interface{}{
									"gte": v,
								},
							},
						}},
					},
				})
			}
		}
		if len(bi.Should) > 0 {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": &bi})
		}
	}

	shouldBoolbody := func(mtype, fieldName string, fieldValues []string, fieldValueMutate func(string) string) *BoolBody {
		bi := BoolBody{MinimumShouldMatch: &mini}
		for _, v := range fieldValues {
			if fieldValueMutate != nil {
				v = fieldValueMutate(v)
			}
			bi.Should = append(bi.Should, map[string]interface{}{
				mtype: map[string]string{fieldName: v},
			})
		}
		if len(bi.Should) == 0 {
			return nil
		}
		return &bi
	}

	if len(f.InvolvedObjectNames) > 0 {
		if bi := shouldBoolbody("match_phrase", "involvedObject.name.keyword",
			f.InvolvedObjectNames, nil); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}
	if len(f.InvolvedObjectNameFuzzy) > 0 {
		if bi := shouldBoolbody("match_phrase_prefix", "involvedObject.name",
			f.InvolvedObjectNameFuzzy, nil); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}
	if len(f.InvolvedObjectkinds) > 0 {
		// involvedObject.kind is single word and here is not field keyword for case ignoring
		if bi := shouldBoolbody("match_phrase", "involvedObject.kind",
			f.InvolvedObjectkinds, nil); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}
	if len(f.Reasons) > 0 {
		// reason is single word and here is not field keyword for case ignoring
		if bi := shouldBoolbody("match_phrase", "reason",
			f.Reasons, nil); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}
	if len(f.ReasonFuzzy) > 0 {
		if bi := shouldBoolbody("wildcard", "reason",
			f.ReasonFuzzy, func(s string) string {
				return fmt.Sprintf("*" + s + "*")
			}); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}
	if len(f.MessageFuzzy) > 0 {
		if bi := shouldBoolbody("match_phrase_prefix", "message",
			f.MessageFuzzy, nil); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}

	if len(f.Type) > 0 {
		// type is single word and here is not field keyword for case ignoring
		if bi := shouldBoolbody("match_phrase", "type",
			[]string{f.Type}, nil); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}

	if f.StartTime != nil || f.EndTime != nil {
		m := make(map[string]*time.Time)
		if f.StartTime != nil {
			m["gte"] = f.StartTime
		}
		if f.EndTime != nil {
			m["lte"] = f.EndTime
		}
		b.Filter = append(b.Filter, map[string]interface{}{
			"range": map[string]interface{}{"lastTimestamp": m},
		})

	}

	return queryBody
}
