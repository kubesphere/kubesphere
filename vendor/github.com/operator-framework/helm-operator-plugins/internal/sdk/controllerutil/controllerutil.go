/*
Copyright 2020 The Operator-SDK Authors.

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

package controllerutil

import (
	"context"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	AddFinalizer      = controllerutil.AddFinalizer
	RemoveFinalizer   = controllerutil.RemoveFinalizer
	ContainsFinalizer = func(obj metav1.Object, finalizer string) bool {
		for _, f := range obj.GetFinalizers() {
			if f == finalizer {
				return true
			}
		}
		return false
	}
)

func WaitForDeletion(ctx context.Context, cl client.Reader, o client.Object) error {
	key := client.ObjectKeyFromObject(o)

	return wait.PollImmediateUntil(time.Millisecond*10, func() (bool, error) {
		err := cl.Get(ctx, key, o)
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}, ctx.Done())
}

func SupportsOwnerReference(restMapper meta.RESTMapper, owner, dependent client.Object) (bool, error) {
	ownerGVK := owner.GetObjectKind().GroupVersionKind()
	ownerMapping, err := restMapper.RESTMapping(ownerGVK.GroupKind(), ownerGVK.Version)
	if err != nil {
		return false, err
	}

	depGVK := dependent.GetObjectKind().GroupVersionKind()
	depMapping, err := restMapper.RESTMapping(depGVK.GroupKind(), depGVK.Version)
	if err != nil {
		return false, err
	}

	ownerClusterScoped := ownerMapping.Scope.Name() == meta.RESTScopeNameRoot
	ownerNamespace := owner.GetNamespace()
	depClusterScoped := depMapping.Scope.Name() == meta.RESTScopeNameRoot

	depNamespace := dependent.GetNamespace()

	if ownerClusterScoped {
		return true, nil
	}

	if depClusterScoped {
		return false, nil
	}

	if ownerNamespace != depNamespace {
		return false, nil
	}

	return true, nil
}
