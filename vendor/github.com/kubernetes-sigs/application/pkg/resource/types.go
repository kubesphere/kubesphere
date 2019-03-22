/*
Copyright 2018 The Kubernetes Authors
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

package resource

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Common const definitions
const (
	LifecycleManaged  = "managed"
	LifecycleReferred = "referred"
)

// ObjectBag abstracts dealing with group of objects
// For now it is a simple list
type ObjectBag struct {
	objects []Object
}

// Object is a container to capture the k8s resource info to be used by controller
type Object struct {
	// Lifecycle can be: managed, reference
	Lifecycle string
	// Obj refers to the resource object  can be: sts, service, secret, pvc, ..
	Obj metav1.Object
	// ObjList refers to the list of resource objects
	ObjList metav1.ListInterface
}

// Observable captures the k8s resource info and selector to fetch child resources
type Observable struct {
	// ObjList refers to the list of resource objects
	ObjList metav1.ListInterface
	// Obj refers to the resource object  can be: sts, service, secret, pvc, ..
	Obj metav1.Object
	// Labels list of labels
	Labels map[string]string
	// Typemeta - needed for go test fake client
	Type metav1.TypeMeta
}

// LocalObjectReference with validation
type LocalObjectReference struct {
	corev1.LocalObjectReference `json:",inline"`
}

// GetObjectFn is a type for any function that returns resource info
type GetObjectFn func(interface{}) (*Object, error)
