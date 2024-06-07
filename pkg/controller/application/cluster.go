/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package application

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type ClusterDeletePredicate struct {
	predicate.Funcs
}

func (ClusterDeletePredicate) Update(e event.UpdateEvent) bool {
	return false
}

func (ClusterDeletePredicate) Create(_ event.CreateEvent) bool {
	return false
}

func (ClusterDeletePredicate) Delete(_ event.DeleteEvent) bool {
	return true
}

func (ClusterDeletePredicate) Generic(_ event.GenericEvent) bool {
	return false
}
