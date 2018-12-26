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
package client

import (
	"context"
	"encoding/json"

	"github.com/olivere/elastic"

	"kubesphere.io/kubesphere/pkg/constants"
)

type Records struct {
	Source json.RawMessage `json:"_source"`
}

type Source struct {
	Log        string          `json:"log"`
	Time       string          `json:"time"`
	Kubernetes json.RawMessage `json:"kubernetes"`
}

type Kubernetes struct {
	Namespace string `json:"namespace_name"`
	Pod       string `json:"pod_name"`
	Container string `json:"container_name"`
	Host      string `json:"host"`
}

type LogRecord struct {
	Time      string `json:"time,omitempty"`
	Log       string `json:"log,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Pod       string `json:"pod,omitempty"`
	Container string `json:"container,omitempty"`
	Host      string `json:"host,omitempty"`
}

type ReadResult struct {
	Total   int64       `json:"total"`
	From    int         `json:"from"`
	Size    int         `json:"size"`
	Records []LogRecord `json:"records,omitempty"`
}

type NamespaceAggregation struct {
	Namespaces []NamespaceStatistics `json:"buckets"`
}

type NamespaceStatistics struct {
	Namespace            string          `json:"Key"`
	Count                int             `json:"doc_count"`
	ContainerAggregation json.RawMessage `json:"Aggregate by container"`
}

type ContainerAggregation struct {
	Containers []ContainerStatistics `json:"buckets"`
}

type ContainerStatistics struct {
	Container string `json:"Key"`
	Count     int    `json:"doc_count"`
}

type NamespaceResult struct {
	Namespace  string            `json:"namespace"`
	Count      int               `json:"count"`
	Containers []ContainerResult `json:"containers"`
}

type ContainerResult struct {
	Container string `json:"container"`
	Count     int    `json:"count"`
}

type StatisticsResult struct {
	Total      int64             `json:"total"`
	Namespaces []NamespaceResult `json:"namespaces"`
}

type HistogramAggregation struct {
	Histograms []HistogramStatistics `json:"buckets"`
}

type HistogramStatistics struct {
	Time  string `json:"key_as_string"`
	Count int    `json:"doc_count"`
}

type HistogramRecord struct {
	Time  string `json:"time"`
	Count int    `json:"count"`
}

type HistogramResult struct {
	Total      int64             `json:"total"`
	StartTime  string            `json:"start_time"`
	EndTime    string            `json:"end_time"`
	Interval   string            `json:"interval"`
	Histograms []HistogramRecord `json:"histograms"`
}

type QueryResult struct {
	Read       *ReadResult       `json:"query,omitempty"`
	Statistics *StatisticsResult `json:"statistics,omitempty"`
	Histogram  *HistogramResult  `json:"histogram,omitempty"`
}

const (
	OperationQuery int = iota
	OperationStatistics
	OperationHistogram
)

func parseQueryResult(operation int, param QueryParameters, esResult *elastic.SearchResult) *QueryResult {
	var queryResult QueryResult

	switch operation {
	case OperationQuery:
		var readResult ReadResult
		readResult.Total = esResult.Hits.TotalHits
		readResult.From = param.From
		readResult.Size = param.Size
		for _, hit := range esResult.Hits.Hits {
			var logRecord LogRecord
			var source Source
			json.Unmarshal(*hit.Source, &source)
			logRecord.Time = source.Time
			logRecord.Log = source.Log
			var kubernetes Kubernetes
			json.Unmarshal(source.Kubernetes, &kubernetes)
			logRecord.Namespace = kubernetes.Namespace
			logRecord.Pod = kubernetes.Pod
			logRecord.Container = kubernetes.Container
			logRecord.Host = kubernetes.Host
			readResult.Records = append(readResult.Records, logRecord)
		}
		queryResult.Read = &readResult

	case OperationStatistics:
		var statisticsResult StatisticsResult
		statisticsResult.Total = esResult.Hits.TotalHits

		var namespaceAggregation NamespaceAggregation
		json.Unmarshal(*esResult.Aggregations["statistics"], &namespaceAggregation)
		for _, namespace := range namespaceAggregation.Namespaces {
			var namespaceResult NamespaceResult
			namespaceResult.Namespace = namespace.Namespace
			namespaceResult.Count = namespace.Count

			var containerAggregation ContainerAggregation
			json.Unmarshal(namespace.ContainerAggregation, &containerAggregation)
			for _, container := range containerAggregation.Containers {
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
		histogramResult.Total = esResult.Hits.TotalHits
		histogramResult.StartTime = param.StartTime
		histogramResult.EndTime = param.EndTime
		histogramResult.Interval = param.Interval

		var histogramAggregation HistogramAggregation
		json.Unmarshal(*esResult.Aggregations["histogram"], &histogramAggregation)
		for _, histogram := range histogramAggregation.Histograms {
			var histogramRecord HistogramRecord
			histogramRecord.Time = histogram.Time
			histogramRecord.Count = histogram.Count

			histogramResult.Histograms = append(histogramResult.Histograms, histogramRecord)
		}

		queryResult.Histogram = &histogramResult
	}

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

	Level     constants.LogQueryLevel
	Operation string
	LogQuery  string
	Interval  string
	StartTime string
	EndTime   string
	From      int
	Size      int
}

func Query(param QueryParameters) *QueryResult {
	// Starting with elastic.v5, you must pass a context to execute each service
	ctx := context.Background()

	// Obtain a client and connect to Elasticsearch
	client, err := elastic.NewClient(
		elastic.SetURL("http://elasticsearch-logging-data.kubesphere-logging-system.svc.cluster.local:9200"), elastic.SetSniff(false),
	)
	if err != nil {
		// Handle error
		// panic(err)
		return nil //Todo: Return error information
	}

	var boolQuery *elastic.BoolQuery = elastic.NewBoolQuery()

	if param.NamespaceFilled {
		var nsQuery *elastic.BoolQuery = elastic.NewBoolQuery()
		if len(param.Namespaces) == 0 {
			matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.namespace_name.key_word", "")
			nsQuery = nsQuery.Should(matchPhraseQuery)
		} else {
			for _, namespace := range param.Namespaces {
				matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.namespace_name.keyword", namespace)
				nsQuery = nsQuery.Should(matchPhraseQuery)
			}
		}
		nsQuery = nsQuery.MinimumNumberShouldMatch(1)
		boolQuery = boolQuery.Must(nsQuery)
	}
	if param.PodFilled {
		var podQuery *elastic.BoolQuery = elastic.NewBoolQuery()
		if len(param.Pods) == 0 {
			matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.pod_name.key_word", "")
			podQuery = podQuery.Should(matchPhraseQuery)
		} else {
			for _, pod := range param.Pods {
				matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.pod_name.keyword", pod)
				podQuery = podQuery.Should(matchPhraseQuery)
			}
		}
		podQuery = podQuery.MinimumNumberShouldMatch(1)
		boolQuery = boolQuery.Must(podQuery)
	}
	if param.ContainerFilled {
		var containerQuery *elastic.BoolQuery = elastic.NewBoolQuery()
		if len(param.Containers) == 0 {
			matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.container_name.key_word", "")
			containerQuery = containerQuery.Should(matchPhraseQuery)
		} else {
			for _, container := range param.Containers {
				matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.container_name.keyword", container)
				containerQuery = containerQuery.Should(matchPhraseQuery)
			}
		}
		containerQuery = containerQuery.MinimumNumberShouldMatch(1)
		boolQuery = boolQuery.Must(containerQuery)
	}

	if param.NamespaceQuery != "" {
		matchQuery := elastic.NewMatchQuery("kubernetes.namespace_name", param.NamespaceQuery)
		boolQuery = boolQuery.Must(matchQuery)
	}
	if param.PodQuery != "" {
		matchQuery := elastic.NewMatchQuery("kubernetes.pod_name", param.PodQuery)
		boolQuery = boolQuery.Must(matchQuery)
	}
	if param.ContainerQuery != "" {
		matchQuery := elastic.NewMatchQuery("kubernetes.container_name", param.ContainerQuery)
		boolQuery = boolQuery.Must(matchQuery)
	}

	if param.LogQuery != "" {
		matchQuery := elastic.NewMatchQuery("log", param.LogQuery)
		boolQuery = boolQuery.Must(matchQuery)
	}

	rangeQuery := elastic.NewRangeQuery("time").From(param.StartTime).To(param.EndTime)
	boolQuery = boolQuery.Must(rangeQuery)

	var queryResult *QueryResult
	var searchResult *elastic.SearchResult
	var searchError error

	if param.Operation == "statistics" {
		nsTermsAgg := elastic.NewTermsAggregation().Field("kubernetes.namespace_name.keyword").Size(2147483647)
		containerTermsAgg := elastic.NewTermsAggregation().Field("kubernetes.container_name.keyword").Size(2147483647)
		resultTermsAgg := nsTermsAgg.SubAggregation("Aggregate by container", containerTermsAgg)

		searchResult, searchError = client.Search().
			Index("logstash-*"). // search in index "logstash-*"
			Query(boolQuery).
			Aggregation("statistics", resultTermsAgg).
			Size(0). // take documents
			Do(ctx)  // execute

		queryResult = parseQueryResult(OperationStatistics, param, searchResult)

		if searchError != nil {
			// Handle error
			// panic(err)
			queryResult = nil //Todo: Add error information
		}
	} else if param.Operation == "histogram" {
		var interval string
		if param.Interval != "" {
			interval = param.Interval
		} else {
			interval = "15m"
		}
		param.Interval = interval
		dateAgg := elastic.NewDateHistogramAggregation().Field("time").Interval(interval)

		searchResult, searchError = client.Search().
			Index("logstash-*"). // search in index "logstash-*"
			Query(boolQuery).
			Aggregation("histogram", dateAgg).
			Size(0). // take documents
			Do(ctx)  // execute

		queryResult = parseQueryResult(OperationHistogram, param, searchResult)

		if searchError != nil {
			// Handle error
			// panic(err)
			queryResult = nil //Todo: Add error information
		}
	} else {
		searchResult, searchError = client.Search().
			Index("logstash-*"). // search in index "logstash-*"
			Query(boolQuery).
			Sort("time", true).                // sort by "time" field, ascending
			From(param.From).Size(param.Size). // take documents
			Do(ctx)                            // execute

		queryResult = parseQueryResult(OperationQuery, param, searchResult)

		if searchError != nil {
			// Handle error
			// panic(err)
			queryResult = nil //Todo: Add error information
		}
	}

	return queryResult
}
