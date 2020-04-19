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
			"label=app.kubernetes.io/name:book&name=foo&status=Running&page=1&limit=10&ascending=true",
			&Query{
				Pagination: newPagination(10, 0),
				SortBy:     FieldCreationTimeStamp,
				Ascending:  true,
				Filters: map[Field]Value{
					FieldLabel:  Value("app.kubernetes.io/name:book"),
					FieldName:   Value("foo"),
					FieldStatus: Value("Running"),
				},
			},
		},
		{
			"test bad case",
			"xxxx=xxxx&dsfsw=xxxx&page=abc&limit=add&ascending=ssss",
			&Query{
				Pagination: NoPagination,
				SortBy:     FieldCreationTimeStamp,
				Ascending:  false,
				Filters: map[Field]Value{
					Field("xxxx"):  Value("xxxx"),
					Field("dsfsw"): Value("xxxx"),
				},
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
