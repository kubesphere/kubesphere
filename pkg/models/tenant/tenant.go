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
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	eventsv1alpha1 "kubesphere.io/kubesphere/pkg/api/events/v1alpha1"
	loggingv1alpha2 "kubesphere.io/kubesphere/pkg/api/logging/v1alpha2"
	clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	tenantv1alpha2 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizerfactory"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/events"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/logging"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	eventsclient "kubesphere.io/kubesphere/pkg/simple/client/events"
	loggingclient "kubesphere.io/kubesphere/pkg/simple/client/logging"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"strings"
	"time"
)

type Interface interface {
	ListWorkspaces(user user.Info, query *query.Query) (*api.ListResult, error)
	ListNamespaces(user user.Info, workspace string, query *query.Query) (*api.ListResult, error)
	CreateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error)
	CreateWorkspace(workspace *tenantv1alpha2.WorkspaceTemplate) (*tenantv1alpha2.WorkspaceTemplate, error)
	DeleteWorkspace(workspace string) error
	UpdateWorkspace(workspace *tenantv1alpha2.WorkspaceTemplate) (*tenantv1alpha2.WorkspaceTemplate, error)
	DescribeWorkspace(workspace string) (*tenantv1alpha2.WorkspaceTemplate, error)
	ListWorkspaceClusters(workspace string) (*api.ListResult, error)

	Events(user user.Info, queryParam *eventsv1alpha1.Query) (*eventsv1alpha1.APIResponse, error)
	QueryLogs(user user.Info, query *loggingv1alpha2.Query) (*loggingv1alpha2.APIResponse, error)
	ExportLogs(user user.Info, query *loggingv1alpha2.Query, writer io.Writer) error
}

type tenantOperator struct {
	am             am.AccessManagementInterface
	authorizer     authorizer.Authorizer
	k8sclient      kubernetes.Interface
	ksclient       kubesphere.Interface
	resourceGetter *resourcesv1alpha3.ResourceGetter
	events         events.Interface
	lo             logging.LoggingOperator
}

func New(informers informers.InformerFactory, k8sclient kubernetes.Interface, ksclient kubesphere.Interface, evtsClient eventsclient.Client, loggingClient loggingclient.Interface) Interface {
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

		result, err := t.resourceGetter.List(tenantv1alpha2.ResourcePluralWorkspaceTemplate, "", queryParam)

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
		workspace, err := t.resourceGetter.Get(tenantv1alpha2.ResourcePluralWorkspaceTemplate, "", workspaceName)

		if errors.IsNotFound(err) {
			klog.Warningf("workspace role binding: %+v found but workspace not exist", roleBinding.ObjectMeta.String())
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

func (t *tenantOperator) CreateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error) {

	_, err := t.resourceGetter.Get(tenantv1alpha1.ResourcePluralWorkspace, "", workspace)

	if err != nil {
		return nil, err
	}

	if namespace.Annotations == nil {
		namespace.Annotations = make(map[string]string, 0)
	}

	namespace.Annotations[tenantv1alpha1.WorkspaceLabel] = workspace

	return t.k8sclient.CoreV1().Namespaces().Create(namespace)
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
	clusters := make([]interface{}, 0)
	for _, cluster := range workspace.Spec.Clusters {
		obj, err := t.resourceGetter.Get(clusterv1alpha1.ResourcesPluralCluster, "", cluster)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		cluster := obj.(*clusterv1alpha1.Cluster)
		clusters = append(clusters, cluster)
	}
	return &api.ListResult{Items: clusters, TotalItems: len(clusters)}, nil
}

func (t *tenantOperator) DeleteWorkspace(workspace string) error {
	return t.ksclient.TenantV1alpha2().WorkspaceTemplates().Delete(workspace, metav1.NewDeleteOptions(0))
}

// listIntersectedNamespaces lists the namespaces which meet all the following conditions at the same time
// 1. the namespace which belongs to user.
// 2. the namespace in workspace which is in workspaces when workspaces is not empty.
// 3. the namespace in workspace which contains one of workspaceSubstrs when workspaceSubstrs is not empty.
// 4. the namespace which is in namespaces when namespaces is not empty.
// 5. the namespace which contains one of namespaceSubstrs when namespaceSubstrs is not empty.
func (t *tenantOperator) listIntersectedNamespaces(user user.Info,
	workspaces, workspaceSubstrs, namespaces, namespaceSubstrs []string) ([]*corev1.Namespace, error) {
	var (
		namespaceSet = stringSet(namespaces)
		workspaceSet = stringSet(workspaces)

		iNamespaces []*corev1.Namespace
	)

	includeNsWithoutWs := len(workspaceSet) == 0 && len(workspaceSubstrs) == 0

	roleBindings, err := t.am.ListRoleBindings(user.GetName(), "")
	if err != nil {
		return nil, err
	}
	for _, rb := range roleBindings {
		if len(namespaceSet) > 0 {
			if _, ok := namespaceSet[rb.Namespace]; !ok {
				continue
			}
		}
		if len(namespaceSubstrs) > 0 && !stringContains(rb.Namespace, namespaceSubstrs) {
			continue
		}
		ns, err := t.resourceGetter.Get("namespaces", "", rb.Namespace)
		if err != nil {
			return nil, err
		}
		if ns, ok := ns.(*corev1.Namespace); ok {
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
	}
	return iNamespaces, nil
}

func (t *tenantOperator) Events(user user.Info, queryParam *eventsv1alpha1.Query) (*eventsv1alpha1.APIResponse, error) {
	iNamespaces, err := t.listIntersectedNamespaces(user,
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
	iNamespaces, err := t.listIntersectedNamespaces(user,
		stringutils.Split(query.WorkspaceFilter, ","),
		stringutils.Split(query.WorkspaceSearch, ","),
		stringutils.Split(query.NamespaceFilter, ","),
		stringutils.Split(query.NamespaceSearch, ","))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	namespaceCreateTimeMap := make(map[string]time.Time)
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
		}
		decision, _, err := t.authorizer.Authorize(podLogs)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if decision == authorizer.DecisionAllow {
			namespaceCreateTimeMap[ns.Name] = ns.CreationTimestamp.Time
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
	switch query.Operation {
	case loggingv1alpha2.OperationStatistics:
		if len(namespaceCreateTimeMap) == 0 {
			ar.Statistics = &loggingclient.Statistics{}
		} else {
			ar, err = t.lo.GetCurrentStats(sf)
		}
	case loggingv1alpha2.OperationHistogram:
		if len(namespaceCreateTimeMap) == 0 {
			ar.Histogram = &loggingclient.Histogram{}
		} else {
			ar, err = t.lo.CountLogsByInterval(sf, query.Interval)
		}
	default:
		if len(namespaceCreateTimeMap) == 0 {
			ar.Logs = &loggingclient.Logs{}
		} else {
			ar, err = t.lo.SearchLogs(sf, query.From, query.Size, query.Sort)
		}
	}
	return &ar, err
}

func (t *tenantOperator) ExportLogs(user user.Info, query *loggingv1alpha2.Query, writer io.Writer) error {
	iNamespaces, err := t.listIntersectedNamespaces(user,
		stringutils.Split(query.WorkspaceFilter, ","),
		stringutils.Split(query.WorkspaceSearch, ","),
		stringutils.Split(query.NamespaceFilter, ","),
		stringutils.Split(query.NamespaceSearch, ","))
	if err != nil {
		klog.Error(err)
		return err
	}

	namespaceCreateTimeMap := make(map[string]time.Time)
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
		}
		decision, _, err := t.authorizer.Authorize(podLogs)
		if err != nil {
			klog.Error(err)
			return err
		}
		if decision == authorizer.DecisionAllow {
			namespaceCreateTimeMap[ns.Name] = ns.CreationTimestamp.Time
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

	if len(namespaceCreateTimeMap) == 0 {
		return nil
	} else {
		return t.lo.ExportLogs(sf, writer)
	}
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
