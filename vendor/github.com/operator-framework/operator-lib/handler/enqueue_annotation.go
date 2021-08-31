// Copyright 2020 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	crtHandler "sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName("event_handler")

const (
	// NamespacedNameAnnotation is an annotation whose value encodes the name and namespace of a resource to
	// reconcile when a resource containing this annotation changes. Valid values are of the form
	// `<namespace>/<name>` for namespace-scoped owners and `<name>` for cluster-scoped owners.
	NamespacedNameAnnotation = "operator-sdk/primary-resource"
	// TypeAnnotation is an annotation whose value encodes the group and kind of a resource to reconcil when a
	// resource containing this annotation changes. Valid values are of the form `<Kind>` for resource in the
	// core group, and `<Kind>.<group>` for all other resources.
	TypeAnnotation = "operator-sdk/primary-resource-type"
)

// EnqueueRequestForAnnotation enqueues Request containing the Name and Namespace specified in the
// annotations of the object that is the source of the Event. The source of the event triggers reconciliation
// of the parent resource which is identified by annotations. `NamespacedNameAnnotation` and
// `TypeAnnotation` together uniquely identify an owner resource to reconcile.
//
// handler.EnqueueRequestForAnnotation can be used to trigger reconciliation of resources which are
// cross-referenced.  This allows a namespace-scoped dependent to trigger reconciliation of an owner
// which is in a different namespace, and a cluster-scoped dependent can trigger the reconciliation
// of a namespace(scoped)-owner.
//
// As an example, consider the case where we would like to watch clusterroles based on which we reconcile
// namespace-scoped replicasets. With native owner references, this would not be possible since the
// cluster-scoped dependent (clusterroles) is trying to specify a namespace-scoped owner (replicasets).
// Whereas in case of annotations-based handlers, we could implement the following:
//
//	if err := c.Watch(&source.Kind{
//		// Watch clusterroles
//		Type: &rbacv1.ClusterRole{}},
//
//		// Enqueue ReplicaSet reconcile requests using the namespacedName annotation value in the request.
//		&handler.EnqueueRequestForAnnotation{schema.GroupKind{Group:"apps", Kind:"ReplicaSet"}}); err != nil {
//			entryLog.Error(err, "unable to watch ClusterRole")
//			os.Exit(1)
//		}
//	}
//
// With this watch, the ReplicaSet reconciler would receive a request to reconcile
// "my-namespace/my-replicaset" based on a change to a ClusterRole that has the following annotations:
//
//	annotations:
//		operator-sdk/primary-resource:"my-namespace/my-replicaset"
//		operator-sdk/primary-resource-type:"ReplicaSet.apps"
//
// Though an annotation-based watch handler removes the boundaries set by native owner reference implementation,
// the garbage collector still respects the scope restrictions. For example,
// if a parent creates a child resource across scopes not supported by owner references, it becomes the
// responsibility of the reconciler to clean up the child resource. Hence, the resource utilizing this handler
// SHOULD ALWAYS BE IMPLEMENTED WITH A FINALIZER.
type EnqueueRequestForAnnotation struct {
	Type schema.GroupKind
}

var _ crtHandler.EventHandler = &EnqueueRequestForAnnotation{}

// Create implements EventHandler
func (e *EnqueueRequestForAnnotation) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	if ok, req := e.getAnnotationRequests(evt.Object); ok {
		q.Add(req)
	}
}

// Update implements EventHandler
func (e *EnqueueRequestForAnnotation) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	if ok, req := e.getAnnotationRequests(evt.ObjectOld); ok {
		q.Add(req)
	}
	if ok, req := e.getAnnotationRequests(evt.ObjectNew); ok {
		q.Add(req)
	}
}

// Delete implements EventHandler
func (e *EnqueueRequestForAnnotation) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	if ok, req := e.getAnnotationRequests(evt.Object); ok {
		q.Add(req)
	}
}

// Generic implements EventHandler
func (e *EnqueueRequestForAnnotation) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	if ok, req := e.getAnnotationRequests(evt.Object); ok {
		q.Add(req)
	}
}

// getAnnotationRequests checks if the provided object has the annotations so as to enqueue the reconcile request.
func (e *EnqueueRequestForAnnotation) getAnnotationRequests(object metav1.Object) (bool, reconcile.Request) {
	if len(object.GetAnnotations()) == 0 {
		return false, reconcile.Request{}
	}

	if typeString, ok := object.GetAnnotations()[TypeAnnotation]; ok && typeString == e.Type.String() {
		namespacedNameString, ok := object.GetAnnotations()[NamespacedNameAnnotation]
		if !ok {
			log.Info("Unable to find namespaced name annotation for resource", "resource", object)
		}
		if strings.TrimSpace(namespacedNameString) == "" {
			return false, reconcile.Request{}
		}
		nsn := parseNamespacedName(namespacedNameString)
		return true, reconcile.Request{NamespacedName: nsn}
	}
	return false, reconcile.Request{}
}

// parseNamespacedName parses the provided string to extract the namespace and name into a
// types.NamespacedName. The edge case of empty string is handled prior to calling this function.
func parseNamespacedName(namespacedNameString string) types.NamespacedName {
	values := strings.SplitN(namespacedNameString, "/", 2)

	switch len(values) {
	case 1:
		return types.NamespacedName{Name: values[0]}
	default:
		return types.NamespacedName{Namespace: values[0], Name: values[1]}
	}
}

// SetOwnerAnnotations helps in adding 'NamespacedNameAnnotation' and 'TypeAnnotation' to object based on
// the values obtained from owner. The object gets the annotations from owner's namespace, name, group
// and kind. In other terms, object can be said to be the dependent having annotations from the owner.
// When a watch is set on the object, the annotations help to identify the owner and trigger reconciliation.
// Annotations are ALWAYS overwritten.
func SetOwnerAnnotations(owner, object client.Object) error {
	if owner.GetName() == "" {
		return fmt.Errorf("%T does not have a name, cannot call SetOwnerAnnotations", owner)
	}

	ownerGK := owner.GetObjectKind().GroupVersionKind().GroupKind()

	if ownerGK.Kind == "" {
		return fmt.Errorf("Owner %s Kind not found, cannot call SetOwnerAnnotations", owner.GetName())
	}

	annotations := object.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[NamespacedNameAnnotation] = fmt.Sprintf("%s/%s", owner.GetNamespace(), owner.GetName())
	annotations[TypeAnnotation] = ownerGK.String()

	object.SetAnnotations(annotations)

	return nil
}
