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

package persistentvolume

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

var (
	testStorageClassName = "test-csi"
)

func TestListPods(t *testing.T) {
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
				Filters:   map[query.Field]query.Value{},
			},
			&api.ListResult{
				Items:      []interface{}{pv3, pv2, pv1},
				TotalItems: len(persistentVolumes),
			},
			nil,
		},
		{
			"test status filter",
			"",
			&query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters: map[query.Field]query.Value{
					query.FieldStatus: query.Value(pv1.Status.Phase),
				},
			},
			&api.ListResult{
				Items:      []interface{}{pv1},
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
			t.Errorf("[%s] %T differ (-got, +want): %s", test.description, test.expected, diff)
		}
	}
}

var (
	pv1 = &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "pv-1",
		},
		Status: corev1.PersistentVolumeStatus{
			Phase: "bound",
		},
	}
	pv2 = &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pv-2",
		},
		Status: corev1.PersistentVolumeStatus{
			Phase: "available",
		},
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName: testStorageClassName,
		},
	}
	pv3 = &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc-3",
			Namespace: "default",
		},
	}

	persistentVolumes = []interface{}{pv1, pv2, pv3}
)

func prepare() v1alpha3.Interface {
	client := fake.NewSimpleClientset()
	informer := informers.NewSharedInformerFactory(client, 0)

	for _, pv := range persistentVolumes {
		_ = informer.Core().V1().PersistentVolumes().Informer().GetIndexer().Add(pv)
	}
	return New(informer)
}
