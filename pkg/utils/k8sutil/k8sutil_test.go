/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package k8sutil

import (
	"reflect"
	"testing"

	"kubesphere.io/api/tenant/v1beta1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
					APIVersion: v1beta1.SchemeGroupVersion.String(),
					Kind:       v1beta1.ResourceKindWorkspace,
					Name:       "workspace-test",
				}},
				kind: v1beta1.ResourceKindWorkspace,
			},
			want: true,
		},
		{
			name: "controlled by workspace-test",
			args: args{
				ownerReferences: []metav1.OwnerReference{{
					APIVersion: v1beta1.SchemeGroupVersion.String(),
					Kind:       v1beta1.ResourceKindWorkspace,
					Name:       "workspace-test",
				}},
				kind: v1beta1.ResourceKindWorkspace,
				name: "workspace-test",
			},
			want: true,
		},
		{
			name: "not controlled by workspace-test",
			args: args{
				ownerReferences: []metav1.OwnerReference{{
					APIVersion: v1beta1.SchemeGroupVersion.String(),
					Kind:       v1beta1.ResourceKindWorkspace,
					Name:       "workspace",
				}},
				kind: v1beta1.ResourceKindWorkspace,
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
				APIVersion: v1beta1.SchemeGroupVersion.String(),
				Kind:       v1beta1.ResourceKindWorkspace,
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
