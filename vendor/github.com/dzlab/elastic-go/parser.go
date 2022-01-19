package elastic

import (
	"encoding/json"
	"errors"
	"log"
)

// Parser an interface for parsing reponses
type Parser interface {
	Parse(data []byte) (interface{}, error)
}

// SuccessParser parses Success responses
type SuccessParser struct{}

// Parse rerturns a parsed Success result structure from the given data
func (parser *SuccessParser) Parse(data []byte) (interface{}, error) {
	success := Success{}
	if err := json.Unmarshal(data, &success); err == nil {
		log.Println("success", success)
		return success, nil
	}
	log.Println("Failed to parse response", string(data))
	return nil, errors.New("Failed to parse response")
}

// FailureParser a parser for search result
type FailureParser struct{}

// Parse rerturns a parsed Failure result structure from the given data
func (parser *FailureParser) Parse(data []byte) (interface{}, error) {
	failure := Failure{}
	if err := json.Unmarshal(data, &failure); err == nil {
		log.Println("failed", failure)
		return failure, nil
	}
	log.Println("Failed to parse response", string(data))
	return nil, errors.New("Failed to parse response")
}

// SearchResultParser a parser for search result
type SearchResultParser struct{}

// Parse rerturns a parsed search result structure from the given data
func (parser *SearchResultParser) Parse(data []byte) (interface{}, error) {
	search := SearchResult{}
	if err := json.Unmarshal(data, &search); err == nil && !deepEqual(search, *new(SearchResult)) {
		log.Println("search", search)
		return search, nil
	}
	next1 := &SuccessParser{}
	if success, err := next1.Parse(data); err == nil && success != *new(Success) {
		return success, nil
	}
	next2 := &FailureParser{}
	if failure, err := next2.Parse(data); err == nil && !deepEqual(failure, *new(Failure)) {
		return failure, nil
	}
	log.Println("Failed to parse response", string(data))
	return nil, errors.New("Failed to parse response")
}

// IndexResultParser a parser for index result
type IndexResultParser struct{}

// Parse returns an index result structure from the given data
func (parser *IndexResultParser) Parse(data []byte) (interface{}, error) {
	next1 := SuccessParser{}
	if success, err := next1.Parse(data); err == nil && success != *new(Success) {
		return success, nil
	}
	next2 := &FailureParser{}
	if failure, err := next2.Parse(data); err == nil && !deepEqual(failure, *new(Failure)) {
		return failure, nil
	}
	log.Println("Failed to parse response", string(data))
	return nil, errors.New("Failed to parse response")
}

// MappingResultParser a parser for mapping result
type MappingResultParser struct{}

// Parse returns an index result structure from the given data
func (parser *MappingResultParser) Parse(data []byte) (interface{}, error) {
	var result interface{}
	/*index := IndexResult{}
	if err := json.Unmarshal(data, &index); err == nil {
		log.Println("index", string(data), index)
	} else {
		log.Println("Failed to parse response", string(data))
	}*/
	log.Println(string(data))
	return result, nil
}

// InsertResultParser a parser for mapping result
type InsertResultParser struct{}

// Parse returns an index result structure from the given data
func (parser *InsertResultParser) Parse(data []byte) (interface{}, error) {
	insert := InsertResult{}
	if err := json.Unmarshal(data, &insert); err == nil && insert != *new(InsertResult) {
		return insert, nil
	}
	next1 := SuccessParser{}
	if success, err := next1.Parse(data); err == nil && success != *new(Success) {
		return success, nil
	}
	next2 := &FailureParser{}
	if failure, err := next2.Parse(data); err == nil && !deepEqual(failure, *new(Failure)) {
		return failure, nil
	}
	log.Println("Failed to parse response", string(data))
	return nil, errors.New("Failed to parse response")
}

// AnalyzeResultParser a parser for analyze result
type AnalyzeResultParser struct{}

// Parse returns an analyze result structure from the given data
func (parser *AnalyzeResultParser) Parse(data []byte) (interface{}, error) {
	analyze := AnalyzeResult{}
	if err := json.Unmarshal(data, &analyze); err == nil && !deepEqual(analyze, *new(AnalyzeResult)) {
		log.Println("analyze", analyze)
		return analyze, nil
	}
	next1 := SuccessParser{}
	if success, err := next1.Parse(data); err == nil && success != *new(Success) {
		return success, nil
	}
	next2 := &FailureParser{}
	if failure, err := next2.Parse(data); err == nil && !deepEqual(failure, *new(Failure)) {
		return failure, nil
	}
	log.Println("Failed to parse response", string(data))
	return nil, errors.New("Failed to parse response")
}

// BulkResultParser a parser for analyze result
type BulkResultParser struct{}

// Parse returns an analyze result structure from the given data
func (parser *BulkResultParser) Parse(data []byte) (interface{}, error) {
	bulk := BulkResult{}
	if err := json.Unmarshal(data, &bulk); err == nil && !deepEqual(bulk, *new(BulkResult)) {
		log.Println("bulk", bulk)
		return bulk, nil
	}
	next := SuccessParser{}
	if success, err := next.Parse(data); err == nil && success != *new(Success) {
		return success, nil
	}
	log.Println("Failed to parse response", string(data))
	return nil, errors.New("Failed to parse response")
}

// AggregationResultParser a parser for aggregation result
type AggregationResultParser struct{}

// Parse returns an index result structure from the given data
func (parser *AggregationResultParser) Parse(data []byte) (interface{}, error) {
	var result interface{}
	agg := AggregationResult{}
	if err := json.Unmarshal(data, &agg); err == nil && !deepEqual(agg, *new(AggregationResult)) {
		log.Println("aggregation", agg)
		return agg, nil
	}
	next2 := &FailureParser{}
	if failure, err := next2.Parse(data); err == nil && !deepEqual(failure, *new(Failure)) {
		return failure, nil
	}
	log.Println(string(data))
	return result, nil
}
