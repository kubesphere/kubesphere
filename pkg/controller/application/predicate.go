/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package application

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type IgnoreAnnotationChangePredicate struct {
	AnnotationKey string
}

func (p IgnoreAnnotationChangePredicate) Create(e event.CreateEvent) bool {
	return true
}

func (p IgnoreAnnotationChangePredicate) Delete(e event.DeleteEvent) bool {
	return true
}

func (p IgnoreAnnotationChangePredicate) Update(e event.UpdateEvent) bool {

	return e.ObjectOld.GetAnnotations()[p.AnnotationKey] == e.ObjectNew.GetAnnotations()[p.AnnotationKey]
}

func (p IgnoreAnnotationChangePredicate) Generic(e event.GenericEvent) bool {
	return true
}
