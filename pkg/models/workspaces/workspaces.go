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
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"net/http"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"

	"log"
	"strings"

	"github.com/jinzhu/gorm"
	core "k8s.io/api/core/v1"

	"errors"
	"regexp"

	"github.com/emicklei/go-restful"
	"k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kubernetes/pkg/util/slice"

	"github.com/golang/glog"

	"sort"

	kserr "kubesphere.io/kubesphere/pkg/errors"
)

func UnBindDevopsProject(workspace string, devops string) error {
	db := mysql.Client()
	defer db.Close()
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
	admins, err := iam.GetWorkspaceUsers(workspace, constants.WorkspaceAdmin)

	if err != nil {
		return err
	}

	for _, admin := range admins {
		createDevopsRoleBinding(workspace, *project.ProjectId, admin, constants.DevopsOwner)
	}

	viewers, err := iam.GetWorkspaceUsers(workspace, constants.WorkspaceViewer)

	if err != nil {
		return err
	}

	for _, viewer := range viewers {
		createDevopsRoleBinding(workspace, *project.ProjectId, viewer, constants.DevopsReporter)
	}

	return nil
}

func deleteDevopsRoleBinding(workspace string, projectId string, user string) {
	projects := make([]string, 0)

	if projectId != "" {
		projects = append(projects, projectId)
	} else {
		p, err := GetDevOpsProjects(workspace)
		if err != nil {
			glog.Warning("delete  devops role binding failed", workspace, projectId, user)
			return
		}
		projects = append(projects, p...)
	}

	for _, project := range projects {
		request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s/api/v1alpha/projects/%s/members/%s", constants.DevopsAPIServer, project, user), nil)
		request.Header.Add("X-Token-Username", "admin")
		resp, err := http.DefaultClient.Do(request)
		if err != nil || resp.StatusCode > 200 {
			glog.Warning("delete  devops role binding failed", workspace, project, user)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
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

	clusterRoles, err := iam.GetClusterRoles(username)

	if err != nil {
		return 0, nil, err
	}

	rules := make([]v1.PolicyRule, 0)

	for _, clusterRole := range clusterRoles {
		rules = append(rules, clusterRole.Rules...)
	}

	namespacesManager := v1.PolicyRule{APIGroups: []string{"kubesphere.io"}, ResourceNames: []string{workspaceName}, Verbs: []string{"get"}, Resources: []string{"workspaces/namespaces"}}

	if !iam.RulesMatchesRequired(rules, namespacesManager) {
		for i := 0; i < len(namespaces); i++ {
			roles, err := iam.GetRoles(namespaces[i].Name, username)
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
	defer db.Close()
	return db.Create(&models.WorkspaceDPBinding{Workspace: workspace, DevOpsProject: devops}).Error
}

func DeleteNamespace(workspace string, namespaceName string) error {
	namespace, err := k8s.Client().CoreV1().Namespaces().Get(namespaceName, meta_v1.GetOptions{})
	if err != nil {
		return err
	}
	if namespace.Labels != nil && namespace.Labels["kubesphere.io/workspace"] == workspace {
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

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/groups/%s", constants.AccountAPIServer, workspace.Name), nil)

	if err != nil {
		return err
	}
	result, err := http.DefaultClient.Do(req)

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

func release(workspace *models.Workspace) error {
	for _, namespace := range workspace.Namespaces {
		err := DeleteNamespace(workspace.Name, namespace)
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	for _, devops := range workspace.DevopsProjects {
		err := DeleteDevopsProject(workspace.Creator, devops)
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

func Create(workspace *models.Workspace) (*models.Workspace, error) {

	data, err := json.Marshal(workspace)

	if err != nil {
		return nil, err
	}

	result, err := http.Post(fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/groups", constants.AccountAPIServer), restful.MIME_JSON, bytes.NewReader(data))

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

	var created models.Workspace

	err = json.Unmarshal(data, &created)

	if err != nil {
		return nil, err
	}

	created.Members = make([]string, 0)
	created.Namespaces = make([]string, 0)
	created.DevopsProjects = make([]string, 0)

	err = WorkspaceRoleInit(workspace)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

func Edit(workspace *models.Workspace) (*models.Workspace, error) {

	data, err := json.Marshal(workspace)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/groups/%s", constants.AccountAPIServer, workspace.Name), bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	result, err := http.DefaultClient.Do(req)

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

	var edited models.Workspace

	err = json.Unmarshal(data, &edited)

	if err != nil {
		return nil, err
	}

	return &edited, nil
}

func Detail(name string) (*models.Workspace, error) {

	result, err := http.Get(fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/groups/%s", constants.AccountAPIServer, name))

	if err != nil {
		return nil, err
	}

	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	if result.StatusCode > 200 {
		return nil, kserr.Parse(data)
	}

	var group models.Group

	err = json.Unmarshal(data, &group)

	if err != nil {
		return nil, err
	}

	db := mysql.Client()
	defer db.Close()

	workspace, err := convertGroupToWorkspace(db, group)

	if err != nil {
		return nil, err
	}

	return workspace, nil
}

// List all workspaces for the current user
func ListWorkspaceByUser(username string, keyword string) ([]*models.Workspace, error) {
	clusterRoles, err := iam.GetClusterRoles(username)

	if err != nil {
		return nil, err
	}

	rules := make([]v1.PolicyRule, 0)

	for _, clusterRole := range clusterRoles {
		rules = append(rules, clusterRole.Rules...)
	}

	workspacesManager := v1.PolicyRule{APIGroups: []string{"kubesphere.io"}, Verbs: []string{"list", "get"}, Resources: []string{"workspaces"}}

	var workspaces []*models.Workspace
	if iam.RulesMatchesRequired(rules, workspacesManager) {
		workspaces, err = fetch(nil)
	} else {
		workspaceNames := make([]string, 0)
		for _, clusterRole := range clusterRoles {
			if groups := regexp.MustCompile(fmt.Sprintf(`^system:(\S+):(%s)$`, strings.Join(constants.WorkSpaceRoles, "|"))).FindStringSubmatch(clusterRole.Name); len(groups) == 3 {
				if !slice.ContainsString(workspaceNames, groups[1], nil) {
					workspaceNames = append(workspaceNames, groups[1])
				}
			}
		}
		workspaces, err = fetch(workspaceNames)
	}

	if keyword != "" {
		for i := 0; i < len(workspaces); i++ {
			if !strings.Contains(workspaces[i].Name, keyword) {
				workspaces = append(workspaces[:i], workspaces[i+1:]...)
				i--
			}
		}
	}
	return workspaces, err
}

func fetch(names []string) ([]*models.Workspace, error) {

	url := fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/groups", constants.AccountAPIServer)

	if names != nil {
		if len(names) == 0 {
			return make([]*models.Workspace, 0), nil
		} else {
			url = url + "?path=" + strings.Join(names, ",")
		}
	}

	result, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	if result.StatusCode > 200 {
		return nil, kserr.Parse(data)
	}

	var groups []models.Group

	err = json.Unmarshal(data, &groups)

	if err != nil {
		return nil, err
	}

	db := mysql.Client()

	defer db.Close()

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

func ListDevopsProjectsByUser(username string, workspace string, keyword string, orderBy string, reverse bool, limit int, offset int) (int, []models.DevopsProject, error) {

	db := mysql.Client()
	defer db.Close()

	var workspaceDOPBindings []models.WorkspaceDPBinding

	if err := db.Where("workspace = ?", workspace).Find(&workspaceDOPBindings).Error; err != nil {
		return 0, nil, err
	}

	devOpsProjects := make([]models.DevopsProject, 0)

	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/api/v1alpha/projects", constants.DevopsAPIServer), nil)
	request.Header.Add(constants.UserNameHeader, username)

	result, err := http.DefaultClient.Do(request)
	if err != nil {
		return 0, nil, err
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)

	if err != nil {
		return 0, nil, err
	}

	if result.StatusCode > 200 {
		return 0, nil, kserr.Parse(data)
	}

	err = json.Unmarshal(data, &devOpsProjects)

	if err != nil {
		return 0, nil, err
	}

	if keyword != "" {
		for i := 0; i < len(devOpsProjects); i++ {
			if !strings.Contains(devOpsProjects[i].Name, keyword) {
				devOpsProjects = append(devOpsProjects[:i], devOpsProjects[i+1:]...)
				i--
			}
		}
	}

	sort.Slice(devOpsProjects, func(i, j int) bool {
		switch orderBy {
		case "name":
			if reverse {
				return devOpsProjects[i].Name < devOpsProjects[j].Name
			} else {
				return devOpsProjects[i].Name > devOpsProjects[j].Name
			}
		default:
			if reverse {
				return devOpsProjects[i].CreateTime.After(*devOpsProjects[j].CreateTime)
			} else {
				return devOpsProjects[i].CreateTime.Before(*devOpsProjects[j].CreateTime)
			}
		}
	})

	for i := 0; i < len(devOpsProjects); i++ {
		inWorkspace := false

		for _, binding := range workspaceDOPBindings {
			if binding.DevOpsProject == *devOpsProjects[i].ProjectId {
				inWorkspace = true
			}
		}
		if !inWorkspace {
			devOpsProjects = append(devOpsProjects[:i], devOpsProjects[i+1:]...)
			i--
		}
	}

	if len(devOpsProjects) < offset {
		return len(devOpsProjects), make([]models.DevopsProject, 0), nil
	} else if len(devOpsProjects) < limit+offset {
		return len(devOpsProjects), devOpsProjects[offset:], nil
	} else {
		return len(devOpsProjects), devOpsProjects[offset : limit+offset], nil
	}
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

func CreateNamespace(namespace *core.Namespace) (*core.Namespace, error) {

	ns, err := k8s.Client().CoreV1().Namespaces().Create(namespace)

	if err != nil {
		return nil, err
	}

	return ns, nil
}

func Invite(workspaceName string, users []models.UserInvite) error {
	for _, user := range users {
		if !slice.ContainsString(constants.WorkSpaceRoles, user.Role, nil) {
			return fmt.Errorf("role %s not exist", user.Role)
		}
	}

	workspace, err := Detail(workspaceName)

	if err != nil {
		return err
	}

	for _, user := range users {
		if !slice.ContainsString(workspace.Members, user.Username, nil) {
			workspace.Members = append(workspace.Members, user.Username)
		}
	}

	workspace, err = Edit(workspace)

	if err != nil {
		return err
	}

	for _, user := range users {
		err := CreateWorkspaceRoleBinding(workspace, user.Username, user.Role)
		if err != nil {
			return err
		}
	}

	return nil
}

func NamespaceExistCheck(namespaceName string) (bool, error) {

	_, err := k8s.Client().CoreV1().Namespaces().Get(namespaceName, meta_v1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func RemoveMembers(workspaceName string, users []string) error {

	workspace, err := Detail(workspaceName)

	if err != nil {
		return err
	}

	err = UnbindWorkspace(workspace, users)

	if err != nil {
		return err
	}

	for i := 0; i < len(workspace.Members); i++ {
		if slice.ContainsString(users, workspace.Members[i], nil) {
			workspace.Members = append(workspace.Members[:i], workspace.Members[i+1:]...)
			i--
		}
	}

	workspace, err = Edit(workspace)

	if err != nil {
		return err
	}

	return nil
}

func Roles(workspace *models.Workspace) ([]*v1.ClusterRole, error) {
	roles := make([]*v1.ClusterRole, 0)
	clusterRoleLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister()
	for _, name := range constants.WorkSpaceRoles {

		clusterRole, err := clusterRoleLister.Get(fmt.Sprintf("system:%s:%s", workspace.Name, name))

		if err != nil {
			if apierrors.IsNotFound(err) {
				go WorkspaceRoleInit(workspace)
			}
			return nil, err
		}

		clusterRole = clusterRole.DeepCopy()

		clusterRole.Name = name
		roles = append(roles, clusterRole)
	}

	return roles, nil
}

func GetWorkspaceMembers(workspace string, keyword string) ([]models.User, error) {

	url := fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/workspaces/%s/members", constants.AccountAPIServer, workspace)

	if keyword != "" {
		url = url + "?keyword=" + keyword
	}

	result, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	if result.StatusCode > 200 {
		return nil, kserr.Parse(data)
	}

	var users []models.User

	err = json.Unmarshal(data, &users)

	if err != nil {
		return nil, err
	}

	return users, nil

}

func WorkspaceRoleInit(workspace *models.Workspace) error {
	k8sClient := k8s.Client()

	admin := new(v1.ClusterRole)
	admin.Name = fmt.Sprintf("system:%s:%s", workspace.Name, constants.WorkspaceAdmin)
	admin.Kind = iam.ClusterRoleKind
	admin.Rules = []v1.PolicyRule{
		{
			Verbs:         []string{"*"},
			APIGroups:     []string{"kubesphere.io", "account.kubesphere.io"},
			ResourceNames: []string{workspace.Name},
			Resources:     []string{"workspaces", "workspaces/*"},
		},
		{
			Verbs:     []string{"*"},
			APIGroups: []string{"devops.kubesphere.io", "jenkins.kubesphere.io"},
			Resources: []string{"*"},
		},
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{"namespaces"},
			Resources:     []string{"status/*", "monitoring/*", "quota/*"},
		},
		{
			Verbs:     []string{"get"},
			APIGroups: []string{"kubesphere.io"},
			Resources: []string{"resources"},
		},
		{
			Verbs:     []string{"list"},
			APIGroups: []string{"account.kubesphere.io"},
			Resources: []string{"users"},
		},
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{"workspaces"},
			Resources:     []string{"monitoring/" + workspace.Name},
		},
	}

	admin.Labels = map[string]string{"creator": "system"}

	regular := new(v1.ClusterRole)
	regular.Name = fmt.Sprintf("system:%s:%s", workspace.Name, constants.WorkspaceRegular)
	regular.Kind = iam.ClusterRoleKind
	regular.Rules = []v1.PolicyRule{
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			Resources:     []string{"workspaces"},
			ResourceNames: []string{workspace.Name},
		}, {
			Verbs:         []string{"create"},
			APIGroups:     []string{"kubesphere.io"},
			Resources:     []string{"workspaces/namespaces", "workspaces/devops"},
			ResourceNames: []string{workspace.Name},
		},
		{
			Verbs:         []string{"delete"},
			APIGroups:     []string{"kubesphere.io"},
			Resources:     []string{"workspaces/namespaces", "workspaces/devops"},
			ResourceNames: []string{workspace.Name},
		},
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{"namespaces"},
			Resources:     []string{"quota/*", "status/*", "monitoring/*"},
		},
		{
			Verbs:     []string{"*"},
			APIGroups: []string{"devops.kubesphere.io"},
			Resources: []string{"*"},
		}, {
			Verbs:     []string{"*"},
			APIGroups: []string{"jenkins.kubesphere.io"},
			Resources: []string{"*"},
		},
		{
			Verbs:     []string{"get"},
			APIGroups: []string{"kubesphere.io"},
			Resources: []string{"resources"},
		},
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{workspace.Name},
			Resources:     []string{"workspaces/members"},
		},
	}

	regular.Labels = map[string]string{"creator": "system"}

	viewer := new(v1.ClusterRole)
	viewer.Name = fmt.Sprintf("system:%s:%s", workspace.Name, constants.WorkspaceViewer)
	viewer.Kind = iam.ClusterRoleKind
	viewer.Rules = []v1.PolicyRule{
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io", "account.kubesphere.io"},
			ResourceNames: []string{workspace.Name},
			Resources:     []string{"workspaces", "workspaces/*"},
		},
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{"namespaces"},
			Resources:     []string{"quota/*", "status/*", "monitoring/*"},
		},
		{
			Verbs:     []string{"get"},
			APIGroups: []string{"kubesphere.io"},
			Resources: []string{"resources"},
		},
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{"workspaces"},
			Resources:     []string{"monitoring/" + workspace.Name},
		},
		{
			Verbs:     []string{"get", "list"},
			APIGroups: []string{"devops.kubesphere.io"},
			Resources: []string{"*"},
		}, {
			Verbs:     []string{"get", "list"},
			APIGroups: []string{"jenkins.kubesphere.io"},
			Resources: []string{"*"},
		},
	}

	viewer.Labels = map[string]string{"creator": "system"}

	_, err := k8sClient.RbacV1().ClusterRoles().Create(admin)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			log.Println("cluster role create failed", admin.Name, err)
			return err
		}
	}

	adminRoleBinding := new(v1.ClusterRoleBinding)
	adminRoleBinding.Name = admin.Name
	adminRoleBinding.RoleRef = v1.RoleRef{Kind: "ClusterRole", Name: admin.Name}
	adminRoleBinding.Subjects = []v1.Subject{{Kind: v1.UserKind, Name: workspace.Creator}}

	_, err = k8sClient.RbacV1().ClusterRoleBindings().Create(adminRoleBinding)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			log.Println("cluster rolebinding create failed", adminRoleBinding.Name, err)
			return err
		}
	}

	_, err = k8sClient.RbacV1().ClusterRoles().Create(regular)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			log.Println("cluster role create failed", viewer.Name, err)
			return err
		}
	}

	regularRoleBinding := new(v1.ClusterRoleBinding)
	regularRoleBinding.Name = regular.Name
	regularRoleBinding.RoleRef = v1.RoleRef{Kind: "ClusterRole", Name: regular.Name}
	regularRoleBinding.Subjects = make([]v1.Subject, 0)
	_, err = k8sClient.RbacV1().ClusterRoleBindings().Create(regularRoleBinding)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			log.Println("cluster rolebinding create failed", regularRoleBinding.Name, err)
			return err
		}
	}

	_, err = k8sClient.RbacV1().ClusterRoles().Create(viewer)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			log.Println("cluster role create failed", viewer.Name, err)
			return err
		}
	}

	viewerRoleBinding := new(v1.ClusterRoleBinding)
	viewerRoleBinding.Name = viewer.Name
	viewerRoleBinding.RoleRef = v1.RoleRef{Kind: "ClusterRole", Name: viewer.Name}
	viewerRoleBinding.Subjects = make([]v1.Subject, 0)
	_, err = k8sClient.RbacV1().ClusterRoleBindings().Create(viewerRoleBinding)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			log.Println("cluster rolebinding create failed", viewerRoleBinding.Name, err)
			return err
		}
	}

	return nil
}

func unbindWorkspaceRole(workspace string, users []string) error {
	k8sClient := k8s.Client()

	for _, name := range constants.WorkSpaceRoles {
		roleBinding, err := k8sClient.RbacV1().ClusterRoleBindings().Get(fmt.Sprintf("system:%s:%s", workspace, name), meta_v1.GetOptions{})

		if err != nil {
			return err
		}

		modify := false

		for i := 0; i < len(roleBinding.Subjects); i++ {
			if roleBinding.Subjects[i].Kind == v1.UserKind && slice.ContainsString(users, roleBinding.Subjects[i].Name, nil) {
				roleBinding.Subjects = append(roleBinding.Subjects[:i], roleBinding.Subjects[i+1:]...)
				i--
				modify = true
			}
		}

		if modify {
			roleBinding, err = k8sClient.RbacV1().ClusterRoleBindings().Update(roleBinding)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func unbindNamespacesRole(namespaces []string, users []string) error {

	k8sClient := k8s.Client()
	for _, namespace := range namespaces {

		roleBindings, err := k8sClient.RbacV1().RoleBindings(namespace).List(meta_v1.ListOptions{})

		if err != nil {
			return err
		}
		for _, roleBinding := range roleBindings.Items {

			modify := false
			for i := 0; i < len(roleBinding.Subjects); i++ {
				if roleBinding.Subjects[i].Kind == v1.UserKind && slice.ContainsString(users, roleBinding.Subjects[i].Name, nil) {
					roleBinding.Subjects = append(roleBinding.Subjects[:i], roleBinding.Subjects[i+1:]...)
					modify = true
				}
			}
			if modify {
				_, err := k8sClient.RbacV1().RoleBindings(namespace).Update(&roleBinding)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func UnbindWorkspace(workspace *models.Workspace, users []string) error {

	err := unbindNamespacesRole(workspace.Namespaces, users)

	if err != nil {
		return err
	}

	err = unbindWorkspaceRole(workspace.Name, users)

	if err != nil {
		return err
	}

	return nil
}

func CreateWorkspaceRoleBinding(workspace *models.Workspace, username string, role string) error {

	k8sClient := k8s.Client()

	for _, roleName := range constants.WorkSpaceRoles {
		roleBinding, err := k8sClient.RbacV1().ClusterRoleBindings().Get(fmt.Sprintf("system:%s:%s", workspace.Name, roleName), meta_v1.GetOptions{})

		if err != nil {
			if apierrors.IsNotFound(err) {
				go WorkspaceRoleInit(workspace)
			}
			return err
		}

		modify := false

		for i, v := range roleBinding.Subjects {
			if v.Kind == v1.UserKind && v.Name == username {
				if roleName == role {
					return nil
				} else {
					modify = true
					roleBinding.Subjects = append(roleBinding.Subjects[:i], roleBinding.Subjects[i+1:]...)
					if roleName == constants.WorkspaceAdmin || roleName == constants.WorkspaceViewer {
						go deleteDevopsRoleBinding(workspace.Name, "", username)
					}
					break
				}
			}
		}

		if roleName == role {
			modify = true
			roleBinding.Subjects = append(roleBinding.Subjects, v1.Subject{Kind: v1.UserKind, Name: username})
			if roleName == constants.WorkspaceAdmin {
				go createDevopsRoleBinding(workspace.Name, "", username, constants.DevopsOwner)
			} else if roleName == constants.WorkspaceViewer {
				go createDevopsRoleBinding(workspace.Name, "", username, constants.DevopsReporter)
			}
		}

		if !modify {
			continue
		}

		_, err = k8sClient.RbacV1().ClusterRoleBindings().Update(roleBinding)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetDevOpsProjects(workspaceName string) ([]string, error) {

	db := mysql.Client()
	defer db.Close()

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

func GetOrgMembers(workspace string) ([]string, error) {
	ws, err := Detail(workspace)
	if err != nil {
		return nil, err
	}
	return ws.Members, nil
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

func Count() (int, error) {

	result, err := http.Get(fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/groups/count", constants.AccountAPIServer))

	if err != nil {
		return 0, err
	}

	defer result.Body.Close()

	data, err := ioutil.ReadAll(result.Body)

	if err != nil {
		return 0, err
	}

	if result.StatusCode > 200 {
		return 0, kserr.Parse(data)
	}
	var count map[string]json.Number

	err = json.Unmarshal(data, &count)

	if err != nil {
		return 0, err
	}

	value := count["total_count"]

	v, err := value.Int64()

	if err != nil {
		return 0, err
	}

	return int(v), nil
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
	defer db.Close()

	var count int
	if err := db.Model(&models.WorkspaceDPBinding{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func GetAllAccountNums() (int, error) {
	result, err := http.Get(fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/users", constants.AccountAPIServer))

	if err != nil {
		return 0, err
	}

	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return 0, err
	}
	if result.StatusCode > 200 {
		return 0, kserr.Parse(data)
	}
	var count map[string]json.Number

	err = json.Unmarshal(data, &count)

	if err != nil {
		return 0, err
	}

	value := count["total_count"]

	v, err := value.Int64()

	if err != nil {
		return 0, err
	}

	return int(v), nil
}
