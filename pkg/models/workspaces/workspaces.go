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
package workspaces

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/db"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/models/iam"
	clientset "kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"strings"

	core "k8s.io/api/core/v1"

	"errors"
	"k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type Interface interface {
	ListNamespaces(workspace string) ([]*core.Namespace, error)
	DeleteNamespace(workspace, namespace string) error
	RemoveUser(user, workspace string) error
	AddUser(workspace string, user *models.User) error
	CountDevopsProjectsInWorkspace(workspace string) (int, error)
	CountUsersInWorkspace(workspace string) (int, error)
	CountOrgRoles() (int, error)
	CountWorkspaces() (int, error)
	CountNamespacesInWorkspace(workspace string) (int, error)
}

type workspaceOperator struct {
	client      kubernetes.Interface
	informers   informers.SharedInformerFactory
	ksInformers externalversions.SharedInformerFactory

	// TODO: use db interface instead of mysql client
	// we can refactor this after rewrite devops using crd
	db *mysql.Database
}

func NewWorkspaceOperator(client kubernetes.Interface, informers informers.SharedInformerFactory, ksinformers externalversions.SharedInformerFactory, db *mysql.Database) Interface {
	return &workspaceOperator{
		client:      client,
		informers:   informers,
		ksInformers: ksinformers,
		db:          db,
	}
}

func (c *workspaceOperator) ListNamespaces(workspace string) ([]*core.Namespace, error) {
	namespaces, err := c.informers.Core().V1().Namespaces().Lister().List(labels.SelectorFromSet(labels.Set{constants.WorkspaceLabelKey: workspace}))

	if err != nil {
		return nil, err
	}

	return namespaces, nil
}

func (c *workspaceOperator) DeleteNamespace(workspace string, namespace string) error {
	ns, err := c.informers.Core().V1().Namespaces().Lister().Get(namespace)
	if err != nil {
		return err
	}

	if ns.Labels[constants.WorkspaceLabelKey] == workspace {
		deletePolicy := metav1.DeletePropagationBackground
		return c.client.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
	} else {
		return errors.New("resource not found")
	}
}

func (c *workspaceOperator) RemoveUser(workspace string, username string) error {
	workspaceRole, err := iam.GetUserWorkspaceRole(workspace, username)
	if err != nil {
		return err
	}

	err = c.deleteWorkspaceRoleBinding(workspace, username, workspaceRole.Annotations[constants.DisplayNameAnnotationKey])
	if err != nil {
		return err
	}

	return nil
}

func (c *workspaceOperator) AddUser(workspaceName string, user *models.User) error {

	workspaceRole, err := iam.GetUserWorkspaceRole(workspaceName, user.Username)

	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("get workspace role failed: %+v", err)
		return err
	}

	workspaceRoleName := fmt.Sprintf("workspace:%s:%s", workspaceName, strings.TrimPrefix(user.WorkspaceRole, "workspace-"))
	var currentWorkspaceRoleName string
	if workspaceRole != nil {
		currentWorkspaceRoleName = workspaceRole.Name
	}

	if currentWorkspaceRoleName != workspaceRoleName && currentWorkspaceRoleName != "" {
		err := c.deleteWorkspaceRoleBinding(workspaceName, user.Username, workspaceRole.Annotations[constants.DisplayNameAnnotationKey])
		if err != nil {
			klog.Errorf("delete workspace role binding failed: %+v", err)
			return err
		}
	} else if currentWorkspaceRoleName != "" {
		return nil
	}

	return c.createWorkspaceRoleBinding(workspaceName, user.Username, user.WorkspaceRole)
}

func (c *workspaceOperator) createWorkspaceRoleBinding(workspace, username string, role string) error {

	if !sliceutil.HasString(constants.WorkSpaceRoles, role) {
		return apierrors.NewNotFound(schema.GroupResource{Resource: "workspace role"}, role)
	}

	roleBindingName := fmt.Sprintf("workspace:%s:%s", workspace, strings.TrimPrefix(role, "workspace-"))
	workspaceRoleBinding, err := c.informers.Rbac().V1().ClusterRoleBindings().Lister().Get(roleBindingName)
	if err != nil {
		return err
	}

	if !k8sutil.ContainsUser(workspaceRoleBinding.Subjects, username) {
		workspaceRoleBinding = workspaceRoleBinding.DeepCopy()
		workspaceRoleBinding.Subjects = append(workspaceRoleBinding.Subjects, v1.Subject{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: username})
		_, err = c.client.RbacV1().ClusterRoleBindings().Update(workspaceRoleBinding)
		if err != nil {
			klog.Errorf("update workspace role binding failed: %+v", err)
			return err
		}
	}

	return nil
}

func (c *workspaceOperator) deleteWorkspaceRoleBinding(workspace, username string, role string) error {

	if !sliceutil.HasString(constants.WorkSpaceRoles, role) {
		return apierrors.NewNotFound(schema.GroupResource{Resource: "workspace role"}, role)
	}

	roleBindingName := fmt.Sprintf("workspace:%s:%s", workspace, strings.TrimPrefix(role, "workspace-"))

	workspaceRoleBinding, err := c.informers.Rbac().V1().ClusterRoleBindings().Lister().Get(roleBindingName)
	if err != nil {
		return err
	}
	workspaceRoleBinding = workspaceRoleBinding.DeepCopy()

	for i, v := range workspaceRoleBinding.Subjects {
		if v.Kind == v1.UserKind && v.Name == username {
			workspaceRoleBinding.Subjects = append(workspaceRoleBinding.Subjects[:i], workspaceRoleBinding.Subjects[i+1:]...)
			i--
		}
	}

	workspaceRoleBinding, err = c.client.RbacV1().ClusterRoleBindings().Update(workspaceRoleBinding)

	return err
}

func (c *workspaceOperator) CountDevopsProjectsInWorkspace(workspaceName string) (int, error) {
	if c.db == nil {
		return 0, clientset.ErrClientSetNotEnabled
	}

	query := c.db.Select(devops.DevOpsProjectIdColumn).
		From(devops.DevOpsProjectTableName).
		Where(db.And(db.Eq(devops.DevOpsProjectWorkSpaceColumn, workspaceName),
			db.Eq(devops.StatusColumn, devops.StatusActive)))

	devOpsProjects := make([]string, 0)

	if _, err := query.Load(&devOpsProjects); err != nil {
		return 0, err
	}
	return len(devOpsProjects), nil
}

func (c *workspaceOperator) CountUsersInWorkspace(workspace string) (int, error) {
	count, err := iam.WorkspaceUsersTotalCount(workspace)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (c *workspaceOperator) CountOrgRoles() (int, error) {
	return len(constants.WorkSpaceRoles), nil
}

func (c *workspaceOperator) CountNamespacesInWorkspace(workspace string) (int, error) {
	ns, err := c.ListNamespaces(workspace)
	if err != nil {
		return 0, err
	}

	return len(ns), nil
}

/*
// TODO: move to metrics package
func GetAllProjectNums() (int, error) {
	namespaceLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	list, err := namespaceLister.List(labels.Everything())
	if err != nil {
		return 0, err
	}
	return len(list), nil
}

func GetAllDevOpsProjectsNums() (int, error) {
	_, err := clientset.ClientSets().Devops()
	if _, notEnabled := err.(clientset.ClientSetNotEnabledError); notEnabled {
		return 0, err
	}

	dbconn, err := clientset.ClientSets().MySQL()
	if err != nil {
		return 0, err
	}

	query := dbconn.Select(devops.DevOpsProjectIdColumn).
		From(devops.DevOpsProjectTableName).
		Where(db.Eq(devops.StatusColumn, devops.StatusActive))

	devOpsProjects := make([]string, 0)

	if _, err := query.Load(&devOpsProjects); err != nil {
		return 0, err
	}
	return len(devOpsProjects), nil
}
*/

func (c *workspaceOperator) CountWorkspaces() (int, error) {
	ws, err := c.ksInformers.Tenant().V1alpha1().Workspaces().Lister().List(labels.Everything())
	if err != nil {
		return 0, err
	}

	return len(ws), nil
}
