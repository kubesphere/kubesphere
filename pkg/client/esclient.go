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

func Query(log_query string, start string, end string) *elastic.SearchResult {
	// Starting with elastic.v5, you must pass a context to execute each service
	ctx := context.Background()

	// Obtain a client and connect to the default Elasticsearch installation
	// on 127.0.0.1:9200. Of course you can configure your client to connect
	// to other hosts and configure it in various other ways.
	client, err := elastic.NewClient(
		elastic.SetURL("http://ks-logging-elasticsearch-data.logging.svc:9200"),
	)
	if err != nil {
		// Handle error
		panic(err)
	}

	matchPhraseQuery1 := elastic.NewMatchPhraseQuery("kubernetes.namespace_name.keyword", "kubesphere-system")
	matchPhraseQuery2 := elastic.NewMatchPhraseQuery("kubernetes.namespace_name.keyword", "logging")
	matchQuery := elastic.NewMatchQuery("log", log_query)
	rangeQuery := elastic.NewRangeQuery("time").From(start).To(end)
	boolQuery := elastic.NewBoolQuery().Must(matchQuery).Must(rangeQuery).Should(matchPhraseQuery1).Should(matchPhraseQuery2).MinimumNumberShouldMatch(1)
	searchResult, err := client.Search().
		Index("logstash-*"). // search in index "logstash-*"
		Query(boolQuery).
		Sort("time", true). // sort by "time" field, ascending
		From(0).Size(10).   // take documents 0-9
		Pretty(true).       // pretty print request and response JSON
		Do(ctx)             // execute
	if err != nil {
		// Handle error
		panic(err)
	}

	return searchResult
}
