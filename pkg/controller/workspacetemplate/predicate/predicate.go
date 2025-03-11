/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package predicate

import (
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	tenantv1alpha1 "kubesphere.io/api/tenant/v1beta1"
)

type WorkspaceStatusChangedPredicate struct {
	predicate.Funcs
}

func (WorkspaceStatusChangedPredicate) Update(e event.UpdateEvent) bool {
	oldWorkspaceTemplate, ok := e.ObjectOld.(*tenantv1alpha1.WorkspaceTemplate)
	if !ok {
		return false
	}
	newWorkspaceTemplate, ok := e.ObjectNew.(*tenantv1alpha1.WorkspaceTemplate)
	if !ok {
		return false
	}
	if !reflect.DeepEqual(oldWorkspaceTemplate.Spec, newWorkspaceTemplate.Spec) {
		return true
	}
	return false
}

func (WorkspaceStatusChangedPredicate) Create(_ event.CreateEvent) bool {
	return false
}

func (WorkspaceStatusChangedPredicate) Delete(_ event.DeleteEvent) bool {
	return false
}

func (WorkspaceStatusChangedPredicate) Generic(_ event.GenericEvent) bool {
	return false
}
