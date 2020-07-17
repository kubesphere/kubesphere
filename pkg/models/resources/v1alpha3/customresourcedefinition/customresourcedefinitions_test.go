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

package customresourcedefinition

import (
	"github.com/google/go-cmp/cmp"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	fakeapiextensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	apiextensionsinformers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"testing"
)

var crds = []*v1beta1.CustomResourceDefinition{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "clusters.cluster.kubesphere.io",
			Labels: map[string]string{
				"controller-tools.k8s.io": "1.0",
			},
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "workspaces.tenant.kubesphere.io",
			Labels: map[string]string{
				"controller-tools.k8s.io": "1.0",
			},
		},
	},
}

func crdsToRuntimeObjects(crds ...*v1beta1.CustomResourceDefinition) []runtime.Object {
	items := make([]runtime.Object, 0)

	for _, crd := range crds {
		items = append(items, crd)
	}

	return items
}

func crdsToInterface(crds ...*v1beta1.CustomResourceDefinition) []interface{} {
	items := make([]interface{}, 0)

	for _, crd := range crds {
		items = append(items, crd)
	}

	return items
}

func TestCrdGetterList(t *testing.T) {
	var testCases = []struct {
		description string
		query       *query.Query
		expected    *api.ListResult
	}{
		{
			description: "Test normal case",
			query: &query.Query{
				Filters: map[query.Field]query.Value{
					query.FieldName: "clusters.cluster.kubesphere.io",
				},
			},
			expected: &api.ListResult{
				TotalItems: 1,
				Items:      crdsToInterface(crds[0]),
			},
		},
	}

	client := fakeapiextensions.NewSimpleClientset(crdsToRuntimeObjects(crds...)...)
	informers := apiextensionsinformers.NewSharedInformerFactory(client, 0)

	for _, crd := range crds {
		informers.Apiextensions().V1beta1().CustomResourceDefinitions().Informer().GetIndexer().Add(crd)
	}

	for _, testCase := range testCases {

		crdGetter := New(informers)

		t.Run(testCase.description, func(t *testing.T) {
			result, err := crdGetter.List("", testCase.query)

			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(result, testCase.expected); len(diff) != 0 {
				t.Errorf("%T, got+ expected-, %s", testCase.expected, diff)
			}
		})
	}
}
