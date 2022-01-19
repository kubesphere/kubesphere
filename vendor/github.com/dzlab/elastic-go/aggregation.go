package elastic

import ()

const (
	// Aggs abreviateed constant name for the Aggregation query.
	Aggs = "aggs"
	// Aggregations constant name for the Aggregation query.
	Aggregations = "aggregations"
	// Terms constant name of terms Bucket
	Terms = "terms"
	// Histogram constant name of the Histogram bucket
	Histogram = "histogram"
	// DateHistogram constant name of the Date Histogram bucket.
	DateHistogram = "date_histogram"
	// Global constant name of the global bucket which is used to by pass aggregation scope.
	Global = "global"
	// FilterBucket constant name of filter bucket which is used to filter aggregation results.
	FilterBucket = "filter"
)

// Constant name of Elasticsearch metrics
const (
	// Count constant name of 'count' metric.
	Count = "count"
	// Sum constant name of 'sum' metric.
	Sum = "sum"
	// Avg constant name of 'avg' metric.
	Avg = "avg"
	// Min constant name of 'min' metric.
	Min = "min"
	// Max constant name of 'max' metric.
	Max = "max"
	// ExtendedStats constant name of a metric that will return a variety of statistics (e.g. stats.avg, stats.count, stats.std_deviation).
	ExtendedStats = "extended_stats"
	// Cardiality constant name of the 'cardinality' approximation metric.
	Cardiality = "cardinality"
	// Percentiles constant name of the 'percentiles' approximation metric.
	Percentiles = "percentiles"
	// PercentileRank constant name of an approximation metric that tells to which percentile the given value belongs.
	PercentileRanks = "percentile_ranks"
	// SignificantTerms constant name of the statistical anomalie aggregation. By default, it will use the entire index as the background group while the foreground will be aggregation query scope.
	SignificantTerms = "significant_terms"
)

const (
	// Field name of parameter that defines the document's field that will be used to create buckets using its unique values.
	Field = "field"
	// Interval name of parameter that define a histogram interval, i.e. the value that Elasticsearch will use to create new buckets.
	Interval = "interval"
	// Size name of parameter that defines how many terms we want to generate. Example of values, for histograms: 10, for date histograms: "month", "quarter".
	Size = "size"
	// Format name of parameter in date histogram, used to define the  dates format for bucket keys.
	Format = "format"
	// MinDocCount name of parameter in date histogram, used to force empty buckets to be returned.
	MinDocCount = "min_doc_count"
	// ExtendedBound name of parameter in date histogram. It is used to extend the boudaries of bucket from the boudaries of actual data. This, it forces all bucket betwen the min and max bound to be returned.
	ExtendedBound = "extended_bound"
	// Order name of an object that defines how the create buckets should be generatedas well as the the ordering mode (e.g. asc). Example of values: _count (sort by document count), _term (sort alphabetically by string value), _key (sort by bucket key, works only for histogram & date_histogram).
	Order = "order"
	// PrecisionThreshold configure the precision of the HyperLogLog algorithm used by the 'cardinality' metric.
	PrecisionThreshold = "precision_threshold"
	// Percents a parameter of the 'percentiles' metric. It's used to define an array of the percentiles that should be calculated instead of the default one (i.e. 5, 25, 50, 75, 95, 99).
	Percents = "percents"
	// Values a parameter of the 'percentile_ranks' metric. It is used to define the values that Elasticsearch should find their percentile.
	Values = "values"
	// Compression a parameter of the 'percentiles' metric (default value is 100). It is used to control the memory footprint (an thus the accuracy) by limiting the number of nodes involved in the calculation.
	Compression = "compression"
)

// Aggregations a structure representing an aggregation request
type Aggregation struct {
	client *Elasticsearch
	parser Parser
	url    string
	params map[string]string
	query  Dict
}

// Aggs creates an aggregation request
func (client *Elasticsearch) Aggs(index, doc string) *Aggregation {
	url := client.request(index, doc, -1, SEARCH)
	return &Aggregation{
		client: client,
		parser: &AggregationResultParser{},
		url:    url,
		params: make(map[string]string),
		query:  make(Dict),
	}
}

// urlString constructs the url of this Search API call
func (agg *Aggregation) urlString() string {
	return urlString(agg.url, agg.params)
}

// String returns a string representation of this Search API call
func (agg *Aggregation) String() string {
	body := ""
	if len(agg.query) > 0 {
		body = String(agg.query)
	}
	return body
}

// Get submits request mappings between the json fields and how Elasticsearch store them
// GET /:index/:type/_search
func (agg *Aggregation) Get() {
	// construct the url
	url := agg.urlString()
	// construct the body
	query := agg.String()

	agg.client.Execute("GET", url, query, agg.parser)
}

// SetMetric sets the search type with the given value (e.g. count)
func (agg *Aggregation) SetMetric(name string) *Aggregation {
	agg.params[SearchType] = name
	return agg
}

// Bucket a structure that defines how Elasticsearch should create Bucket for aggregations.
type Bucket struct {
	name  string
	query Dict
}

// Metric a structure that defines a bucket metric.
type Metric struct {
	name  string
	query Dict
}

// NewBucket creates a new Bucket definition
func NewBucket(name string) *Bucket {
	return &Bucket{
		name:  name,
		query: make(Dict),
	}
}

func (bucket *Bucket) AddTerm(name string, value interface{}) *Bucket {
	bucket.AddMetric(Terms, name, value)
	return bucket
}

func (bucket *Bucket) AddMetric(metric, name string, value interface{}) *Bucket {
	if bucket.query[metric] == nil {
		bucket.query[metric] = make(Dict)
	}
	bucket.query[metric].(Dict)[name] = value
	return bucket
}

func (bucket *Bucket) AddDict(name string, value Dict) *Bucket {
	bucket.query[name] = value
	return bucket
}

// SetOrder set the ordering for this bucket.
// name is the name of ordering, e.g. _count, _term, _key, name of metric
// value defines the sens of ordering, e.g. asc
func (bucket *Bucket) SetOrder(metric, name, value string) *Bucket {
	bucket.AddMetric(metric, Order, Dict{name: value})
	return bucket
}

// AddBucket adds a nested bucket to this bucket
func (bucket *Bucket) AddBucket(b *Bucket) *Bucket {
	if bucket.query[Aggs] == nil {
		bucket.query[Aggs] = make(Dict)
	}
	bucket.query[Aggs].(Dict)[b.name] = b.query
	return bucket
}

// Add adds a bucket definition to this aggregation request
func (agg *Aggregation) Add(bucket *Bucket) *Aggregation {
	if agg.query[Aggs] == nil {
		agg.query[Aggs] = make(Dict)
	}
	aggs := agg.query[Aggs].(Dict)
	aggs[bucket.name] = bucket.query
	return agg
}

// AddQuery defines a scope query for this aggregation request
func (agg *Aggregation) AddQuery(q Query) *Aggregation {
	if agg.query["query"] == nil {
		agg.query["query"] = make(Dict)
	}
	agg.query["query"].(Dict)[q.Name()] = q.KV()
	return agg
}

func (agg *Aggregation) AddPostFilter(q Query) *Aggregation {
	if agg.query[PostFilter] == nil {
		agg.query[PostFilter] = make(Dict)
	}
	agg.query[PostFilter].(Dict)[q.Name()] = q.KV()
	return agg
}
