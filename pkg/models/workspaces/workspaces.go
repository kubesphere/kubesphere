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
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"kubesphere.io/kubesphere/pkg/params"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/kubesphere"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"

	"strings"

	core "k8s.io/api/core/v1"

	"errors"
	"k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func UnBindDevopsProject(workspace string, devops string) error {
	db := mysql.Client()
	return db.Delete(&models.WorkspaceDPBinding{Workspace: workspace, DevOpsProject: devops}).Error
}

func CreateDevopsProject(username string, workspace string, devops *models.DevopsProject) (*models.DevopsProject, error) {

	created, err := kubesphere.Client().CreateDevopsProject(username, devops)

	if err != nil {
		return nil, err
	}

	err = BindingDevopsProject(workspace, created.ProjectId)

	if err != nil {
		return nil, err
	}

	return created, nil
}

func Namespaces(workspaceName string) ([]*core.Namespace, error) {
	namespaceLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	namespaces, err := namespaceLister.List(labels.SelectorFromSet(labels.Set{constants.WorkspaceLabelKey: workspaceName}))

	if err != nil {
		return nil, err
	}

	if namespaces == nil {
		return make([]*core.Namespace, 0), nil
	}

	out := make([]*core.Namespace, len(namespaces))

	for i, v := range namespaces {
		out[i] = v.DeepCopy()
	}

	return out, nil
}

func BindingDevopsProject(workspace string, devops string) error {
	db := mysql.Client()
	return db.Create(&models.WorkspaceDPBinding{Workspace: workspace, DevOpsProject: devops}).Error
}

func DeleteNamespace(workspace string, namespaceName string) error {
	namespace, err := k8s.Client().CoreV1().Namespaces().Get(namespaceName, meta_v1.GetOptions{})
	if err != nil {
		return err
	}
	if namespace.Labels[constants.WorkspaceLabelKey] == workspace {
		deletePolicy := meta_v1.DeletePropagationForeground
		return k8s.Client().CoreV1().Namespaces().Delete(namespaceName, &meta_v1.DeleteOptions{PropagationPolicy: &deletePolicy})
	} else {
		return errors.New("resource not found")
	}
}

func RemoveUser(workspaceName string, username string) error {
	workspaceRole, err := iam.GetUserWorkspaceRole(workspaceName, username)
	if err != nil {
		return err
	}
	err = DeleteWorkspaceRoleBinding(workspaceName, username, workspaceRole.Annotations[constants.DisplayNameAnnotationKey])
	if err != nil {
		return err
	}
	return nil
}

func InviteUser(workspaceName string, user *models.User) error {

	workspaceRole, err := iam.GetUserWorkspaceRole(workspaceName, user.Username)

	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	workspaceRoleName := fmt.Sprintf("workspace:%s:%s", workspaceName, strings.TrimPrefix(user.WorkspaceRole, "workspace-"))

	if workspaceRole != nil && workspaceRole.Name != workspaceRoleName {
		err := DeleteWorkspaceRoleBinding(workspaceName, user.Username, user.WorkspaceRole)
		if err != nil {
			return err
		}
	}

	return CreateWorkspaceRoleBinding(workspaceName, user.Username, user.WorkspaceRole)
}

func CreateWorkspaceRoleBinding(workspace, username string, role string) error {

	if !sliceutil.HasString(constants.WorkSpaceRoles, role) {
		return apierrors.NewNotFound(schema.GroupResource{Resource: "workspace role"}, role)
	}

	roleBindingName := fmt.Sprintf("workspace:%s:%s", workspace, strings.TrimPrefix(role, "workspace-"))
	workspaceRoleBinding, err := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister().Get(roleBindingName)
	if err != nil {
		return err
	}
	workspaceRoleBinding = workspaceRoleBinding.DeepCopy()
	if !k8sutil.ContainsUser(workspaceRoleBinding.Subjects, username) {
		workspaceRoleBinding.Subjects = append(workspaceRoleBinding.Subjects, v1.Subject{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: username})
		_, err = k8s.Client().RbacV1().ClusterRoleBindings().Update(workspaceRoleBinding)
	}

	return err
}

func DeleteWorkspaceRoleBinding(workspace, username string, role string) error {

	if !sliceutil.HasString(constants.WorkSpaceRoles, role) {
		return apierrors.NewNotFound(schema.GroupResource{Resource: "workspace role"}, role)
	}

	roleBindingName := fmt.Sprintf("workspace:%s:%s", workspace, strings.TrimPrefix(role, "workspace-"))

	workspaceRoleBinding, err := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister().Get(roleBindingName)
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

	workspaceRoleBinding, err = k8s.Client().RbacV1().ClusterRoleBindings().Update(workspaceRoleBinding)

	return err
}

func GetDevOpsProjects(workspaceName string) ([]string, error) {

	db := mysql.Client()

	var workspaceDOPBindings []models.WorkspaceDPBinding

	if err := db.Where("workspace = ?", workspaceName).Find(&workspaceDOPBindings).Error; err != nil {
		return nil, err
	}

	devOpsProjects := make([]string, 0)

	for _, workspaceDOPBinding := range workspaceDOPBindings {
		devOpsProjects = append(devOpsProjects, workspaceDOPBinding.DevOpsProject)
	}
	return devOpsProjects, nil
}

func WorkspaceUserCount(workspace string) (int, error) {
	count, err := iam.WorkspaceUsersTotalCount(workspace)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetOrgRoles(name string) ([]string, error) {
	return constants.WorkSpaceRoles, nil
}

func WorkspaceNamespaces(workspaceName string) ([]string, error) {
	ns, err := Namespaces(workspaceName)

	namespaces := make([]string, 0)

	if err != nil {
		return namespaces, err
	}

	for i := 0; i < len(ns); i++ {
		namespaces = append(namespaces, ns[i].Name)
	}

	return namespaces, nil
}

func WorkspaceCount() (int, error) {

	ws, err := resources.ListResources("", resources.Workspaces, &params.Conditions{}, "", false, 1, 0)

	if err != nil {
		return 0, err
	}

	return ws.TotalCount, nil
}

func GetAllProjectNums() (int, error) {
	namespaceLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	list, err := namespaceLister.List(labels.Everything())
	if err != nil {
		return 0, err
	}
	return len(list), nil
}

func GetAllDevOpsProjectsNums() (int, error) {
	db := mysql.Client()
	var count int
	if err := db.Model(&models.WorkspaceDPBinding{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func GetAllAccountNums() (int, error) {
	users, err := iam.ListUsers(&params.Conditions{}, "", false, 1, 0)
	if err != nil {
		return 0, err
	}
	return users.TotalCount, nil
}
