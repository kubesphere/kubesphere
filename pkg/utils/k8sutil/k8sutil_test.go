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

package k8sutil

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"
)

func TestIsControlledBy(t *testing.T) {
	type args struct {
		ownerReferences []metav1.OwnerReference
		kind            string
		name            string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "controlled by Workspace",
			args: args{
				ownerReferences: []metav1.OwnerReference{{
					APIVersion: tenantv1alpha1.SchemeGroupVersion.String(),
					Kind:       tenantv1alpha1.ResourceKindWorkspace,
					Name:       "workspace-test",
				}},
				kind: tenantv1alpha1.ResourceKindWorkspace,
			},
			want: true,
		},
		{
			name: "controlled by workspace-test",
			args: args{
				ownerReferences: []metav1.OwnerReference{{
					APIVersion: tenantv1alpha1.SchemeGroupVersion.String(),
					Kind:       tenantv1alpha1.ResourceKindWorkspace,
					Name:       "workspace-test",
				}},
				kind: tenantv1alpha1.ResourceKindWorkspace,
				name: "workspace-test",
			},
			want: true,
		},
		{
			name: "not controlled by workspace-test",
			args: args{
				ownerReferences: []metav1.OwnerReference{{
					APIVersion: tenantv1alpha1.SchemeGroupVersion.String(),
					Kind:       tenantv1alpha1.ResourceKindWorkspace,
					Name:       "workspace",
				}},
				kind: tenantv1alpha1.ResourceKindWorkspace,
				name: "workspace-test",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsControlledBy(tt.args.ownerReferences, tt.args.kind, tt.args.name); got != tt.want {
				t.Errorf("IsControlledBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveWorkspaceOwnerReference(t *testing.T) {
	type args struct {
		ownerReferences []metav1.OwnerReference
	}
	tests := []struct {
		name string
		args args
		want []metav1.OwnerReference
	}{
		{
			name: "remove workspace owner reference",
			args: args{ownerReferences: []metav1.OwnerReference{{
				APIVersion: tenantv1alpha1.SchemeGroupVersion.String(),
				Kind:       tenantv1alpha1.ResourceKindWorkspace,
				Name:       "workspace-test",
			}}},
			want: []metav1.OwnerReference{},
		},
		{
			name: "remove workspace owner reference",
			args: args{ownerReferences: []metav1.OwnerReference{{
				APIVersion: corev1.SchemeGroupVersion.String(),
				Kind:       "Namespace",
				Name:       "namespace",
			}}},
			want: []metav1.OwnerReference{{
				APIVersion: corev1.SchemeGroupVersion.String(),
				Kind:       "Namespace",
				Name:       "namespace",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveWorkspaceOwnerReference(tt.args.ownerReferences); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveWorkspaceOwnerReference() = %v, want %v", got, tt.want)
			}
		})
	}
}
