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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	quotav1alpha2 "kubesphere.io/api/quota/v1alpha2"
	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"
	tenantv1alpha2 "kubesphere.io/api/tenant/v1alpha2"
	typesv1beta1 "kubesphere.io/api/types/v1beta1"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/api"
	auditingv1alpha1 "kubesphere.io/kubesphere/pkg/api/auditing/v1alpha1"
	eventsv1alpha1 "kubesphere.io/kubesphere/pkg/api/events/v1alpha1"
	loggingv1alpha2 "kubesphere.io/kubesphere/pkg/api/logging/v1alpha2"
	meteringv1alpha1 "kubesphere.io/kubesphere/pkg/api/metering/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/auditing"
	"kubesphere.io/kubesphere/pkg/models/events"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"kubesphere.io/kubesphere/pkg/models/logging"
	"kubesphere.io/kubesphere/pkg/models/metering"
	"kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/models/openpitrix"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	auditingclient "kubesphere.io/kubesphere/pkg/simple/client/auditing"
	eventsclient "kubesphere.io/kubesphere/pkg/simple/client/events"
	loggingclient "kubesphere.io/kubesphere/pkg/simple/client/logging"
	meteringclient "kubesphere.io/kubesphere/pkg/simple/client/metering"
	monitoringclient "kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
	jsonpatchutil "kubesphere.io/kubesphere/pkg/utils/josnpatchutil"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
)

const orphanFinalizer = "orphan.finalizers.kubesphere.io"

type Interface interface {
	ListWorkspaces(user user.Info, queryParam *query.Query) (*api.ListResult, error)
	GetWorkspace(workspace string) (*tenantv1alpha1.Workspace, error)
	ListWorkspaceTemplates(user user.Info, query *query.Query) (*api.ListResult, error)
	CreateWorkspaceTemplate(workspace *tenantv1alpha2.WorkspaceTemplate) (*tenantv1alpha2.WorkspaceTemplate, error)
	DeleteWorkspaceTemplate(workspace string, opts metav1.DeleteOptions) error
	UpdateWorkspaceTemplate(workspace *tenantv1alpha2.WorkspaceTemplate) (*tenantv1alpha2.WorkspaceTemplate, error)
	PatchWorkspaceTemplate(user user.Info, workspace string, data json.RawMessage) (*tenantv1alpha2.WorkspaceTemplate, error)
	DescribeWorkspaceTemplate(workspace string) (*tenantv1alpha2.WorkspaceTemplate, error)
	ListNamespaces(user user.Info, workspace string, query *query.Query) (*api.ListResult, error)
	ListDevOpsProjects(user user.Info, workspace string, query *query.Query) (*api.ListResult, error)
	ListFederatedNamespaces(info user.Info, workspace string, param *query.Query) (*api.ListResult, error)
	CreateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error)
	ListWorkspaceClusters(workspace string) (*api.ListResult, error)
	Events(user user.Info, queryParam *eventsv1alpha1.Query) (*eventsv1alpha1.APIResponse, error)
	QueryLogs(user user.Info, query *loggingv1alpha2.Query) (*loggingv1alpha2.APIResponse, error)
	ExportLogs(user user.Info, query *loggingv1alpha2.Query, writer io.Writer) error
	Auditing(user user.Info, queryParam *auditingv1alpha1.Query) (*auditingv1alpha1.APIResponse, error)
	DescribeNamespace(workspace, namespace string) (*corev1.Namespace, error)
	DeleteNamespace(workspace, namespace string) error
	UpdateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error)
	PatchNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error)
	ListClusters(info user.Info, queryParam *query.Query) (*api.ListResult, error)
	Metering(user user.Info, queryParam *meteringv1alpha1.Query, priceInfo meteringclient.PriceInfo) (monitoring.Metrics, error)
	MeteringHierarchy(user user.Info, queryParam *meteringv1alpha1.Query, priceInfo meteringclient.PriceInfo) (metering.ResourceStatistic, error)
	CreateWorkspaceResourceQuota(workspace string, resourceQuota *quotav1alpha2.ResourceQuota) (*quotav1alpha2.ResourceQuota, error)
	DeleteWorkspaceResourceQuota(workspace string, resourceQuotaName string) error
	UpdateWorkspaceResourceQuota(workspace string, resourceQuota *quotav1alpha2.ResourceQuota) (*quotav1alpha2.ResourceQuota, error)
	DescribeWorkspaceResourceQuota(workspace string, resourceQuotaName string) (*quotav1alpha2.ResourceQuota, error)
}

type tenantOperator struct {
	am             am.AccessManagementInterface
	im             im.IdentityManagementInterface
	authorizer     authorizer.Authorizer
	k8sclient      kubernetes.Interface
	ksclient       kubesphere.Interface
	resourceGetter *resourcesv1alpha3.ResourceGetter
	events         events.Interface
	lo             logging.LoggingOperator
	auditing       auditing.Interface
	mo             monitoring.MonitoringOperator
	opRelease      openpitrix.ReleaseInterface
	clusterClient  clusterclient.ClusterClients
}

func New(informers informers.InformerFactory, k8sclient kubernetes.Interface, ksclient kubesphere.Interface, evtsClient eventsclient.Client, loggingClient loggingclient.Client, auditingclient auditingclient.Client, am am.AccessManagementInterface, im im.IdentityManagementInterface, authorizer authorizer.Authorizer, monitoringclient monitoringclient.Interface, resourceGetter *resourcev1alpha3.ResourceGetter, opClient openpitrix.Interface) Interface {
	return &tenantOperator{
		am:             am,
		im:             im,
		authorizer:     authorizer,
		resourceGetter: resourcesv1alpha3.NewResourceGetter(informers, nil),
		k8sclient:      k8sclient,
		ksclient:       ksclient,
		events:         events.NewEventsOperator(evtsClient),
		lo:             logging.NewLoggingOperator(loggingClient),
		auditing:       auditing.NewEventsOperator(auditingclient),
		mo:             monitoring.NewMonitoringOperator(monitoringclient, nil, k8sclient, informers, resourceGetter, nil),
		opRelease:      opClient,
		clusterClient:  clusterclient.NewClusterClient(informers.KubeSphereSharedInformerFactory().Cluster().V1alpha1().Clusters()),
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
		result, err := t.resourceGetter.List(tenantv1alpha1.ResourcePluralWorkspace, "", queryParam)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return result, nil
	}

	// retrieving associated resources through role binding
	workspaceRoleBindings, err := t.am.ListWorkspaceRoleBindings(user.GetName(), user.GetGroups(), "")
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	workspaces := make([]runtime.Object, 0)
	for _, roleBinding := range workspaceRoleBindings {
		workspaceName := roleBinding.Labels[tenantv1alpha1.WorkspaceLabel]
		obj, err := t.resourceGetter.Get(tenantv1alpha1.ResourcePluralWorkspace, "", workspaceName)
		if errors.IsNotFound(err) {
			klog.Warningf("workspace role binding: %+v found but workspace not exist", roleBinding.Name)
			continue
		}
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		workspace := obj.(*tenantv1alpha1.Workspace)
		// label matching selector, remove duplicate entity
		if queryParam.Selector().Matches(labels.Set(workspace.Labels)) &&
			!contains(workspaces, workspace) {
			workspaces = append(workspaces, workspace)
		}
	}

	// use default pagination search logic
	result := resources.DefaultList(workspaces, queryParam, func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*tenantv1alpha1.Workspace).ObjectMeta, right.(*tenantv1alpha1.Workspace).ObjectMeta, field)
	}, func(workspace runtime.Object, filter query.Filter) bool {
		return resources.DefaultObjectMetaFilter(workspace.(*tenantv1alpha1.Workspace).ObjectMeta, filter)
	})

	return result, nil
}

func (t *tenantOperator) GetWorkspace(workspace string) (*tenantv1alpha1.Workspace, error) {
	obj, err := t.resourceGetter.Get(tenantv1alpha1.ResourcePluralWorkspace, "", workspace)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return obj.(*tenantv1alpha1.Workspace), nil
}

func (t *tenantOperator) ListWorkspaceTemplates(user user.Info, queryParam *query.Query) (*api.ListResult, error) {

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
	workspaceRoleBindings, err := t.am.ListWorkspaceRoleBindings(user.GetName(), user.GetGroups(), "")
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
	roleBindings, err := t.am.ListRoleBindings(user.GetName(), user.GetGroups(), "")
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
	roleBindings, err := t.am.ListRoleBindings(user.GetName(), user.GetGroups(), "")
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
	return t.k8sclient.CoreV1().Namespaces().Create(context.Background(), labelNamespaceWithWorkspaceName(namespace, workspace), metav1.CreateOptions{})
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
	return t.k8sclient.CoreV1().Namespaces().Delete(context.Background(), namespace, *metav1.NewDeleteOptions(0))
}

func (t *tenantOperator) UpdateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	_, err := t.DescribeNamespace(workspace, namespace.Name)
	if err != nil {
		return nil, err
	}
	namespace = labelNamespaceWithWorkspaceName(namespace, workspace)
	return t.k8sclient.CoreV1().Namespaces().Update(context.Background(), namespace, metav1.UpdateOptions{})
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
	return t.k8sclient.CoreV1().Namespaces().Patch(context.Background(), namespace.Name, types.MergePatchType, data, metav1.PatchOptions{})
}

func (t *tenantOperator) PatchWorkspaceTemplate(user user.Info, workspace string, data json.RawMessage) (*tenantv1alpha2.WorkspaceTemplate, error) {
	var manageWorkspaceTemplateRequest bool
	clusterNames := sets.NewString()

	patchs, err := jsonpatchutil.Parse(data)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(patchs) > 0 {
		for _, patch := range patchs {
			path, err := patch.Path()
			if err != nil {
				klog.Error(err)
				return nil, err
			}

			// If the request path is cluster, just collecting cluster name to set and continue to check cluster permission later.
			// Or indicate that want to manage the workspace templates, so check if user has the permission to manage workspace templates.
			if strings.HasPrefix(path, "/spec/placement") {
				if patch.Kind() != "add" && patch.Kind() != "remove" {
					err := errors.NewBadRequest("not support operation type")
					klog.Error(err)
					return nil, err
				}
				clusterValue := make(map[string]interface{})
				err := jsonpatchutil.GetValue(patch, &clusterValue)
				if err != nil {
					klog.Error(err)
					return nil, err
				}

				// if the placement is empty, the first patch need fill with "clusters" field.
				if cName := clusterValue["name"]; cName != nil {
					cn, ok := cName.(string)
					if ok {
						clusterNames.Insert(cn)
					}
				} else if cluster := clusterValue["clusters"]; cluster != nil {
					clusterRefrences := []typesv1beta1.GenericClusterReference{}
					err := mapstructure.Decode(cluster, &clusterRefrences)
					if err != nil {
						klog.Error(err)
						return nil, err
					}
					for _, v := range clusterRefrences {
						clusterNames.Insert(v.Name)
					}
				}

			} else {
				manageWorkspaceTemplateRequest = true
			}
		}
	}

	if manageWorkspaceTemplateRequest {
		deleteWST := authorizer.AttributesRecord{
			User:            user,
			Verb:            authorizer.VerbDelete,
			APIGroup:        tenantv1alpha2.SchemeGroupVersion.Group,
			APIVersion:      tenantv1alpha2.SchemeGroupVersion.Version,
			Resource:        tenantv1alpha2.ResourcePluralWorkspaceTemplate,
			ResourceRequest: true,
			ResourceScope:   request.GlobalScope,
		}
		authorize, reason, err := t.authorizer.Authorize(deleteWST)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if authorize != authorizer.DecisionAllow {
			err := errors.NewForbidden(tenantv1alpha2.Resource(tenantv1alpha2.ResourcePluralWorkspaceTemplate), workspace, fmt.Errorf(reason))
			klog.Error(err)
			return nil, err
		}
	}
	// Checking whether the user can manage the cluster requires authentication from two aspects.
	// First check whether the user has relevant global permissions,
	// and then check whether the user has relevant cluster permissions in the target cluster
	if clusterNames.Len() > 0 {
		for _, clusterName := range clusterNames.List() {
			deleteCluster := authorizer.AttributesRecord{
				User:            user,
				Verb:            authorizer.VerbDelete,
				APIGroup:        clusterv1alpha1.SchemeGroupVersion.Version,
				APIVersion:      clusterv1alpha1.SchemeGroupVersion.Version,
				Resource:        clusterv1alpha1.ResourcesPluralCluster,
				Cluster:         clusterName,
				ResourceRequest: true,
				ResourceScope:   request.GlobalScope,
			}
			authorize, reason, err := t.authorizer.Authorize(deleteCluster)
			if err != nil {
				klog.Error(err)
				return nil, err
			}

			if authorize == authorizer.DecisionAllow {
				continue
			}

			list, err := t.getClusterRoleBindingsByUser(clusterName, user.GetName())
			if err != nil {
				klog.Error(err)
				return nil, err
			}

			allowed := false
			for _, clusterRolebinding := range list.Items {
				if clusterRolebinding.RoleRef.Name == iamv1alpha2.ClusterAdmin {
					allowed = true
					break
				}
			}

			if !allowed {
				err = errors.NewForbidden(clusterv1alpha1.Resource(clusterv1alpha1.ResourcesPluralCluster), clusterName, fmt.Errorf(reason))
				klog.Error(err)
				return nil, err
			}
		}
	}

	return t.ksclient.TenantV1alpha2().WorkspaceTemplates().Patch(context.Background(), workspace, types.JSONPatchType, data, metav1.PatchOptions{})
}

func (t *tenantOperator) CreateWorkspaceTemplate(workspace *tenantv1alpha2.WorkspaceTemplate) (*tenantv1alpha2.WorkspaceTemplate, error) {
	return t.ksclient.TenantV1alpha2().WorkspaceTemplates().Create(context.Background(), workspace, metav1.CreateOptions{})
}

func (t *tenantOperator) UpdateWorkspaceTemplate(workspace *tenantv1alpha2.WorkspaceTemplate) (*tenantv1alpha2.WorkspaceTemplate, error) {
	return t.ksclient.TenantV1alpha2().WorkspaceTemplates().Update(context.Background(), workspace, metav1.UpdateOptions{})
}

func (t *tenantOperator) DescribeWorkspaceTemplate(workspace string) (*tenantv1alpha2.WorkspaceTemplate, error) {
	obj, err := t.resourceGetter.Get(tenantv1alpha2.ResourcePluralWorkspaceTemplate, "", workspace)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return obj.(*tenantv1alpha2.WorkspaceTemplate), nil
}

func (t *tenantOperator) ListWorkspaceClusters(workspaceName string) (*api.ListResult, error) {
	workspace, err := t.DescribeWorkspaceTemplate(workspaceName)
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
				klog.Warning(err)
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

func (t *tenantOperator) ListClusters(user user.Info, queryParam *query.Query) (*api.ListResult, error) {

	listClustersInGlobalScope := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		APIGroup:        "cluster.kubesphere.io",
		Resource:        "clusters",
		ResourceScope:   request.GlobalScope,
		ResourceRequest: true,
	}

	allowedListClustersInGlobalScope, _, err := t.authorizer.Authorize(listClustersInGlobalScope)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize: %s", err)
	}

	if allowedListClustersInGlobalScope == authorizer.DecisionAllow {
		return t.resourceGetter.List(clusterv1alpha1.ResourcesPluralCluster, "", queryParam)
	}

	userDetail, err := t.im.DescribeUser(user.GetName())
	if err != nil {
		return nil, fmt.Errorf("failed to describe user: %s", err)
	}

	grantedClustersAnnotation := userDetail.Annotations[iamv1alpha2.GrantedClustersAnnotation]
	var grantedClusters sets.String
	if len(grantedClustersAnnotation) > 0 {
		grantedClusters = sets.NewString(strings.Split(grantedClustersAnnotation, ",")...)
	} else {
		grantedClusters = sets.NewString()
	}
	var clusters []*clusterv1alpha1.Cluster
	for _, grantedCluster := range grantedClusters.List() {
		obj, err := t.resourceGetter.Get(clusterv1alpha1.ResourcesPluralCluster, "", grantedCluster)
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return nil, fmt.Errorf("failed to fetch cluster: %s", err)
		}
		cluster := obj.(*clusterv1alpha1.Cluster)
		clusters = append(clusters, cluster)
	}

	items := make([]runtime.Object, 0)
	for _, cluster := range clusters {
		items = append(items, cluster)
	}

	// apply additional labelSelector
	if queryParam.LabelSelector != "" {
		queryParam.Filters[query.FieldLabel] = query.Value(queryParam.LabelSelector)
	}

	// use default pagination search logic
	result := resources.DefaultList(items, queryParam, func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*clusterv1alpha1.Cluster).ObjectMeta, right.(*clusterv1alpha1.Cluster).ObjectMeta, field)
	}, func(workspace runtime.Object, filter query.Filter) bool {
		return resources.DefaultObjectMetaFilter(workspace.(*clusterv1alpha1.Cluster).ObjectMeta, filter)
	})

	return result, nil
}

func (t *tenantOperator) DeleteWorkspaceTemplate(workspace string, opts metav1.DeleteOptions) error {

	if opts.PropagationPolicy != nil && *opts.PropagationPolicy == metav1.DeletePropagationOrphan {
		wsp, err := t.DescribeWorkspaceTemplate(workspace)
		if err != nil {
			klog.Error(err)
			return err
		}
		wsp.Finalizers = append(wsp.Finalizers, orphanFinalizer)
		_, err = t.ksclient.TenantV1alpha2().WorkspaceTemplates().Update(context.Background(), wsp, metav1.UpdateOptions{})
		if err != nil {
			klog.Error(err)
			return err
		}
	}
	return t.ksclient.TenantV1alpha2().WorkspaceTemplates().Delete(context.Background(), workspace, opts)
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

func (t *tenantOperator) Metering(user user.Info, query *meteringv1alpha1.Query, priceInfo meteringclient.PriceInfo) (metrics monitoring.Metrics, err error) {

	var opt QueryOptions

	opt, err = t.makeQueryOptions(user, *query, query.Level)
	if err != nil {
		return
	}
	metrics, err = t.ProcessNamedMetersQuery(opt, priceInfo)

	return
}

func (t *tenantOperator) MeteringHierarchy(user user.Info, queryParam *meteringv1alpha1.Query, priceInfo meteringclient.PriceInfo) (metering.ResourceStatistic, error) {
	res, err := t.Metering(user, queryParam, priceInfo)
	if err != nil {
		return metering.ResourceStatistic{}, err
	}

	// get pods stat info under ns
	podsStats := t.transformMetricData(res)

	// classify pods stats
	resourceStats, err := t.classifyPodStats(user, queryParam.Cluster, queryParam.NamespaceName, podsStats)
	if err != nil {
		klog.Error(err)
		return metering.ResourceStatistic{}, err
	}

	return resourceStats, nil
}

func (t *tenantOperator) getClusterRoleBindingsByUser(clusterName, user string) (*rbacv1.ClusterRoleBindingList, error) {
	kubernetesClientSet, err := t.clusterClient.GetKubernetesClientSet(clusterName)
	if err != nil {
		return nil, err
	}
	return kubernetesClientSet.RbacV1().ClusterRoleBindings().
		List(context.Background(),
			metav1.ListOptions{LabelSelector: labels.FormatLabels(map[string]string{"iam.kubesphere.io/user-ref": user})})
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
