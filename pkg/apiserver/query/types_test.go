/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package query

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/google/go-cmp/cmp"
)

func TestParseQueryParameter(t *testing.T) {
	tests := []struct {
		description string
		queryString string
		expected    *Query
	}{
		{
			"test normal case",
			"label=app.kubernetes.io/name=book&name=foo&status=Running&page=1&limit=10&ascending=true",
			&Query{
				Pagination: newPagination(10, 0),
				SortBy:     FieldCreationTimeStamp,
				Ascending:  true,
				Filters: map[Field]Value{
					FieldLabel:  Value("app.kubernetes.io/name=book"),
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
