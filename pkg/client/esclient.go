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

	"github.com/olivere/elastic"
)

type QueryParameters struct {
	Workspaces []string
	Projects []string
	Workloads []string
	Pods []string
	Containers []string

	Workspaces_query string
	Projects_query string
	Workloads_query string
	Pods_query string
	Containers_query string

	Level constants.LogQueryLevel
	Operation string
	Log_query string
	Start_time string
	End_time string
	From int
	Size int
}

func Query(param QueryParameters) *elastic.SearchResult {
	// Starting with elastic.v5, you must pass a context to execute each service
	ctx := context.Background()

	// Obtain a client and connect to Elasticsearch
	client, err := elastic.NewClient(
		elastic.SetURL("http://elasticsearch-logging-data.kubesphere-logging-system.svc.cluster.local:9200"), elastic.SetSniff(false),
	)
	if err != nil {
		// Handle error
		// panic(err)
		return nil //Todo: Add error information
	}

	var boolQuery *elastic.BoolQuery = elastic.NewBoolQuery()
	var hasShould bool = false

	//Todo: Get information from k8s
	switch param.Level {
	case constants.QueryLevelCluster:
		{
			if param.Projects_query != "" {
				matchQuery := elastic.NewMatchQuery("kubernetes.namespace_name", param.Projects_query)
				boolQuery = boolQuery.Must(matchQuery)
			}
			if param.Pods_query != "" {
				matchQuery := elastic.NewMatchQuery("kubernetes.pod_name", param.Pods_query)
				boolQuery = boolQuery.Must(matchQuery)
			}
			if param.Containers_query != "" {
				matchQuery := elastic.NewMatchQuery("kubernetes.container_name", param.Containers_query)
				boolQuery = boolQuery.Must(matchQuery)
			}
		}
	case constants.QueryLevelWorkspace:
		{
			for _, workspace := range param.Workspaces {
				workspace = workspace
			}
			if param.Projects_query != "" {
				matchQuery := elastic.NewMatchQuery("kubernetes.namespace_name", param.Projects_query)
				boolQuery = boolQuery.Must(matchQuery)
			}
			if param.Pods_query != "" {
				matchQuery := elastic.NewMatchQuery("kubernetes.pod_name", param.Pods_query)
				boolQuery = boolQuery.Must(matchQuery)
			}
			if param.Containers_query != "" {
				matchQuery := elastic.NewMatchQuery("kubernetes.container_name", param.Containers_query)
				boolQuery = boolQuery.Must(matchQuery)
			}
		}
	case constants.QueryLevelProject:
		{
			for _, workspace := range param.Workspaces {
				workspace = workspace
			}
			for _, project := range param.Projects {
				matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.namespace_name.keyword", project)
				boolQuery = boolQuery.Should(matchPhraseQuery)
				hasShould = true
			}
			if param.Pods_query != "" {
				matchQuery := elastic.NewMatchQuery("kubernetes.pod_name", param.Pods_query)
				boolQuery = boolQuery.Must(matchQuery)
			}
			if param.Containers_query != "" {
				matchQuery := elastic.NewMatchQuery("kubernetes.container_name", param.Containers_query)
				boolQuery = boolQuery.Must(matchQuery)
			}
		}
	case constants.QueryLevelWorkload:
		{
			for _, workspace := range param.Workspaces {
				workspace = workspace
			}
			for _, project := range param.Projects {
				matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.namespace_name.keyword", project)
				boolQuery = boolQuery.Should(matchPhraseQuery)
				hasShould = true
			}
			for _, workload := range param.Workloads {
				workload = workload
			}
			if param.Pods_query != "" {
				matchQuery := elastic.NewMatchQuery("kubernetes.pod_name", param.Pods_query)
				boolQuery = boolQuery.Must(matchQuery)
			}
			if param.Containers_query != "" {
				matchQuery := elastic.NewMatchQuery("kubernetes.container_name", param.Containers_query)
				boolQuery = boolQuery.Must(matchQuery)
			}
		}
	case constants.QueryLevelPod:
		{
			for _, workspace := range param.Workspaces {
				workspace = workspace
			}
			for _, project := range param.Projects {
				matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.namespace_name.keyword", project)
				boolQuery = boolQuery.Should(matchPhraseQuery)
				hasShould = true
			}
			for _, workload := range param.Workloads {
				workload = workload
			}
			for _, pod := range param.Pods {
				matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.pod_name.key_word", pod)
				boolQuery = boolQuery.Should(matchPhraseQuery)
				hasShould = true
			}
			if param.Containers_query != "" {
				matchQuery := elastic.NewMatchQuery("kubernetes.container_name", param.Containers_query)
				boolQuery = boolQuery.Must(matchQuery)
			}
		}
	case constants.QueryLevelContainer:
		{
			for _, workspace := range param.Workspaces {
				workspace = workspace
			}
			for _, project := range param.Projects {
				matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.namespace_name.keyword", project)
				boolQuery = boolQuery.Should(matchPhraseQuery)
				hasShould = true
			}
			for _, workload := range param.Workloads {
				workload = workload
			}
			for _, pod := range param.Pods {
				matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.pod_name.key_word", pod)
				boolQuery = boolQuery.Should(matchPhraseQuery)
				hasShould = true
			}
			for _, container := range param.Containers {
				matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.container_name.keyword", container)
				boolQuery = boolQuery.Should(matchPhraseQuery)
				hasShould = true
			}
		}
	}

	if(hasShould) {
		boolQuery = boolQuery.MinimumNumberShouldMatch(1)
	}

	matchQuery := elastic.NewMatchQuery("log", param.Log_query)
	rangeQuery := elastic.NewRangeQuery("time").From(param.Start_time).To(param.End_time)
	boolQuery = boolQuery.Must(matchQuery).Must(rangeQuery)
	searchResult, err := client.Search().
		Index("logstash-*"). // search in index "logstash-*"
		Query(boolQuery).
		Sort("time", true). // sort by "time" field, ascending
		From(param.From).Size(param.Size).   // take documents
		Pretty(true).       // pretty print request and response JSON
		Do(ctx)             // execute
	if err != nil {
		// Handle error
		// panic(err)
		searchResult = nil //Todo: Add error information
	}

	return searchResult
}
