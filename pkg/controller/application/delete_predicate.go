/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package application

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type DeletePredicate struct {
	predicate.Funcs
}

func (DeletePredicate) Update(e event.UpdateEvent) bool {
	return false
}

func (DeletePredicate) Create(_ event.CreateEvent) bool {
	return false
}

func (DeletePredicate) Delete(_ event.DeleteEvent) bool {
	return true
}

func (DeletePredicate) Generic(_ event.GenericEvent) bool {
	return false
}
