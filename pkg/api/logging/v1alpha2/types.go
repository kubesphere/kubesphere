package v1alpha2

import (
	"encoding/json"
	"time"
)

const (
	OperationQuery int = iota
	OperationStatistics
	OperationHistogram
	OperationExport
)

// elasticsearch client config
type Config struct {
	Host         string
	Port         string
	Index        string
	VersionMajor string
}

type QueryParameters struct {
	// when true, indicates the provided `namespaces` or `namespace_query` doesn't match any namespace
	NamespaceNotFound bool
	// a map of namespace with creation time
	NamespaceWithCreationTime map[string]string

	// filter for literally matching
	// query for fuzzy matching
	WorkloadFilter  []string
	WorkloadQuery   []string
	PodFilter       []string
	PodQuery        []string
	ContainerFilter []string
	ContainerQuery  []string
	LogQuery        []string

	Operation     int
	Interval      string
	StartTime     string
	EndTime       string
	Sort          string
	From          int64
	Size          int64
	ScrollTimeout time.Duration
}

// elasticsearch request body
type Request struct {
	From      int64       `json:"from"`
	Size      int64       `json:"size"`
	Sorts     []Sort      `json:"sort,omitempty"`
	MainQuery BoolQuery   `json:"query"`
	Aggs      interface{} `json:"aggs,omitempty"`
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

// user filter instead of must
// filter ignores scoring
type BoolFilter struct {
	Filter []interface{} `json:"filter"`
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

type RegexpQuery struct {
	Regexp interface{} `json:"regexp"`
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

// Fore more info, refer to https://www.elastic.co/guide/en/elasticsearch/reference/current/getting-started-search-API.html
// Response body from the elasticsearch engine
type Response struct {
	ScrollId     string          `json:"_scroll_id"`
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
	Source Source  `json:"_source"`
	Sort   []int64 `json:"sort"`
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

type LogRecord struct {
	Time      string `json:"time,omitempty" description:"log timestamp"`
	Log       string `json:"log,omitempty" description:"log message"`
	Namespace string `json:"namespace,omitempty" description:"namespace"`
	Pod       string `json:"pod,omitempty" description:"pod name"`
	Container string `json:"container,omitempty" description:"container name"`
	Host      string `json:"host,omitempty" description:"node id"`
}

type ReadResult struct {
	ScrollID string      `json:"_scroll_id,omitempty"`
	Total    int64       `json:"total" description:"total number of matched results"`
	Records  []LogRecord `json:"records,omitempty" description:"actual array of results"`
}

// StatisticsResponseAggregations, the struct for `aggregations` of type Response, holds return results from the aggregation StatisticsAggs
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
	Histograms []HistogramRecord `json:"histograms" description:"actual array of histogram results"`
}

// Wrap elasticsearch response
type QueryResult struct {
	Read       *ReadResult       `json:"query,omitempty" description:"query results"`
	Statistics *StatisticsResult `json:"statistics,omitempty" description:"statistics results"`
	Histogram  *HistogramResult  `json:"histogram,omitempty" description:"histogram results"`
}
