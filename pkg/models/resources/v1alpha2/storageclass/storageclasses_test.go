package storageclass

import (
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/pointer"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
	"testing"
)

var (
	sc1 = &v1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sc1",
		},
	}

	scs = []interface{}{sc1}

	pvc1 = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc1",
			Namespace: "default",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName:       "pvc1-volume",
			StorageClassName: pointer.StringPtr("sc1"),
		},
	}

	pvc2 = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc2",
			Namespace: "default",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName: "pvc2-volume",
		},
	}

	pvcs = []interface{}{pvc1, pvc2}
)

func prepare() (v1alpha2.Interface, error) {
	client := fake.NewSimpleClientset()
	informer := informers.NewSharedInformerFactory(client, 0)

	for _, sc := range scs {
		err := informer.Storage().V1().StorageClasses().Informer().GetIndexer().Add(sc)
		if err != nil {
			return nil, err
		}
	}

	for _, pvc := range pvcs {
		err := informer.Core().V1().PersistentVolumeClaims().Informer().GetIndexer().Add(pvc)
		if err != nil {
			return nil, err
		}
	}

	return NewStorageClassesSearcher(informer, nil), nil
}

func TestSearch(t *testing.T) {
	tests := []struct {
		namespace   string
		name        string
		conditions  *params.Conditions
		orderBy     string
		reverse     bool
		expected    interface{}
		expectedErr error
	}{
		{
			namespace:   "default",
			name:        sc1.Name,
			conditions:  &params.Conditions{},
			orderBy:     "name",
			reverse:     true,
			expected:    scs,
			expectedErr: nil,
		},
	}

	searcher, err := prepare()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		got, err := searcher.Search(test.namespace, test.conditions, test.orderBy, test.reverse)
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
