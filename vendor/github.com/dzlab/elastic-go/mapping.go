package elastic

import ()

const (
	// MAPPING part of Mapping API path url
	MAPPING = "mapping"
	// MAPPINGS body of Mapping API query
	MAPPINGS = "mappings"
	// TYPE constant name of data type property of field
	TYPE = "type"
	// ANALYZER constant name of language analyzer for a field
	ANALYZER = "analyzer"
	// INDEX constant name of index name
	INDEX = "index"
	// PROPERTIES constant name of Mapping query body that defines properties
	PROPERTIES = "properties"
	// MATCH a query name
	MATCH = "match"
	// MatchMappingType type of matchi mapping (e.g. string)
	MatchMappingType = "match_mapping_type"
	// DynamicTemplates dynamic mapping templates
	DynamicTemplates = "dynamic_templates"
	// DEFAULT default mappings
	DEFAULT = "_default_"
	// PositionOffsetGap constant name for defining acceptable offset gap
	PositionOffsetGap = "position_offset_gap"
	// IndexAnalyzer index-time analyzer
	IndexAnalyzer = "index_analyzer"
	// SearchAnalyzer search-time analyzer
	SearchAnalyzer = "search_analyzer"
	// IndexOptions defines indexing options in Mapping query. Possible values are: docsi (default for 'not_analyzed' string fields), freqs, positions (default for 'analyzed' string fields), offsets.
	IndexOptions = "index_options"
	// Norms constant name for configuring field length normalization
	Norms = "norms"
	// Similarity in an Index mapping query. It defines the similarity algorithm to use. Possible values: default, BM25.
	Similarity = "similarity"
)

// Mapping maps between the json fields and how Elasticsearch store them
type Mapping struct {
	client *Elasticsearch
	parser *MappingResultParser
	url    string
	query  Dict
}

// NewMapping creates a new mapping query
func NewMapping() *Mapping {
	return &Mapping{
		query: make(Dict),
	}
}

// newMapping creates a new mapping query
func newMapping(client *Elasticsearch, url string) *Mapping {
	return &Mapping{
		client: client,
		parser: &MappingResultParser{},
		url:    url,
		query:  make(Dict),
	}
}

// Mapping creates request mappings between the json fields and how Elasticsearch store them
// GET /:index/:type/_mapping
func (client *Elasticsearch) Mapping(index, doctype string) *Mapping {
	url := client.request(index, doctype, -1, MAPPING)
	return newMapping(client, url)
}

// String returns a string representation of this mapping API
func (mapping *Mapping) String() string {
	return String(mapping.query)
}

// AddProperty adds a mapping for a type's property (e.g. type, index, analyzer, etc.)
func (mapping *Mapping) AddProperty(fieldname, propertyname string, propertyvalue interface{}) *Mapping {
	if mapping.query[PROPERTIES] == nil {
		mapping.query[PROPERTIES] = make(Dict)
	}
	property := mapping.query[PROPERTIES].(Dict)[fieldname]
	if property == nil {
		property = make(Dict)
	}
	property.(Dict)[propertyname] = propertyvalue
	mapping.query[PROPERTIES].(Dict)[fieldname] = property
	return mapping
}

// AddField adds a mapping for a field
func (mapping *Mapping) AddField(name string, body Dict) *Mapping {
	if mapping.query[PROPERTIES] == nil {
		mapping.query[PROPERTIES] = make(Dict)
	}
	mapping.query[PROPERTIES].(Dict)[name] = body
	return mapping
}

// AddDocumentType adds a mapping for a type of objects
func (mapping *Mapping) AddDocumentType(class *DocType) *Mapping {
	if mapping.query[MAPPINGS] == nil {
		mapping.query[MAPPINGS] = Dict{}
	}
	mapping.query[MAPPINGS].(Dict)[class.name] = class.dict
	return mapping
}

// Get submits a get request mappings between the json fields and how Elasticsearch store them
// GET /:index/_mapping/:type
func (mapping *Mapping) Get() {
	mapping.client.Execute("GET", mapping.url, "", mapping.parser)
}

// Put submits a request for updating the mappings between the json fields and how Elasticsearch store them
// PUT /:index/_mapping/:type
func (mapping *Mapping) Put() {
	url := mapping.url
	query := mapping.String()
	mapping.client.Execute("PUT", url, query, mapping.parser)
}

// DocType a structure for document type
type DocType struct {
	name string
	dict Dict
}

// NewDefaultType returns a '_default_' type that encapsulates shared/default settings
// e.g. specify index wide dynamic templates
func NewDefaultType() *DocType {
	return NewDocType(DEFAULT)
}

// NewDocType  a new mapping template
func NewDocType(name string) *DocType {
	return &DocType{name: name, dict: make(Dict)}
}

// AddProperty adds a property to this document type
func (doctype *DocType) AddProperty(name string, value interface{}) *DocType {
	doctype.dict[name] = value
	return doctype
}

// AddTemplate adds a template to this document type
func (doctype *DocType) AddTemplate(tmpl *Template) *DocType {
	doctype.dict[tmpl.name] = tmpl.dict
	return doctype
}

// AddDynamicTemplate adds a dynamic template to this mapping
func (doctype *DocType) AddDynamicTemplate(tmpl *Template) *DocType {
	if doctype.dict[DynamicTemplates] == nil {
		doctype.dict[DynamicTemplates] = []Dict{}
	}
	dict := make(Dict)
	dict[tmpl.name] = tmpl.dict
	doctype.dict[DynamicTemplates] = append(doctype.dict[DynamicTemplates].([]Dict), dict)
	return doctype
}

// String returns a string representation of this document type
func (doctype *DocType) String() string {
	dict := make(Dict)
	dict[doctype.name] = doctype.dict
	return String(dict)
}

// Template a structure for mapping template
type Template struct {
	name string
	dict Dict
}

// NewAllTemplate returns an new '_all' template
func NewAllTemplate() *Template {
	return NewTemplate(ALL)
}

// NewTemplate creates a new named mapping template
func NewTemplate(name string) *Template {
	return &Template{name: name, dict: make(Dict)}
}

// AddMatch adds a match string (e.g. '*', '_es')
func (template *Template) AddMatch(match string) *Template {
	template.dict[MATCH] = match
	return template
}

// AddProperty adds a property to this template
func (template *Template) AddProperty(name string, value interface{}) *Template {
	template.dict[name] = value
	return template
}

// AddMappingProperty adds a property to the `mapping` object
func (template *Template) AddMappingProperty(name string, value interface{}) *Template {
	if template.dict[MAPPING] == nil {
		template.dict[MAPPING] = make(Dict)
	}
	template.dict[MAPPING].(Dict)[name] = value
	return template
}

// String returns a string representation of this template
func (template *Template) String() string {
	dict := make(Dict)
	dict[template.name] = template.dict
	return String(dict)
}
