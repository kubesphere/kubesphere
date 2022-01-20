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

package resource

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	fakesnapshot "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned/fake"
	fakeistio "istio.io/client-go/pkg/clientset/versioned/fake"
	fakeapiextensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/informers"
)

func TestResourceGetter_List(t *testing.T) {

	resource := prepare()

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
				Filters:   map[query.Field]query.Value{},
			},
			ExpectError: nil,
			ExpectResponse: &api.ListResult{
				Items:      []interface{}{foo2, foo1, bar1},
				TotalItems: 3,
			},
		},
		{
			Name:      "legecy case",
			Resource:  "pods",
			Namespace: "foo",
			Query: &query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters:   map[query.Field]query.Value{},
			},
			ExpectError: nil,
			ExpectResponse: &api.ListResult{
				Items:      []interface{}{pod3, pod2, pod1},
				TotalItems: 3,
			},
		},
	}

	for _, test := range tests {

		result, err := resource.List(test.Resource, test.Namespace, test.Query)

		if err != test.ExpectError {
			t.Errorf("expected error: %s, got: %s", test.ExpectError, err)
		}
		if diff := cmp.Diff(test.ExpectResponse, result); diff != "" {
			t.Errorf(diff)
		}
	}
}

var (
	foo1 = &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo1",
		},
	}

	foo2 = &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo2",
		},
	}
	bar1 = &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "bar1",
		},
	}

	namespaces = []interface{}{foo1, foo2, bar1}

	pod1 = &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "foo",
		},
	}

	pod2 = &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "foo",
		},
	}

	pod3 = &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod3",
			Namespace: "foo",
		},
	}

	pods = []interface{}{pod1, pod2, pod3}
)

func prepare() ResourceGetter {

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	istioClient := fakeistio.NewSimpleClientset()
	snapshotClient := fakesnapshot.NewSimpleClientset()
	apiextensionsClient := fakeapiextensions.NewSimpleClientset()
	fakeInformerFactory := informers.NewInformerFactories(k8sClient, ksClient, istioClient, snapshotClient, apiextensionsClient, nil)

	for _, pod := range pods {
		fakeInformerFactory.KubernetesSharedInformerFactory().Core().V1().
			Pods().Informer().GetIndexer().Add(pod)
	}

	var resourceToGVK map[string]schema.GroupVersionKind = map[string]schema.GroupVersionKind{
		"namespaces": {Group: "", Version: "v1", Kind: "Namespace"},
	}
	var Scheme = runtime.NewScheme()
	// v1alpha1.AddToScheme(Scheme)
	corev1.AddToScheme(Scheme)

	c := fake.NewClientBuilder().WithScheme(Scheme).WithObjects(foo1, foo2, bar1).Build()
	return NewResourceGetterWithKind(fakeInformerFactory, c.Scheme(), c, resourceToGVK)
}

func TestResourceGetter_Get(t *testing.T) {

	resource := prepare()

	type args struct {
		resource  string
		namespace string
		name      string
	}
	tests := []struct {
		name    string
		args    args
		want    runtime.Object
		wantErr bool
	}{
		{
			name: "Get namespace",
			args: args{
				resource: "namespaces",
				name:     "foo1",
			},
			want: foo1,
		},
		{
			name: "Get Pod",
			args: args{
				resource:  "pods",
				name:      "pod1",
				namespace: "foo",
			},
			want: pod1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := resource.Get(tt.args.resource, tt.args.namespace, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResourceGetter.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
