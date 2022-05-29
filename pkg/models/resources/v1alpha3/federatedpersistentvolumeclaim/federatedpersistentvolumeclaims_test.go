package federatedpersistentvolumeclaim

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/google/go-cmp/cmp"

	fedv1beta1 "kubesphere.io/api/types/v1beta1"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

var (
	testStorageClassName = "sc1"
)

var (
	pvc1 = &fedv1beta1.FederatedPersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc-1",
			Namespace: "default",
		},
	}

	pvc2 = &fedv1beta1.FederatedPersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc-2",
			Namespace: "default",
		},
		Spec: fedv1beta1.FederatedPersistentVolumeClaimSpec{
			Template: fedv1beta1.PersistentVolumeClaimTemplate{
				Spec: corev1.PersistentVolumeClaimSpec{
					StorageClassName: &testStorageClassName,
				},
			},
		},
	}

	pvc3 = &fedv1beta1.FederatedPersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc-3",
			Namespace: "default",
			Labels: map[string]string{
				"kubesphere.io/in-use": "false",
			},
		},
		Spec: fedv1beta1.FederatedPersistentVolumeClaimSpec{
			Template: fedv1beta1.PersistentVolumeClaimTemplate{
				Spec: corev1.PersistentVolumeClaimSpec{
					StorageClassName: &testStorageClassName,
				},
			},
		},
	}

	federatedPersistentVolumeClaims = []*fedv1beta1.FederatedPersistentVolumeClaim{pvc1, pvc2, pvc3}
)

func fedPVCsToInterface(federatedPersistentVolumeClaims ...*fedv1beta1.FederatedPersistentVolumeClaim) []interface{} {
	items := make([]interface{}, 0)

	for _, fedPVC := range federatedPersistentVolumeClaims {
		items = append(items, fedPVC)
	}

	return items
}

func fedPVCsToRuntimeObject(federatedPersistentVolumeClaims ...*fedv1beta1.FederatedPersistentVolumeClaim) []runtime.Object {
	items := make([]runtime.Object, 0)

	for _, fedPVC := range federatedPersistentVolumeClaims {
		items = append(items, fedPVC)
	}

	return items
}

func prepare() (v1alpha3.Interface, error) {
	client := fake.NewSimpleClientset(fedPVCsToRuntimeObject(federatedPersistentVolumeClaims...)...)
	informer := externalversions.NewSharedInformerFactory(client, 0)

	for _, fedPVC := range federatedPersistentVolumeClaims {
		err := informer.Types().V1beta1().FederatedPersistentVolumeClaims().Informer().GetIndexer().Add(fedPVC)
		if err != nil {
			return nil, err
		}
	}

	return New(informer), nil
}

func TestGet(t *testing.T) {
	tests := []struct {
		namespace   string
		name        string
		expected    runtime.Object
		expectedErr error
	}{
		{
			namespace:   "default",
			name:        "pvc-1",
			expected:    fedPVCsToRuntimeObject(pvc1)[0],
			expectedErr: nil,
		},
	}

	getter, err := prepare()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		pvc, err := getter.Get(test.namespace, test.name)
		if test.expectedErr != nil && err != test.expectedErr {
			t.Errorf("expected error, got nothing")
		} else if err != nil {
			t.Fatal(err)
		}

		diff := cmp.Diff(pvc, test.expected)
		if diff != "" {
			t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
		}
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		description string
		namespace   string
		query       *query.Query
		expected    *api.ListResult
		expectedErr error
	}{
		{
			description: "test name filter",
			namespace:   "default",
			query: &query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters:   map[query.Field]query.Value{query.FieldName: query.Value(pvc1.Name)},
			},
			expected: &api.ListResult{
				Items:      fedPVCsToInterface(federatedPersistentVolumeClaims[0]),
				TotalItems: 1,
			},
			expectedErr: nil,
		},
		{
			description: "test storageClass filter",
			namespace:   "default",
			query: &query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters:   map[query.Field]query.Value{query.Field(storageClassName): query.Value(*pvc2.Spec.Template.Spec.StorageClassName)},
			},
			expected: &api.ListResult{
				Items:      fedPVCsToInterface(federatedPersistentVolumeClaims[2], federatedPersistentVolumeClaims[1]),
				TotalItems: 2,
			},
			expectedErr: nil,
		},
		{
			description: "test label filter",
			namespace:   "default",
			query: &query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:        query.FieldName,
				Ascending:     false,
				LabelSelector: "kubesphere.io/in-use=false",
				Filters:       map[query.Field]query.Value{query.Field(storageClassName): query.Value(*pvc2.Spec.Template.Spec.StorageClassName)},
			},
			expected: &api.ListResult{
				Items:      fedPVCsToInterface(federatedPersistentVolumeClaims[2]),
				TotalItems: 1,
			},
			expectedErr: nil,
		},
	}

	lister, err := prepare()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		got, err := lister.List(test.namespace, test.query)
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
