package query

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/google/go-cmp/cmp"
	"net/http"
	"testing"
)

func TestParseQueryParameter(t *testing.T) {
	tests := []struct {
		description string
		queryString string
		expected    *Query
	}{
		{
			"test normal case",
			"name=foo&status=Running&application=book&page=1&limit=10&ascending=true",
			&Query{
				Pagination: newPagination(10, 0),
				SortBy:     FieldCreationTimeStamp,
				Ascending:  true,
				Filters: []Filter{
					{
						FieldName,
						ComparableString("foo"),
					},
					{
						FieldStatus,
						ComparableString("Running"),
					},
					{
						FieldApplication,
						ComparableString("book"),
					},
				},
			},
		},
		{
			"test bad case",
			"xxxx=xxxx&dsfsw=xxxx&page=abc&limit=add&ascending=ssss",
			&Query{
				Pagination: &Pagination{
					Limit: -1,
					Page:  -1,
				},
				SortBy:    FieldCreationTimeStamp,
				Ascending: false,
				Filters:   []Filter{},
			},
		},
	}

	for _, test := range tests {
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost?%s", test.queryString), nil)
		if err != nil {
			t.Fatal(err)
		}

		request := restful.NewRequest(req)

		t.Run(test.description, func(t *testing.T) {
			got := ParseQueryParameter(request)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
				return
			}
		})
	}
}
