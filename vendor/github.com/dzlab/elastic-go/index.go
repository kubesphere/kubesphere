package elastic

import (
	"encoding/json"
	"fmt"
	"log"
)

const (
	// ANALYSIS constant name of analysis part of Index API query
	ANALYSIS = "analysis"
	// SETTINGS constant name of settings attribute in query of Index API
	SETTINGS = "settings"
	// ALIAS constant name of field that defines alias name of this index
	ALIAS = "_alias"
	// ShardsNumber settings param of filed defining number of shards of index
	ShardsNumber = "number_of_shards"
	// ReplicasNumber settings param of field defining replicas number
	ReplicasNumber = "number_of_replicas"
	// RefreshInterval  settings param of field defining the refresh interval
	RefreshInterval = "refresh_interval"
	// TOKENIZER name of the analyzer responsible for tokenisation
	TOKENIZER = "tokenizer" // analyzer params
	// FILTER a parameter name of mapping in an Index API query
	FILTER = "filter"
	// CharFilter name of the analyzer responible for filtering characters.
	CharFilter = "char_filter"
	// MinShingleSize name of field that defines the minimum size of shingle
	MinShingleSize = "min_shingle_size"
	// MaxShingleSize name of field that defines the maximum size of shingle
	MaxShingleSize = "max_shingle_size"
	// OutputUnigrams constant name of field defining output unigrams
	OutputUnigrams = "output_unigrams"
)

// Analyer related constants
const (
	// StemExclusion a property in Analyzer settings used to define words that the analyzer should not stem
	StemExclusion = "stem_exclusion"
	// Stopwords a property in Analyzer settings used to define custom stopwords than the ones used by default by the analyzer
	Stopwords = "stopwords"
	// StopwordsPath a property in Analyzer settings used to define the path to a file containing custom stopwords.
	StopwordsPath = "stopwords_path"
	// Stemmer a value of 'type' propery in Analyzer settings used to define the stemmer
	Stemmer = "stemmer"
	// CommonGrams a vale of 'type' property in Filter settings.
	CommonGrams = "common_grams"
	// Type a property in Analyzer setting used to define the type of the property. Example of values: string (), stop (for stopwords), stemmer, common_grams, etc.
	Type = "type"
	// Language a property in Analyzer setting used to define the type of stemmer to use in order to reduce words to their root form. Possible values: english, english_light, english_possessive_stemmer (removes 's from words), synonym, mapping (e.g. for char filter).
	Language = "language"
	// CommonWords a property in Filter setting, similar to 'shingles' token filter, it makes phrase queries with stopwords more efficient. It accepcts values similar to the 'stopwords' property, example of values: _english_.
	CommonWords = "common_words"
	// CommonWordsPath a property in Analyzer setting used to define the path to a file containing common words.
	CommonWordsPath = "common_words_path"
	// QueryMode a boolean property in Filter settings. Used in conjugtion with common_words. It is set (by default) to false for indexing and to true for searching.
	QueryMode = "query_mode"
	// Synonyms a an array of formatted synonyms in Filter settings. Used when type is set to 'synonym'.
	Synonyms = "synonyms"
	// SynonymsPath a string property in field parameter. It is used to specify a path (absolute or relative to Elasticsearch 'config' directory) to a file containing formatted synonyms.
	SynonymsPath = "synonyms_path"
	// Encoder a property in Filter settings. Used when filter 'type' is set to 'phonetic' to set the name of Phonetic algorithm to use. Possible values: double_metaphone.
	Encoder = "encoder"
)

// Index a strcuture to hold a query for building/deleting indexes
type Index struct {
	client *Elasticsearch
	parser *IndexResultParser
	url    string
	dict   Dict
}

// String returns a JSON representation of the body of this Index
func (idx *Index) String() string {
	result, err := json.Marshal(idx.dict)
	if err != nil {
		log.Println(err)
	}
	return string(result)
}

// Index returns a query for managing indexes
func (client *Elasticsearch) Index(index string) *Index {
	url := fmt.Sprintf("http://%s/%s", client.Addr, index)
	return &Index{
		client: client,
		parser: &IndexResultParser{},
		url:    url,
		dict:   make(Dict),
	}
}

// Settings add a setting parameter to the Index query body
func (idx *Index) Settings(settings Dict) *Index {
	idx.dict[SETTINGS] = settings
	return idx
}

// Mappings set the mapping parameter
func (idx *Index) Mappings(doctype string, mapping *Mapping) *Index {
	if idx.dict[MAPPINGS] == nil {
		idx.dict[MAPPINGS] = make(Dict)
	}
	idx.dict[MAPPINGS].(Dict)[doctype] = mapping.query
	return idx
}

// newIndex creates new index settings
func newIndex() *Index {
	return &Index{dict: make(Dict)}
}

// SetAlias defines an alias for this index
func (idx *Index) SetAlias(alias string) *Index {
	idx.url += fmt.Sprintf("/%s/%s", ALIAS, alias)
	return idx
}

// AddSetting adds a key-value settings
func (idx *Index) AddSetting(name string, value interface{}) *Index {
	if idx.dict[SETTINGS] == nil {
		idx.dict[SETTINGS] = make(Dict)
	}
	idx.dict[SETTINGS].(Dict)[name] = value
	return idx
}

// SetShardsNb sets the number of shards
func (idx *Index) SetShardsNb(number int) *Index {
	idx.AddSetting(ShardsNumber, number)
	return idx
}

// SetReplicasNb sets the number of shards
func (idx *Index) SetReplicasNb(number int) *Index {
	idx.AddSetting(ReplicasNumber, number)
	return idx
}

// SetRefreshInterval sets the refresh interval
func (idx *Index) SetRefreshInterval(interval string) *Index {
	idx.AddSetting(RefreshInterval, interval)
	return idx
}

// Analyzer a structure for representing Analyzers and Filters
type Analyzer struct {
	name string
	kv   map[string]Dict
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer(name string) *Analyzer {
	return &Analyzer{name: name, kv: make(map[string]Dict)}
}

// String returns a JSON string representation of this analyzer
func (analyzer *Analyzer) String() string {
	dict := make(Dict)
	dict[analyzer.name] = analyzer.kv
	return String(dict)
}

// AddAnalyzer adds an anlyzer to the index settings
func (idx *Index) AddAnalyzer(analyzer *Analyzer) *Index {
	// if no "settings" create one
	if idx.dict[SETTINGS] == nil {
		idx.dict[SETTINGS] = make(Dict)
	}
	// if no "settings.analysis" create one
	if idx.dict[SETTINGS].(Dict)[ANALYSIS] == nil {
		idx.dict[SETTINGS].(Dict)[ANALYSIS] = make(Dict) //map[string]*Analyzer)
	}
	// insert the analyser ('name' and 'kv' attributes are taken separately)
	settings := idx.dict[SETTINGS].(Dict)
	analysis := settings[ANALYSIS].(Dict) //map[string]*Analyzer)
	analysis[analyzer.name] = analyzer.kv
	idx.dict[SETTINGS].(Dict)[ANALYSIS] = analysis
	return idx
}

// Add1 adds an attribute to analyzer definition
func (analyzer *Analyzer) Add1(key1, key2 string, value interface{}) *Analyzer {
	if len(analyzer.kv[key1]) == 0 {
		analyzer.kv[key1] = make(Dict)
	}
	analyzer.kv[key1][key2] = value
	return analyzer
}

// Add2 adds a dictionary of attributes to analyzer definition
func (analyzer *Analyzer) Add2(name string, value Dict) *Analyzer {
	if len(analyzer.kv[name]) == 0 {
		analyzer.kv[name] = make(Dict)
	}
	for k, v := range value {
		analyzer.kv[name][k] = v
	}
	return analyzer
}

// Pretty adds a parameter to the query url to pretify elasticsearch result
func (idx *Index) Pretty() *Index {
	idx.url += "?pretty"
	return idx
}

// Put submits to elasticsearch the query to create an index
// PUT /:index
func (idx *Index) Put() {
	url := idx.url
	query := String(idx.dict)

	idx.client.Execute("PUT", url, query, idx.parser)
}

// Delete submits to elasticsearch a query to delete an index
// DELETE /:index
func (idx *Index) Delete() {
	idx.client.Execute("DELETE", idx.url, "", idx.parser)
}
