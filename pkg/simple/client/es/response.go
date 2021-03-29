package es

import (
	jsoniter "github.com/json-iterator/go"
	"k8s.io/klog"
)

type Response struct {
	ScrollId     string `json:"_scroll_id,omitempty"`
	Hits         `json:"hits,omitempty"`
	Aggregations `json:"aggregations,omitempty"`
}

type Hits struct {
	Total   interface{} `json:"total,omitempty"` // `As of Elasticsearch v7.x, hits.total is changed incompatibly
	AllHits []Hit       `json:"hits,omitempty"`
}

type Hit struct {
	Source interface{} `json:"_source,omitempty"`
	Sort   []int64     `json:"sort,omitempty"`
}

type Aggregations struct {
	CardinalityAggregation   `json:"cardinality_aggregation,omitempty"`
	DateHistogramAggregation `json:"date_histogram_aggregation,omitempty"`
}

type CardinalityAggregation struct {
	Value int64 `json:"value,omitempty"`
}

type DateHistogramAggregation struct {
	Buckets []Bucket `json:"buckets,omitempty"`
}

type Bucket struct {
	Key   int64 `json:"key,omitempty"`
	Count int64 `json:"doc_count,omitempty"`
}

func parseResponse(body []byte) (*Response, error) {
	var res Response
	err := jsoniter.Unmarshal(body, &res)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &res, nil
}
