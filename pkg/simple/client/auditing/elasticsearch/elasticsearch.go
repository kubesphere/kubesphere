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
	"fmt"
	"github.com/json-iterator/go"
	"kubesphere.io/kubesphere/pkg/simple/client/auditing"
	"kubesphere.io/kubesphere/pkg/simple/client/es"
	"kubesphere.io/kubesphere/pkg/simple/client/es/query"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type client struct {
	c *es.Client
}

func (c *client) SearchAuditingEvent(filter *auditing.Filter, from, size int64,
	sort string) (*auditing.Events, error) {

	b := query.NewBuilder().
		WithQuery(parseToQueryPart(filter)).
		WithSort("RequestReceivedTimestamp", sort).
		WithFrom(from).
		WithSize(size)

	resp, err := c.c.Search(b, filter.StartTime, filter.EndTime, false)
	if err != nil || resp == nil {
		return nil, err
	}

	events := &auditing.Events{Total: c.c.GetTotalHitCount(resp.Total)}
	for _, hit := range resp.AllHits {
		events.Records = append(events.Records, hit.Source)
	}
	return events, nil
}

func (c *client) CountOverTime(filter *auditing.Filter, interval string) (*auditing.Histogram, error) {

	if interval == "" {
		interval = "15m"
	}

	b := query.NewBuilder().
		WithQuery(parseToQueryPart(filter)).
		WithAggregations(query.NewAggregations().
			WithDateHistogramAggregation("RequestReceivedTimestamp", interval)).
		WithSize(0)

	resp, err := c.c.Search(b, filter.StartTime, filter.EndTime, false)
	if err != nil || resp == nil {
		return nil, err
	}

	h := auditing.Histogram{Total: c.c.GetTotalHitCount(resp.Total)}
	for _, bucket := range resp.Buckets {
		h.Buckets = append(h.Buckets,
			auditing.Bucket{Time: bucket.Key, Count: bucket.Count})
	}
	return &h, nil
}

func (c *client) StatisticsOnResources(filter *auditing.Filter) (*auditing.Statistics, error) {

	b := query.NewBuilder().
		WithQuery(parseToQueryPart(filter)).
		WithAggregations(query.NewAggregations().
			WithCardinalityAggregation("AuditID.keyword")).
		WithSize(0)

	resp, err := c.c.Search(b, filter.StartTime, filter.EndTime, false)
	if err != nil || resp == nil {
		return nil, err
	}

	return &auditing.Statistics{
		Resources: resp.Value,
		Events:    c.c.GetTotalHitCount(resp.Total),
	}, nil
}

func NewClient(options *auditing.Options) (auditing.Client, error) {
	c := &client{}

	var err error
	c.c, err = es.NewClient(options.Host, options.IndexPrefix, options.Version)
	return c, err
}

func parseToQueryPart(f *auditing.Filter) *query.Query {
	if f == nil {
		return nil
	}

	var mini int32 = 1
	b := query.NewBool()

	bi := query.NewBool().WithMinimumShouldMatch(mini)
	for k, v := range f.ObjectRefNamespaceMap {
		bi.AppendShould(query.NewBool().
			AppendFilter(query.NewMatchPhrase("ObjectRef.Namespace.keyword", k)).
			AppendFilter(query.NewRange("RequestReceivedTimestamp").
				WithGTE(v)))
	}

	for k, v := range f.WorkspaceMap {
		bi.AppendShould(query.NewBool().
			AppendFilter(query.NewMatchPhrase("Workspace.keyword", k)).
			AppendFilter(query.NewRange("RequestReceivedTimestamp").
				WithGTE(v)))
	}

	b.AppendFilter(bi)

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrase("ObjectRef.Namespace.keyword", f.ObjectRefNamespaces)).
		WithMinimumShouldMatch(mini))

	bi = query.NewBool().WithMinimumShouldMatch(mini)
	for _, ns := range f.ObjectRefNamespaceFuzzy {
		bi.AppendShould(query.NewWildcard("ObjectRef.Namespace.keyword", fmt.Sprintf("*"+ns+"*")))
	}
	b.AppendFilter(bi)

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrase("Workspace.keyword", f.Workspaces)).
		WithMinimumShouldMatch(mini))

	bi = query.NewBool().WithMinimumShouldMatch(mini)
	for _, ws := range f.WorkspaceFuzzy {
		bi.AppendShould(query.NewWildcard("Workspace.keyword", fmt.Sprintf("*"+ws+"*")))
	}
	b.AppendFilter(bi)

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrase("ObjectRef.Name.keyword", f.ObjectRefNames)).
		WithMinimumShouldMatch(mini))

	bi = query.NewBool().WithMinimumShouldMatch(mini)
	for _, name := range f.ObjectRefNameFuzzy {
		bi.AppendShould(query.NewWildcard("ObjectRef.Name.keyword", fmt.Sprintf("*"+name+"*")))
	}
	b.AppendFilter(bi)

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrase("Verb.keyword", f.Verbs)).
		WithMinimumShouldMatch(mini))
	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrase("Level.keyword", f.Levels)).
		WithMinimumShouldMatch(mini))

	bi = query.NewBool().WithMinimumShouldMatch(mini)
	for _, ip := range f.SourceIpFuzzy {
		bi.AppendShould(query.NewWildcard("SourceIPs.keyword", fmt.Sprintf("*"+ip+"*")))
	}
	b.AppendFilter(bi)

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrase("User.Username.keyword", f.Users)).
		WithMinimumShouldMatch(mini))

	bi = query.NewBool().WithMinimumShouldMatch(mini)
	for _, user := range f.UserFuzzy {
		bi.AppendShould(query.NewWildcard("User.Username.keyword", fmt.Sprintf("*"+user+"*")))
	}
	b.AppendFilter(bi)

	bi = query.NewBool().WithMinimumShouldMatch(mini)
	for _, group := range f.GroupFuzzy {
		bi.AppendShould(query.NewWildcard("User.Groups.keyword", fmt.Sprintf("*"+group+"*")))
	}
	b.AppendFilter(bi)

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrasePrefix("ObjectRef.Resource", f.ObjectRefResources)).
		WithMinimumShouldMatch(mini))
	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrasePrefix("ObjectRef.Subresource", f.ObjectRefSubresources)).
		WithMinimumShouldMatch(mini))
	b.AppendFilter(query.NewBool().
		AppendShould(query.NewTerms("ResponseStatus.code", f.ResponseCodes)).
		WithMinimumShouldMatch(mini))
	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrase("ResponseStatus.status.keyword", f.ResponseStatus)).
		WithMinimumShouldMatch(mini))

	r := query.NewRange("RequestReceivedTimestamp")
	if !f.StartTime.IsZero() {
		r.WithGTE(f.StartTime)
	}
	if !f.EndTime.IsZero() {
		r.WithLTE(f.EndTime)
	}

	b.AppendFilter(r)

	return query.NewQuery().WithBool(b)
}
