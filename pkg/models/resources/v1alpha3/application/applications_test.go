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

package application

import (
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	appv1beta1 "sigs.k8s.io/application/pkg/apis/app/v1beta1"
	"sigs.k8s.io/application/pkg/client/clientset/versioned/fake"
	"sigs.k8s.io/application/pkg/client/informers/externalversions"
	"testing"
)

func applicationsToRuntimeObjects(applications ...*appv1beta1.Application) []runtime.Object {
	var objs []runtime.Object
	for _, app := range applications {
		objs = append(objs, app)
	}
	return objs
}

func TestListApplications(t *testing.T) {
	tests := []struct {
		description string
		namespace   string
		deployments []*appv1beta1.Application
		query       *query.Query
		expected    api.ListResult
		expectedErr error
	}{
		{
			"test name filter",
			"bar2",
			[]*appv1beta1.Application{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-1",
						Namespace: "bar",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-2",
						Namespace: "bar",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bar-2",
						Namespace: "bar2",
					},
				},
			},
			&query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters:   map[query.Field]query.Value{query.FieldNamespace: query.Value("bar2")},
			},
			api.ListResult{
				Items: []interface{}{
					&appv1beta1.Application{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "bar-2",
							Namespace: "bar2",
						},
					},
				},
				TotalItems: 2,
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			objs := applicationsToRuntimeObjects(test.deployments...)
			client := fake.NewSimpleClientset(objs...)

			informer := externalversions.NewSharedInformerFactory(client, 0)

			for _, deployment := range test.deployments {
				informer.App().V1beta1().Applications().Informer().GetIndexer().Add(deployment)
			}

			getter := New(informer)

			got, err := getter.List(test.namespace, test.query)
			if test.expectedErr != nil && err != test.expectedErr {
				t.Errorf("expected error, got nothing")
			} else if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(got.Items, test.expected.Items); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
			}
		})
	}
}
