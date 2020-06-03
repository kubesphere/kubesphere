/*
Copyright 2019 The KubeSphere Authors.

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

package user

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type usersGetter struct {
	ksInformer  ksinformers.SharedInformerFactory
	k8sInformer k8sinformers.SharedInformerFactory
}

func New(ksinformer ksinformers.SharedInformerFactory, k8sinformer k8sinformers.SharedInformerFactory) v1alpha3.Interface {
	return &usersGetter{ksInformer: ksinformer, k8sInformer: k8sinformer}
}

func (d *usersGetter) Get(_, name string) (runtime.Object, error) {
	return d.ksInformer.Iam().V1alpha2().Users().Lister().Get(name)
}

func (d *usersGetter) List(_ string, query *query.Query) (*api.ListResult, error) {

	var users []*iamv1alpha2.User
	var err error

	if namespace := query.Filters[iamv1alpha2.ScopeNamespace]; namespace != "" {
		role := query.Filters[iamv1alpha2.ResourcesSingularRole]
		users, err = d.listAllUsersInNamespace(string(namespace), string(role))
		delete(query.Filters, iamv1alpha2.ScopeNamespace)
		delete(query.Filters, iamv1alpha2.ResourcesSingularRole)
	} else if workspace := query.Filters[iamv1alpha2.ScopeWorkspace]; workspace != "" {
		workspaceRole := query.Filters[iamv1alpha2.ResourcesSingularWorkspaceRole]
		users, err = d.listAllUsersInWorkspace(string(workspace), string(workspaceRole))
		delete(query.Filters, iamv1alpha2.ScopeWorkspace)
		delete(query.Filters, iamv1alpha2.ResourcesSingularWorkspaceRole)
	} else if cluster := query.Filters[iamv1alpha2.ScopeCluster]; cluster == "true" {
		clusterRole := query.Filters[iamv1alpha2.ResourcesSingularClusterRole]
		users, err = d.listAllUsersInCluster(string(clusterRole))
		delete(query.Filters, iamv1alpha2.ScopeCluster)
		delete(query.Filters, iamv1alpha2.ResourcesSingularClusterRole)
	} else if globalRole := query.Filters[iamv1alpha2.ResourcesSingularGlobalRole]; globalRole != "" {
		users, err = d.listAllUsersByGlobalRole(string(globalRole))
		delete(query.Filters, iamv1alpha2.ResourcesSingularGlobalRole)
	} else {
		users, err = d.ksInformer.Iam().V1alpha2().Users().Lister().List(query.Selector())
	}

	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, user := range users {
		result = append(result, user)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *usersGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftUser, ok := left.(*iamv1alpha2.User)
	if !ok {
		return false
	}

	rightUser, ok := right.(*iamv1alpha2.User)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftUser.ObjectMeta, rightUser.ObjectMeta, field)
}

func (d *usersGetter) filter(object runtime.Object, filter query.Filter) bool {
	user, ok := object.(*iamv1alpha2.User)

	if !ok {
		return false
	}

	switch filter.Field {
	case iamv1alpha2.FieldEmail:
		return user.Spec.Email == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(user.ObjectMeta, filter)
	}
}

func (d *usersGetter) listAllUsersInWorkspace(workspace, role string) ([]*iamv1alpha2.User, error) {
	var users []*iamv1alpha2.User
	var err error
	workspaceRoleBindings, err := d.ksInformer.Iam().V1alpha2().
		WorkspaceRoleBindings().Lister().List(labels.SelectorFromValidatedSet(labels.Set{tenantv1alpha1.WorkspaceLabel: workspace}))

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, roleBinding := range workspaceRoleBindings {
		if role != "" && roleBinding.RoleRef.Name != role {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1alpha2.ResourceKindUser {

				if contains(users, subject.Name) {
					klog.Warningf("conflict role binding found: %s, username:%s", roleBinding.ObjectMeta.String(), subject.Name)
					continue
				}

				obj, err := d.Get("", subject.Name)

				if err != nil {
					if errors.IsNotFound(err) {
						klog.Warningf("orphan subject: %s", subject.String())
						continue
					}
					klog.Error(err)
					return nil, err
				}

				user := obj.(*iamv1alpha2.User)
				user = user.DeepCopy()
				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}
				user.Annotations[iamv1alpha2.WorkspaceRoleAnnotation] = roleBinding.RoleRef.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func (d *usersGetter) listAllUsersInNamespace(namespace, role string) ([]*iamv1alpha2.User, error) {
	var users []*iamv1alpha2.User
	var err error

	roleBindings, err := d.k8sInformer.Rbac().V1().
		RoleBindings().Lister().RoleBindings(namespace).List(labels.Everything())

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, roleBinding := range roleBindings {
		if role != "" && roleBinding.RoleRef.Name != role {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1alpha2.ResourceKindUser {
				if contains(users, subject.Name) {
					klog.Warningf("conflict role binding found: %s, username:%s", roleBinding.ObjectMeta.String(), subject.Name)
					continue
				}

				obj, err := d.Get("", subject.Name)

				if err != nil {
					if errors.IsNotFound(err) {
						klog.Warningf("orphan subject: %s", subject.String())
						continue
					}
					klog.Error(err)
					return nil, err
				}

				user := obj.(*iamv1alpha2.User)
				user = user.DeepCopy()
				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}
				user.Annotations[iamv1alpha2.RoleAnnotation] = roleBinding.RoleRef.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func (d *usersGetter) listAllUsersByGlobalRole(globalRole string) ([]*iamv1alpha2.User, error) {
	var users []*iamv1alpha2.User
	var err error

	globalRoleBindings, err := d.ksInformer.Iam().V1alpha2().
		GlobalRoleBindings().Lister().List(labels.Everything())

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, roleBinding := range globalRoleBindings {
		if roleBinding.RoleRef.Name != globalRole {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1alpha2.ResourceKindUser {

				if contains(users, subject.Name) {
					klog.Warningf("conflict role binding found: %s, username:%s", roleBinding.ObjectMeta.String(), subject.Name)
					continue
				}

				obj, err := d.Get("", subject.Name)

				if err != nil {
					if errors.IsNotFound(err) {
						klog.Warningf("orphan subject: %s", subject.String())
						continue
					}
					klog.Error(err)
					return nil, err
				}

				user := obj.(*iamv1alpha2.User)
				user = user.DeepCopy()
				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}
				user.Annotations[iamv1alpha2.GlobalRoleAnnotation] = roleBinding.RoleRef.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func (d *usersGetter) listAllUsersInCluster(clusterRole string) ([]*iamv1alpha2.User, error) {
	var users []*iamv1alpha2.User
	var err error

	roleBindings, err := d.k8sInformer.Rbac().V1().ClusterRoleBindings().Lister().List(labels.Everything())

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, roleBinding := range roleBindings {
		if clusterRole != "" && roleBinding.RoleRef.Name != clusterRole {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1alpha2.ResourceKindUser {
				if contains(users, subject.Name) {
					klog.Warningf("conflict role binding found: %s, username:%s", roleBinding.ObjectMeta.String(), subject.Name)
					continue
				}

				obj, err := d.Get("", subject.Name)

				if err != nil {
					if errors.IsNotFound(err) {
						klog.Warningf("orphan subject: %s", subject.String())
						continue
					}
					klog.Error(err)
					return nil, err
				}

				user := obj.(*iamv1alpha2.User)
				user = user.DeepCopy()
				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}
				user.Annotations[iamv1alpha2.ClusterRoleAnnotation] = roleBinding.RoleRef.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func contains(users []*iamv1alpha2.User, username string) bool {
	for _, user := range users {
		if user.Name == username {
			return true
		}
	}
	return false
}
