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
	"encoding/json"
	"fmt"
	"k8s.io/klog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/json-iterator/go"
)

const (
	OperationQuery int = iota
	OperationStatistics
	OperationHistogram

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

var jsonIter = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	mu     sync.Mutex
	config *Config

	client Client
)

func (cfg *Config) WriteESConfigs() {
	mu.Lock()
	defer mu.Unlock()

	config = cfg
	if err := detectVersionMajor(config); err != nil {
		klog.Errorln(err)
		client = nil
		return
	}

	client = NewForConfig(config)
}

func createQueryRequest(param QueryParameters) (int, []byte, error) {
	var request Request
	var mainBoolQuery BoolFilter

	if len(param.NamespaceWithCreationTime) != 0 {
		var boolShould BoolShould
		for namespace, creationTime := range param.NamespaceWithCreationTime {
			var boolFilter BoolFilter

			matchPhrase := MatchPhrase{MatchPhrase: map[string]string{fieldNamespaceNameKeyword: namespace}}
			rangeQuery := RangeQuery{RangeSpec{TimeRange{creationTime, ""}}}

			boolFilter.Filter = append(boolFilter.Filter, matchPhrase)
			boolFilter.Filter = append(boolFilter.Filter, rangeQuery)

			boolShould.Should = append(boolShould.Should, BoolQuery{Bool: boolFilter})
		}
		boolShould.MinimumShouldMatch = 1
		mainBoolQuery.Filter = append(mainBoolQuery.Filter, BoolQuery{Bool: boolShould})
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

	rangeQuery := RangeQuery{RangeSpec{TimeRange{param.StartTime, param.EndTime}}}
	mainBoolQuery.Filter = append(mainBoolQuery.Filter, rangeQuery)

	var operation int

	if param.Operation == "statistics" {
		operation = OperationStatistics
		containerAgg := AggField{"kubernetes.docker_id.keyword"}
		statisticAggs := StatisticsAggs{ContainerAgg{containerAgg}}
		request.Aggs = statisticAggs
		request.Size = 0
	} else if param.Operation == "histogram" {
		operation = OperationHistogram
		var interval string
		if param.Interval != "" {
			interval = param.Interval
		} else {
			interval = "15m"
		}
		param.Interval = interval
		request.Aggs = HistogramAggs{HistogramAgg{DateHistogram{"time", interval}}}
		request.Size = 0
	} else {
		operation = OperationQuery
		request.From = param.From
		request.Size = param.Size
		var order string
		if strings.Compare(strings.ToLower(param.Sort), "asc") == 0 {
			order = "asc"
		} else {
			order = "desc"
		}
		request.Sorts = append(request.Sorts, Sort{Order{order}})

		var mainHighLight MainHighLight
		mainHighLight.Fields = append(mainHighLight.Fields, LogHighLightField{})
		mainHighLight.Fields = append(mainHighLight.Fields, NamespaceHighLightField{})
		mainHighLight.Fields = append(mainHighLight.Fields, PodHighLightField{})
		mainHighLight.Fields = append(mainHighLight.Fields, ContainerHighLightField{})
		mainHighLight.FragmentSize = 0
		request.MainHighLight = mainHighLight
	}

	request.MainQuery = BoolQuery{mainBoolQuery}

	queryRequest, err := json.Marshal(request)

	return operation, queryRequest, err
}

func makeBoolShould(queryType int, field string, list []string) BoolQuery {
	var should []interface{}
	for _, phrase := range list {

		var q interface{}

		switch queryType {
		case matchPhrase:
			q = MatchPhrase{MatchPhrase: map[string]string{field: phrase}}
		case matchPhrasePrefix:
			q = MatchPhrasePrefix{MatchPhrasePrefix: map[string]string{field: phrase}}
		case regexpQuery:
			q = RegexpQuery{Regexp: map[string]string{field: makePodNameRegexp(phrase)}}
		}

		should = append(should, q)
	}

	return BoolQuery{
		Bool: BoolShould{
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

func calcTimestamp(input string) int64 {
	var t time.Time
	var err error
	var ret int64

	ret = 0

	t, err = time.Parse(time.RFC3339, input)
	if err != nil {
		var i int64
		i, err = strconv.ParseInt(input, 10, 64)
		if err == nil {
			ret = time.Unix(i/1000, (i%1000)*1000000).UnixNano() / 1000000
		}
	} else {
		ret = t.UnixNano() / 1000000
	}

	return ret
}

func parseQueryResult(operation int, param QueryParameters, body []byte) *QueryResult {
	var queryResult QueryResult

	var response Response
	err := jsonIter.Unmarshal(body, &response)
	if err != nil {
		klog.Errorln(err)
		queryResult.Status = http.StatusInternalServerError
		queryResult.Error = err.Error()
		return &queryResult
	}

	if response.Status != 0 {
		//Elastic error, eg, es_rejected_execute_exception
		err := "The query failed with no response"
		queryResult.Status = response.Status
		queryResult.Error = err
		klog.Errorln(err)
		return &queryResult
	}

	if response.Shards.Successful != response.Shards.Total {
		//Elastic some shards error
		klog.Warningf("Not all shards succeed, successful shards: %d, skipped shards: %d, failed shards: %d",
			response.Shards.Successful, response.Shards.Skipped, response.Shards.Failed)
	}

	switch operation {
	case OperationQuery:
		var readResult ReadResult
		readResult.Total = client.GetTotalHitCount(response.Hits.Total)
		readResult.From = param.From
		readResult.Size = param.Size
		for _, hit := range response.Hits.Hits {
			var logRecord LogRecord
			logRecord.Time = calcTimestamp(hit.Source.Time)
			logRecord.Log = hit.Source.Log
			logRecord.Namespace = hit.Source.Kubernetes.Namespace
			logRecord.Pod = hit.Source.Kubernetes.Pod
			logRecord.Container = hit.Source.Kubernetes.Container
			logRecord.Host = hit.Source.Kubernetes.Host
			logRecord.HighLight = hit.HighLight
			readResult.Records = append(readResult.Records, logRecord)
		}
		queryResult.Read = &readResult

	case OperationStatistics:
		var statisticsResponse StatisticsResponseAggregations
		err := jsonIter.Unmarshal(response.Aggregations, &statisticsResponse)
		if err != nil && response.Aggregations != nil {
			klog.Errorln(err)
			queryResult.Status = http.StatusInternalServerError
			queryResult.Error = err.Error()
			return &queryResult
		}
		queryResult.Statistics = &StatisticsResult{Containers: statisticsResponse.ContainerCount.Value, Logs: client.GetTotalHitCount(response.Hits.Total)}

	case OperationHistogram:
		var histogramResult HistogramResult
		histogramResult.Total = client.GetTotalHitCount(response.Hits.Total)
		histogramResult.StartTime = calcTimestamp(param.StartTime)
		histogramResult.EndTime = calcTimestamp(param.EndTime)
		histogramResult.Interval = param.Interval

		var histogramAggregations HistogramAggregations
		err := jsonIter.Unmarshal(response.Aggregations, &histogramAggregations)
		if err != nil && response.Aggregations != nil {
			klog.Errorln(err)
			queryResult.Status = http.StatusInternalServerError
			queryResult.Error = err.Error()
			return &queryResult
		}
		for _, histogram := range histogramAggregations.HistogramAggregation.Histograms {
			var histogramRecord HistogramRecord
			histogramRecord.Time = histogram.Time
			histogramRecord.Count = histogram.Count

			histogramResult.Histograms = append(histogramResult.Histograms, histogramRecord)
		}

		queryResult.Histogram = &histogramResult
	}

	queryResult.Status = http.StatusOK

	return &queryResult
}

func Query(param QueryParameters) *QueryResult {

	var queryResult = new(QueryResult)

	if param.NamespaceNotFound {
		queryResult = new(QueryResult)
		queryResult.Status = http.StatusOK
		switch param.Operation {
		case "statistics":
			queryResult.Statistics = new(StatisticsResult)
		case "histogram":
			queryResult.Histogram = &HistogramResult{
				StartTime: calcTimestamp(param.StartTime),
				EndTime:   calcTimestamp(param.EndTime),
				Interval:  param.Interval}
		default:
			queryResult.Read = new(ReadResult)
		}
		return queryResult
	}

	if client == nil {
		queryResult.Status = http.StatusBadRequest
		queryResult.Error = fmt.Sprintf("Invalid elasticsearch address: host=%s, port=%s", config.Host, config.Port)
		return queryResult
	}

	operation, query, err := createQueryRequest(param)
	if err != nil {
		klog.Errorln(err)
		queryResult.Status = http.StatusInternalServerError
		queryResult.Error = err.Error()
		return queryResult
	}

	body, err := client.Search(query)
	if err != nil {
		klog.Errorln(err)
		queryResult = new(QueryResult)
		queryResult.Status = http.StatusInternalServerError
		queryResult.Error = err.Error()
		return queryResult
	}

	queryResult = parseQueryResult(operation, param, body)

	return queryResult
}
