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
	"github.com/golang/glog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/json-iterator/go"
)

const (
	matchPhrase = iota
	matchPhrasePrefix

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

type Config struct {
	Host         string
	Port         string
	Index        string
	VersionMajor string
}

func (cfg *Config) WriteESConfigs() {
	mu.Lock()
	defer mu.Unlock()

	config = cfg
	if err := detectVersionMajor(config); err != nil {
		glog.Errorln(err)
		client = nil
		return
	}

	client = NewForConfig(config)
}

type Request struct {
	From          int64         `json:"from"`
	Size          int64         `json:"size"`
	Sorts         []Sort        `json:"sort,omitempty"`
	MainQuery     BoolQuery     `json:"query"`
	Aggs          interface{}   `json:"aggs,omitempty"`
	MainHighLight MainHighLight `json:"highlight,omitempty"`
}

type Sort struct {
	Order Order `json:"time"`
}

type Order struct {
	Order string `json:"order"`
}

type BoolQuery struct {
	Bool interface{} `json:"bool"`
}

type FilterContext struct {
	Filter []interface{} `json:"filter"`
}

type BoolMust struct {
	Must []interface{} `json:"must"`
}

type BoolShould struct {
	Should             []interface{} `json:"should"`
	MinimumShouldMatch int64         `json:"minimum_should_match"`
}

type RangeQuery struct {
	RangeSpec RangeSpec `json:"range"`
}

type RangeSpec struct {
	TimeRange TimeRange `json:"time"`
}

type TimeRange struct {
	Gte string `json:"gte,omitempty"`
	Lte string `json:"lte,omitempty"`
}

type MatchPhrase struct {
	MatchPhrase map[string]string `json:"match_phrase"`
}

type MatchPhrasePrefix struct {
	MatchPhrasePrefix interface{} `json:"match_phrase_prefix"`
}

type MainHighLight struct {
	Fields       []interface{} `json:"fields,omitempty"`
	FragmentSize int           `json:"fragment_size"`
}

type LogHighLightField struct {
	FieldContent EmptyField `json:"log"`
}

type NamespaceHighLightField struct {
	FieldContent EmptyField `json:"kubernetes.namespace_name.keyword"`
}

type PodHighLightField struct {
	FieldContent EmptyField `json:"kubernetes.pod_name.keyword"`
}

type ContainerHighLightField struct {
	FieldContent EmptyField `json:"kubernetes.container_name.keyword"`
}

type EmptyField struct {
}

// StatisticsAggs, the struct for `aggs` of type Request, holds a cardinality aggregation for distinct container counting
type StatisticsAggs struct {
	ContainerAgg ContainerAgg `json:"containers"`
}

type ContainerAgg struct {
	Cardinality AggField `json:"cardinality"`
}

type AggField struct {
	Field string `json:"field"`
}

type HistogramAggs struct {
	HistogramAgg HistogramAgg `json:"histogram"`
}

type HistogramAgg struct {
	DateHistogram DateHistogram `json:"date_histogram"`
}

type DateHistogram struct {
	Field    string `json:"field"`
	Interval string `json:"interval"`
}

func createQueryRequest(param QueryParameters) (int, []byte, error) {
	var request Request
	var mainBoolQuery FilterContext

	if len(param.NamespaceWithCreationTime) != 0 {
		var boolShoulds BoolShould
		for namespace, creationTime := range param.NamespaceWithCreationTime {
			var boolMusts BoolMust

			matchPhrase := MatchPhrase{MatchPhrase: map[string]string{fieldNamespaceNameKeyword: namespace}}
			rangeQuery := RangeQuery{RangeSpec{TimeRange{creationTime, ""}}}

			boolMusts.Must = append(boolMusts.Must, matchPhrase)
			boolMusts.Must = append(boolMusts.Must, rangeQuery)

			boolShoulds.Should = append(boolShoulds.Should, BoolQuery{Bool: boolMusts})
		}
		boolShoulds.MinimumShouldMatch = 1
		mainBoolQuery.Filter = append(mainBoolQuery.Filter, BoolQuery{Bool: boolShoulds})
	}
	if param.WorkloadFilter != nil {
		boolQuery := makeBoolShould(matchPhrase, fieldPodNameKeyword, param.WorkloadFilter)
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

// Fore more info, refer to https://www.elastic.co/guide/en/elasticsearch/reference/current/getting-started-search-API.html
// Response from the elasticsearch engine
type Response struct {
	Status       int             `json:"status"`
	Workspace    string          `json:"workspace,omitempty"`
	Shards       Shards          `json:"_shards"`
	Hits         Hits            `json:"hits"`
	Aggregations json.RawMessage `json:"aggregations"`
}

type Shards struct {
	Total      int64 `json:"total"`
	Successful int64 `json:"successful"`
	Skipped    int64 `json:"skipped"`
	Failed     int64 `json:"failed"`
}

type Hits struct {
	// As of ElasticSearch v7.x, hits.total is changed
	Total interface{} `json:"total"`
	Hits  []Hit       `json:"hits"`
}

type Hit struct {
	Source    Source    `json:"_source"`
	HighLight HighLight `json:"highlight"`
	Sort      []int64   `json:"sort"`
}

type Source struct {
	Log        string     `json:"log"`
	Time       string     `json:"time"`
	Kubernetes Kubernetes `json:"kubernetes"`
}

type Kubernetes struct {
	Namespace string `json:"namespace_name"`
	Pod       string `json:"pod_name"`
	Container string `json:"container_name"`
	Host      string `json:"host"`
}

type HighLight struct {
	LogHighLights       []string `json:"log,omitempty" description:"log messages to highlight"`
	NamespaceHighLights []string `json:"kubernetes.namespace_name.keyword,omitempty" description:"namespaces to highlight"`
	PodHighLights       []string `json:"kubernetes.pod_name.keyword,omitempty" description:"pods to highlight"`
	ContainerHighLights []string `json:"kubernetes.container_name.keyword,omitempty" description:"containers to highlight"`
}

type LogRecord struct {
	Time      int64     `json:"time,omitempty" description:"log timestamp"`
	Log       string    `json:"log,omitempty" description:"log message"`
	Namespace string    `json:"namespace,omitempty" description:"namespace"`
	Pod       string    `json:"pod,omitempty" description:"pod name"`
	Container string    `json:"container,omitempty" description:"container name"`
	Host      string    `json:"host,omitempty" description:"node id"`
	HighLight HighLight `json:"highlight,omitempty" description:"highlighted log fragment"`
}

type ReadResult struct {
	Total   int64       `json:"total" description:"total number of matched results"`
	From    int64       `json:"from" description:"the offset from the result set"`
	Size    int64       `json:"size" description:"the amount of hits to be returned"`
	Records []LogRecord `json:"records,omitempty" description:"actual array of results"`
}

// StatisticsResponseAggregations, the struct for `aggregations` of type Reponse, holds return results from the aggregation StatisticsAggs
type StatisticsResponseAggregations struct {
	ContainerCount ContainerCount `json:"containers"`
}

type ContainerCount struct {
	Value int64 `json:"value"`
}

type HistogramAggregations struct {
	HistogramAggregation HistogramAggregation `json:"histogram"`
}

type HistogramAggregation struct {
	Histograms []HistogramStatistics `json:"buckets"`
}

type HistogramStatistics struct {
	Time  int64 `json:"key"`
	Count int64 `json:"doc_count"`
}

type HistogramRecord struct {
	Time  int64 `json:"time" description:"timestamp"`
	Count int64 `json:"count" description:"total number of logs at intervals"`
}

type StatisticsResult struct {
	Containers int64 `json:"containers" description:"total number of containers"`
	Logs       int64 `json:"logs" description:"total number of logs"`
}

type HistogramResult struct {
	Total      int64             `json:"total" description:"total number of logs"`
	StartTime  int64             `json:"start_time" description:"start time"`
	EndTime    int64             `json:"end_time" description:"end time"`
	Interval   string            `json:"interval" description:"interval"`
	Histograms []HistogramRecord `json:"histograms" description:"actual array of histogram results"`
}

// Wrap elasticsearch response
type QueryResult struct {
	Status     int               `json:"status,omitempty" description:"query status"`
	Error      string            `json:"error,omitempty" description:"debugging information"`
	Read       *ReadResult       `json:"query,omitempty" description:"query results"`
	Statistics *StatisticsResult `json:"statistics,omitempty" description:"statistics results"`
	Histogram  *HistogramResult  `json:"histogram,omitempty" description:"histogram results"`
}

const (
	OperationQuery int = iota
	OperationStatistics
	OperationHistogram
)

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
		glog.Errorln(err)
		queryResult.Status = http.StatusInternalServerError
		queryResult.Error = err.Error()
		return &queryResult
	}

	if response.Status != 0 {
		//Elastic error, eg, es_rejected_execute_exception
		err := "The query failed with no response"
		queryResult.Status = response.Status
		queryResult.Error = err
		glog.Errorln(err)
		return &queryResult
	}

	if response.Shards.Successful != response.Shards.Total {
		//Elastic some shards error
		glog.Warningf("Not all shards succeed, successful shards: %d, skipped shards: %d, failed shards: %d",
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
			glog.Errorln(err)
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
			glog.Errorln(err)
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

type QueryParameters struct {
	// when true, indicates the provided `namespaces` or `namespace_query` doesn't match any namespace
	NamespaceNotFound bool
	// a map of namespace with creation time
	NamespaceWithCreationTime map[string]string

	// when true, indicates the provided `workloads` or `workload_query` doesn't match any workload
	WorkloadNotFound bool
	WorkloadFilter   []string

	PodFilter []string
	PodQuery  []string

	ContainerFilter []string
	ContainerQuery  []string

	LogQuery []string

	Operation string
	Interval  string
	StartTime string
	EndTime   string
	Sort      string
	From      int64
	Size      int64
}

func Query(param QueryParameters) *QueryResult {

	var queryResult = new(QueryResult)

	if param.NamespaceNotFound || param.WorkloadNotFound {
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
		glog.Errorln(err)
		queryResult.Status = http.StatusInternalServerError
		queryResult.Error = err.Error()
		return queryResult
	}

	body, err := client.Search(query)
	if err != nil {
		glog.Errorln(err)
		queryResult = new(QueryResult)
		queryResult.Status = http.StatusInternalServerError
		queryResult.Error = err.Error()
		return queryResult
	}

	queryResult = parseQueryResult(operation, param, body)

	return queryResult
}
