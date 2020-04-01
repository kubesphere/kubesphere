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
package tenant

import (
	"k8s.io/api/core/v1"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
)

type Interface interface {
	CreateNamespace(workspace string, namespace *v1.Namespace, username string) (*v1.Namespace, error)
	DeleteNamespace(workspace, namespace string) error
	DescribeWorkspace(username, workspace string) (*v1alpha1.Workspace, error)
	ListWorkspaces(username string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	ListNamespaces(username string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	ListDevopsProjects(conditions *params.Conditions, orderBy string, reverse bool, limit int, offset int) (*models.PageableResponse, error)
}

type tenantOperator struct {
	workspaces WorkspaceInterface
	namespaces NamespaceInterface
	am         am.AccessManagementInterface
	devops     DevOpsProjectLister
}

func (t *tenantOperator) ListDevopsProjects(conditions *params.Conditions, orderBy string, reverse bool, limit int, offset int) (*models.PageableResponse, error) {
	return t.devops.ListDevOpsProjects(conditions.Match["workspace"], conditions, orderBy, reverse, limit, offset)
}

func (t *tenantOperator) DeleteNamespace(workspace, namespace string) error {
	return t.workspaces.DeleteNamespace(workspace, namespace)
}

func New(client kubernetes.Interface, informers k8sinformers.SharedInformerFactory, ksinformers ksinformers.SharedInformerFactory, db *mysql.Database) Interface {
	amOperator := am.NewAMOperator(client, informers)
	return &tenantOperator{
		workspaces: newWorkspaceOperator(client, informers, ksinformers, amOperator, db),
		namespaces: newNamespaceOperator(client, informers, amOperator),
		am:         amOperator,
	}
}

func (t *tenantOperator) CreateNamespace(workspaceName string, namespace *v1.Namespace, username string) (*v1.Namespace, error) {
	return t.namespaces.CreateNamespace(workspaceName, namespace, username)
}

func (t *tenantOperator) DescribeWorkspace(username, workspaceName string) (*v1alpha1.Workspace, error) {
	workspace, err := t.workspaces.GetWorkspace(workspaceName)

	if err != nil {
		return nil, err
	}

	return workspace, nil
}

func (t *tenantOperator) ListWorkspaces(username string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	panic("implement me")
}

func (t *tenantOperator) ListNamespaces(username string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	namespaces, err := t.namespaces.Search(username, conditions, orderBy, reverse)

	if err != nil {
		return nil, err
	}

	// limit offset
	result := make([]interface{}, 0)
	for i, v := range namespaces {
		if len(result) < limit && i >= offset {
			result = append(result, v)
		}
	}

	return &models.PageableResponse{Items: result, TotalCount: len(namespaces)}, nil
}
