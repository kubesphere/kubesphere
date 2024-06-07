/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package storageclass

import (
	"context"
	"testing"

	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newStorageClass(name string, provisioner string) *storagev1.StorageClass {
	isExpansion := true
	return &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Provisioner:          provisioner,
		AllowVolumeExpansion: &isExpansion,
	}
}

func newCSIDriver(name string) *storagev1.CSIDriver {
	return &storagev1.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func Test(t *testing.T) {
	storageClass1 := newStorageClass("csi-example", "csi.example.com")
	storageClass2 := storageClass1.DeepCopy()
	storageClass2.Annotations = map[string]string{annotationAllowSnapshot: "true", annotationAllowClone: "false"}
	csiDriver := newCSIDriver("csi.example.com")

	tests := []struct {
		name  string
		objs  []client.Object
		req   types.NamespacedName
		isErr bool
	}{
		{
			name: "has csi driver test",
			objs: []client.Object{storageClass1, csiDriver},
			req: types.NamespacedName{
				Name: storageClass1.Name,
			},
			isErr: false,
		},
		{
			name: "no csi driver test",
			objs: []client.Object{storageClass2},
			req: types.NamespacedName{
				Name: storageClass2.Name,
			},
			isErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(tt.objs...).Build()
			reconciler := &Reconciler{}
			reconciler.Client = fakeClient

			if _, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: tt.req}); tt.isErr != (err != nil) {
				t.Errorf("%s Reconcile() unexpected error: %v", tt.name, err)
			}
		})
	}
}
