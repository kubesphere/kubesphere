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

package persistentvolumeclaim

import (
	"github.com/google/go-cmp/cmp"
	snapshot "github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	snapshotefakeclient "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/clientset/versioned/fake"
	snapshotinformers "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/informers/externalversions"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"testing"
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
			"default",
			&query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters:   map[query.Field]query.Value{query.FieldNamespace: query.Value("default")},
			},
			&api.ListResult{
				Items:      []interface{}{pvc3, pvc2, pvc1},
				TotalItems: len(persistentVolumeClaims),
			},
			nil,
		},
		{
			"test status filter",
			"default",
			&query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters: map[query.Field]query.Value{
					query.FieldNamespace: query.Value("default"),
					query.FieldStatus:    query.Value(pvc1.Status.Phase),
				},
			},
			&api.ListResult{
				Items:      []interface{}{pvc1},
				TotalItems: 1,
			},
			nil,
		},
		{
			"test StorageClass filter and allow snapshot",
			"default",
			&query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters: map[query.Field]query.Value{
					query.FieldNamespace:          query.Value("default"),
					query.Field(storageClassName): query.Value(*pvc2.Spec.StorageClassName),
				},
			},
			&api.ListResult{
				Items:      []interface{}{pvcGet2},
				TotalItems: 1,
			},
			nil,
		},
		{
			"test pvc in use",
			"default",
			&query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters: map[query.Field]query.Value{
					query.FieldNamespace: query.Value("default"),
					query.FieldName:      query.Value(pvc3.Name),
				},
			},
			&api.ListResult{
				Items:      []interface{}{pvcGet3},
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
	pvc1 = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc-1",
			Namespace: "default",
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimPending,
		},
	}
	pvc2 = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc-2",
			Namespace: "default",
			Annotations: map[string]string{
				annotationStorageProvisioner: testStorageClassName,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &testStorageClassName,
		},
	}

	pvcGet2 = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc-2",
			Namespace: "default",
			Annotations: map[string]string{
				annotationInUse:              "false",
				annotationAllowSnapshot:      "true",
				annotationStorageProvisioner: testStorageClassName,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &testStorageClassName,
		},
	}

	pvc3 = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc-3",
			Namespace: "default",
		},
	}

	pvcGet3 = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc-3",
			Namespace: "default",
			Annotations: map[string]string{
				annotationInUse:         "true",
				annotationAllowSnapshot: "false",
			},
		},
	}

	pod1 = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvc3.Name,
						},
					},
				},
			},
		},
	}

	volumeSnapshotClass1 = &snapshot.VolumeSnapshotClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "VolumeSnapshotClass-1",
			Namespace: "default",
		},
		Driver: testStorageClassName,
	}

	persistentVolumeClaims = []interface{}{pvc1, pvc2, pvc3}
	pods                   = []interface{}{pod1}
	volumeSnapshotClasses  = []interface{}{volumeSnapshotClass1}
)

func prepare() v1alpha3.Interface {
	client := fake.NewSimpleClientset()
	informer := informers.NewSharedInformerFactory(client, 0)
	snapshotClient := snapshotefakeclient.NewSimpleClientset()
	snapshotInformers := snapshotinformers.NewSharedInformerFactory(snapshotClient, 0)

	for _, pvc := range persistentVolumeClaims {
		_ = informer.Core().V1().PersistentVolumeClaims().Informer().GetIndexer().Add(pvc)
	}
	for _, pod := range pods {
		_ = informer.Core().V1().Pods().Informer().GetIndexer().Add(pod)
	}
	for _, volumeSnapshotClass := range volumeSnapshotClasses {
		_ = snapshotInformers.Snapshot().V1beta1().VolumeSnapshotClasses().Informer().GetIndexer().Add(volumeSnapshotClass)
	}

	return New(informer, snapshotInformers)
}
