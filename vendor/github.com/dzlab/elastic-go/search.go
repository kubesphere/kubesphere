package elastic

import ()

// Dict a dictionary with string keys and values of any type
type Dict map[string]interface{}

// fields of a Search API call
const (
	// EXPLAIN constant name of Explain API request
	EXPLAIN = "explain"
	// VALIDATE constant name of Validate API request
	VALIDATE = "validate"
	// SEARCH constant name of Search API request
	SEARCH = "search"
	// ALL a query element
	ALL = "_all"
	// INCLUDE a query element
	INCLUDE = "include_in_all"
	// SOURCE a query element
	SOURCE = "_source"
	// SearchType a url param
	SearchType = "search_type"
	// SCROLL a url param
	SCROLL = "scroll"
	// PostFilter contant name of post_filter, a top level search parameter that is executed after the search query.
	PostFilter = "post_filter"
	// Filter a query name.
	Filter = "filter"
	// DisMax query name.
	DisMax = "dis_max"
	// MultiMatch a match query on multiple terms
	MultiMatch = "multi_match"
	// Common a query name.
	Common = "common"
	// Boosting a query param that include additional results but donwgrade them
	Boosting = "boosting"
	// ConstantScore a query param that assings 1 as score to any matching document
	ConstantScore = "constant_score"
	// FunctionScore a query for customizing the scoring with predefined functions: weight, field_value_factor, random_score
	FunctionScore = "function_score"
	// Fuzzy 'fuzzy' qearch query. It's a term-level query that doesn't do analysis.
	Fuzzy = "fuzzy"
	// MatchPhrase 'phrase' search query
	MatchPhrase = "match_phrase"
	// MatchPhrasePrefix 'phrase' search query
	MatchPhrasePrefix = "match_phrase_prefix"
	// Prefix search terms with given prefix
	Prefix = "prefix"
	// Wildcard search terms with widcard
	Wildcard = "wildcard"
	// RegExp filter terms application to regular expression
	RegExp = "regexp"
	// RESCORE rescores result of previous query
	RESCORE = "rescore"
	// RescoreQuery
	RescoreQuery = "rescore_query"

	// CutOffFrequency query params. It is used to split query terms into 2 categories: low frequency terms for matching, and high frequency terms for sorting only.
	CutOffFrequency = "cutoff_frequency"
	// MinimumShouldMatch query params. It is used to reduce the number of low qualitymatches.
	MinimumShouldMatch = "minimum_should_match"
	// SLOP in 'phrase' queries to describe proximity/word ordering
	SLOP = "slop"
	// MaxExpansions controls how many terms the prefix is allowed to match
	MaxExpansions = "max_expansions"
	// WindowSize number of document from each shard
	WindowSize = "window_size"
	// DisableCoord a boolean value to enable/disable the use of Query Coordination in 'bool' queries
	DisableCoord = "disable_coord"
	// Boost an Int value in query clauses to give it more importance
	Boost = "boost"
	// IndicesBoost in mutli-index search, a dictionary for each index name it's boost value. For instance, it can be used to specify a language preference if there is an index defined per language (e.g. blogs-en, blogs-fr)
	IndicesBoost = "indices_boost"
	// NegativeBoost in boosting query, a float representing negative boost value
	NegativeBoost = "negative_boost"
	// Fuzziness a query parameter in 'fuzzy' (and also 'match', 'multi_match') query. It's used to set the maximum edit distance between a potentially mispelled word and the index words.
	Fuzziness = "fuzziness"
	// PrefixLength an integer query parameter in the 'fuzzy' query. It is used to fix the initial characters, of a word, which will not be fuzzified.
	PrefixLength = "prefix_length"
	// Operator a query parameter in the 'match' query. Possible values: and.
	Operator = "operator"

	// Weight a predifined scoring function that can be used in any query. It assigns a non normalized boost to each document (i.e. is used as it is an not alterned like 'boost')
	Weight = "weight"
	// FieldValueFactor a predifined scoring function that uses a value of a field from the given document to alter _score
	FieldValueFactor = "field_value_factor"
	// RandomScore a predifined scoring function to randomly sort documents for different users
	RandomScore = "random_score"
	// Seed is a parameter used in comabination with 'random_score'. It is used to ensure same document ordering when same seed is used (e.g. session identifier).
	Seed = "seed"
	// ScriptScore a predifined scoring function that uses a custom script
	ScriptScore = "script_score"
	// Modifer a parameter of 'field_value_factor' in a FunctionScore query. It is used to alter the calculation of the new document score, possible values log1p, etc.
	Modifer = "modifier"
	// Factor a parameter of 'field_value_factor' in a FunctionScore query. It is used to multiply the value of the concerned field (e.g. votes) to alter the final score calculation.
	Factor = "factor"
	// BoostMode is a parameter in a FunctionScore query. It is used to specify how the calculated score will affect final document score.
	// Possible values: multiply (mulitply _score by calculated result), sum (sum _score with calculated), min (lower of _score and calculated), max (higher of _score and calculated), replace (replace _score with calculated)
	BoostMode = "boost_mode"
	// MaxBoost is a parameter in a FunctionScore query. It is used to cap the maximum effect of the scoring function.
	MaxBoost = "max_boost"
	// ScoreMode is a parameter in a FunctionScore query. It defines, when there is many 'functions', how to reduce multiple results into single value.
	// Possible values are multiply, sum, avg, max, min, first.
	ScoreMode = "score_mode"
)

// Search a request representing a search
type Search struct {
	client *Elasticsearch
	parser *SearchResultParser
	url    string
	params map[string]string
	query  Dict
}

// Query defines an interfece of an object from an Elasticsearch query
type Query interface {
	Name() string
	KV() Dict
}

// Object a general purpose query
type Object struct {
	name string
	kv   Dict
}

// Name returns the name of this query object
func (obj *Object) Name() string {
	return obj.name
}

// KV returns the key-value store representing the body of this query
func (obj *Object) KV() Dict {
	return obj.kv
}

// NewQuery Create a new query object
func NewQuery(name string) *Object {
	return &Object{name: name, kv: make(Dict)}
}

// NewFilter returns a new filter query
func NewFilter() *Object {
	return NewQuery(Filter)
}

// NewMatch Create a new match query
func NewMatch() *Object {
	return NewQuery(MATCH)
}

// NewMultiMatch Create a new multi_match query
func NewMultiMatch() *Object {
	return NewQuery(MultiMatch)
}

// NewMatchPhrase Create a `match_phrase` query to find words that are near each other
func NewMatchPhrase() *Object {
	return NewQuery(MatchPhrase)
}

// NewRescore Create a `rescore` query
func NewRescore() *Object {
	return NewQuery(RESCORE)
}

// NewRescoreQuery Create a `rescore` query algorithm
func NewRescoreQuery() *Object {
	return NewQuery(RescoreQuery)
}

// NewConstantScore creates a new 'constant_score' query
func NewConstantScore() *Object {
	return NewQuery(ConstantScore)
}

// NewFunctionScore creates a new 'function_score' query
func NewFunctionScore() *Object {
	return NewQuery(FunctionScore)
}

// NewFuzzyQuery create a new 'fuzzy' query
func NewFuzzyQuery() *Object {
	return NewQuery(Fuzzy)
}

// newQuery used for test purpose
func newQuery() *Object {
	return &Object{name: "", kv: make(Dict)}
}

// Dict return a dictionarry representation of this object
func (obj *Object) Dict() Dict {
	return Dict{obj.name: obj.KV()}
}

// String returns a string representation of this object
func (obj *Object) String() string {
	return String(obj.KV())
}

// Explain creates an Explaination request, that will return explanation for why a document is returned by the query
func (client *Elasticsearch) Explain(index, class string, id int64) *Search {
	url := client.request(index, class, id, EXPLAIN)
	return newSearch(client, url)
}

// Validate creates a Validation request
func (client *Elasticsearch) Validate(index, class string, explain bool) *Search {
	url := client.request(index, class, -1, VALIDATE) + "/query"
	if explain {
		url += "?" + EXPLAIN
	}
	return newSearch(client, url)
}

// Search creates a Search request
func (client *Elasticsearch) Search(index, class string) *Search {
	url := client.request(index, class, -1, SEARCH)
	return newSearch(client, url)
}

// newSearch creates a new Search API call
func newSearch(client *Elasticsearch, url string) *Search {
	return &Search{
		client: client,
		parser: &SearchResultParser{},
		url:    url,
		params: make(map[string]string),
		query:  make(Dict),
	}
}

// AddParam adds a url parameter/value, e.g. search_type (count, query_and_fetch, dfs_query_then_fetch/dfs_query_and_fetch, scan)
func (search *Search) AddParam(name, value string) *Search {
	search.params[name] = value
	return search
}

// Pretty pretiffies the response result
func (search *Search) Pretty() *Search {
	search.AddParam("pretty", "")
	return search
}

// AddQuery adds a query to this search request
func (search *Search) AddQuery(query Query) *Search {
	search.query[query.Name()] = query.KV()
	return search
}

// AddSource adds to _source (i.e. specify another field that should be extracted)
func (search *Search) AddSource(source string) *Search {
	var sources []string
	if search.query[SOURCE] == nil {
		sources = []string{}
	} else {
		sources = search.query[SOURCE].([]string)
	}
	sources = append(sources, source)
	search.query[SOURCE] = sources
	return search
}

// Add adds a query argument/value, e.g. size, from, etc.
func (search *Search) Add(argument string, value interface{}) *Search {
	search.query[argument] = value
	return search
}

// String returns a string representation of this Search API call
func (search *Search) String() string {
	body := ""
	if len(search.query) > 0 {
		body = String(search.query)
	}
	return body
}

// urlString constructs the url of this Search API call
func (search *Search) urlString() string {
	return urlString(search.url, search.params)
}

// Get submits request mappings between the json fields and how Elasticsearch store them
// GET /:index/:type/_search
func (search *Search) Get() {
	// construct the url
	url := search.urlString()
	// construct the body
	query := search.String()

	search.client.Execute("GET", url, query, search.parser)
}

// Add adds a query argument/value
func (obj *Object) Add(argument string, value interface{}) *Object {
	obj.kv[argument] = value
	return obj
}

// AddMultiple specify multiple values to match
func (obj *Object) AddMultiple(argument string, values ...interface{}) *Object {
	obj.kv[argument] = values
	return obj
}

// AddQueries adds multiple queries, under given `name`
func (obj *Object) AddQueries(name string, queries ...Query) *Object {
	for _, q := range queries {
		parent := NewQuery(name)
		parent.AddQuery(q)
		obj.AddQuery(parent)
	}
	return obj
}

// AddQuery adds a sub query (e.g. a field query)
func (obj *Object) AddQuery(query Query) *Object {
	collection := obj.kv[query.Name()]
	// check if query.Name exists, otherwise transform the map to array
	if collection == nil {
		// at first the collection is a map
		collection = query.KV()
	} else {
		// when more items are added, then it becomes an array
		dict := query.KV()
		// check if it is a map
		if _, ok := collection.(Dict); ok {
			array := []Dict{} // transform previous map into array
			for k, v := range collection.(Dict) {
				d := make(Dict)
				d[k] = v
				array = append(array, d)
			}
			collection = array
		}
		collection = append(collection.([]Dict), dict)
	}
	obj.kv[query.Name()] = collection
	return obj
}

// Bool represents a boolean clause, it is a complex clause that allows to combine other clauses as 'must' match, 'must_not' match, 'should' match.
type Bool struct {
	name string
	kv   Dict
}

// Name returns the name of this 'bool' query
func (b *Bool) Name() string {
	return b.name
}

// KV returns the key-value store representing the body of this 'bool' query
func (b *Bool) KV() Dict {
	return b.kv
}

// NewBool creates a new 'bool' clause
func NewBool() *Bool {
	kv := make(Dict)
	return &Bool{name: "bool", kv: kv}
}

// AddMust adds a 'must' clause to this 'bool' clause
func (b *Bool) AddMust(query Query) *Bool {
	b.add("must", query)
	return b
}

// AddMustNot adds a 'must_not' clause to this 'bool' clause
func (b *Bool) AddMustNot(query Query) *Bool {
	b.add("must_not", query)
	return b
}

// AddShould adds a 'should' clause to this 'bool' clause
func (b *Bool) AddShould(query Query) *Bool {
	b.add("should", query)
	return b
}

// Add adds a parameter to this `bool` query
func (b *Bool) Add(name string, value interface{}) *Bool {
	b.kv[name] = value
	return b
}

// add adds a clause
func (b *Bool) add(key string, query Query) {
	collection := b.kv[key]
	// check if query.Name exists, otherwise transform the map to array
	if collection == nil {
		// at first the collection is a map
		collection = make(Dict)
		collection.(Dict)[query.Name()] = query.KV()
	} else {
		// when more items are added, then it becomes an array
		dict := make(Dict)
		dict[query.Name()] = query.KV()
		// check if it is a map
		if _, ok := collection.(Dict); ok {
			array := []Dict{} // transform previous map into array
			for k, v := range collection.(Dict) {
				d := make(Dict)
				d[k] = v
				array = append(array, d)
			}
			collection = array
		}
		collection = append(collection.([]Dict), dict)
	}
	b.kv[key] = collection
}

// NewTerms creates a new 'terms' filter, it is like 'term' but can match multiple values
func NewTerms() *Object {
	return NewQuery("terms")
}

// NewTerm creates a new 'term' filter
func NewTerm() *Object {
	return NewQuery("term")
}

// NewExists creates a new `exists` filter.
func NewExists() *Object {
	return NewQuery("exists")
}

// NewMissing creates a new `missing` filter (the inverse of `exists`)
func NewMissing() *Object {
	return NewQuery("missing")
}

// BoostingQuery a strcuture representing the 'boosting' query
type BoostingQuery struct {
	positive      Dict
	negative      Dict
	negativeBoost float32
}

// NewBoosting returns a new Boosting query
func NewBoosting() *BoostingQuery {
	return &BoostingQuery{
		positive: make(Dict),
		negative: make(Dict),
	}
}

// Name returns the name of boosting query
func (boosting *BoostingQuery) Name() string {
	return Boosting
}

// KV returns the body of this boosting query as a dictionary
func (boosting *BoostingQuery) KV() Dict {
	dict := make(Dict)
	dict["positive"] = boosting.positive
	dict["negative"] = boosting.negative
	dict[NegativeBoost] = boosting.negativeBoost
	return dict
}

// SetNegativeBoost sets the negative boost
func (boosting *BoostingQuery) SetNegativeBoost(value float32) *BoostingQuery {
	boosting.negativeBoost = value
	return boosting
}

// AddPositive adds a positive clause to boosting query
func (boosting *BoostingQuery) AddPositive(name string, value interface{}) *BoostingQuery {
	boosting.positive[name] = value
	return boosting
}

// AddNegative adds a negative clause to boosting query
func (boosting *BoostingQuery) AddNegative(name string, value interface{}) *BoostingQuery {
	boosting.negative[name] = value
	return boosting
}
