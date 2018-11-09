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

package log

import (
	"context"
	//"encoding/json"
	//"regexp"
	//"strings"

	"github.com/emicklei/go-restful"
	//"github.com/golang/glog"

	//"time"

	//"k8s.io/api/core/v1"
	//metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//"kubesphere.io/kubesphere/pkg/client"
	//"kubesphere.io/kubesphere/pkg/models"
	"github.com/olivere/elastic"
)

func LogQuery(request *restful.Request) *elastic.SearchResult {
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

	matchPhraseQuery := elastic.NewMatchPhraseQuery("kubernetes.namespace_name.keyword", "kubesphere-system")
	matchQuery := elastic.NewMatchQuery("log", "认证 校验")
	rangeQuery := elastic.NewRangeQuery("time").From("2018-11-09T00:00:00.000").To("2018-11-09T12:00:00.000")
	boolQuery := elastic.NewBoolQuery().Must(matchPhraseQuery).Must(matchQuery).Must(rangeQuery)
	searchResult, err := client.Search().
		Index("logstash-*").   // search in index "logstash-*"
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