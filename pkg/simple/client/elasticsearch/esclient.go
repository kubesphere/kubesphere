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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/json-iterator/go"
)

var jsonIter = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	mu        sync.RWMutex
	esConfigs *ESConfigs
)

type ESConfigs struct {
	Host  string
	Port  string
	Index string
}

func readESConfigs() *ESConfigs {
	mu.RLock()
	defer mu.RUnlock()

	return esConfigs
}

func (configs *ESConfigs) WriteESConfigs() {
	mu.Lock()
	defer mu.Unlock()

	esConfigs = configs
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
	BoolMusts BoolMusts `json:"bool"`
}

type BoolMusts struct {
	Musts []interface{} `json:"must"`
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

type BoolShouldMatchPhrase struct {
	ShouldMatchPhrase ShouldMatchPhrase `json:"bool"`
}

type ShouldMatchPhrase struct {
	Shoulds            []interface{} `json:"should"`
	MinimumShouldMatch int64         `json:"minimum_should_match"`
}

type MatchPhrase struct {
	MatchPhrase interface{} `json:"match_phrase"`
}

type Match struct {
	Match interface{} `json:"match"`
}

type QueryWord struct {
	Word string `json:"query"`
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

// The aggs object holds two aggregations to be computed by Elasticsearch
// ContainterAgg is a cardinality aggregation to calculate the count of distinct containers
// StartTimeAgg is a top hits aggregation to retrieve the first record
type StatisticsAggs struct {
	ContainerAgg ContainerAgg `json:"containers"`
	StartTimeAgg StartTimeAgg `json:"starttime"`
}

type ContainerAgg struct {
	Cardinality AggField `json:"cardinality"`
}

type AggField struct {
	Field string `json:"field"`
}

type StartTimeAgg struct {
	TopHits TopHits `json:"top_hits"`
}

type TopHits struct {
	Sort []Sort `json:"sort"`
	Size int    `json:"size"`
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
	var mainBoolQuery BoolMusts

	if param.NamespaceFilled {
		var shouldMatchPhrase ShouldMatchPhrase
		if len(param.NamespaceWithCreationTime) == 0 {
			matchPhrase := MatchPhrase{map[string]interface{}{"kubernetes.namespace_name.key_word": QueryWord{""}}}
			shouldMatchPhrase.Shoulds = append(shouldMatchPhrase.Shoulds, matchPhrase)
		} else {
			for namespace, creationTime := range param.NamespaceWithCreationTime {
				var boolQuery BoolQuery

				matchPhrase := MatchPhrase{map[string]interface{}{"kubernetes.namespace_name.keyword": QueryWord{namespace}}}
				rangeQuery := RangeQuery{RangeSpec{TimeRange{creationTime, ""}}}

				boolQuery.BoolMusts.Musts = append(boolQuery.BoolMusts.Musts, matchPhrase)
				boolQuery.BoolMusts.Musts = append(boolQuery.BoolMusts.Musts, rangeQuery)

				shouldMatchPhrase.Shoulds = append(shouldMatchPhrase.Shoulds, boolQuery)
			}
		}
		shouldMatchPhrase.MinimumShouldMatch = 1
		mainBoolQuery.Musts = append(mainBoolQuery.Musts, BoolShouldMatchPhrase{shouldMatchPhrase})
	}
	if param.PodFilled {
		var shouldMatchPhrase ShouldMatchPhrase
		if len(param.Pods) == 0 {
			matchPhrase := MatchPhrase{map[string]interface{}{"kubernetes.pod_name.key_word": QueryWord{""}}}
			shouldMatchPhrase.Shoulds = append(shouldMatchPhrase.Shoulds, matchPhrase)
		} else {
			for _, pod := range param.Pods {
				matchPhrase := MatchPhrase{map[string]interface{}{"kubernetes.pod_name.keyword": QueryWord{pod}}}
				shouldMatchPhrase.Shoulds = append(shouldMatchPhrase.Shoulds, matchPhrase)
			}
		}
		shouldMatchPhrase.MinimumShouldMatch = 1
		mainBoolQuery.Musts = append(mainBoolQuery.Musts, BoolShouldMatchPhrase{shouldMatchPhrase})
	}
	if param.ContainerFilled {
		var shouldMatchPhrase ShouldMatchPhrase
		if len(param.Containers) == 0 {
			matchPhrase := MatchPhrase{map[string]interface{}{"kubernetes.container_name.key_word": QueryWord{""}}}
			shouldMatchPhrase.Shoulds = append(shouldMatchPhrase.Shoulds, matchPhrase)
		} else {
			for _, container := range param.Containers {
				matchPhrase := MatchPhrase{map[string]interface{}{"kubernetes.container_name.keyword": QueryWord{container}}}
				shouldMatchPhrase.Shoulds = append(shouldMatchPhrase.Shoulds, matchPhrase)
			}
		}
		shouldMatchPhrase.MinimumShouldMatch = 1
		mainBoolQuery.Musts = append(mainBoolQuery.Musts, BoolShouldMatchPhrase{shouldMatchPhrase})
	}

	if param.NamespaceQuery != "" {
		match := Match{map[string]interface{}{"kubernetes.namespace_name": QueryWord{param.NamespaceQuery}}}
		mainBoolQuery.Musts = append(mainBoolQuery.Musts, match)
	}
	if param.PodQuery != "" {
		match := Match{map[string]interface{}{"kubernetes.pod_name": QueryWord{param.PodQuery}}}
		mainBoolQuery.Musts = append(mainBoolQuery.Musts, match)
	}
	if param.ContainerQuery != "" {
		match := Match{map[string]interface{}{"kubernetes.container_name": QueryWord{param.ContainerQuery}}}
		mainBoolQuery.Musts = append(mainBoolQuery.Musts, match)
	}

	if param.LogQuery != "" {
		match := Match{map[string]interface{}{"log": QueryWord{param.LogQuery}}}
		mainBoolQuery.Musts = append(mainBoolQuery.Musts, match)
	}

	rangeQuery := RangeQuery{RangeSpec{TimeRange{param.StartTime, param.EndTime}}}
	mainBoolQuery.Musts = append(mainBoolQuery.Musts, rangeQuery)

	var operation int

	if param.Operation == "statistics" {
		operation = OperationStatistics
		containerAgg := AggField{"kubernetes.docker_id.keyword"}
		startTimeAgg := TopHits{[]Sort{{Order{"asc"}}}, 1}
		statisticAggs := StatisticsAggs{ContainerAgg{containerAgg}, StartTimeAgg{startTimeAgg}}
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
	Total int64 `json:"total"`
	Hits  []Hit `json:"hits"`
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
	LogHighLights       []string `json:"log,omitempty"`
	NamespaceHighLights []string `json:"kubernetes.namespace_name.keyword,omitempty"`
	PodHighLights       []string `json:"kubernetes.pod_name.keyword,omitempty"`
	ContainerHighLights []string `json:"kubernetes.container_name.keyword,omitempty"`
}

type LogRecord struct {
	Time      int64     `json:"time,omitempty"`
	Log       string    `json:"log,omitempty"`
	Namespace string    `json:"namespace,omitempty"`
	Pod       string    `json:"pod,omitempty"`
	Container string    `json:"container,omitempty"`
	Host      string    `json:"host,omitempty"`
	HighLight HighLight `json:"highlight,omitempty"`
}

type ReadResult struct {
	Total   int64       `json:"total"`
	From    int64       `json:"from"`
	Size    int64       `json:"size"`
	Records []LogRecord `json:"records,omitempty"`
}

// The aggregations object represents the return from an aggregation (see StatisticsAggs type)
type StatisticsResponseAggregations struct {
	ContainerCount ContainerCount  `json:"containers"`
	StartTime      StartTimeTopHit `json:"starttime"`
}

type ContainerCount struct {
	Value int64 `json:"value"`
}

type StartTimeTopHit struct {
	Hits Hits `json:"hits"`
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
	Time  int64 `json:"time"`
	Count int64 `json:"count"`
}

type StatisticsResult struct {
	Containers int64 `json:"containers"`
	Logs       int64 `json:"logs"`
	StartTime  int64 `json:"starttime"`
}

type HistogramResult struct {
	Total      int64             `json:"total"`
	StartTime  int64             `json:"start_time"`
	EndTime    int64             `json:"end_time"`
	Interval   string            `json:"interval"`
	Histograms []HistogramRecord `json:"histograms"`
}

type QueryResult struct {
	Status     int               `json:"status,omitempty"`
	Workspace  string            `json:"workspace,omitempty"`
	Read       *ReadResult       `json:"query,omitempty"`
	Statistics *StatisticsResult `json:"statistics,omitempty"`
	Histogram  *HistogramResult  `json:"histogram,omitempty"`
	Request    string            `json:"request,omitempty"`
	Response   string            `json:"response,omitempty"`
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

func parseQueryResult(operation int, param QueryParameters, body []byte, query []byte) *QueryResult {
	var queryResult QueryResult
	//queryResult.Request = string(query)
	//queryResult.Response = string(body)

	var response Response
	err := jsonIter.Unmarshal(body, &response)
	if err != nil {
		glog.Errorln(err)
		queryResult.Status = http.StatusInternalServerError
		return &queryResult
	}

	if response.Status != 0 {
		//Elastic error, eg, es_rejected_execute_exception
		queryResult.Status = response.Status
		return &queryResult
	}

	if response.Shards.Successful != response.Shards.Total {
		//Elastic some shards error
		queryResult.Status = http.StatusInternalServerError
		return &queryResult
	}

	switch operation {
	case OperationQuery:
		var readResult ReadResult
		readResult.Total = response.Hits.Total
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
		if err != nil {
			glog.Errorln(err)
			queryResult.Status = http.StatusInternalServerError
			return &queryResult
		}
		queryResult.Statistics = &StatisticsResult{Containers: statisticsResponse.ContainerCount.Value,
			Logs: statisticsResponse.StartTime.Hits.Total, StartTime: statisticsResponse.StartTime.Hits.Hits[0].Sort[0]}

	case OperationHistogram:
		var histogramResult HistogramResult
		histogramResult.Total = response.Hits.Total
		histogramResult.StartTime = calcTimestamp(param.StartTime)
		histogramResult.EndTime = calcTimestamp(param.EndTime)
		histogramResult.Interval = param.Interval

		var histogramAggregations HistogramAggregations
		err := jsonIter.Unmarshal(response.Aggregations, &histogramAggregations)
		if err != nil {
			glog.Errorln(err)
			queryResult.Status = http.StatusInternalServerError
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
	queryResult.Workspace = param.Workspace

	return &queryResult
}

type QueryParameters struct {
	NamespaceFilled           bool
	Namespaces                []string
	NamespaceWithCreationTime map[string]string
	PodFilled                 bool
	Pods                      []string
	ContainerFilled           bool
	Containers                []string

	NamespaceQuery string
	PodQuery       string
	ContainerQuery string

	Workspace string

	Operation string
	LogQuery  string
	Interval  string
	StartTime string
	EndTime   string
	Sort      string
	From      int64
	Size      int64
}

func stubResult() *QueryResult {
	var queryResult QueryResult

	queryResult.Status = http.StatusOK

	return &queryResult
}

func Query(param QueryParameters) *QueryResult {
	var queryResult *QueryResult

	//queryResult = stubResult()
	//return queryResult

	client := &http.Client{}

	operation, query, err := createQueryRequest(param)
	if err != nil {
		//fmt.Println("Create query error ", err.Error())
		queryResult = new(QueryResult)
		queryResult.Status = http.StatusNotFound
		return queryResult
	}

	es := readESConfigs()
	if es == nil {
		queryResult = new(QueryResult)
		queryResult.Status = http.StatusNotFound
		return queryResult
	}

	url := fmt.Sprintf("http://%s:%s/%s*/_search", es.Host, es.Port, es.Index)
	request, err := http.NewRequest("GET", url, bytes.NewBuffer(query))
	if err != nil {
		glog.Errorln(err)
		queryResult = new(QueryResult)
		queryResult.Status = http.StatusNotFound
		return queryResult
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := client.Do(request)
	if err != nil {
		glog.Errorln(err)
		queryResult = new(QueryResult)
		queryResult.Status = http.StatusNotFound
		return queryResult
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		glog.Errorln(err)
		queryResult = new(QueryResult)
		queryResult.Status = http.StatusNotFound
		return queryResult
	}

	queryResult = parseQueryResult(operation, param, body, query)

	return queryResult
}
