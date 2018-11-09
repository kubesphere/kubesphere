package workspaces

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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
	clientV1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/kubernetes/pkg/util/slice"

	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/controllers"
	"kubesphere.io/kubesphere/pkg/models/iam"
	ksErr "kubesphere.io/kubesphere/pkg/util/errors"
)

const (
	WorkspaceKey = "kubesphere.io/workspace"
)

var WorkSpaceRoles = []string{"admin", "operator", "viewer"}

func UnBindDevopsProject(workspace string, devops string) error {
	db := client.NewSharedDBClient()
	defer db.Close()
	return db.Delete(&WorkspaceDPBinding{Workspace: workspace, DevOpsProject: devops}).Error
}

func DeleteDevopsProject(username string, devops string) error {
	request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s/api/v1alpha/projects/%s", constants.DevopsAPIServer, devops), nil)
	request.Header.Add("X-Token-Username", username)

	result, err := http.DefaultClient.Do(request)

	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return err
	}
	if result.StatusCode > 200 {
		return ksErr.Wrap(data)
	}
	return nil
}

func CreateDevopsProject(username string, devops DevopsProject) (*DevopsProject, error) {

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

	data, err = ioutil.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	if result.StatusCode > 200 {
		return nil, ksErr.Wrap(data)
	}

	var project DevopsProject

	err = json.Unmarshal(data, &project)

	if err != nil {
		return nil, err
	}

	return &project, nil
}

func ListNamespaceByUser(workspaceName string, username string) ([]*core.Namespace, error) {

	namespaces, err := Namespaces(workspaceName)

	if err != nil {
		return nil, err
	}

	clusterRoles, err := iam.GetClusterRoles(username)

	if err != nil {
		return nil, err
	}

	rules := make([]v1.PolicyRule, 0)

	for _, clusterRole := range clusterRoles {
		rules = append(rules, clusterRole.Rules...)
	}

	namespacesManager := v1.PolicyRule{APIGroups: []string{"kubesphere.io"}, ResourceNames: []string{workspaceName}, Verbs: []string{"get"}, Resources: []string{"workspaces/namespaces"}}

	if iam.RulesMatchesRequired(rules, namespacesManager) {
		return namespaces, nil
	} else {
		for i := 0; i < len(namespaces); i++ {
			roles, err := iam.GetRoles(namespaces[i].Name, username)
			if err != nil {
				return nil, err
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

	return namespaces, nil
}

func Namespaces(workspaceName string) ([]*core.Namespace, error) {

	lister := controllers.ResourceControllers.Controllers[controllers.Namespaces].Lister().(clientV1.NamespaceLister)

	namespaces, err := lister.List(labels.SelectorFromSet(labels.Set{"kubesphere.io/workspace": workspaceName}))

	if err != nil {
		return nil, err
	}

	if namespaces == nil {
		return make([]*core.Namespace, 0), nil
	}

	return namespaces, nil
}

func BindingDevopsProject(workspace string, devops string) error {
	db := client.NewSharedDBClient()
	defer db.Close()
	return db.Create(&WorkspaceDPBinding{Workspace: workspace, DevOpsProject: devops}).Error
}

func DeleteNamespace(workspace string, namespaceName string) error {
	namespace, err := client.NewK8sClient().CoreV1().Namespaces().Get(namespaceName, meta_v1.GetOptions{})
	if err != nil {
		return err
	}
	if namespace.Labels != nil && namespace.Labels["kubesphere.io/workspace"] == workspace {
		deletePolicy := meta_v1.DeletePropagationBackground
		return client.NewK8sClient().CoreV1().Namespaces().Delete(namespaceName, &meta_v1.DeleteOptions{PropagationPolicy: &deletePolicy})
	} else {
		return errors.New("resource not found")
	}

}

func Delete(workspace *Workspace) error {

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

	data, err := ioutil.ReadAll(result.Body)

	if err != nil {
		return err
	}

	if result.StatusCode > 200 {
		return ksErr.Wrap(data)
	}

	return nil
}

func release(workspace *Workspace) error {
	for _, namespace := range workspace.Namespaces {
		err := DeleteNamespace(workspace.Name, namespace)
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	for _, devops := range workspace.DevopsProjects {
		err := DeleteDevopsProject(workspace.Creator, devops)
		if err != nil {
			return err
		}
	}

	err := workspaceRoleRelease(workspace.Name)

	return err
}
func workspaceRoleRelease(workspace string) error {
	k8sClient := client.NewK8sClient()
	deletePolicy := meta_v1.DeletePropagationForeground

	for _, role := range WorkSpaceRoles {
		err := k8sClient.RbacV1().ClusterRoles().Delete(fmt.Sprintf("system:%s:%s", workspace, role), &meta_v1.DeleteOptions{PropagationPolicy: &deletePolicy})

		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	for _, role := range WorkSpaceRoles {
		err := k8sClient.RbacV1().ClusterRoleBindings().Delete(fmt.Sprintf("system:%s:%s", workspace, role), &meta_v1.DeleteOptions{PropagationPolicy: &deletePolicy})

		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func Create(workspace *Workspace) (*Workspace, error) {

	data, err := json.Marshal(workspace)

	if err != nil {
		return nil, err
	}

	result, err := http.Post(fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/groups", constants.AccountAPIServer), restful.MIME_JSON, bytes.NewReader(data))

	if err != nil {
		return nil, err
	}

	data, err = ioutil.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	if result.StatusCode > 200 {
		return nil, ksErr.Wrap(data)
	}

	var created Workspace

	err = json.Unmarshal(data, &created)

	if err != nil {
		return nil, err
	}

	created.Members = make([]string, 0)
	created.Namespaces = make([]string, 0)
	created.DevopsProjects = make([]string, 0)

	go WorkspaceRoleInit(workspace)

	return &created, nil
}

func Edit(workspace *Workspace) (*Workspace, error) {

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

	data, err = ioutil.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	if result.StatusCode > 200 {
		return nil, ksErr.Wrap(data)
	}

	var edited Workspace

	err = json.Unmarshal(data, &edited)

	if err != nil {
		return nil, err
	}

	return &edited, nil
}

func Detail(name string) (*Workspace, error) {

	result, err := http.Get(fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/groups/%s", constants.AccountAPIServer, name))

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	if result.StatusCode > 200 {
		return nil, ksErr.Wrap(data)
	}

	var group Group

	err = json.Unmarshal(data, &group)

	if err != nil {
		return nil, err
	}

	db := client.NewSharedDBClient()
	defer db.Close()

	workspace, err := convertGroupToWorkspace(db, group)

	if err != nil {
		return nil, err
	}

	return workspace, nil
}

// List all workspaces for the current user
func ListByUser(username string) ([]*Workspace, error) {

	clusterRoles, err := iam.GetClusterRoles(username)

	if err != nil {
		return nil, err
	}

	rules := make([]v1.PolicyRule, 0)

	for _, clusterRole := range clusterRoles {
		rules = append(rules, clusterRole.Rules...)
	}

	workspacesManager := v1.PolicyRule{APIGroups: []string{"kubesphere.io"}, Verbs: []string{"list", "get"}, Resources: []string{"workspaces"}}

	if iam.RulesMatchesRequired(rules, workspacesManager) {
		return fetch(nil)
	} else {
		workspaceNames := make([]string, 0)

		for _, clusterRole := range clusterRoles {
			if regexp.MustCompile("^system:\\w+:(admin|operator|viewer)$").MatchString(clusterRole.Name) {
				arr := strings.Split(clusterRole.Name, ":")
				workspaceNames = append(workspaceNames, arr[1])
			}
		}

		if len(workspaceNames) == 0 {
			return make([]*Workspace, 0), nil
		}

		return fetch(workspaceNames)
	}
}

func fetch(names []string) ([]*Workspace, error) {

	url := fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/groups", constants.AccountAPIServer)

	if names != nil {
		url = url + "?path=" + strings.Join(names, ",")
	}

	result, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	if result.StatusCode > 200 {
		return nil, ksErr.Wrap(data)
	}

	var groups []Group

	err = json.Unmarshal(data, &groups)

	if err != nil {
		return nil, err
	}

	db := client.NewSharedDBClient()

	defer db.Close()

	workspaces := make([]*Workspace, 0)
	for _, group := range groups {
		workspace, err := convertGroupToWorkspace(db, group)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, workspace)
	}

	return workspaces, nil
}

func DevopsProjects(workspace string) ([]DevopsProject, error) {

	db := client.NewSharedDBClient()
	defer db.Close()

	var workspaceDOPBindings []WorkspaceDPBinding

	if err := db.Where("workspace = ?", workspace).Find(&workspaceDOPBindings).Error; err != nil {
		return nil, err
	}

	devOpsProjects := make([]DevopsProject, 0)

	for _, workspaceDOPBinding := range workspaceDOPBindings {
		request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/api/v1alpha/projects/%s", constants.DevopsAPIServer, workspaceDOPBinding.DevOpsProject), nil)
		request.Header.Add("X-Token-Username", "admin")

		result, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(result.Body)

		if err != nil {
			return nil, err
		}

		if result.StatusCode == 403 || result.StatusCode == 404 {
			if err := db.Delete(&workspaceDOPBinding).Error; err != nil {
				return nil, err
			}
			continue
		}

		if result.StatusCode > 200 {
			return nil, ksErr.Wrap(data)
		}

		var project DevopsProject

		err = json.Unmarshal(data, &project)

		if err != nil {
			return nil, err
		}
		devOpsProjects = append(devOpsProjects, project)
	}

	return devOpsProjects, nil

}
func convertGroupToWorkspace(db *gorm.DB, group Group) (*Workspace, error) {
	namespaces, err := Namespaces(group.Name)

	if err != nil {
		return nil, err
	}

	namespacesNames := make([]string, 0)

	for _, namespace := range namespaces {
		namespacesNames = append(namespacesNames, namespace.Name)
	}

	var workspaceDOPBindings []WorkspaceDPBinding

	if err := db.Where("workspace = ?", group.Name).Find(&workspaceDOPBindings).Error; err != nil {
		return nil, err
	}

	devOpsProjects := make([]string, 0)

	for _, workspaceDOPBinding := range workspaceDOPBindings {
		devOpsProjects = append(devOpsProjects, workspaceDOPBinding.DevOpsProject)
	}

	workspace := Workspace{Group: group}
	workspace.Namespaces = namespacesNames
	workspace.DevopsProjects = devOpsProjects
	return &workspace, nil
}

func CreateNamespace(namespace *core.Namespace) (*core.Namespace, error) {
	return client.NewK8sClient().CoreV1().Namespaces().Create(namespace)
}

func Invite(workspaceName string, users []UserInvite) error {
	for _, user := range users {
		if !slice.ContainsString(WorkSpaceRoles, user.Role, nil) {
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

func Roles(workspace *Workspace) ([]*v1.ClusterRole, error) {
	roles := make([]*v1.ClusterRole, 0)

	k8sClient := client.NewK8sClient()

	for _, name := range WorkSpaceRoles {
		role, err := k8sClient.RbacV1().ClusterRoles().Get(fmt.Sprintf("system:%s:%s", workspace.Name, name), meta_v1.GetOptions{})

		if err != nil {
			if apierrors.IsNotFound(err) {
				go WorkspaceRoleInit(workspace)
			}
			return nil, err
		}

		role.Name = name
		roles = append(roles, role)
	}

	return roles, nil
}

func GetWorkspaceMembers(workspace string) ([]iam.User, error) {

	result, err := http.Get(fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/groups/%s/users", constants.AccountAPIServer, workspace))

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	if result.StatusCode > 200 {
		return nil, ksErr.Wrap(data)
	}

	var users []iam.User

	err = json.Unmarshal(data, &users)

	if err != nil {
		return nil, err
	}

	return users, nil

}

func WorkspaceRoleInit(workspace *Workspace) error {
	k8sClient := client.NewK8sClient()

	admin := new(v1.ClusterRole)
	admin.Name = fmt.Sprintf("system:%s:admin", workspace.Name)
	admin.Kind = iam.ClusterRoleKind
	admin.Rules = []v1.PolicyRule{
		// apis/kubesphere.io/v1alpha1/workspaces/sample
		// apis/kubesphere.io/v1alpha1/workspaces/sample/namespaces
		// apis/kubesphere.io/v1alpha1/workspaces/sample/devops
		// apis/kubesphere.io/v1alpha1/workspaces/sample/roles
		// apis/kubesphere.io/v1alpha1/workspaces/sample/members
		// apis/kubesphere.io/v1alpha1/workspaces/sample/members/admin

		{
			Verbs:         []string{"*"},
			APIGroups:     []string{"kubesphere.io", "account.kubesphere.io"},
			ResourceNames: []string{workspace.Name},
			Resources:     []string{"workspaces", "workspaces/*"},
		},

		// post apis/kubesphere.io/v1alpha1/workspaces/sample/namespaces

		{
			Verbs:         []string{"create"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{workspace.Name},
			Resources:     []string{"workspaces/namespaces"},
		},

		// post apis/kubesphere.io/v1alpha1/workspaces/sample/members

		{
			Verbs:         []string{"create"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{workspace.Name},
			Resources:     []string{"workspaces/members"},
		},

		// post apis/kubesphere.io/v1alpha1/workspaces/sample/devops
		{
			Verbs:         []string{"create"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{workspace.Name},
			Resources:     []string{"workspaces/devops"},
		},
		// TODO have risks
		// get apis/apps/v1/namespaces/proj1/deployments/?labelSelector
		// post api/v1/namespaces/project-0vya57/limitranges
		{
			Verbs:     []string{"*"},
			APIGroups: []string{"", "apps", "extensions", "batch"},
			Resources: []string{"limitranges", "deployments", "configmaps", "secrets", "jobs", "cronjobs", "persistentvolumes", "statefulsets", "daemonsets", "ingresses", "services", "pods/*", "pods", "events", "deployments/scale"},
		},
		// get apis/kubesphere.io/v1alpha1/quota/namespaces/proj1
		{
			Verbs:     []string{"get"},
			APIGroups: []string{"kubesphere.io"},
			Resources: []string{"quota/*"},
		},
		// get api/v1/namespaces/proj1
		{
			Verbs:     []string{"get"},
			APIGroups: []string{""},
			Resources: []string{"namespaces", "serviceaccounts", "configmaps"},
		},
		// get api/v1/namespaces/proj1/serviceaccounts
		// get api/v1/namespaces/proj1/configmaps
		// get api/v1/namespaces/proj1/secrets

		{
			Verbs:     []string{"list"},
			APIGroups: []string{""},
			Resources: []string{"serviceaccounts", "configmaps", "secrets"},
		},

		// get apis/kubesphere.io/v1alpha1/status/namespaces/proj1
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{"namespaces"},
			Resources:     []string{"status/*"},
		},
		// apis/kubesphere.io/v1alpha1/namespaces/proj1/router
		{
			Verbs:     []string{"list"},
			APIGroups: []string{"kubesphere.io"},
			Resources: []string{"router"},
		},
		// get apis/kubesphere.io/v1alpha1/registries/proj1
		{
			Verbs:     []string{"get"},
			APIGroups: []string{"kubesphere.io"},
			Resources: []string{"registries"},
		},

		// get apis/kubesphere.io/v1alpha1/monitoring/namespaces/proj1

		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{"namespaces"},
			Resources:     []string{"monitoring/*"},
		},

		// get apis/kubesphere.io/v1alpha1/resources/persistent-volume-claims
		// get apis/kubesphere.io/v1alpha1/resources/deployments
		// get apis/kubesphere.io/v1alpha1/resources/statefulsets
		// get apis/kubesphere.io/v1alpha1/resources/daemonsets
		// get apis/kubesphere.io/v1alpha1/resources/jobs
		// get apis/kubesphere.io/v1alpha1/resources/cronjobs
		// get apis/kubesphere.io/v1alpha1/resources/persistent-volume-claims
		// get apis/kubesphere.io/v1alpha1/resources/services
		// get apis/kubesphere.io/v1alpha1/resources/ingresses
		// get apis/kubesphere.io/v1alpha1/resources/secrets
		// get apis/kubesphere.io/v1alpha1/resources/configmaps
		// get apis/kubesphere.io/v1alpha1/resources/roles

		{
			Verbs:     []string{"get"},
			APIGroups: []string{"kubesphere.io"},
			Resources: []string{"resources"},
		},

		// apis/account.kubesphere.io/v1alpha1/users
		// apis/account.kubesphere.io/v1alpha1/namespaces/proj1/users
		{
			Verbs:     []string{"list"},
			APIGroups: []string{"account.kubesphere.io"},
			Resources: []string{"users"},
		},

		// apis/kubesphere.io/v1alpha1/monitoring/workspaces/sample?metrics_filter=
		// apis/kubesphere.io/v1alpha1/monitoring/workspaces/sample/pods?step=30m
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{"workspaces"},
			Resources:     []string{"monitoring/" + workspace.Name},
		},
	}

	admin.Labels = map[string]string{"creator": "system"}

	operator := new(v1.ClusterRole)
	operator.Name = fmt.Sprintf("system:%s:operator", workspace.Name)
	operator.Kind = iam.ClusterRoleKind
	operator.Rules = []v1.PolicyRule{
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			Resources:     []string{"workspaces"},
			ResourceNames: []string{workspace.Name},
		}, {
			Verbs:         []string{"create", "get"},
			APIGroups:     []string{"kubesphere.io"},
			Resources:     []string{"workspaces/namespaces", "workspaces/devops"},
			ResourceNames: []string{workspace.Name},
		},
		{
			Verbs:     []string{"get"},
			APIGroups: []string{"kubesphere.io"},
			Resources: []string{"registries"},
		},
	}

	operator.Labels = map[string]string{"creator": "system"}

	viewer := new(v1.ClusterRole)
	viewer.Name = fmt.Sprintf("system:%s:viewer", workspace.Name)
	viewer.Kind = iam.ClusterRoleKind
	viewer.Rules = []v1.PolicyRule{
		// apis/kubesphere.io/v1alpha1/workspaces/sample
		// apis/kubesphere.io/v1alpha1/workspaces/sample/namespaces
		// apis/kubesphere.io/v1alpha1/workspaces/sample/devops
		// apis/kubesphere.io/v1alpha1/workspaces/sample/roles
		// apis/kubesphere.io/v1alpha1/workspaces/sample/members
		// apis/kubesphere.io/v1alpha1/workspaces/sample/members/admin

		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io", "account.kubesphere.io"},
			ResourceNames: []string{workspace.Name},
			Resources:     []string{"workspaces", "workspaces/*"},
		},

		// post apis/kubesphere.io/v1alpha1/workspaces/sample/namespaces

		//{
		//	Verbs:         []string{"create"},
		//	APIGroups:     []string{"kubesphere.io"},
		//	ResourceNames: []string{workspace.Name},
		//	Resources:     []string{"workspaces/namespaces"},
		//},

		// post apis/kubesphere.io/v1alpha1/workspaces/sample/members

		//{
		//	Verbs:         []string{"create"},
		//	APIGroups:     []string{"kubesphere.io"},
		//	ResourceNames: []string{workspace.Name},
		//	Resources:     []string{"workspaces/members"},
		//},

		// post apis/kubesphere.io/v1alpha1/workspaces/sample/devops
		//{
		//	Verbs:         []string{"create"},
		//	APIGroups:     []string{"kubesphere.io"},
		//	ResourceNames: []string{workspace.Name},
		//	Resources:     []string{"workspaces/devops"},
		//},
		// TODO have risks
		// get apis/apps/v1/namespaces/proj1/deployments/?labelSelector
		// post api/v1/namespaces/project-0vya57/limitranges
		{
			Verbs:     []string{"get", "list"},
			APIGroups: []string{"", "apps", "extensions", "batch"},
			Resources: []string{"limitranges", "deployments", "configmaps", "secrets", "jobs", "cronjobs", "persistentvolumes", "statefulsets", "daemonsets", "ingresses", "services", "pods/*", "pods", "events", "deployments/scale"},
		},
		// get apis/kubesphere.io/v1alpha1/quota/namespaces/proj1
		{
			Verbs:     []string{"get"},
			APIGroups: []string{"kubesphere.io"},
			Resources: []string{"quota/*"},
		},
		// get api/v1/namespaces/proj1
		{
			Verbs:     []string{"get"},
			APIGroups: []string{""},
			Resources: []string{"namespaces", "serviceaccounts", "configmaps"},
		},
		// get api/v1/namespaces/proj1/serviceaccounts
		// get api/v1/namespaces/proj1/configmaps
		// get api/v1/namespaces/proj1/secrets

		{
			Verbs:     []string{"list"},
			APIGroups: []string{""},
			Resources: []string{"serviceaccounts", "configmaps", "secrets"},
		},

		// get apis/kubesphere.io/v1alpha1/status/namespaces/proj1
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{"namespaces"},
			Resources:     []string{"status/*"},
		},
		// apis/kubesphere.io/v1alpha1/namespaces/proj1/router
		{
			Verbs:     []string{"list"},
			APIGroups: []string{"kubesphere.io"},
			Resources: []string{"router"},
		},
		// get apis/kubesphere.io/v1alpha1/registries/proj1
		{
			Verbs:     []string{"get"},
			APIGroups: []string{"kubesphere.io"},
			Resources: []string{"registries"},
		},

		// get apis/kubesphere.io/v1alpha1/monitoring/namespaces/proj1

		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{"namespaces"},
			Resources:     []string{"monitoring/*"},
		},

		// get apis/kubesphere.io/v1alpha1/resources/persistent-volume-claims
		// get apis/kubesphere.io/v1alpha1/resources/deployments
		// get apis/kubesphere.io/v1alpha1/resources/statefulsets
		// get apis/kubesphere.io/v1alpha1/resources/daemonsets
		// get apis/kubesphere.io/v1alpha1/resources/jobs
		// get apis/kubesphere.io/v1alpha1/resources/cronjobs
		// get apis/kubesphere.io/v1alpha1/resources/persistent-volume-claims
		// get apis/kubesphere.io/v1alpha1/resources/services
		// get apis/kubesphere.io/v1alpha1/resources/ingresses
		// get apis/kubesphere.io/v1alpha1/resources/secrets
		// get apis/kubesphere.io/v1alpha1/resources/configmaps
		// get apis/kubesphere.io/v1alpha1/resources/roles

		{
			Verbs:     []string{"get"},
			APIGroups: []string{"kubesphere.io"},
			Resources: []string{"resources"},
		},

		// apis/account.kubesphere.io/v1alpha1/users
		// apis/account.kubesphere.io/v1alpha1/namespaces/proj1/users
		{
			Verbs:     []string{"list"},
			APIGroups: []string{"account.kubesphere.io"},
			Resources: []string{"users"},
		},

		// apis/kubesphere.io/v1alpha1/monitoring/workspaces/sample?metrics_filter=
		// apis/kubesphere.io/v1alpha1/monitoring/workspaces/sample/pods?step=30m
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"kubesphere.io"},
			ResourceNames: []string{"workspaces"},
			Resources:     []string{"monitoring/" + workspace.Name},
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

	_, err = k8sClient.RbacV1().ClusterRoles().Create(operator)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			log.Println("cluster role create failed", viewer.Name, err)
			return err
		}
	}

	operatorRoleBinding := new(v1.ClusterRoleBinding)
	operatorRoleBinding.Name = operator.Name
	operatorRoleBinding.RoleRef = v1.RoleRef{Kind: "ClusterRole", Name: operator.Name}
	operatorRoleBinding.Subjects = make([]v1.Subject, 0)
	_, err = k8sClient.RbacV1().ClusterRoleBindings().Create(operatorRoleBinding)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			log.Println("cluster rolebinding create failed", operatorRoleBinding.Name, err)
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
	k8sClient := client.NewK8sClient()

	for _, name := range WorkSpaceRoles {
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

	k8sClient := client.NewK8sClient()
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

func UnbindWorkspace(workspace *Workspace, users []string) error {

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

func CreateWorkspaceRoleBinding(workspace *Workspace, username string, role string) error {

	k8sClient := client.NewK8sClient()

	for _, roleName := range WorkSpaceRoles {
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
					if err != nil {
						return err
					}
					break
				}
			}
		}

		if roleName == role {
			modify = true
			roleBinding.Subjects = append(roleBinding.Subjects, v1.Subject{Kind: v1.UserKind, Name: username})
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

func GetDevOpsProjects(name string) ([]string, error) {

	db := client.NewSharedDBClient()
	defer db.Close()

	var workspaceDOPBindings []WorkspaceDPBinding

	if err := db.Where("workspace = ?", name).Find(&workspaceDOPBindings).Error; err != nil {
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
	return []string{"admin", "operator", "user"}, nil
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

func CountAll() (int, error) {
	result, err := http.Get(fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/groups/count", constants.AccountAPIServer))

	if err != nil {
		return 0, err
	}

	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return 0, err
	}
	if result.StatusCode > 200 {
		return 0, ksErr.Wrap(data)
	}
	var count map[string]interface{}

	err = json.Unmarshal(data, &count)

	if err != nil {
		return 0, err
	}
	val, ok := count["total"]

	if !ok {
		return 0, errors.New("not found")
	}

	switch val.(type) {
	case int:
		return val.(int), nil
	case float32:
		return int(val.(float32)), nil
	case float64:
		return int(val.(float64)), nil
	}

	return 0, errors.New("not found")
}

func GetAllOrgNums() (int, error) {
	count, err := CountAll()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetAllProjectNums() (int, error) {
	return controllers.ResourceControllers.Controllers[controllers.Namespaces].CountWithConditions(""), nil
}

func GetAllDevOpsProjectsNums() (int, error) {
	db := client.NewSharedDBClient()
	defer db.Close()

	var count int
	if err := db.Find(&WorkspaceDPBinding{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func GetAllAccountNums() (int, error) {
	result, err := http.Get(fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/users", constants.AccountAPIServer))

	if err != nil {
		return 0, err
	}

	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return 0, err
	}
	if result.StatusCode > 200 {
		return 0, ksErr.Wrap(data)
	}
	var count map[string]interface{}

	err = json.Unmarshal(data, &count)

	if err != nil {
		return 0, err
	}
	val, ok := count["total"]

	if !ok {
		return 0, errors.New("not found")
	}

	switch val.(type) {
	case int:
		return val.(int), nil
	case float32:
		return int(val.(float32)), nil
	case float64:
		return int(val.(float64)), nil
	}
	return 0, errors.New("not found")
}

// get cluster organizations name which contains at least one namespace,
func GetAllOrgAndProjList() (map[string][]string, map[string]string, error) {
	nsList, err := client.NewK8sClient().CoreV1().Namespaces().List(meta_v1.ListOptions{})
	if err != nil {
		glog.Errorln(err)
		return nil, nil, err
	}

	var workspaceNamespaceMap = make(map[string][]string)
	var namespaceWorkspaceMap = make(map[string]string)

	for _, item := range nsList.Items {
		ws, exist := item.Labels[WorkspaceKey]
		ns := item.Name
		if exist {
			if nsArray, exist := workspaceNamespaceMap[ws]; exist {
				nsArray = append(nsArray, ns)
				workspaceNamespaceMap[ws] = nsArray
			} else {
				var nsArray []string
				nsArray = append(nsArray, ns)
				workspaceNamespaceMap[ws] = nsArray
			}

			namespaceWorkspaceMap[ns] = ws
		} else {
			// this namespace do not belong to any workspaces
			namespaceWorkspaceMap[ns] = ""
		}
	}

	return workspaceNamespaceMap, namespaceWorkspaceMap, nil
}
