/*
Copyright 2016 The Kubernetes Authors.

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

package util

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Copies cluster-independent, user provided data from the given ObjectMeta struct. If in
// the future the ObjectMeta structure is expanded then any field that is not populated
// by the api server should be included here.
func copyObjectMeta(obj metav1.ObjectMeta) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:            obj.Name,
		Namespace:       obj.Namespace,
		Labels:          obj.Labels,
		Annotations:     obj.Annotations,
		ResourceVersion: obj.ResourceVersion,
	}
}

// Deep copies cluster-independent, user provided data from the given ObjectMeta struct. If in
// the future the ObjectMeta structure is expanded then any field that is not populated
// by the api server should be included here.
func DeepCopyRelevantObjectMeta(obj metav1.ObjectMeta) metav1.ObjectMeta {
	copyMeta := copyObjectMeta(obj)
	if obj.Labels != nil {
		copyMeta.Labels = make(map[string]string)
		for key, val := range obj.Labels {
			copyMeta.Labels[key] = val
		}
	}
	if obj.Annotations != nil {
		copyMeta.Annotations = make(map[string]string)
		for key, val := range obj.Annotations {
			copyMeta.Annotations[key] = val
		}
	}
	return copyMeta
}

// Checks if cluster-independent, user provided data in two given ObjectMeta are equal. If in
// the future the ObjectMeta structure is expanded then any field that is not populated
// by the api server should be included here.
func ObjectMetaEquivalent(a, b metav1.ObjectMeta) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Namespace != b.Namespace {
		return false
	}
	if !reflect.DeepEqual(a.Labels, b.Labels) && (len(a.Labels) != 0 || len(b.Labels) != 0) {
		return false
	}
	if !reflect.DeepEqual(a.Annotations, b.Annotations) && (len(a.Annotations) != 0 || len(b.Annotations) != 0) {
		return false
	}
	return true
}

// Checks if cluster-independent, user provided data in two given ObjectMeta are equal. If in
// the future the ObjectMeta structure is expanded then any field that is not populated
// by the api server should be included here.
func ObjectMetaObjEquivalent(a, b metav1.Object) bool {
	if a.GetName() != b.GetName() {
		return false
	}
	if a.GetNamespace() != b.GetNamespace() {
		return false
	}
	aLabels := a.GetLabels()
	bLabels := b.GetLabels()
	if !reflect.DeepEqual(aLabels, bLabels) && (len(aLabels) != 0 || len(bLabels) != 0) {
		return false
	}
	aAnnotations := a.GetAnnotations()
	bAnnotations := b.GetAnnotations()
	if !reflect.DeepEqual(aAnnotations, bAnnotations) && (len(aAnnotations) != 0 || len(bAnnotations) != 0) {
		return false
	}
	return true
}

// Checks if cluster-independent, user provided data in ObjectMeta and Spec in two given top
// level api objects are equivalent.
func ObjectMetaAndSpecEquivalent(a, b runtimeclient.Object) bool {
	objectMetaA := reflect.ValueOf(a).Elem().FieldByName("ObjectMeta").Interface().(metav1.ObjectMeta)
	objectMetaB := reflect.ValueOf(b).Elem().FieldByName("ObjectMeta").Interface().(metav1.ObjectMeta)
	specA := reflect.ValueOf(a).Elem().FieldByName("Spec").Interface()
	specB := reflect.ValueOf(b).Elem().FieldByName("Spec").Interface()
	return ObjectMetaEquivalent(objectMetaA, objectMetaB) && reflect.DeepEqual(specA, specB)
}

func MetaAccessor(obj runtimeclient.Object) metav1.Object {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		// This should always succeed if obj is not nil.  Also,
		// adapters are slated for replacement by unstructured.
		return nil
	}
	return accessor
}

// GetUnstructured return Unstructured for any given kubernetes type
func GetUnstructured(resource interface{}) (*unstructured.Unstructured, error) {
	content, err := json.Marshal(resource)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to JSON Marshal")
	}
	unstructuredResource := &unstructured.Unstructured{}
	err = unstructuredResource.UnmarshalJSON(content)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to UnmarshalJSON into unstructured content")
	}
	return unstructuredResource, nil
}
