/*
Copyright 2022 The KubeSphere Authors.

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

package crds

import (
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/diff"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

type fakeClient struct {
	Client client.Client
}

// Get retrieves an obj for the given object key from the Kubernetes Cluster.
// obj must be a struct pointer so that obj can be updated with the response
// returned by the Server.
func (f *fakeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return f.Client.Get(ctx, key, obj)
}

// List retrieves list of objects for a given namespace and list options. On a
// successful call, Items field in the list will be populated with the
// result returned from the server.
func (f *fakeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return f.Client.List(ctx, list, opts...)
}

// GetInformer fetches or constructs an informer for the given object that corresponds to a single
// API kind and resource.
func (f *fakeClient) GetInformer(ctx context.Context, obj client.Object) (cache.Informer, error) {
	return nil, nil
}

// GetInformerForKind is similar to GetInformer, except that it takes a group-version-kind, instead
// of the underlying object.
func (f *fakeClient) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind) (cache.Informer, error) {
	return nil, nil
}

// Start runs all the informers known to this cache until the context is closed.
// It blocks.
func (f *fakeClient) Start(ctx context.Context) error {
	return nil
}

// WaitForCacheSync waits for all the caches to sync.  Returns false if it could not sync a cache.
func (f *fakeClient) WaitForCacheSync(ctx context.Context) bool {
	return false
}

func (f *fakeClient) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	return nil
}

var (
	cm1 = &corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "cm1",
			Namespace: "default",
		},
	}

	cm2 = &corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "cm2",
			Namespace: "default",
		},
	}

	cm3 = &corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "cm3",
			Namespace: "default",
		},
	}
)

func TestHandler_Get(t *testing.T) {

	var Scheme = runtime.NewScheme()
	// v1alpha1.AddToScheme(Scheme)
	corev1.AddToScheme(Scheme)

	c := fake.NewClientBuilder().WithScheme(Scheme).WithRuntimeObjects(cm1).Build()

	h := NewTyped(&fakeClient{c}, cm1.GroupVersionKind(), Scheme)

	type args struct {
		gvk schema.GroupVersionKind
		key types.NamespacedName
	}
	tests := []struct {
		name    string
		handler Client
		args    args
		want    client.Object
		wantErr bool
	}{
		{
			name:    "test",
			handler: h,
			args: args{
				key: types.NamespacedName{
					Namespace: "default",
					Name:      "cm1",
				},
			},
			want: cm1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler
			got, err := h.Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler.Get() \nDiff:\n %s", diff.ObjectGoPrintSideBySide(tt.want, got))
			}
		})
	}
}

func TestResourceGetter_List(t *testing.T) {

	var Scheme = runtime.NewScheme()
	// v1alpha1.AddToScheme(Scheme)
	corev1.AddToScheme(Scheme)

	c := fake.NewClientBuilder().WithScheme(Scheme).WithRuntimeObjects(cm1, cm2, cm3).Build()

	h := NewTyped(&fakeClient{c}, cm1.GroupVersionKind(), Scheme)

	tests := []struct {
		name      string
		Resource  string
		Namespace string
		Query     *query.Query
		wantErr   bool
		want      *api.ListResult
	}{
		{
			name:      "list configmaps",
			Resource:  "configmaps",
			Namespace: "default",
			Query: &query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters:   map[query.Field]query.Value{},
			},
			want: &api.ListResult{
				Items:      []interface{}{cm3, cm2, cm1},
				TotalItems: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.List(tt.Namespace, tt.Query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler.Get() \nDiff:\n %s", diff.ObjectGoPrintSideBySide(tt.want, got))
			}
		})
	}
}
