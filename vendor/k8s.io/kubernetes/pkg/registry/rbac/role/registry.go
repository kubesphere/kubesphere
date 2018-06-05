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

package role

import (
	"context"

	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/kubernetes/pkg/apis/rbac"
)

// Registry is an interface for things that know how to store Roles.
type Registry interface {
	ListRoles(ctx context.Context, options *metainternalversion.ListOptions) (*rbac.RoleList, error)
	CreateRole(ctx context.Context, role *rbac.Role, createValidation rest.ValidateObjectFunc) error
	UpdateRole(ctx context.Context, role *rbac.Role, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc) error
	GetRole(ctx context.Context, name string, options *metav1.GetOptions) (*rbac.Role, error)
	DeleteRole(ctx context.Context, name string) error
	WatchRoles(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error)
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

func (s *storage) ListRoles(ctx context.Context, options *metainternalversion.ListOptions) (*rbac.RoleList, error) {
	obj, err := s.List(ctx, options)
	if err != nil {
		return nil, err
	}

	return obj.(*rbac.RoleList), nil
}

func (s *storage) CreateRole(ctx context.Context, role *rbac.Role, createValidation rest.ValidateObjectFunc) error {
	_, err := s.Create(ctx, role, createValidation, false)
	return err
}

func (s *storage) UpdateRole(ctx context.Context, role *rbac.Role, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc) error {
	// TODO: any admission?
	_, _, err := s.Update(ctx, role.Name, rest.DefaultUpdatedObjectInfo(role), createValidation, updateValidation)
	return err
}

func (s *storage) WatchRoles(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	return s.Watch(ctx, options)
}

func (s *storage) GetRole(ctx context.Context, name string, options *metav1.GetOptions) (*rbac.Role, error) {
	obj, err := s.Get(ctx, name, options)
	if err != nil {
		return nil, err
	}
	return obj.(*rbac.Role), nil
}

func (s *storage) DeleteRole(ctx context.Context, name string) error {
	_, _, err := s.Delete(ctx, name, nil)
	return err
}

// AuthorizerAdapter adapts the registry to the authorizer interface
type AuthorizerAdapter struct {
	Registry Registry
}

func (a AuthorizerAdapter) GetRole(namespace, name string) (*rbac.Role, error) {
	return a.Registry.GetRole(genericapirequest.WithNamespace(genericapirequest.NewContext(), namespace), name, &metav1.GetOptions{})
}
