package query

import (
	"github.com/emicklei/go-restful"
	"strconv"
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

	// offset
	Offset int
}

var NoPagination = newPagination(-1, 0)

// make sure that pagination is valid
func newPagination(limit int, offset int) *Pagination {
	return &Pagination{
		Limit:  limit,
		Offset: offset,
	}
}

func (p *Pagination) GetValidPagination(total int) (startIndex, endIndex int) {

	if p.Limit == NoPagination.Limit {
		return 0, total
	}

	if p.Limit < 0 || p.Offset < 0 || total == 0 {
		return 0, 0
	}

	startIndex = p.Limit * p.Offset
	endIndex = startIndex + p.Limit

	if endIndex > total {
		endIndex = total
	}

	return startIndex, endIndex
}

func New() *Query {
	return &Query{
		Pagination: NoPagination,
		SortBy:     "",
		Ascending:  false,
		Filters:    []Filter{},
	}
}

type Filter struct {
	Field Field
	Value Value
}

func ParseQueryParameter(request *restful.Request) *Query {
	query := New()

	limit, err := strconv.Atoi(request.QueryParameter("limit"))
	// equivalent to undefined, use the default value
	if err != nil {
		limit = -1
	}
	page, err := strconv.Atoi(request.QueryParameter("page"))
	// equivalent to undefined, use the default value
	if err != nil {
		page = 1
	}

	query.Pagination = newPagination(limit, page-1)

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
				Value: Value(f),
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
