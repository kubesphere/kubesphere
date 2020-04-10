/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */
package am

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"net/http"
)

type AccessManagementInterface interface {
	ListRolesOfUser(scope iamv1alpha2.Scope, username string) ([]iamv1alpha2.Role, error)
	GetRoleOfUserInTargetScope(scope iamv1alpha2.Scope, target string, username string) (*iamv1alpha2.Role, error)
	GetPolicyRule(name string) (*iamv1alpha2.PolicyRule, error)
}

type amOperator struct {
	informers informers.SharedInformerFactory
	ksClient  kubesphere.Interface
}

func NewAMOperator(ksClient kubesphere.Interface, informers informers.SharedInformerFactory) AccessManagementInterface {
	return &amOperator{
		informers: informers,
		ksClient:  ksClient,
	}
}

func containsUser(subjets []iamv1alpha2.Subject, username string) bool {
	for _, sub := range subjets {
		if sub.Kind == iamv1alpha2.UserKind && sub.Name == username {
			return true
		}
	}
	return false
}

func (am *amOperator) ListRolesOfUser(scope iamv1alpha2.Scope, username string) ([]iamv1alpha2.Role, error) {

	lister := am.informers.Iam().V1alpha2().RoleBindings().Lister()

	roleBindings, err := lister.List(labels.Everything())

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	roleBindingsInScope := filterRoleBindingByScope(roleBindings, scope)

	roles := make([]iamv1alpha2.Role, 0)

	for _, roleBinding := range roleBindingsInScope {
		if containsUser(roleBinding.Subjects, username) {
			role, err := am.informers.
				Iam().V1alpha2().Roles().Lister().Get(roleBinding.RoleRef.Name)

			if err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return nil, err
			}

			roles = append(roles, *role)
		}
	}

	return roles, nil
}

func filterRoleBindingByScope(roles []*iamv1alpha2.RoleBinding, scope iamv1alpha2.Scope) []*iamv1alpha2.RoleBinding {
	result := make([]*iamv1alpha2.RoleBinding, 0)
	for _, role := range roles {
		if role.Scope == scope {
			result = append(result, role)
		}
	}
	return result
}

func (am *amOperator) GetPolicyRule(name string) (*iamv1alpha2.PolicyRule, error) {
	lister := am.informers.Iam().V1alpha2().PolicyRules().Lister()
	return lister.Get(name)
}

// Users can only bind one role at each level
func (am *amOperator) GetRoleOfUserInTargetScope(scope iamv1alpha2.Scope, target, username string) (*iamv1alpha2.Role, error) {
	roles, err := am.ListRolesOfUser(scope, username)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	for _, role := range roles {
		if role.Target.Name == iamv1alpha2.TargetAll ||
			role.Target.Name == target {
			return &role, nil
		}
	}

	err = &errors.StatusError{ErrStatus: metav1.Status{
		Status: metav1.StatusFailure,
		Code:   http.StatusNotFound,
		Reason: metav1.StatusReasonNotFound,
		Details: &metav1.StatusDetails{
			Group: iamv1alpha2.SchemeGroupVersion.Group,
			Kind:  iamv1alpha2.RoleBindingKind,
		},
		Message: fmt.Sprintf("role bind not found in %s %s scope", target, scope),
	}}

	klog.Error(err)
	return nil, err
}
