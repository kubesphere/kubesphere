/*
Copyright 2018 The KubeSphere Authors.
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
package esclient

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/logging/v1alpha2"
	v5 "kubesphere.io/kubesphere/pkg/simple/client/elasticsearch/versions/v5"
	v6 "kubesphere.io/kubesphere/pkg/simple/client/elasticsearch/versions/v6"
	v7 "kubesphere.io/kubesphere/pkg/simple/client/elasticsearch/versions/v7"
	"strings"
	"time"

	"github.com/json-iterator/go"
)

const (
	matchPhrase = iota
	matchPhrasePrefix
	regexpQuery

	podNameMaxLength = 63
	// max 10 characters + 1 hyphen
	replicaSetSuffixMaxLength = 11
	// a unique random string as suffix, 5 characters + 1 hyphen
	randSuffixLength = 6

	fieldPodName       = "kubernetes.pod_name"
	fieldContainerName = "kubernetes.container_name"
	fieldLog           = "log"

	fieldNamespaceNameKeyword = "kubernetes.namespace_name.keyword"
	fieldPodNameKeyword       = "kubernetes.pod_name.keyword"
	fieldContainerNameKeyword = "kubernetes.container_name.keyword"
)

const (
	ElasticV5 = "5"
	ElasticV6 = "6"
	ElasticV7 = "7"
)

var jsonIter = jsoniter.ConfigCompatibleWithStandardLibrary

type ElasticSearchClient struct {
	client Client
}

func NewLoggingClient(options *ElasticSearchOptions) (*ElasticSearchClient, error) {
	var version, index string
	esClient := &ElasticSearchClient{}

	if options.Version == "" {
		var err error
		version, err = detectVersionMajor(options.Host)
		if err != nil {
			return nil, err
		}
	} else {
		version = options.Version
	}

	if options.IndexPrefix != "" {
		index = options.IndexPrefix
	} else {
		index = "logstash"
	}

	switch version {
	case ElasticV5:
		esClient.client = v5.New(options.Host, index)
	case ElasticV6:
		esClient.client = v6.New(options.Host, index)
	case ElasticV7:
		esClient.client = v7.New(options.Host, index)
	default:
		return nil, fmt.Errorf("unsupported elasticsearch version %s", version)
	}

	return esClient, nil
}

func (c *ElasticSearchClient) ES() *Client {
	return &c.client
}

func detectVersionMajor(host string) (string, error) {

	// Info APIs are backward compatible with versions of v5.x, v6.x and v7.x
	es := v6.New(host, "")
	res, err := es.Client.Info(
		es.Client.Info.WithContext(context.Background()),
	)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	var b map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&b); err != nil {
		return "", err
	}
	if res.IsError() {
		// Print the response status and error information.
		e, _ := b["error"].(map[string]interface{})
		return "", fmt.Errorf("[%s] type: %v, reason: %v", res.Status(), e["type"], e["reason"])
	}

	// get the major version
	version, _ := b["version"].(map[string]interface{})
	number, _ := version["number"].(string)
	if number == "" {
		return "", fmt.Errorf("failed to detect elastic version number")
	}

	v := strings.Split(number, ".")[0]
	return v, nil
}

func createQueryRequest(param v1alpha2.QueryParameters) ([]byte, error) {
	var request v1alpha2.Request
	var mainBoolQuery v1alpha2.BoolFilter

	if len(param.NamespaceWithCreationTime) != 0 {
		var boolShould v1alpha2.BoolShould
		for namespace, creationTime := range param.NamespaceWithCreationTime {
			var boolFilter v1alpha2.BoolFilter

			matchPhrase := v1alpha2.MatchPhrase{MatchPhrase: map[string]string{fieldNamespaceNameKeyword: namespace}}
			rangeQuery := v1alpha2.RangeQuery{RangeSpec: v1alpha2.RangeSpec{TimeRange: v1alpha2.TimeRange{Gte: creationTime, Lte: ""}}}

			boolFilter.Filter = append(boolFilter.Filter, matchPhrase)
			boolFilter.Filter = append(boolFilter.Filter, rangeQuery)

			boolShould.Should = append(boolShould.Should, v1alpha2.BoolQuery{Bool: boolFilter})
		}
		boolShould.MinimumShouldMatch = 1
		mainBoolQuery.Filter = append(mainBoolQuery.Filter, v1alpha2.BoolQuery{Bool: boolShould})
	}
	if param.WorkloadFilter != nil {
		boolQuery := makeBoolShould(regexpQuery, fieldPodNameKeyword, param.WorkloadFilter)
		mainBoolQuery.Filter = append(mainBoolQuery.Filter, boolQuery)
	}
	if param.PodFilter != nil {
		boolQuery := makeBoolShould(matchPhrase, fieldPodNameKeyword, param.PodFilter)
		mainBoolQuery.Filter = append(mainBoolQuery.Filter, boolQuery)
	}
	if param.ContainerFilter != nil {
		boolQuery := makeBoolShould(matchPhrase, fieldContainerNameKeyword, param.ContainerFilter)
		mainBoolQuery.Filter = append(mainBoolQuery.Filter, boolQuery)
	}

	if param.WorkloadQuery != nil {
		boolQuery := makeBoolShould(matchPhrasePrefix, fieldPodName, param.WorkloadQuery)
		mainBoolQuery.Filter = append(mainBoolQuery.Filter, boolQuery)
	}
	if param.PodQuery != nil {
		boolQuery := makeBoolShould(matchPhrasePrefix, fieldPodName, param.PodQuery)
		mainBoolQuery.Filter = append(mainBoolQuery.Filter, boolQuery)
	}
	if param.ContainerQuery != nil {
		boolQuery := makeBoolShould(matchPhrasePrefix, fieldContainerName, param.ContainerQuery)
		mainBoolQuery.Filter = append(mainBoolQuery.Filter, boolQuery)
	}
	if param.LogQuery != nil {
		boolQuery := makeBoolShould(matchPhrasePrefix, fieldLog, param.LogQuery)
		mainBoolQuery.Filter = append(mainBoolQuery.Filter, boolQuery)
	}

	rangeQuery := v1alpha2.RangeQuery{RangeSpec: v1alpha2.RangeSpec{TimeRange: v1alpha2.TimeRange{Gte: param.StartTime, Lte: param.EndTime}}}
	mainBoolQuery.Filter = append(mainBoolQuery.Filter, rangeQuery)

	if param.Operation == v1alpha2.OperationStatistics {
		containerAgg := v1alpha2.AggField{Field: "kubernetes.docker_id.keyword"}
		statisticAggs := v1alpha2.StatisticsAggs{ContainerAgg: v1alpha2.ContainerAgg{Cardinality: containerAgg}}
		request.Aggs = statisticAggs
		request.Size = 0
	} else if param.Operation == v1alpha2.OperationHistogram {
		var interval string
		if param.Interval != "" {
			interval = param.Interval
		} else {
			interval = "15m"
		}
		param.Interval = interval
		request.Aggs = v1alpha2.HistogramAggs{HistogramAgg: v1alpha2.HistogramAgg{DateHistogram: v1alpha2.DateHistogram{Field: "time", Interval: interval}}}
		request.Size = 0
	} else {
		request.From = param.From
		request.Size = param.Size
		var order string
		if strings.Compare(strings.ToLower(param.Sort), "asc") == 0 {
			order = "asc"
		} else {
			order = "desc"
		}
		request.Sorts = append(request.Sorts, v1alpha2.Sort{Order: v1alpha2.Order{Order: order}})
	}

	request.MainQuery = v1alpha2.BoolQuery{Bool: mainBoolQuery}

	return json.Marshal(request)
}

func makeBoolShould(queryType int, field string, list []string) v1alpha2.BoolQuery {
	var should []interface{}
	for _, phrase := range list {

		var q interface{}

		switch queryType {
		case matchPhrase:
			q = v1alpha2.MatchPhrase{MatchPhrase: map[string]string{field: phrase}}
		case matchPhrasePrefix:
			q = v1alpha2.MatchPhrasePrefix{MatchPhrasePrefix: map[string]string{field: phrase}}
		case regexpQuery:
			q = v1alpha2.RegexpQuery{Regexp: map[string]string{field: makePodNameRegexp(phrase)}}
		}

		should = append(should, q)
	}

	return v1alpha2.BoolQuery{
		Bool: v1alpha2.BoolShould{
			Should:             should,
			MinimumShouldMatch: 1,
		},
	}
}

func makePodNameRegexp(workloadName string) string {
	var regexp string
	if len(workloadName) <= podNameMaxLength-replicaSetSuffixMaxLength-randSuffixLength {
		// match deployment pods, eg. <deploy>-579dfbcddd-24znw
		// replicaset rand string is limited to vowels
		// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/util/rand/rand.go#L83
		regexp += workloadName + "-[bcdfghjklmnpqrstvwxz2456789]{1,10}-[a-z0-9]{5}|"
		// match statefulset pods, eg. <sts>-0
		regexp += workloadName + "-[0-9]+|"
		// match pods of daemonset or job, eg. <ds>-29tdk, <job>-5xqvl
		regexp += workloadName + "-[a-z0-9]{5}"
	} else if len(workloadName) <= podNameMaxLength-randSuffixLength {
		replicaSetSuffixLength := podNameMaxLength - randSuffixLength - len(workloadName)
		regexp += fmt.Sprintf("%s%d%s", workloadName+"-[bcdfghjklmnpqrstvwxz2456789]{", replicaSetSuffixLength, "}[a-z0-9]{5}|")
		regexp += workloadName + "-[0-9]+|"
		regexp += workloadName + "-[a-z0-9]{5}"
	} else {
		// Rand suffix may overwrites the workload name if the name is too long
		// This won't happen for StatefulSet because a statefulset pod will fail to create
		regexp += workloadName[:podNameMaxLength-randSuffixLength+1] + "[a-z0-9]{5}|"
		regexp += workloadName + "-[0-9]+"
	}
	return regexp
}

func (c *ElasticSearchClient) parseQueryResult(operation int, body []byte) (*v1alpha2.QueryResult, error) {
	var queryResult v1alpha2.QueryResult

	var response v1alpha2.Response
	err := jsonIter.Unmarshal(body, &response)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if response.Shards.Successful != response.Shards.Total {
		//Elastic some shards error
		klog.Warningf("Not all shards succeed, successful shards: %d, skipped shards: %d, failed shards: %d",
			response.Shards.Successful, response.Shards.Skipped, response.Shards.Failed)
	}

	switch operation {
	case v1alpha2.OperationQuery:
		var readResult v1alpha2.ReadResult
		readResult.Total = c.client.GetTotalHitCount(response.Hits.Total)
		for _, hit := range response.Hits.Hits {
			var logRecord v1alpha2.LogRecord
			logRecord.Time = hit.Source.Time
			logRecord.Log = hit.Source.Log
			logRecord.Namespace = hit.Source.Kubernetes.Namespace
			logRecord.Pod = hit.Source.Kubernetes.Pod
			logRecord.Container = hit.Source.Kubernetes.Container
			logRecord.Host = hit.Source.Kubernetes.Host
			readResult.Records = append(readResult.Records, logRecord)
		}
		queryResult.Read = &readResult
	case v1alpha2.OperationStatistics:
		var statisticsResponse v1alpha2.StatisticsResponseAggregations
		err := jsonIter.Unmarshal(response.Aggregations, &statisticsResponse)
		if err != nil && response.Aggregations != nil {
			klog.Error(err)
			return nil, err
		}
		queryResult.Statistics = &v1alpha2.StatisticsResult{Containers: statisticsResponse.ContainerCount.Value, Logs: c.client.GetTotalHitCount(response.Hits.Total)}
	case v1alpha2.OperationHistogram:
		var histogramResult v1alpha2.HistogramResult
		histogramResult.Total = c.client.GetTotalHitCount(response.Hits.Total)

		var histogramAggregations v1alpha2.HistogramAggregations
		err = jsonIter.Unmarshal(response.Aggregations, &histogramAggregations)
		if err != nil && response.Aggregations != nil {
			klog.Error(err)
			return nil, err
		}
		for _, histogram := range histogramAggregations.HistogramAggregation.Histograms {
			var histogramRecord v1alpha2.HistogramRecord
			histogramRecord.Time = histogram.Time
			histogramRecord.Count = histogram.Count

			histogramResult.Histograms = append(histogramResult.Histograms, histogramRecord)
		}

		queryResult.Histogram = &histogramResult
	case v1alpha2.OperationExport:
		var readResult v1alpha2.ReadResult
		readResult.ScrollID = response.ScrollId
		for _, hit := range response.Hits.Hits {
			var logRecord v1alpha2.LogRecord
			logRecord.Log = hit.Source.Log
			readResult.Records = append(readResult.Records, logRecord)
		}
		queryResult.Read = &readResult
	}

	return &queryResult, nil
}

func (c *ElasticSearchClient) Query(param v1alpha2.QueryParameters) (*v1alpha2.QueryResult, error) {

	var queryResult = new(v1alpha2.QueryResult)

	if param.NamespaceNotFound {
		queryResult = new(v1alpha2.QueryResult)
		switch param.Operation {
		case v1alpha2.OperationStatistics:
			queryResult.Statistics = new(v1alpha2.StatisticsResult)
		case v1alpha2.OperationHistogram:
			queryResult.Histogram = new(v1alpha2.HistogramResult)
		default:
			queryResult.Read = new(v1alpha2.ReadResult)
		}
		return queryResult, nil
	}

	query, err := createQueryRequest(param)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	body, err := c.client.Search(query, param.ScrollTimeout)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return c.parseQueryResult(param.Operation, body)
}

func (c *ElasticSearchClient) Scroll(scrollId string) (*v1alpha2.QueryResult, error) {
	body, err := c.client.Scroll(scrollId, time.Minute)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return c.parseQueryResult(v1alpha2.OperationExport, body)
}

func (c *ElasticSearchClient) ClearScroll(scrollId string) {
	c.client.ClearScroll(scrollId)
}
