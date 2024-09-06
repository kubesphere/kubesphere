/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package job

import (
	"context"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"kubesphere.io/kubesphere/pkg/scheme"
)

func newJob(name string, spec batchv1.JobSpec) *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{APIVersion: batchv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: spec,
	}
}

func TestAddAnnotation(t *testing.T) {
	job := newJob("test", batchv1.JobSpec{})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(job).Build()

	reconciler := &Reconciler{}
	reconciler.Client = fakeClient

	tests := []struct {
		name  string
		req   types.NamespacedName
		isErr bool
	}{
		{
			name: "normal test",
			req: types.NamespacedName{
				Namespace: job.Namespace,
				Name:      job.Name,
			},
			isErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: tt.req}); tt.isErr != (err != nil) {
				t.Errorf("%s Reconcile() unexpected error: %v", tt.name, err)
			}
		})
	}
}
