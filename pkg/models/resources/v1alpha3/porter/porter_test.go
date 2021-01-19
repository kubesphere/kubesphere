/*
Copyright 2021 The KubeSphere Authors.

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

package porter

import (
	"context"
	"path/filepath"
	"testing"

	porterv1alpha2 "github.com/kubesphere/porter/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestListEips(t *testing.T) {
	tests := []struct {
		description string
		namespace   string
		query       *query.Query
		expected    *api.ListResult
		expectedErr error
	}{
		{
			"test name list",
			"",
			&query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters: map[query.Field]query.Value{
					query.FieldName: query.Value("eip1"),
				},
			},
			&api.ListResult{
				Items:      []interface{}{eip1},
				TotalItems: 1,
			},
			nil,
		},
		{
			"test list",
			"",
			&query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
			},
			&api.ListResult{
				Items:      []interface{}{eip1, eip2},
				TotalItems: 2,
			},
			nil,
		},
	}

	getter := prepare(t)

	for _, test := range tests {
		got, err := getter.List(test.namespace, test.query)
		if test.expectedErr != nil && err != test.expectedErr {
			t.Errorf("expected error, got nothing")
		} else if err != nil {
			t.Fatal(err)
		}

		if len(got.Items) != test.expected.TotalItems {
			t.Errorf("expect %d, got %d", test.expected.TotalItems, len(got.Items))
		}
	}
}

var (
	eip1 = &porterv1alpha2.Eip{
		ObjectMeta: metav1.ObjectMeta{
			Name: "eip1",
		},
	}
	eip2 = &porterv1alpha2.Eip{
		ObjectMeta: metav1.ObjectMeta{
			Name: "eip2",
		},
	}
)

func prepare(t *testing.T) v1alpha3.Interface {
	e := &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("crds")},
	}
	cfg, err := e.Start()
	if err != nil {
		t.Fatal(err)
	}

	sch := scheme.Scheme
	if err := porterv1alpha2.AddToScheme(sch); err != nil {
		t.Fatalf("unable add APIs to scheme: %v", err)
	}

	stopCh := make(chan struct{})

	ce, err := cache.New(cfg, cache.Options{Scheme: sch})
	if err != nil {
		t.Fatalf("failed to create cache")
	}
	go ce.Start(stopCh)
	ce.WaitForCacheSync(stopCh)

	c, err := client.New(cfg, client.Options{Scheme: sch})
	if err != nil {
		t.Fatalf("failed to create client")
	}
	err = c.Create(context.Background(), eip1)
	if err != nil {
		t.Fatalf("failed to create eip1")
	}
	err = c.Create(context.Background(), eip2)
	if err != nil {
		t.Fatalf("failed to create eip2")
	}

	return NewEipGetter(ce)
}
