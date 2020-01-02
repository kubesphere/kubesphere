package query

import (
	"github.com/emicklei/go-restful"
	"strconv"
	"strings"
)

const (
	ParameterName          = "name"
	ParameterLabelSelector = "labelSelector"
	ParameterFieldSelector = "fieldSelector"
	ParameterPage          = "page"
	ParameterLimit         = "limit"
	ParameterOrderBy       = "sortBy"
	ParameterAscending     = "ascending"
)

type Comparable interface {
	Compare(Comparable) int

	Contains(Comparable) bool
}

type ComparableString string

func (c ComparableString) Compare(comparable Comparable) int {
	other := comparable.(ComparableString)
	return strings.Compare(string(c), string(other))
}

func (c ComparableString) Contains(comparable Comparable) bool {
	other := comparable.(ComparableString)
	return strings.Contains(string(c), string(other))
}

// Query represents api search terms
type Query struct {
	Pagination *Pagination

	// sort result in which field, default to FieldCreationTimeStamp
	SortBy Field

	// sort result in ascending or descending order, default to descending
	Ascending bool

	//
	Filters []Filter
}

type Pagination struct {
	// items per page
	Limit int

	// page number
	Page int
}

var NoPagination = newPagination(-1, -1)

func newPagination(limit int, page int) *Pagination {
	return &Pagination{
		Limit: limit,
		Page:  page,
	}
}

func (p *Pagination) IsValidPagintaion() bool {
	return p.Limit >= 0 && p.Page >= 0
}

func (p *Pagination) IsPageAvailable(total, startIndex int) bool {
	return total > startIndex && p.Limit > 0
}

func (p *Pagination) GetPaginationSettings(total int) (startIndex, endIndex int) {
	startIndex = p.Limit * p.Page
	endIndex = startIndex + p.Limit

	if endIndex > total {
		endIndex = total
	}

	return startIndex, endIndex
}

func New() *Query {
	return &Query{
		Pagination: &Pagination{
			Limit: -1,
			Page:  -1,
		},
		SortBy:    "",
		Ascending: false,
		Filters:   []Filter{},
	}
}

type Filter struct {
	Field Field
	Value Comparable
}

func ParseQueryParameter(request *restful.Request) *Query {
	query := New()

	limit, err := strconv.ParseInt(request.QueryParameter("limit"), 10, 0)
	if err != nil {
		query.Pagination = NoPagination
	}

	page, err := strconv.ParseInt(request.QueryParameter("page"), 10, 0)
	if err == nil {
		query.Pagination = newPagination(int(limit), int(page-1))
	}

	query.SortBy = Field(defaultString(request.QueryParameter("sortBy"), FieldCreationTimeStamp))

	ascending, err := strconv.ParseBool(defaultString(request.QueryParameter("ascending"), "false"))
	if err != nil {
		query.Ascending = false
	} else {
		query.Ascending = ascending
	}

	for _, field := range ComparableFields {
		f := request.QueryParameter(string(field))
		if len(f) != 0 {
			query.Filters = append(query.Filters, Filter{
				Field: field,
				Value: ComparableString(f),
			})
		}
	}

	return query

}

func defaultString(value, defaultValue string) string {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
