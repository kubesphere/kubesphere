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
	MainQuery     MainQuery     `json:"query"`
	Aggs          interface{}   `json:"aggs,omitempty"`
	MainHighLight MainHighLight `json:"highlight,omitempty"`
}

type Sort struct {
	Order Order `json:"time"`
}

type Order struct {
	Order string `json:"order"`
}

type MainQuery struct {
	MainBoolQuery MainBoolQuery `json:"bool"`
}

type MainBoolQuery struct {
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
	Fields []interface{} `json:"fields,omitempty"`
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

type StatisticsAggs struct {
	NamespaceAgg NamespaceAgg `json:"Namespace"`
}

type NamespaceAgg struct {
	Terms         StatisticsAggTerm `json:"terms"`
	ContainerAggs ContainerAggs     `json:"aggs"`
}

type ContainerAggs struct {
	ContainerAgg ContainerAgg `json:"Container"`
}

type ContainerAgg struct {
	Terms StatisticsAggTerm `json:"terms"`
}

type StatisticsAggTerm struct {
	Field string `json:"field"`
	Size  int64  `json:"size"`
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
	var mainBoolQuery MainBoolQuery

	if param.NamespaceFilled {
		var shouldMatchPhrase ShouldMatchPhrase
		if len(param.Namespaces) == 0 {
			matchPhrase := MatchPhrase{map[string]interface{}{"kubernetes.namespace_name.key_word": QueryWord{""}}}
			shouldMatchPhrase.Shoulds = append(shouldMatchPhrase.Shoulds, matchPhrase)
		} else {
			for _, namespace := range param.Namespaces {
				matchPhrase := MatchPhrase{map[string]interface{}{"kubernetes.namespace_name.keyword": QueryWord{namespace}}}
				shouldMatchPhrase.Shoulds = append(shouldMatchPhrase.Shoulds, matchPhrase)
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
		containerAggs := ContainerAggs{ContainerAgg{StatisticsAggTerm{"kubernetes.container_name.keyword", 2147483647}}}
		namespaceAgg := NamespaceAgg{StatisticsAggTerm{"kubernetes.namespace_name.keyword", 2147483647}, containerAggs}
		request.Aggs = StatisticsAggs{namespaceAgg}
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
		request.MainHighLight = mainHighLight
	}

	request.MainQuery = MainQuery{mainBoolQuery}

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

type NamespaceAggregations struct {
	NamespaceAggregation NamespaceAggregation `json:"Namespace"`
}

type NamespaceAggregation struct {
	Namespaces []NamespaceStatistics `json:"buckets"`
}

type NamespaceStatistics struct {
	Namespace            string               `json:"Key"`
	Count                int64                `json:"doc_count"`
	ContainerAggregation ContainerAggregation `json:"Container"`
}

type ContainerAggregation struct {
	Containers []ContainerStatistics `json:"buckets"`
}

type ContainerStatistics struct {
	Container string `json:"Key"`
	Count     int64  `json:"doc_count"`
}

type NamespaceResult struct {
	Namespace  string            `json:"namespace"`
	Count      int64             `json:"count"`
	Containers []ContainerResult `json:"containers"`
}

type ContainerResult struct {
	Container string `json:"container"`
	Count     int64  `json:"count"`
}

type StatisticsResult struct {
	Total      int64             `json:"total"`
	Namespaces []NamespaceResult `json:"namespaces"`
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
		//fmt.Println("Parse response error ", err.Error())
		queryResult.Status = http.StatusNotFound
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
		var statisticsResult StatisticsResult
		statisticsResult.Total = response.Hits.Total

		var namespaceAggregations NamespaceAggregations
		jsonIter.Unmarshal(response.Aggregations, &namespaceAggregations)

		for _, namespace := range namespaceAggregations.NamespaceAggregation.Namespaces {
			var namespaceResult NamespaceResult
			namespaceResult.Namespace = namespace.Namespace
			namespaceResult.Count = namespace.Count

			for _, container := range namespace.ContainerAggregation.Containers {
				var containerResult ContainerResult
				containerResult.Container = container.Container
				containerResult.Count = container.Count
				namespaceResult.Containers = append(namespaceResult.Containers, containerResult)
			}

			statisticsResult.Namespaces = append(statisticsResult.Namespaces, namespaceResult)
		}

		queryResult.Statistics = &statisticsResult

	case OperationHistogram:
		var histogramResult HistogramResult
		histogramResult.Total = response.Hits.Total
		histogramResult.StartTime = calcTimestamp(param.StartTime)
		histogramResult.EndTime = calcTimestamp(param.EndTime)
		histogramResult.Interval = param.Interval

		var histogramAggregations HistogramAggregations
		jsonIter.Unmarshal(response.Aggregations, &histogramAggregations)
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
	NamespaceFilled bool
	Namespaces      []string
	PodFilled       bool
	Pods            []string
	ContainerFilled bool
	Containers      []string

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
