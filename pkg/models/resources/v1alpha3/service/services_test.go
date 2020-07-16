/*
Copyright 2019 The KubeSphere Authors.

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

package service

import (
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"testing"
)

func TestListServices(t *testing.T) {
	tests := []struct {
		description string
		namespace   string
		query       *query.Query
		expected    *api.ListResult
		expectedErr error
	}{
		{
			"test name filter",
			"bar",
			&query.Query{
				Pagination: &query.Pagination{
					Limit:  1,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters:   map[query.Field]query.Value{query.FieldName: query.Value("foo2")},
			},
			&api.ListResult{
				Items: []interface{}{
					foo2,
				},
				TotalItems: 1,
			},
			nil,
		},
	}

	getter := prepare()

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			got, err := getter.List(test.namespace, test.query)

			if test.expectedErr != nil && err != test.expectedErr {
				t.Errorf("expected error, got nothing")
			} else if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
			}
		})
	}
}

var (
	foo1 = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo1",
			Namespace: "bar",
		},
	}

	foo2 = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo2",
			Namespace: "bar",
		},
	}
	bar1 = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar1",
			Namespace: "bar",
		},
	}

	services = []interface{}{foo1, foo2, bar1}
)

func prepare() v1alpha3.Interface {
	client := fake.NewSimpleClientset()
	informer := informers.NewSharedInformerFactory(client, 0)

	for _, service := range services {
		informer.Core().V1().Services().Informer().GetIndexer().Add(service)
	}

	return New(informer)
}
