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

// If you make changes to this file, you should also make the corresponding change in ReplicationController.

package replicaset

import (
	"context"
	"fmt"

	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

// Registry is an interface for things that know how to store ReplicaSets.
type Registry interface {
	ListReplicaSets(ctx context.Context, options *metainternalversion.ListOptions) (*extensions.ReplicaSetList, error)
	WatchReplicaSets(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error)
	GetReplicaSet(ctx context.Context, replicaSetID string, options *metav1.GetOptions) (*extensions.ReplicaSet, error)
	CreateReplicaSet(ctx context.Context, replicaSet *extensions.ReplicaSet, createValidation rest.ValidateObjectFunc) (*extensions.ReplicaSet, error)
	UpdateReplicaSet(ctx context.Context, replicaSet *extensions.ReplicaSet, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc) (*extensions.ReplicaSet, error)
	DeleteReplicaSet(ctx context.Context, replicaSetID string) error
}

// storage puts strong typing around storage calls
type storage struct {
	rest.StandardStorage
}

// NewRegistry returns a new Registry interface for the given Storage. Any mismatched
// types will panic.
func NewRegistry(s rest.StandardStorage) Registry {
	return &storage{s}
}

func (s *storage) ListReplicaSets(ctx context.Context, options *metainternalversion.ListOptions) (*extensions.ReplicaSetList, error) {
	if options != nil && options.FieldSelector != nil && !options.FieldSelector.Empty() {
		return nil, fmt.Errorf("field selector not supported yet")
	}
	obj, err := s.List(ctx, options)
	if err != nil {
		return nil, err
	}
	return obj.(*extensions.ReplicaSetList), err
}

func (s *storage) WatchReplicaSets(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	return s.Watch(ctx, options)
}

func (s *storage) GetReplicaSet(ctx context.Context, replicaSetID string, options *metav1.GetOptions) (*extensions.ReplicaSet, error) {
	obj, err := s.Get(ctx, replicaSetID, options)
	if err != nil {
		return nil, err
	}
	return obj.(*extensions.ReplicaSet), nil
}

func (s *storage) CreateReplicaSet(ctx context.Context, replicaSet *extensions.ReplicaSet, createValidation rest.ValidateObjectFunc) (*extensions.ReplicaSet, error) {
	obj, err := s.Create(ctx, replicaSet, createValidation, false)
	if err != nil {
		return nil, err
	}
	return obj.(*extensions.ReplicaSet), nil
}

func (s *storage) UpdateReplicaSet(ctx context.Context, replicaSet *extensions.ReplicaSet, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc) (*extensions.ReplicaSet, error) {
	obj, _, err := s.Update(ctx, replicaSet.Name, rest.DefaultUpdatedObjectInfo(replicaSet), createValidation, updateValidation)
	if err != nil {
		return nil, err
	}
	return obj.(*extensions.ReplicaSet), nil
}

func (s *storage) DeleteReplicaSet(ctx context.Context, replicaSetID string) error {
	_, _, err := s.Delete(ctx, replicaSetID, nil)
	return err
}
