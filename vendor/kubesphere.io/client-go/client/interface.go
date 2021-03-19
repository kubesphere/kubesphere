/*
Copyright 2020 KubeSphere Authors

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

package client

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// ObjectKey identifies a Kubernetes Object.
type ObjectKey = types.NamespacedName

// ObjectKeyFromObject returns the ObjectKey given a runtime.Object
func ObjectKeyFromObject(obj runtime.Object) (ObjectKey, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return ObjectKey{}, err
	}
	return ObjectKey{Namespace: accessor.GetNamespace(), Name: accessor.GetName()}, nil
}

// Patch is a patch that can be applied to a Kubernetes object.
type Patch interface {
	// Type is the PatchType of the patch.
	Type() types.PatchType
	// Data is the raw data representing the patch.
	Data(obj runtime.Object) ([]byte, error)
}

// Reader knows how to read and list Kubernetes objects.
type Reader interface {
	// Get retrieves an obj for the given object key from the Kubernetes Cluster.
	// obj must be a struct pointer so that obj can be updated with the response
	// returned by the Server.
	Get(ctx context.Context, key ObjectKey, obj runtime.Object, opts ...GetOption) error

	// List retrieves list of objects for a given namespace and list options. On a
	// successful call, Items field in the list will be populated with the
	// result returned from the server.
	List(ctx context.Context, list runtime.Object, opts ...ListOption) error
}

// Writer knows how to create, delete, and update Kubernetes objects.
type Writer interface {
	// Create saves the object obj in the Kubernetes cluster.
	Create(ctx context.Context, obj runtime.Object, opts ...CreateOption) error

	// Delete deletes the given obj from Kubernetes cluster.
	Delete(ctx context.Context, obj runtime.Object, opts ...DeleteOption) error

	// Update updates the given obj in the Kubernetes cluster. obj must be a
	// struct pointer so that obj can be updated with the content returned by the Server.
	Update(ctx context.Context, obj runtime.Object, opts ...UpdateOption) error

	// Patch patches the given obj in the Kubernetes cluster. obj must be a
	// struct pointer so that obj can be updated with the content returned by the Server.
	Patch(ctx context.Context, obj runtime.Object, patch Patch, opts ...PatchOption) error

	// DeleteAllOf deletes all objects of the given type matching the given options.
	DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...DeleteAllOfOption) error
}

// StatusClient knows how to create a client which can update status subresource
// for kubernetes objects.
type StatusClient interface {
	Status() StatusWriter
}

// StatusWriter knows how to update status subresource of a Kubernetes object.
type StatusWriter interface {
	// Update updates the fields corresponding to the status subresource for the
	// given obj. obj must be a struct pointer so that obj can be updated
	// with the content returned by the Server.
	Update(ctx context.Context, obj runtime.Object, opts ...UpdateOption) error

	// Patch patches the given object's subresource. obj must be a struct
	// pointer so that obj can be updated with the content returned by the
	// Server.
	Patch(ctx context.Context, obj runtime.Object, patch Patch, opts ...PatchOption) error
}

// Client knows how to perform CRUD operations on Kubernetes objects.
type Client interface {
	Reader
	Writer
	StatusClient
}

// IgnoreNotFound returns nil on NotFound errors.
// All other values that are not NotFound errors or nil are returned unmodified.
func IgnoreNotFound(err error) error {
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}
