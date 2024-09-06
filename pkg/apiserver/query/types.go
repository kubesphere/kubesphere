/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package query

import (
	"strconv"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/labels"

	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
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
	Filters map[Field]Value

	LabelSelector string
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

func (q *Query) Selector() labels.Selector {
	if selector, err := labels.Parse(q.LabelSelector); err != nil {
		return labels.Everything()
	} else {
		return selector
	}
}

func (q *Query) AppendLabelSelector(ls map[string]string) error {
	labelsMap, err := labels.ConvertSelectorToLabelsMap(q.LabelSelector)
	if err != nil {
		return err
	}
	q.LabelSelector = labels.Merge(labelsMap, ls).String()
	return nil
}

func (p *Pagination) GetValidPagination(total int) (startIndex, endIndex int) {
	// no pagination
	if p.Limit == NoPagination.Limit {
		return 0, total
	}

	// out of range
	if p.Limit < 0 || p.Offset < 0 || p.Offset > total {
		return 0, 0
	}

	startIndex = p.Offset
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
		Filters:    map[Field]Value{},
	}
}

type Filter struct {
	Field Field `json:"field"`
	Value Value `json:"value"`
}

func ParseQueryParameter(request *restful.Request) *Query {
	query := New()

	limit, err := strconv.Atoi(request.QueryParameter(ParameterLimit))
	// equivalent to undefined, use the default value
	if err != nil {
		limit = -1
	}
	page, err := strconv.Atoi(request.QueryParameter(ParameterPage))
	// equivalent to undefined, use the default value
	if err != nil {
		page = 1
	}

	query.Pagination = newPagination(limit, (page-1)*limit)

	query.SortBy = Field(defaultString(request.QueryParameter(ParameterOrderBy), FieldCreationTimeStamp))

	ascending, err := strconv.ParseBool(defaultString(request.QueryParameter(ParameterAscending), "false"))
	if err != nil {
		query.Ascending = false
	} else {
		query.Ascending = ascending
	}

	query.LabelSelector = request.QueryParameter(ParameterLabelSelector)

	for key, values := range request.Request.URL.Query() {
		if !sliceutil.HasString([]string{ParameterPage, ParameterLimit, ParameterOrderBy, ParameterAscending, ParameterLabelSelector}, key) {
			value := ""
			if len(values) > 0 {
				value = values[0]
			}
			query.Filters[Field(key)] = Value(value)
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
