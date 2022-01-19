package elastic

import (
//"encoding/json"
)

// Failure is a structure representing the Elasticsearch failure response
// e.g.:{"error":{"root_cause":[{"type":"no_shard_available_action_exception","reason":"No shard available for [org.elasticsearch.action.admin.indices.analyze.AnalyzeRequest@74508901]"}],"type":"no_shard_available_action_exception","reason":"No shard available for [org.elasticsearch.action.admin.indices.analyze.AnalyzeRequest@74508901]"},"status":503}
// e.g.:{"error":{"root_cause":[{"type":"index_already_exists_exception","reason":"already exists","index":"my_index"}],"type":"index_already_exists_exception","reason":"already exists","index":"my_index"},"status":400}
type Failure struct {
	Err    Error `json:"error"`
	Status int   `json:"status"`
}

// Error is a structure representing the Elasticsearch error response
type Error struct {
	RootCause    []Dict `json:"root_cause"`
	Type         string `json:"type"`
	Reason       string `json:"reason"`
	CausedBy     Dict   `json:"caused_by"`
	ResourceType string `json:"resource.type"`
	ResourceId   string `json:"resource.id"`
	// Index the name of index involved in this error incase of an Index API response
	Index string `json:"index"`
}

// Success is a structure representing an Elasticsearch success response
// e.g.: {"acknowledged":true}
type Success struct {
	Acknowledged bool `json:"acknowledged"`
}

// Unvalid is a structure representing an Elasticsearch unvalid response
// e.g.: {"valid":false,"_shards":{"total":1,"successful":1,"failed":0},"explanations":[{"index":"gb","valid":false,"error":"org.elasticsearch.index.query.QueryParsingException: No query registered for [tweet]"}]}
type Unvalid struct {
	Valid       bool   `json:"valid"`
	Shards      Dict   `json:"_shards"`
	Explanation []Dict `json:"explanations"`
}

/////////////////////////////////// Search Query

// Shard is a structure representing the Elasticsearch shard part of Search query response
type Shard struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
}

// Hits is a structure representing the Elasticsearch hits part of Search query response
type Hits struct {
	Total    int          `json:"total"`
	MaxScore interface{}  `json:"max_score"`
	Hits     []SearchHits `json:"hits"`
}

// SearchHits is a structure represennting the hitted document
type SearchHits struct {
	Index  string  `json:"_index"`
	Type   string  `json:"_type"`
	ID     string  `json:"_id"`
	Score  float32 `json:"_score"`
	Source Dict    `json:"_source"`
}

// ExplainResult Elasticsearch explain result
// e.g. {"valid":true,"_shards":{"total":1,"successful":1,"failed":0},"explanations":[{"index":"my_index","valid":true,"explanation":"+((name:b name:br name:bro name:brow name:brown) (name:f name:fo)) #ConstantScore(+ConstantScore(_type:my_type))"}]}
type ExplainResult struct {
	Valid        bool          `json:"valid"`
	Shards       Shard         `json:"_shards"`
	Explanations []Explanation `json:"explanations"`
}

// Explanation the details of explanation
type Explanation struct {
	Index       string `json:"index"`
	Valid       bool   `json:"valid"`
	Explanation string `json:"explanation"`
}

// SearchResult is a structure representing the Elastisearch search result
// e.g. {"took":1,"timed_out":false,"_shards":{"total":5,"successful":5,"failed":0},"hits":{"total":0,"max_score":null,"hits":[]}}
// e.g. {"took":3,"timed_out":false,"_shards":{"total":1,"successful":1,"failed":0},"hits":{"total":1,"max_score":0.50741017,"hits":[{"_index":"my_index","_type":"my_type","_id":"1","_score":0.50741017,"_source":{"name":"Brown foxes"}}]}}
type SearchResult struct {
	Took     int   `json:"took"`
	TimedOut bool  `json:"timed_out"`
	Shards   Shard `json:"_shards"`
	Hits     Hits  `json:"hits"`
}

/////////////////////////////////// Analyze Query

// AnalyzeResult is a structure representing the Elasticsearch analyze query result
// e.g. {"tokens":[{"token":"quick","start_offset":0,"end_offset":5,"type":"<ALPHANUM>","position":0},{"token":"brown","start_offset":6,"end_offset":11,"type":"<ALPHANUM>","position":1},{"token":"fox","start_offset":12,"end_offset":15,"type":"<ALPHANUM>","position":2}]}
type AnalyzeResult struct {
	Tokens []AnalyzeToken `json:"tokens"`
}

// AnalyzeToken is a structure representing part of the Elasticsearch analyze query response
type AnalyzeToken struct {
	Token       string `json:"token"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
	TokenType   string `json:"type"`
	Position    int    `json:"position"`
}

/////////////////////////////////// Insert Query

// InsertResult is a strucuture representing the Elasticsearch insert query result
// e.g. {"_index":"my_index","_type":"groups","_id":"1","_version":1,"_shards":{"total":2,"successful":1,"failed":0},"created":true}
type InsertResult struct {
	Index   string `json:"_index"`
	Doctype string `json:"_type"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Shards  Shard  `json:"_shards"`
	Created bool   `json:"created"`
	//Status  int    `json:"status"`
}

/////////////////////////////////// Bulk Query

// BulkResult is a structure representing the Elasticsearch bulk query result
// e.g. {"took":118,"errors":false,"items":[{"index":{"_index":"my_index","_type":"my_type","_id":"1","_version":1,"_shards":{"total":2,"successful":1,"failed":0},"status":201}},{"index":{"_index":"my_index","_type":"my_type","_id":"2","_version":1,"_shards":{"total":2,"successful":1,"failed":0},"status":201}}]}
type BulkResult struct {
	Took   int            `josn:"took"`
	Errors bool           `json:"errors"`
	Items  []InsertResult `json:"items"`
}

/////////////////////////////////// Aggregation Query
// AggregationResult is a structure representing the Elasticsearch aggregation query result
// e.g. {"took":4,"timed_out":false,"_shards":{"total":5,"successful":5,"failed":0},"hits":{"total":7,"max_score":0.0,"hits":[]},"aggregations":{"colors":{"doc_count_error_upper_bound":0,"sum_other_doc_count":0,"buckets":[{"key":"blue","doc_count":1,"avg_price":{"value":15000.0}},{"key":"green","doc_count":2,"avg_price":{"value":21000.0}},{"key":"red","doc_count":4,"avg_price":{"value":32500.0}}]}}}
// e.g. {"took":3,"timed_out":false,"_shards":{"total":5,"successful":5,"failed":0},"hits":{"total":7,"max_score":0.0,"hits":[]},"aggregations":{"distinct_colors":{"value":3}}}
type AggregationResult struct {
	SearchResult
	Aggregations map[string]AggSubResult `json:"aggregations"`
}

// AggSubResult is a structure representing a sub result of the aggregation query result
type AggSubResult struct {
	DocCountErrorUpperBound int    `json:"doc_count_error_upper_bound"`
	SumOtherDocCount        int    `json:"sum_other_doc_count"`
	Buckets                 []Dict `json:"buckets"`
	//Buckets                 []BucketResult `json:"buckets"`
}

type BucketResult struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
	Dict     `json`
}
