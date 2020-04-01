/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package resource

import (
	"github.com/google/go-cmp/cmp"
	fakeapp "github.com/kubernetes-sigs/application/pkg/client/clientset/versioned/fake"
	fakeistio "istio.io/client-go/pkg/clientset/versioned/fake"
	v1 "k8s.io/api/core/v1"
	corev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/informers"
	"testing"
)

func TestResourceGetter(t *testing.T) {

	namespaces := make([]interface{}, 0)
	defaultNamespace := &v1.Namespace{
		ObjectMeta: corev1.ObjectMeta{
			Name:   "default",
			Labels: map[string]string{"kubesphere.io/workspace": "system-workspace"},
		},
	}
	kubesphereNamespace := &v1.Namespace{
		ObjectMeta: corev1.ObjectMeta{
			Name:   "kubesphere-system",
			Labels: map[string]string{"kubesphere.io/workspace": "system-workspace"},
		},
	}

	namespaces = append(namespaces, defaultNamespace, kubesphereNamespace)

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset(defaultNamespace, kubesphereNamespace)
	istioClient := fakeistio.NewSimpleClientset()
	appClient := fakeapp.NewSimpleClientset()
	fakeInformerFactory := informers.NewInformerFactories(k8sClient, ksClient, istioClient, appClient)

	k8sInformerFactory := fakeInformerFactory.KubernetesSharedInformerFactory()
	for _, namespace := range namespaces {
		err := k8sInformerFactory.Core().V1().Namespaces().Informer().GetIndexer().Add(namespace)
		if err != nil {
			t.Fatal(err)
		}
	}

	resource := New(fakeInformerFactory)

	tests := []struct {
		Name           string
		Resource       string
		Namespace      string
		Query          *query.Query
		ExpectError    error
		ExpectResponse *api.ListResult
	}{
		{
			Name:      "normal case",
			Resource:  "namespaces",
			Namespace: "",
			Query: &query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters:   []query.Filter{},
			},
			ExpectError: nil,
			ExpectResponse: &api.ListResult{
				Items:      namespaces,
				TotalItems: 2,
			},
		},
	}

	for _, test := range tests {

		result, err := resource.List(test.Resource, test.Namespace, test.Query)

		t.Logf("%+v", result)
		if err != test.ExpectError {
			t.Errorf("expected error: %s, got: %s", test.ExpectError, err)
		}
		if diff := cmp.Diff(test.ExpectResponse, result); diff != "" {
			t.Errorf(diff)
		}
	}

}
