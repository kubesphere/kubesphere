package elasticsearch

import "time"

// --------------------------------------------- Request Body ---------------------------------------------

// More info: https://www.elastic.co/guide/en/elasticsearch/reference/current/getting-started-search-API.html
type Body struct {
	From   int64               `json:"from,omitempty"`
	Size   int64               `json:"size,omitempty"`
	Sorts  []map[string]string `json:"sort,omitempty"`
	*Query `json:"query,omitempty"`
	*Aggs  `json:"aggs,omitempty"`
}

type Query struct {
	Bool `json:"bool,omitempty"`
}

// Example:
// {bool: {filter: <[]Match>}}
// {bool: {should: <[]Match>, minimum_should_match: 1}}
type Bool struct {
	Filter             []Match `json:"filter,omitempty"`
	Should             []Match `json:"should,omitempty"`
	MinimumShouldMatch int32   `json:"minimum_should_match,omitempty"`
}

// Example: []Match
// [
//   {
//     bool: <Bool>
//   },
//   {
//     match_phrase: {
//       <string>: <string>
//     }
//   },
//   ...
// ]
type Match struct {
	*Bool             `json:"bool,omitempty"`
	MatchPhrase       map[string]string `json:"match_phrase,omitempty"`
	MatchPhrasePrefix map[string]string `json:"match_phrase_prefix,omitempty"`
	Regexp            map[string]string `json:"regexp,omitempty"`
	*Range            `json:"range,omitempty"`
}

type Range struct {
	*Time `json:"time,omitempty"`
}

type Time struct {
	Gte *time.Time `json:"gte,omitempty"`
	Lte *time.Time `json:"lte,omitempty"`
}

type Aggs struct {
	*CardinalityAggregation   `json:"container_count,omitempty"`
	*DateHistogramAggregation `json:"log_count_over_time,omitempty"`
}

type CardinalityAggregation struct {
	*Cardinality `json:"cardinality,omitempty"`
}

type Cardinality struct {
	Field string `json:"field,omitempty"`
}

type DateHistogramAggregation struct {
	*DateHistogram `json:"date_histogram,omitempty"`
}

type DateHistogram struct {
	Field    string `json:"field,omitempty"`
	Interval string `json:"interval,omitempty"`
}

// --------------------------------------------- Response Body ---------------------------------------------

type Response struct {
	ScrollId     string `json:"_scroll_id,omitempty"`
	Hits         `json:"hits,omitempty"`
	Aggregations `json:"aggregations,omitempty"`
}

type Hits struct {
	Total   interface{} `json:"total"` // `As of Elasticsearch v7.x, hits.total is changed incompatibly
	AllHits []Hit       `json:"hits"`
}

type Hit struct {
	Source `json:"_source"`
	Sort   []int64 `json:"sort"`
}

type Source struct {
	Log        string `json:"log"`
	Time       string `json:"time"`
	Kubernetes `json:"kubernetes"`
}

type Kubernetes struct {
	Namespace string `json:"namespace_name"`
	Pod       string `json:"pod_name"`
	Container string `json:"container_name"`
	Host      string `json:"host"`
}

type Aggregations struct {
	ContainerCount   `json:"container_count"`
	LogCountOverTime `json:"log_count_over_time"`
}

type ContainerCount struct {
	Value int64 `json:"value"`
}

type LogCountOverTime struct {
	Buckets []Bucket `json:"buckets"`
}

type Bucket struct {
	Time  int64 `json:"key"`
	Count int64 `json:"doc_count"`
}
