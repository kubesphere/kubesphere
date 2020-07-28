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
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/constants"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

func (t *tenantOperator) ListDevOpsProjects(user user.Info, workspace string, queryParam *query.Query) (*api.ListResult, error) {
	scope := request.ClusterScope
	if workspace != "" {
		scope = request.WorkspaceScope
		// filter by workspace
		queryParam.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s=%s", tenantv1alpha1.WorkspaceLabel, workspace))
	}

	listDevOps := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		Workspace:       workspace,
		Resource:        "devops",
		ResourceRequest: true,
		ResourceScope:   scope,
	}

	decision, _, err := t.authorizer.Authorize(listDevOps)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// allowed list devops project in the specified scope
	if decision == authorizer.DecisionAllow {
		result, err := t.resourceGetter.List(devopsv1alpha3.ResourcePluralDevOpsProject, "", queryParam)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return result, nil
	}

	roleBindings, err := t.am.ListRoleBindings(user.GetName(), "")
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// list the devops projects that the user joined
	devopsProjects := make([]runtime.Object, 0)
	for _, roleBinding := range roleBindings {
		// the namespace to which role binding belongs
		obj, err := t.resourceGetter.Get("namespaces", "", roleBinding.Namespace)
		if err != nil {
			klog.Error(err)
			return nil, err
		}

		controlledDevOpsProject := obj.(*corev1.Namespace).Labels[constants.DevOpsProjectLabelKey]
		// skip if not controlled by devops project
		if controlledDevOpsProject == "" {
			continue
		}

		devopsProject, err := t.resourceGetter.Get(devopsv1alpha3.ResourcePluralDevOpsProject, "", controlledDevOpsProject)
		if err != nil {
			if errors.IsNotFound(err) {
				klog.Warningf("orphan devops project found: %s", roleBinding.Namespace)
				continue
			}
			klog.Error(err)
			return nil, err
		}

		// avoid duplication
		if !contains(devopsProjects, devopsProject) {
			devopsProjects = append(devopsProjects, devopsProject)
		}
	}

	// devops project filtering
	result := resources.DefaultList(devopsProjects, queryParam, func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*devopsv1alpha3.DevOpsProject).ObjectMeta, right.(*devopsv1alpha3.DevOpsProject).ObjectMeta, field)
	}, func(object runtime.Object, filter query.Filter) bool {
		devopsProject := object.(*devopsv1alpha3.DevOpsProject)
		return resources.DefaultObjectMetaFilter(devopsProject.ObjectMeta, filter)
	})
	return result, nil
}
