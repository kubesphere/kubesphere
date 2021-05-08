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

package ippool

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"kubesphere.io/api/network/v1alpha1"
	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

func TestListIPPools(t *testing.T) {
	tests := []struct {
		description string
		namespace   string
		query       *query.Query
		expected    *api.ListResult
		expectedErr error
	}{
		{
			"test name filter",
			"",
			&query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters: map[query.Field]query.Value{
					query.FieldName: query.Value("foo2"),
				},
			},
			&api.ListResult{
				Items:      []interface{}{foo2},
				TotalItems: 1,
			},
			nil,
		},
		{
			"test namespace filter",
			"ns1",
			&query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
			},
			&api.ListResult{
				Items:      []interface{}{foo1},
				TotalItems: 1,
			},
			nil,
		},
	}

	getter := prepare()

	for _, test := range tests {
		got, err := getter.List(test.namespace, test.query)
		if test.expectedErr != nil && err != test.expectedErr {
			t.Errorf("expected error, got nothing")
		} else if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(got, test.expected); diff != "" {
			t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
		}
	}
}

var (
	foo1 = &v1alpha1.IPPool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo1",
			Labels: map[string]string{
				constants.WorkspaceLabelKey: "wk1",
			},
		},
	}
	foo2 = &v1alpha1.IPPool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo2",
		},
	}
	foo3 = &v1alpha1.IPPool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo3",
		},
	}
	ns = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns1",
			Labels: map[string]string{
				constants.WorkspaceLabelKey: "wk1",
			},
		},
	}
	wk = &tenantv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wk1",
		},
	}
	ps = []interface{}{foo1, foo2, foo3}
)

func prepare() v1alpha3.Interface {

	client := fake.NewSimpleClientset()
	k8sClient := k8sfake.NewSimpleClientset()
	informer := informers.NewSharedInformerFactory(client, 0)
	k8sInformer := k8sinformers.NewSharedInformerFactory(k8sClient, 0)

	for _, p := range ps {
		informer.Network().V1alpha1().IPPools().Informer().GetIndexer().Add(p)
	}

	informer.Tenant().V1alpha1().Workspaces().Informer().GetIndexer().Add(wk)
	k8sInformer.Core().V1().Namespaces().Informer().GetIndexer().Add(ns)

	return New(informer, k8sInformer)
}
