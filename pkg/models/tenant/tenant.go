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
	"encoding/json"
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	auditingv1alpha1 "kubesphere.io/kubesphere/pkg/api/auditing/v1alpha1"
	eventsv1alpha1 "kubesphere.io/kubesphere/pkg/api/events/v1alpha1"
	loggingv1alpha2 "kubesphere.io/kubesphere/pkg/api/logging/v1alpha2"
	clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	tenantv1alpha2 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha2"
	typesv1beta1 "kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizerfactory"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/auditing"
	"kubesphere.io/kubesphere/pkg/models/events"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/logging"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	auditingclient "kubesphere.io/kubesphere/pkg/simple/client/auditing"
	eventsclient "kubesphere.io/kubesphere/pkg/simple/client/events"
	loggingclient "kubesphere.io/kubesphere/pkg/simple/client/logging"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"strings"
	"time"
)

type Interface interface {
	ListWorkspaces(user user.Info, query *query.Query) (*api.ListResult, error)
	ListNamespaces(user user.Info, workspace string, query *query.Query) (*api.ListResult, error)
	ListDevOpsProjects(user user.Info, workspace string, query *query.Query) (*api.ListResult, error)
	ListFederatedNamespaces(info user.Info, workspace string, param *query.Query) (*api.ListResult, error)
	CreateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error)
	CreateWorkspace(workspace *tenantv1alpha2.WorkspaceTemplate) (*tenantv1alpha2.WorkspaceTemplate, error)
	DeleteWorkspace(workspace string) error
	UpdateWorkspace(workspace *tenantv1alpha2.WorkspaceTemplate) (*tenantv1alpha2.WorkspaceTemplate, error)
	DescribeWorkspace(workspace string) (*tenantv1alpha2.WorkspaceTemplate, error)
	ListWorkspaceClusters(workspace string) (*api.ListResult, error)
	Events(user user.Info, queryParam *eventsv1alpha1.Query) (*eventsv1alpha1.APIResponse, error)
	QueryLogs(user user.Info, query *loggingv1alpha2.Query) (*loggingv1alpha2.APIResponse, error)
	ExportLogs(user user.Info, query *loggingv1alpha2.Query, writer io.Writer) error
	Auditing(user user.Info, queryParam *auditingv1alpha1.Query) (*auditingv1alpha1.APIResponse, error)
	DescribeNamespace(workspace, namespace string) (*corev1.Namespace, error)
	DeleteNamespace(workspace, namespace string) error
	UpdateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error)
	PatchNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error)
	PatchWorkspace(workspace string, data json.RawMessage) (*tenantv1alpha2.WorkspaceTemplate, error)
	ListClusters(info user.Info) (*api.ListResult, error)
}

type tenantOperator struct {
	am             am.AccessManagementInterface
	authorizer     authorizer.Authorizer
	k8sclient      kubernetes.Interface
	ksclient       kubesphere.Interface
	resourceGetter *resourcesv1alpha3.ResourceGetter
	events         events.Interface
	lo             logging.LoggingOperator
	auditing       auditing.Interface
}

func New(informers informers.InformerFactory, k8sclient kubernetes.Interface, ksclient kubesphere.Interface, evtsClient eventsclient.Client, loggingClient loggingclient.Interface, auditingclient auditingclient.Client) Interface {
	amOperator := am.NewReadOnlyOperator(informers)
	authorizer := authorizerfactory.NewRBACAuthorizer(amOperator)
	return &tenantOperator{
		am:             amOperator,
		authorizer:     authorizer,
		resourceGetter: resourcesv1alpha3.NewResourceGetter(informers),
		k8sclient:      k8sclient,
		ksclient:       ksclient,
		events:         events.NewEventsOperator(evtsClient),
		lo:             logging.NewLoggingOperator(loggingClient),
		auditing:       auditing.NewEventsOperator(auditingclient),
	}
}

func (t *tenantOperator) ListWorkspaces(user user.Info, queryParam *query.Query) (*api.ListResult, error) {

	listWS := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		APIGroup:        "*",
		Resource:        "workspaces",
		ResourceRequest: true,
		ResourceScope:   request.GlobalScope,
	}

	decision, _, err := t.authorizer.Authorize(listWS)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// allowed to list all workspaces
	if decision == authorizer.DecisionAllow {
		result, err := t.resourceGetter.List(tenantv1alpha2.ResourcePluralWorkspaceTemplate, "", queryParam)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return result, nil
	}

	// retrieving associated resources through role binding
	workspaceRoleBindings, err := t.am.ListWorkspaceRoleBindings(user.GetName(), "")
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	workspaces := make([]runtime.Object, 0)
	for _, roleBinding := range workspaceRoleBindings {
		workspaceName := roleBinding.Labels[tenantv1alpha1.WorkspaceLabel]
		obj, err := t.resourceGetter.Get(tenantv1alpha2.ResourcePluralWorkspaceTemplate, "", workspaceName)
		if errors.IsNotFound(err) {
			klog.Warningf("workspace role binding: %+v found but workspace not exist", roleBinding.Name)
			continue
		}
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		workspace := obj.(*tenantv1alpha2.WorkspaceTemplate)
		// label matching selector, remove duplicate entity
		if queryParam.Selector().Matches(labels.Set(workspace.Labels)) &&
			!contains(workspaces, workspace) {
			workspaces = append(workspaces, workspace)
		}
	}

	// use default pagination search logic
	result := resources.DefaultList(workspaces, queryParam, func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*tenantv1alpha2.WorkspaceTemplate).ObjectMeta, right.(*tenantv1alpha2.WorkspaceTemplate).ObjectMeta, field)
	}, func(workspace runtime.Object, filter query.Filter) bool {
		return resources.DefaultObjectMetaFilter(workspace.(*tenantv1alpha2.WorkspaceTemplate).ObjectMeta, filter)
	})

	return result, nil
}

func (t *tenantOperator) ListFederatedNamespaces(user user.Info, workspace string, queryParam *query.Query) (*api.ListResult, error) {

	nsScope := request.ClusterScope
	if workspace != "" {
		nsScope = request.WorkspaceScope
		// filter by workspace
		queryParam.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s=%s", tenantv1alpha1.WorkspaceLabel, workspace))
	}

	listNS := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		Workspace:       workspace,
		Resource:        "namespaces",
		ResourceRequest: true,
		ResourceScope:   nsScope,
	}

	decision, _, err := t.authorizer.Authorize(listNS)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// allowed to list all namespaces in the specified scope
	if decision == authorizer.DecisionAllow {
		result, err := t.resourceGetter.List(typesv1beta1.ResourcePluralFederatedNamespace, "", queryParam)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return result, nil
	}

	// retrieving associated resources through role binding
	roleBindings, err := t.am.ListRoleBindings(user.GetName(), "")
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	namespaces := make([]runtime.Object, 0)
	for _, roleBinding := range roleBindings {
		obj, err := t.resourceGetter.Get(typesv1beta1.ResourcePluralFederatedNamespace, roleBinding.Namespace, roleBinding.Namespace)
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			klog.Error(err)
			return nil, err
		}
		namespace := obj.(*typesv1beta1.FederatedNamespace)
		// label matching selector, remove duplicate entity
		if queryParam.Selector().Matches(labels.Set(namespace.Labels)) &&
			!contains(namespaces, namespace) {
			namespaces = append(namespaces, namespace)
		}
	}

	// use default pagination search logic
	result := resources.DefaultList(namespaces, queryParam, func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*typesv1beta1.FederatedNamespace).ObjectMeta, right.(*typesv1beta1.FederatedNamespace).ObjectMeta, field)
	}, func(object runtime.Object, filter query.Filter) bool {
		return resources.DefaultObjectMetaFilter(object.(*typesv1beta1.FederatedNamespace).ObjectMeta, filter)
	})

	return result, nil
}

func (t *tenantOperator) ListNamespaces(user user.Info, workspace string, queryParam *query.Query) (*api.ListResult, error) {
	nsScope := request.ClusterScope
	if workspace != "" {
		nsScope = request.WorkspaceScope
		// filter by workspace
		queryParam.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s=%s", tenantv1alpha1.WorkspaceLabel, workspace))
	}

	listNS := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		Workspace:       workspace,
		Resource:        "namespaces",
		ResourceRequest: true,
		ResourceScope:   nsScope,
	}

	decision, _, err := t.authorizer.Authorize(listNS)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// allowed to list all namespaces in the specified scope
	if decision == authorizer.DecisionAllow {
		result, err := t.resourceGetter.List("namespaces", "", queryParam)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return result, nil
	}

	// retrieving associated resources through role binding
	roleBindings, err := t.am.ListRoleBindings(user.GetName(), "")
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	namespaces := make([]runtime.Object, 0)
	for _, roleBinding := range roleBindings {
		obj, err := t.resourceGetter.Get("namespaces", "", roleBinding.Namespace)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		namespace := obj.(*corev1.Namespace)
		// label matching selector, remove duplicate entity
		if queryParam.Selector().Matches(labels.Set(namespace.Labels)) &&
			!contains(namespaces, namespace) {
			namespaces = append(namespaces, namespace)
		}
	}

	// use default pagination search logic
	result := resources.DefaultList(namespaces, queryParam, func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*corev1.Namespace).ObjectMeta, right.(*corev1.Namespace).ObjectMeta, field)
	}, func(object runtime.Object, filter query.Filter) bool {
		return resources.DefaultObjectMetaFilter(object.(*corev1.Namespace).ObjectMeta, filter)
	})

	return result, nil
}

// CreateNamespace adds a workspace label to namespace which indicates namespace is under the workspace
// The reason here why don't check the existence of workspace anymore is this function is only executed in host cluster.
// but if the host cluster is not authorized to workspace, there will be no workspace in host cluster.
func (t *tenantOperator) CreateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	return t.k8sclient.CoreV1().Namespaces().Create(labelNamespaceWithWorkspaceName(namespace, workspace))
}

// labelNamespaceWithWorkspaceName adds a kubesphere.io/workspace=[workspaceName] label to namespace which
// indicates namespace is under the workspace
func labelNamespaceWithWorkspaceName(namespace *corev1.Namespace, workspaceName string) *corev1.Namespace {
	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string, 0)
	}

	namespace.Labels[tenantv1alpha1.WorkspaceLabel] = workspaceName // label namespace with workspace name

	return namespace
}

func (t *tenantOperator) DescribeNamespace(workspace, namespace string) (*corev1.Namespace, error) {
	obj, err := t.resourceGetter.Get("namespaces", "", namespace)
	if err != nil {
		return nil, err
	}
	ns := obj.(*corev1.Namespace)
	if ns.Labels[tenantv1alpha1.WorkspaceLabel] != workspace {
		err := errors.NewNotFound(corev1.Resource("namespace"), namespace)
		klog.Error(err)
		return nil, err
	}
	return ns, nil
}

func (t *tenantOperator) DeleteNamespace(workspace, namespace string) error {
	_, err := t.DescribeNamespace(workspace, namespace)
	if err != nil {
		return err
	}
	return t.k8sclient.CoreV1().Namespaces().Delete(namespace, metav1.NewDeleteOptions(0))
}

func (t *tenantOperator) UpdateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	_, err := t.DescribeNamespace(workspace, namespace.Name)
	if err != nil {
		return nil, err
	}
	namespace = labelNamespaceWithWorkspaceName(namespace, workspace)
	return t.k8sclient.CoreV1().Namespaces().Update(namespace)
}

func (t *tenantOperator) PatchNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	_, err := t.DescribeNamespace(workspace, namespace.Name)
	if err != nil {
		return nil, err
	}
	if namespace.Labels != nil {
		namespace.Labels[tenantv1alpha1.WorkspaceLabel] = workspace
	}
	data, err := json.Marshal(namespace)
	if err != nil {
		return nil, err
	}
	return t.k8sclient.CoreV1().Namespaces().Patch(namespace.Name, types.MergePatchType, data)
}

func (t *tenantOperator) PatchWorkspace(workspace string, data json.RawMessage) (*tenantv1alpha2.WorkspaceTemplate, error) {
	return t.ksclient.TenantV1alpha2().WorkspaceTemplates().Patch(workspace, types.MergePatchType, data)
}

func (t *tenantOperator) CreateWorkspace(workspace *tenantv1alpha2.WorkspaceTemplate) (*tenantv1alpha2.WorkspaceTemplate, error) {
	return t.ksclient.TenantV1alpha2().WorkspaceTemplates().Create(workspace)
}

func (t *tenantOperator) UpdateWorkspace(workspace *tenantv1alpha2.WorkspaceTemplate) (*tenantv1alpha2.WorkspaceTemplate, error) {
	return t.ksclient.TenantV1alpha2().WorkspaceTemplates().Update(workspace)
}

func (t *tenantOperator) DescribeWorkspace(workspace string) (*tenantv1alpha2.WorkspaceTemplate, error) {
	obj, err := t.resourceGetter.Get(tenantv1alpha2.ResourcePluralWorkspaceTemplate, "", workspace)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return obj.(*tenantv1alpha2.WorkspaceTemplate), nil
}
func (t *tenantOperator) ListWorkspaceClusters(workspaceName string) (*api.ListResult, error) {
	workspace, err := t.DescribeWorkspace(workspaceName)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// In this case, spec.placement.clusterSelector will be ignored, since spec.placement.clusters is provided.
	if workspace.Spec.Placement.Clusters != nil {
		clusters := make([]interface{}, 0)
		for _, cluster := range workspace.Spec.Placement.Clusters {
			obj, err := t.resourceGetter.Get(clusterv1alpha1.ResourcesPluralCluster, "", cluster.Name)
			if err != nil {
				klog.Error(err)
				if errors.IsNotFound(err) {
					continue
				}
				return nil, err
			}
			clusters = append(clusters, obj)
		}
		return &api.ListResult{Items: clusters, TotalItems: len(clusters)}, nil
	}

	if workspace.Spec.Placement.ClusterSelector != nil {
		// In this case, the resource will be propagated to all member clusters.
		if workspace.Spec.Placement.ClusterSelector.MatchLabels == nil {
			return t.resourceGetter.List(clusterv1alpha1.ResourcesPluralCluster, "", query.New())
		} else {
			// In this case, the resource will only be propagated to member clusters that are labeled with foo: bar.
			return t.resourceGetter.List(clusterv1alpha1.ResourcesPluralCluster, "", &query.Query{
				Pagination:    query.NoPagination,
				Ascending:     false,
				LabelSelector: labels.SelectorFromSet(workspace.Spec.Placement.ClusterSelector.MatchLabels).String(),
			})
		}
	}

	// In this case, you can either set spec: {} as above or remove spec field from your placement policy. The resource will not be propagated to member clusters.
	return &api.ListResult{Items: []interface{}{}, TotalItems: 0}, nil
}
func (t *tenantOperator) ListClusters(user user.Info) (*api.ListResult, error) {

	listClustersInGlobalScope := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		Resource:        "clusters",
		ResourceScope:   request.GlobalScope,
		ResourceRequest: true,
	}

	allowedListClustersInGlobalScope, _, err := t.authorizer.Authorize(listClustersInGlobalScope)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	listWorkspacesInGlobalScope := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		Resource:        "workspaces",
		ResourceScope:   request.GlobalScope,
		ResourceRequest: true,
	}

	allowedListWorkspacesInGlobalScope, _, err := t.authorizer.Authorize(listWorkspacesInGlobalScope)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if allowedListClustersInGlobalScope == authorizer.DecisionAllow ||
		allowedListWorkspacesInGlobalScope == authorizer.DecisionAllow {
		result, err := t.resourceGetter.List(clusterv1alpha1.ResourcesPluralCluster, "", query.New())
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

	clusters := map[string]*clusterv1alpha1.Cluster{}

	for _, roleBinding := range workspaceRoleBindings {
		workspaceName := roleBinding.Labels[tenantv1alpha1.WorkspaceLabel]
		workspace, err := t.DescribeWorkspace(workspaceName)
		if err != nil {
			klog.Error(err)
			return nil, err
		}

		for _, grantedCluster := range workspace.Spec.Placement.Clusters {
			// skip if cluster exist
			if clusters[grantedCluster.Name] != nil {
				continue
			}
			obj, err := t.resourceGetter.Get(clusterv1alpha1.ResourcesPluralCluster, "", grantedCluster.Name)
			if err != nil {
				klog.Error(err)
				if errors.IsNotFound(err) {
					continue
				}
				return nil, err
			}
			cluster := obj.(*clusterv1alpha1.Cluster)
			clusters[cluster.Name] = cluster
		}
	}

	items := make([]interface{}, 0)
	for _, cluster := range clusters {
		items = append(items, cluster)
	}

	return &api.ListResult{Items: items, TotalItems: len(items)}, nil
}

func (t *tenantOperator) DeleteWorkspace(workspace string) error {
	return t.ksclient.TenantV1alpha2().WorkspaceTemplates().Delete(workspace, metav1.NewDeleteOptions(0))
}

// listIntersectedNamespaces returns a list of namespaces that MUST meet ALL the following filters:
// 1. If `workspaces` is not empty, the namespace SHOULD belong to one of the specified workpsaces.
// 2. If `workspaceSubstrs` is not empty, the namespace SHOULD belong to a workspace whose name contains one of the specified substrings.
// 3. If `namespaces` is not empty, the namespace SHOULD be one of the specified namespacs.
// 4. If `namespaceSubstrs` is not empty, the namespace's name SHOULD contain one of the specified substrings.
// 5. If ALL of the filters above are empty, returns all namespaces.
func (t *tenantOperator) listIntersectedNamespaces(workspaces, workspaceSubstrs,
	namespaces, namespaceSubstrs []string) ([]*corev1.Namespace, error) {
	var (
		namespaceSet = stringSet(namespaces)
		workspaceSet = stringSet(workspaces)

		iNamespaces []*corev1.Namespace
	)
	includeNsWithoutWs := len(workspaceSet) == 0 && len(workspaceSubstrs) == 0

	result, err := t.resourceGetter.List("namespaces", "", query.New())
	if err != nil {
		return nil, err
	}
	for _, obj := range result.Items {
		ns, ok := obj.(*corev1.Namespace)
		if !ok {
			continue
		}

		if len(namespaceSet) > 0 {
			if _, ok := namespaceSet[ns.Name]; !ok {
				continue
			}
		}
		if len(namespaceSubstrs) > 0 && !stringContains(ns.Name, namespaceSubstrs) {
			continue
		}
		if ws := ns.Labels[tenantv1alpha1.WorkspaceLabel]; ws != "" {
			if len(workspaceSet) > 0 {
				if _, ok := workspaceSet[ws]; !ok {
					continue
				}
			}
			if len(workspaceSubstrs) > 0 && !stringContains(ws, workspaceSubstrs) {
				continue
			}
		} else if !includeNsWithoutWs {
			continue
		}
		iNamespaces = append(iNamespaces, ns)
	}
	return iNamespaces, nil
}

// listIntersectedWorkspaces returns a list of workspaces that MUST meet ALL the following filters:
// 1. If `workspaces` is not empty, the workspace SHOULD be one of the specified workpsaces.
// 2. Else if `workspaceSubstrs` is not empty, the workspace SHOULD be contains one of the specified substrings.
// 3. Else, return all workspace in the cluster.
func (t *tenantOperator) listIntersectedWorkspaces(workspaces, workspaceSubstrs []string) ([]*tenantv1alpha1.Workspace, error) {
	var (
		workspaceSet = stringSet(workspaces)
		iWorkspaces  []*tenantv1alpha1.Workspace
	)

	result, err := t.resourceGetter.List("workspaces", "", query.New())
	if err != nil {
		return nil, err
	}
	for _, obj := range result.Items {
		ws, ok := obj.(*tenantv1alpha1.Workspace)
		if !ok {
			continue
		}

		if len(workspaceSet) > 0 {
			if _, ok := workspaceSet[ws.Name]; !ok {
				continue
			}
		}
		if len(workspaceSubstrs) > 0 && !stringContains(ws.Name, workspaceSubstrs) {
			continue
		}

		iWorkspaces = append(iWorkspaces, ws)
	}
	return iWorkspaces, nil
}

func (t *tenantOperator) Events(user user.Info, queryParam *eventsv1alpha1.Query) (*eventsv1alpha1.APIResponse, error) {
	iNamespaces, err := t.listIntersectedNamespaces(
		stringutils.Split(queryParam.WorkspaceFilter, ","),
		stringutils.Split(queryParam.WorkspaceSearch, ","),
		stringutils.Split(queryParam.InvolvedObjectNamespaceFilter, ","),
		stringutils.Split(queryParam.InvolvedObjectNamespaceSearch, ","))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	namespaceCreateTimeMap := make(map[string]time.Time)

	for _, ns := range iNamespaces {
		listEvts := authorizer.AttributesRecord{
			User:            user,
			Verb:            "list",
			APIGroup:        "",
			APIVersion:      "v1",
			Namespace:       ns.Name,
			Resource:        "events",
			ResourceRequest: true,
			ResourceScope:   request.NamespaceScope,
		}
		decision, _, err := t.authorizer.Authorize(listEvts)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if decision == authorizer.DecisionAllow {
			namespaceCreateTimeMap[ns.Name] = ns.CreationTimestamp.Time
		}
	}
	// If there are no ns and ws query conditions,
	// those events with empty `involvedObject.namespace` will also be listed when user can list all events
	if len(queryParam.WorkspaceFilter) == 0 && len(queryParam.InvolvedObjectNamespaceFilter) == 0 &&
		len(queryParam.WorkspaceSearch) == 0 && len(queryParam.InvolvedObjectNamespaceSearch) == 0 {
		listEvts := authorizer.AttributesRecord{
			User:            user,
			Verb:            "list",
			APIGroup:        "",
			APIVersion:      "v1",
			Resource:        "events",
			ResourceRequest: true,
			ResourceScope:   request.ClusterScope,
		}
		decision, _, err := t.authorizer.Authorize(listEvts)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if decision == authorizer.DecisionAllow {
			namespaceCreateTimeMap[""] = time.Time{}
		}
	}

	return t.events.Events(queryParam, func(filter *eventsclient.Filter) {
		filter.InvolvedObjectNamespaceMap = namespaceCreateTimeMap
	})
}

func (t *tenantOperator) QueryLogs(user user.Info, query *loggingv1alpha2.Query) (*loggingv1alpha2.APIResponse, error) {
	iNamespaces, err := t.listIntersectedNamespaces(nil, nil,
		stringutils.Split(query.NamespaceFilter, ","),
		stringutils.Split(query.NamespaceSearch, ","))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	namespaceCreateTimeMap := make(map[string]*time.Time)

	var isGlobalAdmin bool

	// If it is a global admin, the user can view logs from any namespace.
	podLogs := authorizer.AttributesRecord{
		User:            user,
		Verb:            "get",
		APIGroup:        "",
		APIVersion:      "v1",
		Resource:        "pods",
		Subresource:     "log",
		ResourceRequest: true,
		ResourceScope:   request.ClusterScope,
	}
	decision, _, err := t.authorizer.Authorize(podLogs)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	if decision == authorizer.DecisionAllow {
		isGlobalAdmin = true
		if query.NamespaceFilter != "" || query.NamespaceSearch != "" {
			for _, ns := range iNamespaces {
				namespaceCreateTimeMap[ns.Name] = nil
			}
		}
	}

	// If it is a regular user, this user can only view logs of namespaces the user belongs to.
	if !isGlobalAdmin {
		for _, ns := range iNamespaces {
			podLogs := authorizer.AttributesRecord{
				User:            user,
				Verb:            "get",
				APIGroup:        "",
				APIVersion:      "v1",
				Namespace:       ns.Name,
				Resource:        "pods",
				Subresource:     "log",
				ResourceRequest: true,
				ResourceScope:   request.NamespaceScope,
			}
			decision, _, err := t.authorizer.Authorize(podLogs)
			if err != nil {
				klog.Error(err)
				return nil, err
			}
			if decision == authorizer.DecisionAllow {
				namespaceCreateTimeMap[ns.Name] = &ns.CreationTimestamp.Time
			}
		}
	}

	sf := loggingclient.SearchFilter{
		NamespaceFilter: namespaceCreateTimeMap,
		WorkloadSearch:  stringutils.Split(query.WorkloadSearch, ","),
		WorkloadFilter:  stringutils.Split(query.WorkloadFilter, ","),
		PodSearch:       stringutils.Split(query.PodSearch, ","),
		PodFilter:       stringutils.Split(query.PodFilter, ","),
		ContainerSearch: stringutils.Split(query.ContainerSearch, ","),
		ContainerFilter: stringutils.Split(query.ContainerFilter, ","),
		LogSearch:       stringutils.Split(query.LogSearch, ","),
		Starttime:       query.StartTime,
		Endtime:         query.EndTime,
	}

	var ar loggingv1alpha2.APIResponse
	noHit := !isGlobalAdmin && len(namespaceCreateTimeMap) == 0 ||
		isGlobalAdmin && len(namespaceCreateTimeMap) == 0 && (query.NamespaceFilter != "" || query.NamespaceSearch != "")

	switch query.Operation {
	case loggingv1alpha2.OperationStatistics:
		if noHit {
			ar.Statistics = &loggingclient.Statistics{}
		} else {
			ar, err = t.lo.GetCurrentStats(sf)
		}
	case loggingv1alpha2.OperationHistogram:
		if noHit {
			ar.Histogram = &loggingclient.Histogram{}
		} else {
			ar, err = t.lo.CountLogsByInterval(sf, query.Interval)
		}
	default:
		if noHit {
			ar.Logs = &loggingclient.Logs{}
		} else {
			ar, err = t.lo.SearchLogs(sf, query.From, query.Size, query.Sort)
		}
	}
	return &ar, err
}

func (t *tenantOperator) ExportLogs(user user.Info, query *loggingv1alpha2.Query, writer io.Writer) error {
	iNamespaces, err := t.listIntersectedNamespaces(nil, nil,
		stringutils.Split(query.NamespaceFilter, ","),
		stringutils.Split(query.NamespaceSearch, ","))
	if err != nil {
		klog.Error(err)
		return err
	}

	namespaceCreateTimeMap := make(map[string]*time.Time)

	var isGlobalAdmin bool

	// If it is a global admin, the user can view logs from any namespace.
	podLogs := authorizer.AttributesRecord{
		User:            user,
		Verb:            "get",
		APIGroup:        "",
		APIVersion:      "v1",
		Resource:        "pods",
		Subresource:     "log",
		ResourceRequest: true,
		ResourceScope:   request.ClusterScope,
	}
	decision, _, err := t.authorizer.Authorize(podLogs)
	if err != nil {
		klog.Error(err)
		return err
	}
	if decision == authorizer.DecisionAllow {
		isGlobalAdmin = true
		if query.NamespaceFilter != "" || query.NamespaceSearch != "" {
			for _, ns := range iNamespaces {
				namespaceCreateTimeMap[ns.Name] = nil
			}
		}
	}

	// If it is a regular user, this user can only view logs of namespaces the user belongs to.
	if !isGlobalAdmin {
		for _, ns := range iNamespaces {
			podLogs := authorizer.AttributesRecord{
				User:            user,
				Verb:            "get",
				APIGroup:        "",
				APIVersion:      "v1",
				Namespace:       ns.Name,
				Resource:        "pods",
				Subresource:     "log",
				ResourceRequest: true,
				ResourceScope:   request.NamespaceScope,
			}
			decision, _, err := t.authorizer.Authorize(podLogs)
			if err != nil {
				klog.Error(err)
				return err
			}
			if decision == authorizer.DecisionAllow {
				namespaceCreateTimeMap[ns.Name] = &ns.CreationTimestamp.Time
			}
		}
	}

	sf := loggingclient.SearchFilter{
		NamespaceFilter: namespaceCreateTimeMap,
		WorkloadSearch:  stringutils.Split(query.WorkloadSearch, ","),
		WorkloadFilter:  stringutils.Split(query.WorkloadFilter, ","),
		PodSearch:       stringutils.Split(query.PodSearch, ","),
		PodFilter:       stringutils.Split(query.PodFilter, ","),
		ContainerSearch: stringutils.Split(query.ContainerSearch, ","),
		ContainerFilter: stringutils.Split(query.ContainerFilter, ","),
		LogSearch:       stringutils.Split(query.LogSearch, ","),
		Starttime:       query.StartTime,
		Endtime:         query.EndTime,
	}

	noHit := !isGlobalAdmin && len(namespaceCreateTimeMap) == 0 ||
		isGlobalAdmin && len(namespaceCreateTimeMap) == 0 && (query.NamespaceFilter != "" || query.NamespaceSearch != "")

	if noHit {
		return nil
	} else {
		return t.lo.ExportLogs(sf, writer)
	}
}

func (t *tenantOperator) Auditing(user user.Info, queryParam *auditingv1alpha1.Query) (*auditingv1alpha1.APIResponse, error) {
	iNamespaces, err := t.listIntersectedNamespaces(
		stringutils.Split(queryParam.WorkspaceFilter, ","),
		stringutils.Split(queryParam.WorkspaceSearch, ","),
		stringutils.Split(queryParam.ObjectRefNamespaceFilter, ","),
		stringutils.Split(queryParam.ObjectRefNamespaceSearch, ","))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	iWorkspaces, err := t.listIntersectedWorkspaces(
		stringutils.Split(queryParam.WorkspaceFilter, ","),
		stringutils.Split(queryParam.WorkspaceSearch, ","))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	namespaceCreateTimeMap := make(map[string]time.Time)
	workspaceCreateTimeMap := make(map[string]time.Time)

	// Now auditing and event have the same authorization mechanism, so we can determine whether the user
	// has permission to view the auditing log in ns by judging whether the user has the permission to view the event in ns.
	for _, ns := range iNamespaces {
		listEvts := authorizer.AttributesRecord{
			User:            user,
			Verb:            "list",
			APIGroup:        "",
			APIVersion:      "v1",
			Namespace:       ns.Name,
			Resource:        "events",
			ResourceRequest: true,
			ResourceScope:   request.NamespaceScope,
		}
		decision, _, err := t.authorizer.Authorize(listEvts)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if decision == authorizer.DecisionAllow {
			namespaceCreateTimeMap[ns.Name] = ns.CreationTimestamp.Time
		}
	}

	// Now auditing and event have the same authorization mechanism, so we can determine whether the user
	// has permission to view the auditing log in ws by judging whether the user has the permission to view the event in ws.
	for _, ws := range iWorkspaces {
		listEvts := authorizer.AttributesRecord{
			User:            user,
			Verb:            "list",
			APIGroup:        "",
			APIVersion:      "v1",
			Workspace:       ws.Name,
			Resource:        "events",
			ResourceRequest: true,
			ResourceScope:   request.WorkspaceScope,
		}
		decision, _, err := t.authorizer.Authorize(listEvts)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if decision == authorizer.DecisionAllow {
			workspaceCreateTimeMap[ws.Name] = ws.CreationTimestamp.Time
		}
	}

	// If there are no ns and ws query conditions,
	// those events with empty `objectRef.namespace` will also be listed when user can list all events
	if len(queryParam.WorkspaceFilter) == 0 && len(queryParam.ObjectRefNamespaceFilter) == 0 &&
		len(queryParam.WorkspaceSearch) == 0 && len(queryParam.ObjectRefNamespaceSearch) == 0 {
		listEvts := authorizer.AttributesRecord{
			User:            user,
			Verb:            "list",
			APIGroup:        "",
			APIVersion:      "v1",
			Resource:        "events",
			ResourceRequest: true,
			ResourceScope:   request.ClusterScope,
		}
		decision, _, err := t.authorizer.Authorize(listEvts)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if decision == authorizer.DecisionAllow {
			namespaceCreateTimeMap[""] = time.Time{}
			workspaceCreateTimeMap[""] = time.Time{}
		}
	}

	return t.auditing.Events(queryParam, func(filter *auditingclient.Filter) {
		filter.ObjectRefNamespaceMap = namespaceCreateTimeMap
		filter.WorkspaceMap = workspaceCreateTimeMap
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

func stringSet(strs []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, str := range strs {
		m[str] = struct{}{}
	}
	return m
}

func stringContains(str string, subStrs []string) bool {
	for _, sub := range subStrs {
		if strings.Contains(str, sub) {
			return true
		}
	}
	return false
}
