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
	"bytes"
	"fmt"
	"strings"
	"time"

	es5 "github.com/elastic/go-elasticsearch/v5"
	es6 "github.com/elastic/go-elasticsearch/v6"
	es7 "github.com/elastic/go-elasticsearch/v7"
	jsoniter "github.com/json-iterator/go"
	"kubesphere.io/kubesphere/pkg/simple/client/auditing"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Elasticsearch struct {
	c    client
	opts struct {
		index string
	}
}

func (es *Elasticsearch) SearchAuditingEvent(filter *auditing.Filter, from, size int64,
	sort string) (*auditing.Events, error) {
	queryPart := parseToQueryPart(filter)
	if sort == "" {
		sort = "desc"
	}
	sortPart := []map[string]interface{}{{
		"RequestReceivedTimestamp": map[string]string{"order": sort},
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
		*auditing.Event `json:"_source"`
	}
	if err := json.Unmarshal(resp.Hits.Hits, &innerHits); err != nil {
		return nil, err
	}
	evts := auditing.Events{Total: resp.Hits.Total}
	for _, hit := range innerHits {
		evts.Records = append(evts.Records, hit.Event)
	}
	return &evts, nil
}

func (es *Elasticsearch) CountOverTime(filter *auditing.Filter, interval string) (*auditing.Histogram, error) {
	if interval == "" {
		interval = "15m"
	}

	queryPart := parseToQueryPart(filter)
	aggName := "events_count_over_timestamp"
	aggsPart := map[string]interface{}{
		aggName: map[string]interface{}{
			"date_histogram": map[string]string{
				"field":    "RequestReceivedTimestamp",
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
	histo := auditing.Histogram{Total: int64(len(agg.Buckets))}
	for _, b := range agg.Buckets {
		histo.Buckets = append(histo.Buckets,
			auditing.Bucket{Time: b.Key, Count: b.DocCount})
	}
	return &histo, nil
}

func (es *Elasticsearch) StatisticsOnResources(filter *auditing.Filter) (*auditing.Statistics, error) {
	queryPart := parseToQueryPart(filter)
	aggName := "resources_count"
	aggsPart := map[string]interface{}{
		aggName: map[string]interface{}{
			"cardinality": map[string]string{
				"field": "AuditID.keyword",
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

	return &auditing.Statistics{
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

func parseToQueryPart(f *auditing.Filter) interface{} {
	if f == nil {
		return nil
	}
	type BoolBody struct {
		Filter             []map[string]interface{} `json:"filter,omitempty"`
		Should             []map[string]interface{} `json:"should,omitempty"`
		MinimumShouldMatch *int                     `json:"minimum_should_match,omitempty"`
	}
	var mini = 1
	b := BoolBody{}
	queryBody := map[string]interface{}{
		"bool": &b,
	}

	if len(f.ObjectRefNamespaceMap) > 0 {
		bi := BoolBody{MinimumShouldMatch: &mini}
		for k, v := range f.ObjectRefNamespaceMap {
			bi.Should = append(bi.Should, map[string]interface{}{
				"bool": &BoolBody{
					Filter: []map[string]interface{}{{
						"match_phrase": map[string]string{"ObjectRef.Namespace.keyword": k},
					}, {
						"range": map[string]interface{}{
							"RequestReceivedTimestamp": map[string]interface{}{
								"gte": v,
							},
						},
					}},
				},
			})
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

	if len(f.ObjectRefNames) > 0 {
		if bi := shouldBoolbody("match_phrase_prefix", "ObjectRef.Name.keyword",
			f.ObjectRefNames, nil); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}
	if len(f.ObjectRefNameFuzzy) > 0 {
		if bi := shouldBoolbody("wildcard", "ObjectRef.Name",
			f.ObjectRefNameFuzzy, func(s string) string {
				return fmt.Sprintf("*" + s + "*")
			}); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}

	if len(f.Verbs) > 0 {
		if bi := shouldBoolbody("match_phrase", "Verb",
			f.Verbs, nil); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}
	if len(f.Levels) > 0 {
		if bi := shouldBoolbody("match_phrase", "Level",
			f.Levels, nil); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}

	if len(f.SourceIpFuzzy) > 0 {
		if bi := shouldBoolbody("wildcard", "SourceIPs",
			f.SourceIpFuzzy, func(s string) string {
				return fmt.Sprintf("*" + s + "*")
			}); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}

	if len(f.Users) > 0 {
		if bi := shouldBoolbody("match_phrase", "User.Username.keyword",
			f.Users, nil); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}
	if len(f.UserFuzzy) > 0 {
		if bi := shouldBoolbody("wildcard", "User.Username",
			f.UserFuzzy, func(s string) string {
				return fmt.Sprintf("*" + s + "*")
			}); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}

	if len(f.GroupFuzzy) > 0 {
		if bi := shouldBoolbody("wildcard", "User.Groups",
			f.GroupFuzzy, func(s string) string {
				return fmt.Sprintf("*" + s + "*")
			}); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}

	if len(f.ObjectRefResources) > 0 {
		if bi := shouldBoolbody("match_phrase_prefix", "ObjectRef.Resource.keyword",
			f.ObjectRefResources, nil); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}

	if len(f.ObjectRefSubresources) > 0 {
		if bi := shouldBoolbody("match_phrase_prefix", "ObjectRef.Subresource.keyword",
			f.ObjectRefSubresources, nil); bi != nil {
			b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
		}
	}

	if f.ResponseCodes != nil && len(f.ResponseCodes) > 0 {

		bi := BoolBody{MinimumShouldMatch: &mini}
		for _, v := range f.ResponseCodes {
			bi.Should = append(bi.Should, map[string]interface{}{
				"term": map[string]int32{"ResponseStatus.code": v},
			})
		}

		b.Filter = append(b.Filter, map[string]interface{}{"bool": bi})
	}

	if len(f.ResponseStatus) > 0 {
		if bi := shouldBoolbody("match_phrase", "ResponseStatus.status",
			f.ResponseStatus, nil); bi != nil {
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
			"range": map[string]interface{}{"RequestReceivedTimestamp": m},
		})

	}

	return queryBody
}
