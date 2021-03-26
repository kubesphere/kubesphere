package query

import (
	"reflect"

	jsoniter "github.com/json-iterator/go"
)

// TODO: elastic/go-elasticsearch is working on Query DSL support.
//  See https://github.com/elastic/go-elasticsearch/issues/42.
//  We need refactor our query body builder when that is ready.
type Builder struct {
	From          int64               `json:"from,omitempty"`
	Size          int64               `json:"size,omitempty"`
	Sorts         []map[string]string `json:"sort,omitempty"`
	*Query        `json:",inline"`
	*Aggregations `json:"aggs,omitempty"`
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) Bytes() ([]byte, error) {
	return jsoniter.Marshal(b)
}

func (b *Builder) WithQuery(q *Query) *Builder {

	if q == nil || q.Bool == nil || !q.IsValid() {
		return b
	}

	b.Query = q
	return b
}

func (b *Builder) WithAggregations(aggs *Aggregations) *Builder {

	b.Aggregations = aggs
	return b
}

func (b *Builder) WithFrom(n int64) *Builder {
	b.From = n
	return b
}

func (b *Builder) WithSize(n int64) *Builder {
	b.Size = n
	return b
}

func (b *Builder) WithSort(key, order string) *Builder {
	if order == "" {
		order = "desc"
	}
	b.Sorts = []map[string]string{{key: order}}
	return b
}

// Query

type Query struct {
	*Bool `json:"query,omitempty"`
}

func NewQuery() *Query {
	return &Query{}
}

func (q *Query) WithBool(b *Bool) *Query {
	if b == nil || !b.IsValid() {
		return q
	}

	q.Bool = b
	return q
}

// Aggregations

type Aggregations struct {
	*CardinalityAggregation   `json:"cardinality_aggregation,omitempty"`
	*DateHistogramAggregation `json:"date_histogram_aggregation,omitempty"`
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

func NewAggregations() *Aggregations {
	return &Aggregations{}
}

func (a *Aggregations) WithCardinalityAggregation(field string) *Aggregations {

	a.CardinalityAggregation = &CardinalityAggregation{
		&Cardinality{
			Field: field,
		},
	}

	return a
}

func (a *Aggregations) WithDateHistogramAggregation(field string, interval string) *Aggregations {

	a.DateHistogramAggregation = &DateHistogramAggregation{
		&DateHistogram{
			Field:    field,
			Interval: interval,
		},
	}

	return a
}

type Item interface {
	IsValid() bool
}

// Example:
// {bool: {filter: <[]Match>}}
// {bool: {should: <[]Match>, minimum_should_match: 1}}
type Bool struct {
	*Parameter `json:"bool,omitempty"`
}

type Parameter struct {
	Filter             []interface{} `json:"filter,omitempty"`
	Should             []interface{} `json:"should,omitempty"`
	MustNot            []interface{} `json:"must_not,omitempty"`
	MinimumShouldMatch int32         `json:"minimum_should_match,omitempty"`
}

func NewBool() *Bool {
	return &Bool{
		&Parameter{},
	}
}

func (b *Bool) IsValid() bool {
	if (b.Filter == nil || len(b.Filter) == 0) &&
		(b.Should == nil || len(b.Should) == 0) &&
		(b.MustNot == nil || len(b.MustNot) == 0) {
		return false
	}

	return true
}

func (b *Bool) AppendFilter(item Item) *Bool {

	if reflect.ValueOf(item).IsNil() || !item.IsValid() {
		return b
	}

	b.Filter = append(b.Filter, item)
	return b
}

func (b *Bool) AppendMultiFilter(items []Item) *Bool {

	if items == nil || len(items) == 0 {
		return b
	}

	for _, item := range items {
		if item.IsValid() {
			b.Filter = append(b.Filter, item)
		}
	}

	return b
}

func (b *Bool) AppendShould(item Item) *Bool {

	if reflect.ValueOf(item).IsNil() || !item.IsValid() {
		return b
	}

	b.Should = append(b.Should, item)
	return b
}

func (b *Bool) AppendMultiShould(items []Item) *Bool {

	if items == nil || len(items) == 0 {
		return b
	}

	for _, item := range items {
		if item.IsValid() {
			b.Should = append(b.Should, item)
		}
	}
	return b
}

func (b *Bool) AppendMustNot(item Item) *Bool {

	if reflect.ValueOf(item).IsNil() || !item.IsValid() {
		return b
	}

	b.MustNot = append(b.MustNot, item)
	return b
}

func (b *Bool) AppendMultiMustNot(items []Item) *Bool {

	if items == nil || len(items) == 0 {
		return b
	}

	for _, item := range items {
		if item.IsValid() {
			b.MustNot = append(b.MustNot, item)
		}
	}
	return b
}

func (b *Bool) WithMinimumShouldMatch(min int32) *Bool {

	b.MinimumShouldMatch = min
	return b
}

type MatchPhrase struct {
	MatchPhrase map[string]string `json:"match_phrase,omitempty"`
}

func (m *MatchPhrase) IsValid() bool {

	if m.MatchPhrase == nil || len(m.MatchPhrase) == 0 {
		return false
	}

	return true
}

func NewMatchPhrase(key, val string) *MatchPhrase {
	return &MatchPhrase{
		MatchPhrase: map[string]string{
			key: val,
		},
	}
}

func NewMultiMatchPhrase(key string, val []string) []Item {

	var array []Item

	if val == nil || len(val) == 0 {
		return nil
	}

	for _, v := range val {
		array = append(array, &MatchPhrase{
			MatchPhrase: map[string]string{
				key: v,
			},
		})
	}

	return array
}

type MatchPhrasePrefix struct {
	MatchPhrasePrefix map[string]string `json:"match_phrase_prefix,omitempty"`
}

func (m *MatchPhrasePrefix) IsValid() bool {

	if m.MatchPhrasePrefix == nil || len(m.MatchPhrasePrefix) == 0 {
		return false
	}

	return true
}

func NewMatchPhrasePrefix(key, val string) *MatchPhrasePrefix {
	return &MatchPhrasePrefix{
		MatchPhrasePrefix: map[string]string{
			key: val,
		},
	}
}

func NewMultiMatchPhrasePrefix(key string, val []string) []Item {

	var array []Item

	if val == nil || len(val) == 0 {
		return nil
	}

	for _, v := range val {
		array = append(array, &MatchPhrasePrefix{
			MatchPhrasePrefix: map[string]string{
				key: v,
			},
		})
	}

	return array
}

type Regexp struct {
	Regexp map[string]string `json:"regexp,omitempty"`
}

func (m *Regexp) IsValid() bool {

	if m.Regexp == nil || len(m.Regexp) == 0 {
		return false
	}

	return true
}

func NewRegex(key, val string) *Regexp {
	return &Regexp{
		Regexp: map[string]string{
			key: val,
		},
	}
}

type Range struct {
	Range map[string]map[string]interface{} `json:"range,omitempty"`
}

func NewRange(key string) *Range {
	return &Range{
		Range: map[string]map[string]interface{}{
			key: make(map[string]interface{}),
		},
	}
}

func (r *Range) WithGT(val interface{}) *Range {
	r.withRange("gt", val)
	return r
}

func (r *Range) WithGTE(val interface{}) *Range {
	r.withRange("gte", val)
	return r
}

func (r *Range) WithLT(val interface{}) *Range {
	r.withRange("lt", val)
	return r
}

func (r *Range) WithLTE(val interface{}) *Range {
	r.withRange("lte", val)
	return r
}

func (r *Range) IsValid() bool {
	if r.Range == nil {
		return false
	}

	if len(r.Range) == 0 {
		return false
	}

	for _, v := range r.Range {
		if len(v) != 0 {
			return true
		}
	}

	return false
}

func (r *Range) withRange(operator string, val interface{}) {
	if r.Range == nil {
		return
	}

	for _, v := range r.Range {
		v[operator] = val
	}
}

type Wildcard struct {
	Wildcard map[string]string `json:"wildcard,omitempty"`
}

func (m *Wildcard) IsValid() bool {

	if m.Wildcard == nil || len(m.Wildcard) == 0 {
		return false
	}

	return true
}

func NewWildcard(key, val string) *Wildcard {

	return &Wildcard{
		Wildcard: map[string]string{
			key: val,
		},
	}
}

func NewMultiWildcard(key string, val []string) []Item {

	var array []Item

	if val == nil || len(val) == 0 {
		return nil
	}

	for _, v := range val {
		array = append(array, &Wildcard{
			Wildcard: map[string]string{
				key: v,
			},
		})
	}

	return array
}

type Terms struct {
	Terms map[string]interface{} `json:"terms,omitempty"`
}

func (m *Terms) IsValid() bool {

	if m.Terms == nil || len(m.Terms) == 0 {
		return false
	}

	return true
}

func NewTerms(key string, val interface{}) *Terms {

	if reflect.ValueOf(val).IsNil() {
		return nil
	}

	return &Terms{
		Terms: map[string]interface{}{
			key: val,
		},
	}
}

type Exists struct {
	Exists map[string]string `json:"exists,omitempty"`
}

func (m *Exists) IsValid() bool {

	if m.Exists == nil || len(m.Exists) == 0 {
		return false
	}

	return true
}

func NewExists(key, val string) *Exists {
	return &Exists{
		Exists: map[string]string{
			key: val,
		},
	}
}
