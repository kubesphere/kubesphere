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
	"kubesphere.io/kubesphere/pkg/simple/client/es"
	"kubesphere.io/kubesphere/pkg/simple/client/es/query"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
)

type client struct {
	c *es.Client
}

func NewClient(options *events.Options) (events.Client, error) {
	c := &client{}

	var err error
	c.c, err = es.NewClient(options.Host, options.IndexPrefix, options.Version)
	return c, err
}

func (c *client) SearchEvents(filter *events.Filter, from, size int64,
	sort string) (*events.Events, error) {

	b := query.NewBuilder().
		WithQuery(parseToQueryPart(filter)).
		WithSort("lastTimestamp", sort).
		WithFrom(from).
		WithSize(size)

	resp, err := c.c.Search(b, filter.StartTime, filter.EndTime, false)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.AllHits) == 0 {
		return &events.Events{}, nil
	}

	evts := events.Events{Total: c.c.GetTotalHitCount(resp.Total)}
	for _, hit := range resp.AllHits {
		evts.Records = append(evts.Records, hit.Source)
	}
	return &evts, nil
}

func (c *client) CountOverTime(filter *events.Filter, interval string) (*events.Histogram, error) {
	if interval == "" {
		interval = "15m"
	}

	b := query.NewBuilder().
		WithQuery(parseToQueryPart(filter)).
		WithAggregations(query.NewAggregations().
			WithDateHistogramAggregation("lastTimestamp", interval)).
		WithSize(0)

	resp, err := c.c.Search(b, filter.StartTime, filter.EndTime, false)
	if err != nil {
		return nil, err
	}

	histo := events.Histogram{Total: c.c.GetTotalHitCount(resp.Total)}
	for _, bucket := range resp.Buckets {
		histo.Buckets = append(histo.Buckets,
			events.Bucket{Time: bucket.Key, Count: bucket.Count})
	}
	return &histo, nil
}

func (c *client) StatisticsOnResources(filter *events.Filter) (*events.Statistics, error) {

	b := query.NewBuilder().
		WithQuery(parseToQueryPart(filter)).
		WithAggregations(query.NewAggregations().
			WithCardinalityAggregation("involvedObject.uid.keyword")).
		WithSize(0)

	resp, err := c.c.Search(b, filter.StartTime, filter.EndTime, false)
	if err != nil {
		return nil, err
	}

	return &events.Statistics{
		Resources: resp.Value,
		Events:    c.c.GetTotalHitCount(resp.Total),
	}, nil
}

func parseToQueryPart(f *events.Filter) *query.Query {
	if f == nil {
		return nil
	}

	var mini int32 = 1
	b := query.NewBool()

	bi := query.NewBool().WithMinimumShouldMatch(mini)
	for k, v := range f.InvolvedObjectNamespaceMap {
		if k == "" {
			bi.AppendShould(query.NewBool().
				AppendMustNot(query.NewExists("field", "involvedObject.namespace")))
		} else {
			bi.AppendShould(query.NewBool().
				AppendFilter(query.NewMatchPhrase("involvedObject.namespace.keyword", k)).
				AppendFilter(query.NewRange("lastTimestamp").
					WithGTE(v)))
		}
	}
	b.AppendFilter(bi)

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrase("involvedObject.name.keyword", f.InvolvedObjectNames)).
		WithMinimumShouldMatch(mini))

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrasePrefix("involvedObject.name", f.InvolvedObjectNameFuzzy)).
		WithMinimumShouldMatch(mini))

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrase("involvedObject.kind", f.InvolvedObjectkinds)).
		WithMinimumShouldMatch(mini))

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrase("reason", f.Reasons)).
		WithMinimumShouldMatch(mini))

	bi = query.NewBool().WithMinimumShouldMatch(mini)
	for _, r := range f.ReasonFuzzy {
		bi.AppendShould(query.NewWildcard("reason.keyword", fmt.Sprintf("*"+r+"*")))
	}
	b.AppendFilter(bi)

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrasePrefix("message", f.MessageFuzzy)).
		WithMinimumShouldMatch(mini))

	if f.Type != "" {
		b.AppendFilter(query.NewBool().
			AppendShould(query.NewMatchPhrase("type", f.Type)))
	}

	r := query.NewRange("lastTimestamp")
	if !f.StartTime.IsZero() {
		r.WithGTE(f.StartTime)
	}
	if !f.EndTime.IsZero() {
		r.WithLTE(f.EndTime)
	}

	b.AppendFilter(r)

	return query.NewQuery().WithBool(b)
}
