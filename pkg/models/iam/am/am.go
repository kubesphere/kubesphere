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
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/clusterrole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/resource"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/role"
)

const (
	ClusterRoleKind             = "ClusterRole"
	NamespaceAdminRoleBindName  = "admin"
	NamespaceViewerRoleBindName = "viewer"
)

type AccessManagementInterface interface {
	GetPlatformRole(username string) (Role, error)
	GetClusterRole(cluster, username string) (Role, error)
	GetWorkspaceRole(workspace, username string) (Role, error)
	GetNamespaceRole(namespace, username string) (Role, error)
	GetDevOpsRole(project, username string) (Role, error)
}

type Role interface {
	GetName() string
	GetRego() string
}

type amOperator struct {
	informers  informers.SharedInformerFactory
	resources  resource.ResourceGetter
	kubeClient kubernetes.Interface
}

func (am *amOperator) ListClusterRoleBindings(clusterRole string) ([]*rbacv1.ClusterRoleBinding, error) {
	panic("implement me")
}

func (am *amOperator) GetRoles(namespace, username string) ([]*rbacv1.Role, error) {
	panic("implement me")
}

func (am *amOperator) GetClusterPolicyRules(username string) ([]rbacv1.PolicyRule, error) {
	panic("implement me")
}

func (am *amOperator) GetPolicyRules(namespace, username string) ([]rbacv1.PolicyRule, error) {
	panic("implement me")
}

func (am *amOperator) GetWorkspaceRole(workspace, username string) (Role, error) {
	panic("implement me")
}

func NewAMOperator(kubeClient kubernetes.Interface, informers informers.SharedInformerFactory) AccessManagementInterface {
	resourceGetter := resource.ResourceGetter{}
	resourceGetter.Add(v1alpha2.Role, role.NewRoleSearcher(informers))
	resourceGetter.Add(v1alpha2.ClusterRoles, clusterrole.NewClusterRoleSearcher(informers))
	return &amOperator{
		informers:  informers,
		resources:  resourceGetter,
		kubeClient: kubeClient,
	}
}

func (am *amOperator) GetPlatformRole(username string) (Role, error) {
	panic("implement me")
}

func (am *amOperator) GetClusterRole(cluster, username string) (Role, error) {
	panic("implement me")
}

func (am *amOperator) GetNamespaceRole(namespace, username string) (Role, error) {
	panic("implement me")
}

func (am *amOperator) GetDevOpsRole(namespace, username string) (Role, error) {
	panic("implement me")
}
