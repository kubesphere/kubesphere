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
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	ws "kubesphere.io/kubesphere/pkg/models/workspaces"
	"kubesphere.io/kubesphere/pkg/params"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"strconv"
)

var (
	workspaces = workspaceSearcher{}
	namespaces = namespaceSearcher{}
)

func CreateNamespace(workspaceName string, namespace *v1.Namespace, username string) (*v1.Namespace, error) {
	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string, 0)
	}
	if username != "" {
		namespace.Labels[constants.CreatorLabelKey] = username
	}

	namespace.Labels[constants.WorkspaceLabelKey] = workspaceName

	return k8s.Client().CoreV1().Namespaces().Create(namespace)
}

func ListWorkspaces(username string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	workspaces, err := workspaces.search(username, conditions, orderBy, reverse)

	if err != nil {
		return nil, err
	}

	// limit offset
	result := make([]interface{}, 0)
	for i, workspace := range workspaces {
		if len(result) < limit && i >= offset {
			workspace := workspace.DeepCopy()
			ns, err := ListNamespaces(username, &params.Conditions{Match: map[string]string{"kubesphere.io/workspace": workspace.Name}}, "", false, 1, 0)
			if err != nil {
				return nil, err
			}
			if workspace.Annotations == nil {
				workspace.Annotations = make(map[string]string)
			}
			workspace.Annotations["kubesphere.io/namespace-count"] = strconv.Itoa(ns.TotalCount)
			devops, err := ListDevopsProjects(workspace.Name, username, &params.Conditions{}, "", false, 1, 0)
			if err != nil {
				return nil, err
			}
			workspace.Annotations["kubesphere.io/devops-count"] = strconv.Itoa(devops.TotalCount)
			userCount, err := ws.WorkspaceUserCount(workspace.Name)
			if err != nil {
				return nil, err
			}
			workspace.Annotations["kubesphere.io/member-count"] = strconv.Itoa(userCount)
			result = append(result, workspace)
		}
	}

	return &models.PageableResponse{Items: result, TotalCount: len(workspaces)}, nil
}

func ListNamespaces(username string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	namespaces, err := namespaces.search(username, conditions, orderBy, reverse)

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
