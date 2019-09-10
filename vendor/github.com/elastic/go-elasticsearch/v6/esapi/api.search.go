// Code generated from specification version 6.8.2: DO NOT EDIT

package esapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func newSearchFunc(t Transport) Search {
	return func(o ...func(*SearchRequest)) (*Response, error) {
		var r = SearchRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// Search returns results matching a query.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/search-search.html.
//
type Search func(o ...func(*SearchRequest)) (*Response, error)

// SearchRequest configures the Search API request.
//
type SearchRequest struct {
	Index        []string
	DocumentType []string

	Body io.Reader

	AllowNoIndices             *bool
	AllowPartialSearchResults  *bool
	Analyzer                   string
	AnalyzeWildcard            *bool
	BatchedReduceSize          *int
	DefaultOperator            string
	Df                         string
	DocvalueFields             []string
	ExpandWildcards            string
	Explain                    *bool
	From                       *int
	IgnoreThrottled            *bool
	IgnoreUnavailable          *bool
	Lenient                    *bool
	MaxConcurrentShardRequests *int
	Preference                 string
	PreFilterShardSize         *int
	Query                      string
	RequestCache               *bool
	RestTotalHitsAsInt         *bool
	Routing                    []string
	Scroll                     time.Duration
	SearchType                 string
	SeqNoPrimaryTerm           *bool
	Size                       *int
	Sort                       []string
	Source                     []string
	SourceExcludes             []string
	SourceIncludes             []string
	Stats                      []string
	StoredFields               []string
	SuggestField               string
	SuggestMode                string
	SuggestSize                *int
	SuggestText                string
	TerminateAfter             *int
	Timeout                    time.Duration
	TrackScores                *bool
	TrackTotalHits             interface{}
	TypedKeys                  *bool
	Version                    *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r SearchRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len(strings.Join(r.DocumentType, ",")) + 1 + len("_search"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	if len(r.DocumentType) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.DocumentType, ","))
	}
	path.WriteString("/")
	path.WriteString("_search")

	params = make(map[string]string)

	if r.AllowNoIndices != nil {
		params["allow_no_indices"] = strconv.FormatBool(*r.AllowNoIndices)
	}

	if r.AllowPartialSearchResults != nil {
		params["allow_partial_search_results"] = strconv.FormatBool(*r.AllowPartialSearchResults)
	}

	if r.Analyzer != "" {
		params["analyzer"] = r.Analyzer
	}

	if r.AnalyzeWildcard != nil {
		params["analyze_wildcard"] = strconv.FormatBool(*r.AnalyzeWildcard)
	}

	if r.BatchedReduceSize != nil {
		params["batched_reduce_size"] = strconv.FormatInt(int64(*r.BatchedReduceSize), 10)
	}

	if r.DefaultOperator != "" {
		params["default_operator"] = r.DefaultOperator
	}

	if r.Df != "" {
		params["df"] = r.Df
	}

	if len(r.DocvalueFields) > 0 {
		params["docvalue_fields"] = strings.Join(r.DocvalueFields, ",")
	}

	if r.ExpandWildcards != "" {
		params["expand_wildcards"] = r.ExpandWildcards
	}

	if r.Explain != nil {
		params["explain"] = strconv.FormatBool(*r.Explain)
	}

	if r.From != nil {
		params["from"] = strconv.FormatInt(int64(*r.From), 10)
	}

	if r.IgnoreThrottled != nil {
		params["ignore_throttled"] = strconv.FormatBool(*r.IgnoreThrottled)
	}

	if r.IgnoreUnavailable != nil {
		params["ignore_unavailable"] = strconv.FormatBool(*r.IgnoreUnavailable)
	}

	if r.Lenient != nil {
		params["lenient"] = strconv.FormatBool(*r.Lenient)
	}

	if r.MaxConcurrentShardRequests != nil {
		params["max_concurrent_shard_requests"] = strconv.FormatInt(int64(*r.MaxConcurrentShardRequests), 10)
	}

	if r.Preference != "" {
		params["preference"] = r.Preference
	}

	if r.PreFilterShardSize != nil {
		params["pre_filter_shard_size"] = strconv.FormatInt(int64(*r.PreFilterShardSize), 10)
	}

	if r.Query != "" {
		params["q"] = r.Query
	}

	if r.RequestCache != nil {
		params["request_cache"] = strconv.FormatBool(*r.RequestCache)
	}

	if r.RestTotalHitsAsInt != nil {
		params["rest_total_hits_as_int"] = strconv.FormatBool(*r.RestTotalHitsAsInt)
	}

	if len(r.Routing) > 0 {
		params["routing"] = strings.Join(r.Routing, ",")
	}

	if r.Scroll != 0 {
		params["scroll"] = formatDuration(r.Scroll)
	}

	if r.SearchType != "" {
		params["search_type"] = r.SearchType
	}

	if r.SeqNoPrimaryTerm != nil {
		params["seq_no_primary_term"] = strconv.FormatBool(*r.SeqNoPrimaryTerm)
	}

	if r.Size != nil {
		params["size"] = strconv.FormatInt(int64(*r.Size), 10)
	}

	if len(r.Sort) > 0 {
		params["sort"] = strings.Join(r.Sort, ",")
	}

	if len(r.Source) > 0 {
		params["_source"] = strings.Join(r.Source, ",")
	}

	if len(r.SourceExcludes) > 0 {
		params["_source_excludes"] = strings.Join(r.SourceExcludes, ",")
	}

	if len(r.SourceIncludes) > 0 {
		params["_source_includes"] = strings.Join(r.SourceIncludes, ",")
	}

	if len(r.Stats) > 0 {
		params["stats"] = strings.Join(r.Stats, ",")
	}

	if len(r.StoredFields) > 0 {
		params["stored_fields"] = strings.Join(r.StoredFields, ",")
	}

	if r.SuggestField != "" {
		params["suggest_field"] = r.SuggestField
	}

	if r.SuggestMode != "" {
		params["suggest_mode"] = r.SuggestMode
	}

	if r.SuggestSize != nil {
		params["suggest_size"] = strconv.FormatInt(int64(*r.SuggestSize), 10)
	}

	if r.SuggestText != "" {
		params["suggest_text"] = r.SuggestText
	}

	if r.TerminateAfter != nil {
		params["terminate_after"] = strconv.FormatInt(int64(*r.TerminateAfter), 10)
	}

	if r.Timeout != 0 {
		params["timeout"] = formatDuration(r.Timeout)
	}

	if r.TrackScores != nil {
		params["track_scores"] = strconv.FormatBool(*r.TrackScores)
	}

	if r.TrackTotalHits != nil {
		params["track_total_hits"] = fmt.Sprintf("%v", r.TrackTotalHits)
	}

	if r.TypedKeys != nil {
		params["typed_keys"] = strconv.FormatBool(*r.TypedKeys)
	}

	if r.Version != nil {
		params["version"] = strconv.FormatBool(*r.Version)
	}

	if r.Pretty {
		params["pretty"] = "true"
	}

	if r.Human {
		params["human"] = "true"
	}

	if r.ErrorTrace {
		params["error_trace"] = "true"
	}

	if len(r.FilterPath) > 0 {
		params["filter_path"] = strings.Join(r.FilterPath, ",")
	}

	req, _ := newRequest(method, path.String(), r.Body)

	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	if r.Body != nil {
		req.Header[headerContentType] = headerContentTypeJSON
	}

	if len(r.Header) > 0 {
		if len(req.Header) == 0 {
			req.Header = r.Header
		} else {
			for k, vv := range r.Header {
				for _, v := range vv {
					req.Header.Add(k, v)
				}
			}
		}
	}

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	res, err := transport.Perform(req)
	if err != nil {
		return nil, err
	}

	response := Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	return &response, nil
}

// WithContext sets the request context.
//
func (f Search) WithContext(v context.Context) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.ctx = v
	}
}

// WithBody - The search definition using the Query DSL.
//
func (f Search) WithBody(v io.Reader) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Body = v
	}
}

// WithIndex - a list of index names to search; use _all to perform the operation on all indices.
//
func (f Search) WithIndex(v ...string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Index = v
	}
}

// WithDocumentType - a list of document types to search; leave empty to perform the operation on all types.
//
func (f Search) WithDocumentType(v ...string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.DocumentType = v
	}
}

// WithAllowNoIndices - whether to ignore if a wildcard indices expression resolves into no concrete indices. (this includes `_all` string or when no indices have been specified).
//
func (f Search) WithAllowNoIndices(v bool) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.AllowNoIndices = &v
	}
}

// WithAllowPartialSearchResults - indicate if an error should be returned if there is a partial search failure or timeout.
//
func (f Search) WithAllowPartialSearchResults(v bool) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.AllowPartialSearchResults = &v
	}
}

// WithAnalyzer - the analyzer to use for the query string.
//
func (f Search) WithAnalyzer(v string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Analyzer = v
	}
}

// WithAnalyzeWildcard - specify whether wildcard and prefix queries should be analyzed (default: false).
//
func (f Search) WithAnalyzeWildcard(v bool) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.AnalyzeWildcard = &v
	}
}

// WithBatchedReduceSize - the number of shard results that should be reduced at once on the coordinating node. this value should be used as a protection mechanism to reduce the memory overhead per search request if the potential number of shards in the request can be large..
//
func (f Search) WithBatchedReduceSize(v int) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.BatchedReduceSize = &v
	}
}

// WithDefaultOperator - the default operator for query string query (and or or).
//
func (f Search) WithDefaultOperator(v string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.DefaultOperator = v
	}
}

// WithDf - the field to use as default where no field prefix is given in the query string.
//
func (f Search) WithDf(v string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Df = v
	}
}

// WithDocvalueFields - a list of fields to return as the docvalue representation of a field for each hit.
//
func (f Search) WithDocvalueFields(v ...string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.DocvalueFields = v
	}
}

// WithExpandWildcards - whether to expand wildcard expression to concrete indices that are open, closed or both..
//
func (f Search) WithExpandWildcards(v string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.ExpandWildcards = v
	}
}

// WithExplain - specify whether to return detailed information about score computation as part of a hit.
//
func (f Search) WithExplain(v bool) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Explain = &v
	}
}

// WithFrom - starting offset (default: 0).
//
func (f Search) WithFrom(v int) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.From = &v
	}
}

// WithIgnoreThrottled - whether specified concrete, expanded or aliased indices should be ignored when throttled.
//
func (f Search) WithIgnoreThrottled(v bool) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.IgnoreThrottled = &v
	}
}

// WithIgnoreUnavailable - whether specified concrete indices should be ignored when unavailable (missing or closed).
//
func (f Search) WithIgnoreUnavailable(v bool) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.IgnoreUnavailable = &v
	}
}

// WithLenient - specify whether format-based query failures (such as providing text to a numeric field) should be ignored.
//
func (f Search) WithLenient(v bool) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Lenient = &v
	}
}

// WithMaxConcurrentShardRequests - the number of concurrent shard requests this search executes concurrently. this value should be used to limit the impact of the search on the cluster in order to limit the number of concurrent shard requests.
//
func (f Search) WithMaxConcurrentShardRequests(v int) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.MaxConcurrentShardRequests = &v
	}
}

// WithPreference - specify the node or shard the operation should be performed on (default: random).
//
func (f Search) WithPreference(v string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Preference = v
	}
}

// WithPreFilterShardSize - a threshold that enforces a pre-filter roundtrip to prefilter search shards based on query rewriting if the number of shards the search request expands to exceeds the threshold. this filter roundtrip can limit the number of shards significantly if for instance a shard can not match any documents based on it's rewrite method ie. if date filters are mandatory to match but the shard bounds and the query are disjoint..
//
func (f Search) WithPreFilterShardSize(v int) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.PreFilterShardSize = &v
	}
}

// WithQuery - query in the lucene query string syntax.
//
func (f Search) WithQuery(v string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Query = v
	}
}

// WithRequestCache - specify if request cache should be used for this request or not, defaults to index level setting.
//
func (f Search) WithRequestCache(v bool) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.RequestCache = &v
	}
}

// WithRestTotalHitsAsInt - this parameter is ignored in this version. it is used in the next major version to control whether the rest response should render the total.hits as an object or a number.
//
func (f Search) WithRestTotalHitsAsInt(v bool) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.RestTotalHitsAsInt = &v
	}
}

// WithRouting - a list of specific routing values.
//
func (f Search) WithRouting(v ...string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Routing = v
	}
}

// WithScroll - specify how long a consistent view of the index should be maintained for scrolled search.
//
func (f Search) WithScroll(v time.Duration) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Scroll = v
	}
}

// WithSearchType - search operation type.
//
func (f Search) WithSearchType(v string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.SearchType = v
	}
}

// WithSeqNoPrimaryTerm - specify whether to return sequence number and primary term of the last modification of each hit.
//
func (f Search) WithSeqNoPrimaryTerm(v bool) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.SeqNoPrimaryTerm = &v
	}
}

// WithSize - number of hits to return (default: 10).
//
func (f Search) WithSize(v int) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Size = &v
	}
}

// WithSort - a list of <field>:<direction> pairs.
//
func (f Search) WithSort(v ...string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Sort = v
	}
}

// WithSource - true or false to return the _source field or not, or a list of fields to return.
//
func (f Search) WithSource(v ...string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Source = v
	}
}

// WithSourceExcludes - a list of fields to exclude from the returned _source field.
//
func (f Search) WithSourceExcludes(v ...string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.SourceExcludes = v
	}
}

// WithSourceIncludes - a list of fields to extract and return from the _source field.
//
func (f Search) WithSourceIncludes(v ...string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.SourceIncludes = v
	}
}

// WithStats - specific 'tag' of the request for logging and statistical purposes.
//
func (f Search) WithStats(v ...string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Stats = v
	}
}

// WithStoredFields - a list of stored fields to return as part of a hit.
//
func (f Search) WithStoredFields(v ...string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.StoredFields = v
	}
}

// WithSuggestField - specify which field to use for suggestions.
//
func (f Search) WithSuggestField(v string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.SuggestField = v
	}
}

// WithSuggestMode - specify suggest mode.
//
func (f Search) WithSuggestMode(v string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.SuggestMode = v
	}
}

// WithSuggestSize - how many suggestions to return in response.
//
func (f Search) WithSuggestSize(v int) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.SuggestSize = &v
	}
}

// WithSuggestText - the source text for which the suggestions should be returned.
//
func (f Search) WithSuggestText(v string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.SuggestText = v
	}
}

// WithTerminateAfter - the maximum number of documents to collect for each shard, upon reaching which the query execution will terminate early..
//
func (f Search) WithTerminateAfter(v int) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.TerminateAfter = &v
	}
}

// WithTimeout - explicit operation timeout.
//
func (f Search) WithTimeout(v time.Duration) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Timeout = v
	}
}

// WithTrackScores - whether to calculate and return scores even if they are not used for sorting.
//
func (f Search) WithTrackScores(v bool) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.TrackScores = &v
	}
}

// WithTrackTotalHits - indicate if the number of documents that match the query should be tracked.
//
func (f Search) WithTrackTotalHits(v interface{}) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.TrackTotalHits = v
	}
}

// WithTypedKeys - specify whether aggregation and suggester names should be prefixed by their respective types in the response.
//
func (f Search) WithTypedKeys(v bool) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.TypedKeys = &v
	}
}

// WithVersion - specify whether to return document version as part of a hit.
//
func (f Search) WithVersion(v bool) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Version = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f Search) WithPretty() func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f Search) WithHuman() func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f Search) WithErrorTrace() func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f Search) WithFilterPath(v ...string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f Search) WithHeader(h map[string]string) func(*SearchRequest) {
	return func(r *SearchRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}
