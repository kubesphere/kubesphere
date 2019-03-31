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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"kubesphere.io/kubesphere/pkg/params"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"net/http"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"

	"strings"

	"github.com/jinzhu/gorm"
	core "k8s.io/api/core/v1"

	"errors"
	"github.com/golang/glog"
	"k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"sort"

	kserr "kubesphere.io/kubesphere/pkg/errors"
)

func UnBindDevopsProject(workspace string, devops string) error {
	db := mysql.Client()
	return db.Delete(&models.WorkspaceDPBinding{Workspace: workspace, DevOpsProject: devops}).Error
}

func DeleteDevopsProject(username string, devops string) error {
	request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s/api/v1alpha/projects/%s", constants.DevopsAPIServer, devops), nil)
	request.Header.Add("X-Token-Username", username)

	result, err := http.DefaultClient.Do(request)

	if err != nil {
		return err
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return err
	}
	if result.StatusCode > 200 {
		return kserr.Parse(data)
	}
	return nil
}

func CreateDevopsProject(username string, workspace string, devops models.DevopsProject) (*models.DevopsProject, error) {

	data, err := json.Marshal(devops)

	if err != nil {
		return nil, err
	}

	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/api/v1alpha/projects", constants.DevopsAPIServer), bytes.NewReader(data))
	request.Header.Add("X-Token-Username", username)
	request.Header.Add("Content-Type", "application/json")
	result, err := http.DefaultClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer result.Body.Close()
	data, err = ioutil.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	if result.StatusCode > 200 {
		return nil, kserr.Parse(data)
	}

	var project models.DevopsProject

	err = json.Unmarshal(data, &project)

	if err != nil {
		return nil, err
	}

	err = BindingDevopsProject(workspace, *project.ProjectId)

	if err != nil {
		DeleteDevopsProject(username, *project.ProjectId)
		return nil, err
	}

	go createDefaultDevopsRoleBinding(workspace, project)

	return &project, nil
}

func createDefaultDevopsRoleBinding(workspace string, project models.DevopsProject) error {
	admins := []string{""}

	for _, admin := range admins {
		createDevopsRoleBinding(workspace, *project.ProjectId, admin, constants.DevopsOwner)
	}

	viewers := []string{""}

	for _, viewer := range viewers {
		createDevopsRoleBinding(workspace, *project.ProjectId, viewer, constants.DevopsReporter)
	}

	return nil
}

func createDevopsRoleBinding(workspace string, projectId string, user string, role string) {

	projects := make([]string, 0)

	if projectId != "" {
		projects = append(projects, projectId)
	} else {
		p, err := GetDevOpsProjects(workspace)
		if err != nil {
			glog.Warning("create  devops role binding failed", workspace, projectId, user, role)
			return
		}
		projects = append(projects, p...)
	}

	for _, project := range projects {
		data := []byte(fmt.Sprintf(`{"username":"%s","role":"%s"}`, user, role))
		request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/api/v1alpha/projects/%s/members", constants.DevopsAPIServer, project), bytes.NewReader(data))
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("X-Token-Username", "admin")
		resp, err := http.DefaultClient.Do(request)
		if err != nil || resp.StatusCode > 200 {
			glog.Warning(fmt.Sprintf("create  devops role binding failed %s,%s,%s,%s", workspace, project, user, role))
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
}

func ListNamespaceByUser(workspaceName string, username string, keyword string, orderBy string, reverse bool, limit int, offset int) (int, []*core.Namespace, error) {

	namespaces, err := Namespaces(workspaceName)

	if err != nil {
		return 0, nil, err
	}

	if keyword != "" {
		for i := 0; i < len(namespaces); i++ {
			if !strings.Contains(namespaces[i].Name, keyword) {
				namespaces = append(namespaces[:i], namespaces[i+1:]...)
				i--
			}
		}
	}

	sort.Slice(namespaces, func(i, j int) bool {
		switch orderBy {
		case "name":
			if reverse {
				return namespaces[i].Name < namespaces[j].Name
			} else {
				return namespaces[i].Name > namespaces[j].Name
			}
		default:
			if reverse {
				return namespaces[i].CreationTimestamp.Time.After(namespaces[j].CreationTimestamp.Time)
			} else {
				return namespaces[i].CreationTimestamp.Time.Before(namespaces[j].CreationTimestamp.Time)
			}
		}
	})

	rules, err := iam.GetUserClusterRules(username)

	if err != nil {
		return 0, nil, err
	}

	namespacesManager := v1.PolicyRule{APIGroups: []string{"kubesphere.io"}, ResourceNames: []string{workspaceName}, Verbs: []string{"get"}, Resources: []string{"workspaces/namespaces"}}

	if !iam.RulesMatchesRequired(rules, namespacesManager) {
		for i := 0; i < len(namespaces); i++ {
			roles, err := iam.GetUserRoles(namespaces[i].Name, username)
			if err != nil {
				return 0, nil, err
			}
			rules := make([]v1.PolicyRule, 0)
			for _, role := range roles {
				rules = append(rules, role.Rules...)
			}
			if !iam.RulesMatchesRequired(rules, v1.PolicyRule{APIGroups: []string{""}, ResourceNames: []string{namespaces[i].Name}, Verbs: []string{"get"}, Resources: []string{"namespaces"}}) {
				namespaces = append(namespaces[:i], namespaces[i+1:]...)
				i--
			}
		}
	}

	if len(namespaces) < offset {
		return len(namespaces), namespaces, nil
	} else if len(namespaces) < limit+offset {
		return len(namespaces), namespaces[offset:], nil
	} else {
		return len(namespaces), namespaces[offset : limit+offset], nil
	}
}

func Namespaces(workspaceName string) ([]*core.Namespace, error) {
	namespaceLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	namespaces, err := namespaceLister.List(labels.SelectorFromSet(labels.Set{"kubesphere.io/workspace": workspaceName}))

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

func Delete(workspace *models.Workspace) error {

	err := release(workspace)

	if err != nil {
		return err
	}

	err = iam.DeleteGroup(workspace.Name)

	if err != nil {
		return err
	}

	return nil
}

// TODO
func release(workspace *models.Workspace) error {
	for _, namespace := range workspace.Namespaces {
		err := DeleteNamespace(workspace.Name, namespace)
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	for _, devops := range workspace.DevopsProjects {
		err := DeleteDevopsProject("admin", devops)
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return err
		}
	}

	err := workspaceRoleRelease(workspace.Name)

	return err
}
func workspaceRoleRelease(workspace string) error {
	k8sClient := k8s.Client()
	deletePolicy := meta_v1.DeletePropagationForeground

	for _, role := range constants.WorkSpaceRoles {
		err := k8sClient.RbacV1().ClusterRoles().Delete(fmt.Sprintf("system:%s:%s", workspace, role), &meta_v1.DeleteOptions{PropagationPolicy: &deletePolicy})

		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	for _, role := range constants.WorkSpaceRoles {
		err := k8sClient.RbacV1().ClusterRoleBindings().Delete(fmt.Sprintf("system:%s:%s", workspace, role), &meta_v1.DeleteOptions{PropagationPolicy: &deletePolicy})

		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func Edit(workspace *models.Workspace) (*models.Workspace, error) {

	group, err := iam.UpdateGroup(&workspace.Group)

	if err != nil {
		return nil, err
	}

	workspace.Group = *group

	return workspace, nil
}

func DescribeWorkspace(workspaceName string) (*v1alpha1.Workspace, error) {
	workspace, err := informers.KsSharedInformerFactory().Tenant().V1alpha1().Workspaces().Lister().Get(workspaceName)

	if err != nil {
		return nil, err
	}

	return workspace, nil
}

func fetch(names []string) ([]*models.Workspace, error) {

	if names != nil && len(names) == 0 {
		return make([]*models.Workspace, 0), nil
	}
	var groups []models.Group
	var err error
	if names == nil {
		groups, err = iam.ChildList("")
		if err != nil {
			return nil, err
		}
	} else {
		conn, err := ldap.Client()
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		for _, name := range names {
			group, err := iam.DescribeGroup(name)
			if err != nil {
				return nil, err
			}
			groups = append(groups, *group)
		}
	}

	db := mysql.Client()

	workspaces := make([]*models.Workspace, 0)
	for _, group := range groups {
		workspace, err := convertGroupToWorkspace(db, group)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, workspace)
	}

	return workspaces, nil
}

func convertGroupToWorkspace(db *gorm.DB, group models.Group) (*models.Workspace, error) {
	namespaces, err := Namespaces(group.Name)

	if err != nil {
		return nil, err
	}

	namespacesNames := make([]string, 0)

	for _, namespace := range namespaces {
		namespacesNames = append(namespacesNames, namespace.Name)
	}

	var workspaceDOPBindings []models.WorkspaceDPBinding

	if err := db.Where("workspace = ?", group.Name).Find(&workspaceDOPBindings).Error; err != nil {
		return nil, err
	}

	devOpsProjects := make([]string, 0)

	for _, workspaceDOPBinding := range workspaceDOPBindings {
		devOpsProjects = append(devOpsProjects, workspaceDOPBinding.DevOpsProject)
	}

	workspace := models.Workspace{Group: group}
	workspace.Namespaces = namespacesNames
	workspace.DevopsProjects = devOpsProjects
	return &workspace, nil
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
	workspaceRoleBinding = workspaceRoleBinding.DeepCopy()
	if err != nil {
		return err
	}

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
	workspaceRoleBinding = workspaceRoleBinding.DeepCopy()

	if err != nil {
		return err
	}

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
