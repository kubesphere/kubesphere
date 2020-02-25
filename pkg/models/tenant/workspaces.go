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
package tenant

import (
	"fmt"
	core "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/db"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
	clientset "kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sort"
	"strings"

	"k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type InWorkspaceUser struct {
	*iam.User
	WorkspaceRole string `json:"workspaceRole"`
}

type WorkspaceInterface interface {
	GetWorkspace(workspace string) (*v1alpha1.Workspace, error)
	SearchWorkspace(username string, conditions *params.Conditions, orderBy string, reverse bool) ([]*v1alpha1.Workspace, error)
	ListNamespaces(workspace string) ([]*core.Namespace, error)
	DeleteNamespace(workspace, namespace string) error
	RemoveUser(user, workspace string) error
	AddUser(workspace string, user *InWorkspaceUser) error
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
	am          iam.AccessManagementInterface

	// TODO: use db interface instead of mysql client
	// we can refactor this after rewrite devops using crd
	db *mysql.Database
}

func newWorkspaceOperator(client kubernetes.Interface, informers informers.SharedInformerFactory, ksinformers externalversions.SharedInformerFactory, am iam.AccessManagementInterface, db *mysql.Database) WorkspaceInterface {
	return &workspaceOperator{
		client:      client,
		informers:   informers,
		ksInformers: ksinformers,
		am:          am,
		db:          db,
	}
}

func (w *workspaceOperator) ListNamespaces(workspace string) ([]*core.Namespace, error) {
	namespaces, err := w.informers.Core().V1().Namespaces().Lister().List(labels.SelectorFromSet(labels.Set{constants.WorkspaceLabelKey: workspace}))

	if err != nil {
		return nil, err
	}

	return namespaces, nil
}

func (w *workspaceOperator) DeleteNamespace(workspace string, namespace string) error {
	ns, err := w.informers.Core().V1().Namespaces().Lister().Get(namespace)
	if err != nil {
		return err
	}

	if ns.Labels[constants.WorkspaceLabelKey] == workspace {
		deletePolicy := metav1.DeletePropagationBackground
		return w.client.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
	} else {
		return apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "workspace"}, workspace)
	}
}

func (w *workspaceOperator) RemoveUser(workspace string, username string) error {
	workspaceRole, err := w.am.GetWorkspaceRole(workspace, username)
	if err != nil {
		return err
	}

	err = w.deleteWorkspaceRoleBinding(workspace, username, workspaceRole.Annotations[constants.DisplayNameAnnotationKey])
	if err != nil {
		return err
	}

	return nil
}

func (w *workspaceOperator) AddUser(workspaceName string, user *InWorkspaceUser) error {

	workspaceRole, err := w.am.GetWorkspaceRole(workspaceName, user.Username)

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
		err := w.deleteWorkspaceRoleBinding(workspaceName, user.Username, workspaceRole.Annotations[constants.DisplayNameAnnotationKey])
		if err != nil {
			klog.Errorf("delete workspace role binding failed: %+v", err)
			return err
		}
	} else if currentWorkspaceRoleName != "" {
		return nil
	}

	return w.createWorkspaceRoleBinding(workspaceName, user.Username, user.WorkspaceRole)
}

func (w *workspaceOperator) createWorkspaceRoleBinding(workspace, username string, role string) error {

	if !sliceutil.HasString(constants.WorkSpaceRoles, role) {
		return apierrors.NewNotFound(schema.GroupResource{Resource: "workspace role"}, role)
	}

	roleBindingName := fmt.Sprintf("workspace:%s:%s", workspace, strings.TrimPrefix(role, "workspace-"))
	workspaceRoleBinding, err := w.informers.Rbac().V1().ClusterRoleBindings().Lister().Get(roleBindingName)
	if err != nil {
		return err
	}

	if !iam.ContainsUser(workspaceRoleBinding.Subjects, username) {
		workspaceRoleBinding = workspaceRoleBinding.DeepCopy()
		workspaceRoleBinding.Subjects = append(workspaceRoleBinding.Subjects, v1.Subject{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: username})
		_, err = w.client.RbacV1().ClusterRoleBindings().Update(workspaceRoleBinding)
		if err != nil {
			klog.Errorf("update workspace role binding failed: %+v", err)
			return err
		}
	}

	return nil
}

func (w *workspaceOperator) deleteWorkspaceRoleBinding(workspace, username string, role string) error {

	if !sliceutil.HasString(constants.WorkSpaceRoles, role) {
		return apierrors.NewNotFound(schema.GroupResource{Resource: "workspace role"}, role)
	}

	roleBindingName := fmt.Sprintf("workspace:%s:%s", workspace, strings.TrimPrefix(role, "workspace-"))

	workspaceRoleBinding, err := w.informers.Rbac().V1().ClusterRoleBindings().Lister().Get(roleBindingName)
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

	workspaceRoleBinding, err = w.client.RbacV1().ClusterRoleBindings().Update(workspaceRoleBinding)

	return err
}

func (w *workspaceOperator) CountDevopsProjectsInWorkspace(workspaceName string) (int, error) {
	if w.db == nil {
		return 0, clientset.ErrClientSetNotEnabled
	}

	query := w.db.Select(devops.DevOpsProjectIdColumn).
		From(devops.DevOpsProjectTableName).
		Where(db.And(db.Eq(devops.DevOpsProjectWorkSpaceColumn, workspaceName),
			db.Eq(devops.StatusColumn, devops.StatusActive)))

	devOpsProjects := make([]string, 0)

	if _, err := query.Load(&devOpsProjects); err != nil {
		return 0, err
	}
	return len(devOpsProjects), nil
}

func (w *workspaceOperator) CountUsersInWorkspace(workspace string) (int, error) {
	count, err := w.CountUsersInWorkspace(workspace)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (w *workspaceOperator) CountOrgRoles() (int, error) {
	return len(constants.WorkSpaceRoles), nil
}

func (w *workspaceOperator) CountNamespacesInWorkspace(workspace string) (int, error) {
	ns, err := w.ListNamespaces(workspace)
	if err != nil {
		return 0, err
	}

	return len(ns), nil
}

func (*workspaceOperator) match(match map[string]string, item *v1alpha1.Workspace) bool {
	for k, v := range match {
		switch k {
		case v1alpha2.Name:
			names := strings.Split(v, "|")
			if !sliceutil.HasString(names, item.Name) {
				return false
			}
		case v1alpha2.Keyword:
			if !strings.Contains(item.Name, v) && !contains(item.Labels, "", v) && !contains(item.Annotations, "", v) {
				return false
			}
		default:
			// label not exist or value not equal
			if val, ok := item.Labels[k]; !ok || val != v {
				return false
			}
		}
	}
	return true
}

func (*workspaceOperator) fuzzy(fuzzy map[string]string, item *v1alpha1.Workspace) bool {

	for k, v := range fuzzy {
		switch k {
		case v1alpha2.Name:
			if !strings.Contains(item.Name, v) && !strings.Contains(item.Annotations[constants.DisplayNameAnnotationKey], v) {
				return false
			}
		default:
			return false
		}
	}

	return true
}

func (*workspaceOperator) compare(a, b *v1alpha1.Workspace, orderBy string) bool {
	switch orderBy {
	case v1alpha2.CreateTime:
		return a.CreationTimestamp.Time.Before(b.CreationTimestamp.Time)
	case v1alpha2.Name:
		fallthrough
	default:
		return strings.Compare(a.Name, b.Name) <= 0
	}
}

func (w *workspaceOperator) SearchWorkspace(username string, conditions *params.Conditions, orderBy string, reverse bool) ([]*v1alpha1.Workspace, error) {
	rules, err := w.am.GetClusterPolicyRules(username)

	if err != nil {
		return nil, err
	}

	workspaces := make([]*v1alpha1.Workspace, 0)

	if iam.RulesMatchesRequired(rules, rbacv1.PolicyRule{Verbs: []string{"list"}, APIGroups: []string{"tenant.kubesphere.io"}, Resources: []string{"workspaces"}}) {
		workspaces, err = w.ksInformers.Tenant().V1alpha1().Workspaces().Lister().List(labels.Everything())
		if err != nil {
			return nil, err
		}
	} else {
		workspaceRoles, err := w.am.GetWorkspaceRoleMap(username)
		if err != nil {
			return nil, err
		}
		for k := range workspaceRoles {
			workspace, err := w.ksInformers.Tenant().V1alpha1().Workspaces().Lister().Get(k)
			if err != nil {
				return nil, err
			}
			workspaces = append(workspaces, workspace)
		}
	}

	result := make([]*v1alpha1.Workspace, 0)

	for _, workspace := range workspaces {
		if w.match(conditions.Match, workspace) && w.fuzzy(conditions.Fuzzy, workspace) {
			result = append(result, workspace)
		}
	}

	// order & reverse
	sort.Slice(result, func(i, j int) bool {
		if reverse {
			i, j = j, i
		}
		return w.compare(result[i], result[j], orderBy)
	})

	return result, nil
}

func (w *workspaceOperator) GetWorkspace(workspaceName string) (*v1alpha1.Workspace, error) {
	return w.ksInformers.Tenant().V1alpha1().Workspaces().Lister().Get(workspaceName)
}

func contains(m map[string]string, key, value string) bool {
	for k, v := range m {
		if key == "" {
			if strings.Contains(k, value) || strings.Contains(v, value) {
				return true
			}
		} else if k == key && strings.Contains(v, value) {
			return true
		}
	}
	return false
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

func (w *workspaceOperator) CountWorkspaces() (int, error) {
	ws, err := w.ksInformers.Tenant().V1alpha1().Workspaces().Lister().List(labels.Everything())
	if err != nil {
		return 0, err
	}

	return len(ws), nil
}
