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

	"kubesphere.io/kubesphere/pkg/constants"
)

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
	StartTime string
	EndTime   string
	From      int
	Size      int
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
		return nil //Todo: Return error information
	}

	var boolQuery *elastic.BoolQuery = elastic.NewBoolQuery()
	var hasShould bool = false

	if param.NamespaceFilled {
		for _, namespace := range param.Namespaces {
			matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.namespace_name.keyword", namespace)
			boolQuery = boolQuery.Should(matchPhraseQuery)
		}
		hasShould = true
	}
	if param.PodFilled {
		for _, pod := range param.Pods {
			matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.pod_name.key_word", pod)
			boolQuery = boolQuery.Should(matchPhraseQuery)
		}
		hasShould = true
	}
	if param.ContainerFilled {
		for _, container := range param.Containers {
			matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.container_name.keyword", container)
			boolQuery = boolQuery.Should(matchPhraseQuery)
		}
		hasShould = true
	}

	if hasShould {
		boolQuery = boolQuery.MinimumNumberShouldMatch(1)
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
	searchResult, err := client.Search().
		Index("logstash-*"). // search in index "logstash-*"
		Query(boolQuery).
		Sort("time", true).                // sort by "time" field, ascending
		From(param.From).Size(param.Size). // take documents
		Pretty(true).                      // pretty print request and response JSON
		Do(ctx)                            // execute
	if err != nil {
		// Handle error
		// panic(err)
		searchResult = nil //Todo: Add error information
	}

	return searchResult
}
