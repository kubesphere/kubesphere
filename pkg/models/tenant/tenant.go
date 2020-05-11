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
	eventsv1alpha1 "kubesphere.io/kubesphere/pkg/api/events/v1alpha1"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizerfactory"
	unionauthorizer "kubesphere.io/kubesphere/pkg/apiserver/authorization/union"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/events"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	eventsclient "kubesphere.io/kubesphere/pkg/simple/client/events"
)

type Interface interface {
	ListWorkspaces(user user.Info, query *query.Query) (*api.ListResult, error)
	ListNamespaces(user user.Info, workspace string, query *query.Query) (*api.ListResult, error)
	Events(user user.Info, queryParam *eventsv1alpha1.Query) (*eventsv1alpha1.APIResponse, error)
}

type tenantOperator struct {
	am             am.AccessManagementInterface
	authorizer     authorizer.Authorizer
	resourceGetter *resourcesv1alpha3.ResourceGetter
	events         events.Interface
}

func New(informers informers.InformerFactory, evtsClient eventsclient.Client) Interface {
	amOperator := am.NewAMOperator(informers)
	rbacAuthorizer := authorizerfactory.NewRBACAuthorizer(amOperator)
	opaAuthorizer := authorizerfactory.NewOPAAuthorizer(amOperator)
	authorizers := unionauthorizer.New(opaAuthorizer, rbacAuthorizer)
	return &tenantOperator{
		am:             amOperator,
		authorizer:     authorizers,
		resourceGetter: resourcesv1alpha3.NewResourceGetter(informers),
		events:         events.NewEventsOperator(evtsClient),
	}
}

func (t *tenantOperator) ListWorkspaces(user user.Info, queryParam *query.Query) (*api.ListResult, error) {

	listWS := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		APIGroup:        "tenant.kubesphere.io",
		APIVersion:      "v1alpha2",
		Resource:        "workspaces",
		ResourceRequest: true,
	}

	decision, _, err := t.authorizer.Authorize(listWS)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if decision == authorizer.DecisionAllow {

		result, err := t.resourceGetter.List(tenantv1alpha1.ResourcePluralWorkspace, "", queryParam)

		if err != nil {
			klog.Error(err)
			return nil, err
		}

		return result, nil
	}

	workspaceRoleBindings, err := t.am.ListWorkspaceRoleBindings(user.GetName(), "")

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	workspaces := make([]runtime.Object, 0)

	for _, roleBinding := range workspaceRoleBindings {

		workspaceName := roleBinding.Labels[tenantv1alpha1.WorkspaceLabel]
		workspace, err := t.resourceGetter.Get(tenantv1alpha1.ResourcePluralWorkspace, "", workspaceName)

		if errors.IsNotFound(err) {
			klog.Warningf("workspace role binding: %+v found but workspace not exist", roleBinding.ObjectMeta)
			continue
		}

		if err != nil {
			klog.Error(err)
			return nil, err
		}

		if !contains(workspaces, workspace) {
			workspaces = append(workspaces, workspace)
		}
	}

	result := resources.DefaultList(workspaces, queryParam, func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*tenantv1alpha1.Workspace).ObjectMeta, right.(*tenantv1alpha1.Workspace).ObjectMeta, field)
	}, func(workspace runtime.Object, filter query.Filter) bool {
		return resources.DefaultObjectMetaFilter(workspace.(*tenantv1alpha1.Workspace).ObjectMeta, filter)
	})

	return result, nil
}

func (t *tenantOperator) ListNamespaces(user user.Info, workspace string, queryParam *query.Query) (*api.ListResult, error) {

	listNSInWS := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		APIGroup:        "",
		APIVersion:      "v1",
		Workspace:       workspace,
		Resource:        "namespaces",
		ResourceRequest: true,
	}

	decision, _, err := t.authorizer.Authorize(listNSInWS)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if decision == authorizer.DecisionAllow {

		queryParam.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s=%s", tenantv1alpha1.WorkspaceLabel, workspace))

		result, err := t.resourceGetter.List("namespaces", "", queryParam)

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

	namespaces := make([]runtime.Object, 0)

	for _, roleBinding := range roleBindings {
		namespace, err := t.resourceGetter.Get("namespaces", "", roleBinding.Namespace)

		if err != nil {
			klog.Error(err)
			return nil, err
		}

		// skip if not controlled by the specified workspace
		if ns := namespace.(*corev1.Namespace); ns.Labels[tenantv1alpha1.WorkspaceLabel] != workspace {
			continue
		}

		if !contains(namespaces, namespace) {
			namespaces = append(namespaces, namespace)
		}
	}

	result := resources.DefaultList(namespaces, queryParam, func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*corev1.Namespace).ObjectMeta, right.(*corev1.Namespace).ObjectMeta, field)
	}, func(object runtime.Object, filter query.Filter) bool {
		namespace := object.(*corev1.Namespace).ObjectMeta
		if workspaceLabel, ok := namespace.Labels[tenantv1alpha1.WorkspaceLabel]; !ok || workspaceLabel != workspace {
			return false
		}
		return resources.DefaultObjectMetaFilter(namespace, filter)
	})

	return result, nil
}

func (t *tenantOperator) Events(user user.Info, queryParam *eventsv1alpha1.Query) (*eventsv1alpha1.APIResponse, error) {
	listEvts := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		APIGroup:        "",
		APIVersion:      "v1",
		Resource:        "events",
		ResourceRequest: true,
	}
	decision, _, err := t.authorizer.Authorize(listEvts)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var workspaces []*tenantv1alpha1.Workspace
	lw, err := t.ListWorkspaces(user, query.New())
	if err != nil {
		return nil, err
	}
	for _, r := range lw.Items {
		if ws, ok := r.(*tenantv1alpha1.Workspace); ok {
			workspaces = append(workspaces, ws)
		}
	}
	var namespaces []*corev1.Namespace
	ln, err := t.ListNamespaces(user, "", query.New())
	if err != nil {
		return nil, err
	}
	for _, r := range ln.Items {
		if ns, ok := r.(*corev1.Namespace); ok {
			namespaces = append(namespaces, ns)
		}
	}

	return t.events.Events(queryParam, func(filter *eventsclient.Filter) {
		filter.InvolvedObjectNamespaceMap =
			events.IntersectedNamespaces(queryParam, workspaces, namespaces, decision == authorizer.DecisionAllow)
	})
}

func contains(objects []runtime.Object, object runtime.Object) bool {
	for _, item := range objects {
		if item == object {
			return true
		}
	}
	return false
}
