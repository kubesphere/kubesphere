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
