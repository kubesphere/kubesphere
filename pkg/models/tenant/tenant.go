/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/mitchellh/mapstructure"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	quotav1alpha2 "kubesphere.io/api/quota/v1alpha2"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
	jsonpatchutil "kubesphere.io/kubesphere/pkg/utils/josnpatchutil"
)

const orphanFinalizer = "orphan.finalizers.kubesphere.io"

type Interface interface {
	ListWorkspaces(user user.Info, queryParam *query.Query) (*api.ListResult, error)
	GetWorkspace(workspace string) (*tenantv1beta1.Workspace, error)
	ListWorkspaceTemplates(user user.Info, query *query.Query) (*api.ListResult, error)
	CreateWorkspaceTemplate(user user.Info, workspace *tenantv1beta1.WorkspaceTemplate) (*tenantv1beta1.WorkspaceTemplate, error)
	DeleteWorkspaceTemplate(workspace string, opts metav1.DeleteOptions) error
	UpdateWorkspaceTemplate(user user.Info, workspace *tenantv1beta1.WorkspaceTemplate) (*tenantv1beta1.WorkspaceTemplate, error)
	PatchWorkspaceTemplate(user user.Info, workspace string, data json.RawMessage) (*tenantv1beta1.WorkspaceTemplate, error)
	DescribeWorkspaceTemplate(workspace string) (*tenantv1beta1.WorkspaceTemplate, error)
	ListNamespaces(user user.Info, workspace string, query *query.Query) (*api.ListResult, error)
	CreateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error)
	ListWorkspaceClusters(workspace string) (*api.ListResult, error)
	DescribeNamespace(workspace, namespace string) (*corev1.Namespace, error)
	DeleteNamespace(workspace, namespace string) error
	UpdateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error)
	PatchNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error)
	ListClusters(info user.Info, queryParam *query.Query) (*api.ListResult, error)
	CreateWorkspaceResourceQuota(workspace string, resourceQuota *quotav1alpha2.ResourceQuota) (*quotav1alpha2.ResourceQuota, error)
	DeleteWorkspaceResourceQuota(workspace string, resourceQuotaName string) error
	UpdateWorkspaceResourceQuota(workspace string, resourceQuota *quotav1alpha2.ResourceQuota) (*quotav1alpha2.ResourceQuota, error)
	DescribeWorkspaceResourceQuota(workspace string, resourceQuotaName string) (*quotav1alpha2.ResourceQuota, error)
}

type tenantOperator struct {
	am             am.AccessManagementInterface
	im             im.IdentityManagementInterface
	authorizer     authorizer.Authorizer
	resourceGetter *resourcesv1alpha3.Getter
	clusterClient  clusterclient.Interface
	client         runtimeclient.Client
}

func New(cacheClient runtimeclient.Client, k8sVersion *semver.Version, clusterClient clusterclient.Interface,
	am am.AccessManagementInterface, im im.IdentityManagementInterface, authorizer authorizer.Authorizer) Interface {
	return &tenantOperator{
		am:             am,
		im:             im,
		authorizer:     authorizer,
		resourceGetter: resourcesv1alpha3.NewResourceGetter(cacheClient, k8sVersion),
		client:         cacheClient,
		clusterClient:  clusterClient,
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
		result, err := t.resourceGetter.List(tenantv1beta1.ResourcePluralWorkspace, "", queryParam)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return result, nil
	}

	// retrieving associated resources through role binding
	workspaceRoleBindings, err := t.am.ListWorkspaceRoleBindings(user.GetName(), "", user.GetGroups(), "")
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	workspaces := make([]runtime.Object, 0)
	for _, roleBinding := range workspaceRoleBindings {
		workspaceName := roleBinding.Labels[tenantv1beta1.WorkspaceLabel]
		obj, err := t.resourceGetter.Get(tenantv1beta1.ResourcePluralWorkspace, "", workspaceName)
		if errors.IsNotFound(err) {
			klog.Warningf("workspace role binding: %+v found but workspace not exist", roleBinding.Name)
			continue
		}
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		workspace := obj.(*tenantv1beta1.Workspace)
		// label matching selector, remove duplicate entity
		if queryParam.Selector().Matches(labels.Set(workspace.Labels)) &&
			!contains(workspaces, workspace) {
			workspaces = append(workspaces, workspace)
		}
	}

	// use default pagination search logic
	result := resources.DefaultList(workspaces, queryParam, func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*tenantv1beta1.Workspace).ObjectMeta, right.(*tenantv1beta1.Workspace).ObjectMeta, field)
	}, func(workspace runtime.Object, filter query.Filter) bool {
		return resources.DefaultObjectMetaFilter(workspace.(*tenantv1beta1.Workspace).ObjectMeta, filter)
	})

	return result, nil
}

func (t *tenantOperator) GetWorkspace(workspace string) (*tenantv1beta1.Workspace, error) {
	obj, err := t.resourceGetter.Get(tenantv1beta1.ResourcePluralWorkspace, "", workspace)
	if err != nil {
		return nil, err
	}
	return obj.(*tenantv1beta1.Workspace), nil
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
		result, err := t.resourceGetter.List(tenantv1beta1.ResourcePluralWorkspaceTemplate, "", queryParam)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return result, nil
	}

	// retrieving associated resources through role binding
	workspaceRoleBindings, err := t.am.ListWorkspaceRoleBindings(user.GetName(), "", user.GetGroups(), "")
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	workspaces := make([]runtime.Object, 0)
	for _, roleBinding := range workspaceRoleBindings {
		workspaceName := roleBinding.Labels[tenantv1beta1.WorkspaceLabel]
		obj, err := t.resourceGetter.Get(tenantv1beta1.ResourcePluralWorkspaceTemplate, "", workspaceName)
		if errors.IsNotFound(err) {
			klog.Warningf("workspace role binding: %+v found but workspace not exist", roleBinding.Name)
			continue
		}
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		workspace := obj.(*tenantv1beta1.WorkspaceTemplate)
		// label matching selector, remove duplicate entity
		if queryParam.Selector().Matches(labels.Set(workspace.Labels)) &&
			!contains(workspaces, workspace) {
			workspaces = append(workspaces, workspace)
		}
	}

	// use default pagination search logic
	result := resources.DefaultList(workspaces, queryParam, func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*tenantv1beta1.WorkspaceTemplate).ObjectMeta, right.(*tenantv1beta1.WorkspaceTemplate).ObjectMeta, field)
	}, func(workspace runtime.Object, filter query.Filter) bool {
		return resources.DefaultObjectMetaFilter(workspace.(*tenantv1beta1.WorkspaceTemplate).ObjectMeta, filter)
	})

	return result, nil
}

func (t *tenantOperator) ListNamespaces(user user.Info, workspace string, queryParam *query.Query) (*api.ListResult, error) {
	nsScope := request.ClusterScope
	if workspace != "" {
		nsScope = request.WorkspaceScope
		// filter by workspace
		queryParam.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s=%s", tenantv1beta1.WorkspaceLabel, workspace))
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
	roleBindings, err := t.am.ListRoleBindings(user.GetName(), "", user.GetGroups(), "")
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
	namespace = labelNamespaceWithWorkspaceName(namespace, workspace)
	return namespace, t.client.Create(context.Background(), namespace)
}

// labelNamespaceWithWorkspaceName adds a kubesphere.io/workspace=[workspaceName] label to namespace which
// indicates namespace is under the workspace
func labelNamespaceWithWorkspaceName(namespace *corev1.Namespace, workspaceName string) *corev1.Namespace {
	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string, 0)
	}

	namespace.Labels[tenantv1beta1.WorkspaceLabel] = workspaceName // label namespace with workspace name

	return namespace
}

func (t *tenantOperator) DescribeNamespace(workspace, namespace string) (*corev1.Namespace, error) {
	obj, err := t.resourceGetter.Get("namespaces", "", namespace)
	if err != nil {
		return nil, err
	}
	ns := obj.(*corev1.Namespace)
	if ns.Labels[tenantv1beta1.WorkspaceLabel] != workspace {
		return nil, errors.NewNotFound(corev1.Resource("namespace"), namespace)
	}
	return ns, nil
}

func (t *tenantOperator) DeleteNamespace(workspace, namespaceName string) error {
	namespace, err := t.DescribeNamespace(workspace, namespaceName)
	if err != nil {
		return err
	}
	return t.client.Delete(context.Background(), namespace, &runtimeclient.DeleteOptions{GracePeriodSeconds: ptr.To[int64](0)})
}

func (t *tenantOperator) UpdateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	_, err := t.DescribeNamespace(workspace, namespace.Name)
	if err != nil {
		return nil, err
	}
	namespace = labelNamespaceWithWorkspaceName(namespace, workspace)
	return namespace, t.client.Update(context.Background(), namespace)
}

func (t *tenantOperator) PatchNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	if _, err := t.DescribeNamespace(workspace, namespace.Name); err != nil {
		return nil, err
	}
	if namespace.Labels != nil {
		namespace.Labels[tenantv1beta1.WorkspaceLabel] = workspace
	}
	data, err := json.Marshal(namespace)
	if err != nil {
		return nil, err
	}
	return namespace, t.client.Patch(context.Background(), namespace, runtimeclient.RawPatch(types.MergePatchType, data))
}

func (t *tenantOperator) PatchWorkspaceTemplate(user user.Info, workspace string, data json.RawMessage) (*tenantv1beta1.WorkspaceTemplate, error) {
	var manageWorkspaceTemplateRequest bool
	clusterNames := sets.New[string]()

	patchData, err := jsonpatchutil.Parse(data)
	if err != nil {
		return nil, err
	}

	if len(patchData) > 0 {
		for _, patch := range patchData {
			path, err := patch.Path()
			if err != nil {
				return nil, err
			}

			// If the request path is cluster, just collecting cluster name to set and continue to check cluster permission later.
			// Or indicate that want to manage the workspace templates, so check if user has the permission to manage workspace templates.
			if strings.HasPrefix(path, "/spec/placement") {
				if patch.Kind() != "add" && patch.Kind() != "remove" {
					return nil, errors.NewBadRequest("not support operation type")
				}
				clusterValue := make(map[string]interface{})
				if err := jsonpatchutil.GetValue(patch, &clusterValue); err != nil {
					return nil, err
				}

				// if the placement is empty, the first patch need fill with "clusters" field.
				if cName := clusterValue["name"]; cName != nil {
					cn, ok := cName.(string)
					if ok {
						clusterNames.Insert(cn)
					}
				} else if cluster := clusterValue["clusters"]; cluster != nil {
					var clusterReferences []tenantv1beta1.GenericClusterReference
					if err := mapstructure.Decode(cluster, &clusterReferences); err != nil {
						return nil, err
					}
					for _, v := range clusterReferences {
						clusterNames.Insert(v.Name)
					}
				}

			} else {
				manageWorkspaceTemplateRequest = true
			}
		}
	}

	if manageWorkspaceTemplateRequest {
		if err := t.checkWorkspaceTemplatePermission(user, workspace); err != nil {
			return nil, err
		}
	}

	if clusterNames.Len() > 0 {
		if err := t.checkClusterPermission(user, clusterNames.UnsortedList()); err != nil {
			return nil, err
		}
	}

	workspaceTemplate := &tenantv1beta1.WorkspaceTemplate{}
	if err := t.client.Get(context.Background(), types.NamespacedName{Name: workspace}, workspaceTemplate); err != nil {
		return nil, err
	}

	return workspaceTemplate, t.client.Patch(context.Background(), workspaceTemplate, runtimeclient.RawPatch(types.JSONPatchType, data))
}

func (t *tenantOperator) CreateWorkspaceTemplate(user user.Info, workspace *tenantv1beta1.WorkspaceTemplate) (*tenantv1beta1.WorkspaceTemplate, error) {
	workspace = workspace.DeepCopy()
	if len(workspace.Spec.Placement.Clusters) != 0 {
		clusters := make([]string, 0)
		for _, v := range workspace.Spec.Placement.Clusters {
			clusters = append(clusters, v.Name)
		}
		if err := t.checkClusterPermission(user, clusters); err != nil {
			return nil, err
		}

	}
	return workspace, t.client.Create(context.Background(), workspace)
}

func (t *tenantOperator) UpdateWorkspaceTemplate(user user.Info, workspace *tenantv1beta1.WorkspaceTemplate) (*tenantv1beta1.WorkspaceTemplate, error) {
	workspace = workspace.DeepCopy()
	if len(workspace.Spec.Placement.Clusters) != 0 {
		clusters := make([]string, 0)
		for _, v := range workspace.Spec.Placement.Clusters {
			clusters = append(clusters, v.Name)
		}
		if err := t.checkClusterPermission(user, clusters); err != nil {
			return nil, err
		}
	}
	return workspace, t.client.Update(context.Background(), workspace)
}

func (t *tenantOperator) DescribeWorkspaceTemplate(workspaceName string) (*tenantv1beta1.WorkspaceTemplate, error) {
	workspace := &tenantv1beta1.WorkspaceTemplate{}
	return workspace, t.client.Get(context.Background(), types.NamespacedName{Name: workspaceName}, workspace)
}

func (t *tenantOperator) ListWorkspaceClusters(workspaceName string) (*api.ListResult, error) {
	workspace, err := t.DescribeWorkspaceTemplate(workspaceName)
	if err != nil {
		return nil, err
	}

	// In this case, spec.placement.clusterSelector will be ignored, since spec.placement.clusters is provided.
	if workspace.Spec.Placement.Clusters != nil {
		clusters := make([]runtime.Object, 0)
		for _, cluster := range workspace.Spec.Placement.Clusters {
			obj, err := t.resourceGetter.Get(clusterv1alpha1.ResourcesPluralCluster, "", cluster.Name)
			if err != nil {
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
	return &api.ListResult{Items: []runtime.Object{}, TotalItems: 0}, nil
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

	grantedClustersAnnotation := userDetail.Annotations[iamv1beta1.GrantedClustersAnnotation]
	var grantedClusters sets.Set[string]
	if len(grantedClustersAnnotation) > 0 {
		grantedClusters = sets.New(strings.Split(grantedClustersAnnotation, ",")...)
	} else {
		grantedClusters = sets.New[string]()
	}
	var clusters []*clusterv1alpha1.Cluster
	for _, grantedCluster := range grantedClusters.UnsortedList() {
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

func (t *tenantOperator) DeleteWorkspaceTemplate(workspaceName string, opts metav1.DeleteOptions) error {
	workspace := &tenantv1beta1.WorkspaceTemplate{}
	if err := t.client.Get(context.Background(), types.NamespacedName{Name: workspaceName}, workspace); err != nil {
		return err
	}
	if opts.PropagationPolicy != nil && *opts.PropagationPolicy == metav1.DeletePropagationOrphan {

		workspace.Finalizers = append(workspace.Finalizers, orphanFinalizer)
		if err := t.client.Update(context.Background(), workspace); err != nil {
			return err
		}
	}
	return t.client.Delete(context.Background(), workspace, &runtimeclient.DeleteOptions{Raw: &opts})
}

func (t *tenantOperator) getClusterRoleBindingsByUser(clusterName, username string) (*iamv1beta1.ClusterRoleBindingList, error) {
	clusterClient, err := t.clusterClient.GetRuntimeClient(clusterName)
	if err != nil {
		return nil, err
	}

	clusterRoleBindings := &iamv1beta1.ClusterRoleBindingList{}
	if err := clusterClient.List(context.Background(), clusterRoleBindings, runtimeclient.MatchingLabels{iamv1beta1.UserReferenceLabel: username}); err != nil {
		return nil, err
	}
	return clusterRoleBindings, nil
}

func contains(objects []runtime.Object, object runtime.Object) bool {
	for _, item := range objects {
		if item == object {
			return true
		}
	}
	return false
}

func (t *tenantOperator) checkWorkspaceTemplatePermission(user user.Info, workspace string) error {
	deleteWST := authorizer.AttributesRecord{
		User:            user,
		Verb:            authorizer.VerbDelete,
		APIGroup:        tenantv1beta1.SchemeGroupVersion.Group,
		APIVersion:      tenantv1beta1.SchemeGroupVersion.Version,
		Resource:        tenantv1beta1.ResourcePluralWorkspaceTemplate,
		ResourceRequest: true,
		ResourceScope:   request.GlobalScope,
	}
	authorize, reason, err := t.authorizer.Authorize(deleteWST)
	if err != nil {
		return err
	}
	if authorize != authorizer.DecisionAllow {
		return errors.NewForbidden(tenantv1beta1.Resource(tenantv1beta1.ResourcePluralWorkspaceTemplate), workspace, fmt.Errorf(reason))
	}
	return nil
}

func (t *tenantOperator) checkClusterPermission(user user.Info, clusters []string) error {
	// Checking whether the user can manage the cluster requires authentication from two aspects.
	// First check whether the user has relevant global permissions,
	// and then check whether the user has relevant cluster permissions in the target cluster

	for _, clusterName := range clusters {
		cluster := &clusterv1alpha1.Cluster{}
		if err := t.client.Get(context.Background(), types.NamespacedName{Name: clusterName}, cluster); err != nil {
			return err
		}
		if cluster.Labels[clusterv1alpha1.ClusterVisibilityLabel] == clusterv1alpha1.ClusterVisibilityPublic {
			continue
		}

		deleteCluster := authorizer.AttributesRecord{
			User:            user,
			Verb:            authorizer.VerbDelete,
			APIGroup:        clusterv1alpha1.SchemeGroupVersion.Group,
			APIVersion:      clusterv1alpha1.SchemeGroupVersion.Version,
			Resource:        clusterv1alpha1.ResourcesPluralCluster,
			Cluster:         clusterName,
			ResourceRequest: true,
			ResourceScope:   request.GlobalScope,
		}
		authorize, _, err := t.authorizer.Authorize(deleteCluster)
		if err != nil {
			return err
		}

		if authorize == authorizer.DecisionAllow {
			continue
		}

		clusterRoleBindings, err := t.getClusterRoleBindingsByUser(clusterName, user.GetName())
		if err != nil {
			return err
		}

		allowed := false
		for _, clusterRoleBinding := range clusterRoleBindings.Items {
			// TODO fix me
			if clusterRoleBinding.RoleRef.Name == iamv1beta1.ClusterAdmin {
				allowed = true
				break
			}
		}

		if !allowed {
			return errors.NewForbidden(clusterv1alpha1.Resource(clusterv1alpha1.ResourcesPluralCluster), clusterName, fmt.Errorf("user is not allowed to use the cluster %s", clusterName))
		}
	}

	return nil
}
