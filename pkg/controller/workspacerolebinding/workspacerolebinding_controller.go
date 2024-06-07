/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package workspacerolebinding

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	errorutils "k8s.io/apimachinery/pkg/util/errors"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/constants"
	clusterutils "kubesphere.io/kubesphere/pkg/controller/cluster/utils"
	"kubesphere.io/kubesphere/pkg/controller/workspacetemplate/utils"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
)

const (
	controllerName = "workspacerolebinding"
	finalizer      = "finalizers.kubesphere.io/workspacerolebindings"
)

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

func (r *Reconciler) Name() string {
	return controllerName
}

// Reconciler reconciles a WorkspaceRoleBinding object
type Reconciler struct {
	client.Client
	cache         cache.Cache
	logger        logr.Logger
	recorder      record.EventRecorder
	ClusterClient clusterclient.Interface
}

func (r *Reconciler) Start(ctx context.Context) error {
	informer, err := r.cache.GetInformer(ctx, &clusterv1alpha1.Cluster{})
	if err != nil {
		return err
	}
	var hostCluster *clusterv1alpha1.Cluster
	clusters, err := r.ClusterClient.ListClusters(ctx)
	if err != nil {
		return err
	}
	for i := range clusters {
		if clusterutils.IsHostCluster(&clusters[i]) {
			hostCluster = &clusters[i]
			break
		}
	}
	_, err = informer.AddEventHandler(toolscache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldCluster := oldObj.(*clusterv1alpha1.Cluster)
			newCluster := newObj.(*clusterv1alpha1.Cluster)
			// cluster is ready
			if !clusterutils.IsHostCluster(newCluster) &&
				(!clusterutils.IsClusterReady(oldCluster) && clusterutils.IsClusterReady(newCluster)) {
				err := r.CompletelySync(ctx, *hostCluster, *newCluster)
				if err != nil {
					r.logger.Error(err, "failed to sync workspacerolebinding for cluster",
						"cluster", newCluster.Name)
					return
				}
			}
		},
	})
	return err
}

func (r *Reconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.ClusterClient = mgr.ClusterClient
	r.Client = mgr.GetClient()
	r.cache = mgr.GetCache()
	r.logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	r.recorder = mgr.GetEventRecorderFor(controllerName)

	err := mgr.Add(r)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		For(&iamv1beta1.WorkspaceRoleBinding{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=iam.kubesphere.io,resources=workspacerolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=types.kubefed.io,resources=federatedworkspacerolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=workspaces,verbs=get;list;watch;

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	workspaceRoleBinding := &iamv1beta1.WorkspaceRoleBinding{}
	if err := r.Get(ctx, req.NamespacedName, workspaceRoleBinding); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if workspaceRoleBinding.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !controllerutil.ContainsFinalizer(workspaceRoleBinding, finalizer) {
			expected := workspaceRoleBinding.DeepCopy()
			controllerutil.AddFinalizer(expected, finalizer)
			return ctrl.Result{}, r.Patch(ctx, expected, client.MergeFrom(workspaceRoleBinding))
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(workspaceRoleBinding, finalizer) {
			if err := r.deleteRelatedResources(ctx, workspaceRoleBinding); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to delete related resources: %s", err)
			}
			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(workspaceRoleBinding, finalizer)
			if err := r.Update(ctx, workspaceRoleBinding, &client.UpdateOptions{}); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if err := r.bindWorkspace(ctx, workspaceRoleBinding); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.multiClusterSync(ctx, workspaceRoleBinding); err != nil {
		return ctrl.Result{}, err
	}

	r.recorder.Event(workspaceRoleBinding, corev1.EventTypeNormal, kscontroller.Synced, kscontroller.MessageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *Reconciler) deleteRelatedResources(ctx context.Context, workspaceRoleBinding *iamv1beta1.WorkspaceRoleBinding) error {
	clusters, err := r.ClusterClient.ListClusters(ctx)
	if err != nil {
		return fmt.Errorf("failed to list clusters: %s", err)
	}
	var notReadyClusters []string
	for _, cluster := range clusters {
		if clusterutils.IsHostCluster(&cluster) {
			continue
		}
		// skip if cluster is not ready
		if !clusterutils.IsClusterReady(&cluster) {
			notReadyClusters = append(notReadyClusters, cluster.Name)
			continue
		}
		clusterClient, err := r.ClusterClient.GetRuntimeClient(cluster.Name)
		if err != nil {
			return fmt.Errorf("failed to get cluster client: %s", err)
		}
		if err = clusterClient.Delete(ctx, &iamv1beta1.WorkspaceRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: workspaceRoleBinding.Name}}); err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			logr.FromContextOrDiscard(ctx).Error(err, "failed to delete related resources")
		}
	}
	if len(notReadyClusters) > 0 {
		err = fmt.Errorf("cluster not ready: %s", strings.Join(notReadyClusters, ","))
		logr.FromContextOrDiscard(ctx).Error(err, "failed to delete related resources")
	}
	return nil
}

func (r *Reconciler) multiClusterSync(ctx context.Context, workspaceRoleBinding *iamv1beta1.WorkspaceRoleBinding) error {
	clusters, err := r.ClusterClient.ListClusters(ctx)
	if err != nil {
		return fmt.Errorf("failed to list clusters: %s", err)
	}
	var notReadyClusters []string
	for _, cluster := range clusters {
		// skip if cluster is not ready
		if !clusterutils.IsClusterReady(&cluster) {
			notReadyClusters = append(notReadyClusters, cluster.Name)
			continue
		}
		if clusterutils.IsHostCluster(&cluster) {
			continue
		}
		// skip the sync failed , cause error will break process
		if err := r.syncWorkspaceRoleBinding(ctx, cluster, workspaceRoleBinding); err != nil {
			r.recorder.Event(workspaceRoleBinding, corev1.EventTypeWarning, kscontroller.SyncFailed,
				fmt.Sprintf("failed to sync workspace role binding %s to cluster %s: %s", workspaceRoleBinding.Name, cluster.Name, err))
		}
	}
	if len(notReadyClusters) > 0 {
		klog.FromContext(ctx).V(4).Info("cluster not ready", "clusters", strings.Join(notReadyClusters, ","))
		r.recorder.Event(workspaceRoleBinding, corev1.EventTypeWarning, kscontroller.SyncFailed, fmt.Sprintf("cluster not ready: %s", strings.Join(notReadyClusters, ",")))
	}
	return nil
}

func (r *Reconciler) syncWorkspaceRoleBinding(ctx context.Context, cluster clusterv1alpha1.Cluster, workspaceRoleBinding *iamv1beta1.WorkspaceRoleBinding) error {
	clusterClient, err := r.ClusterClient.GetRuntimeClient(cluster.Name)
	if err != nil {
		return err
	}

	workspaceTemplate := &tenantv1beta1.WorkspaceTemplate{}
	if err := r.Get(ctx, types.NamespacedName{Name: workspaceRoleBinding.Labels[tenantv1beta1.WorkspaceLabel]}, workspaceTemplate); err != nil {
		return client.IgnoreNotFound(err)
	}

	if utils.WorkspaceTemplateMatchTargetCluster(workspaceTemplate, &cluster) {
		target := &iamv1beta1.WorkspaceRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: workspaceRoleBinding.Name}}
		op, err := controllerutil.CreateOrUpdate(ctx, clusterClient, target, func() error {
			target.Labels = workspaceRoleBinding.Labels
			target.Annotations = workspaceRoleBinding.Annotations
			target.RoleRef = workspaceRoleBinding.RoleRef
			target.Subjects = workspaceRoleBinding.Subjects
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to update workspace role binding: %s", err)
		}
		klog.FromContext(ctx).V(4).Info("workspace role binding successfully synced", "cluster", cluster.Name, "operation", op, "name", workspaceRoleBinding.Name)
	} else {
		return client.IgnoreNotFound(clusterClient.DeleteAllOf(ctx, &iamv1beta1.WorkspaceRole{}, client.MatchingLabels{tenantv1beta1.WorkspaceLabel: workspaceTemplate.Name}))
	}
	return nil
}

func (r *Reconciler) bindWorkspace(ctx context.Context, workspaceRoleBinding *iamv1beta1.WorkspaceRoleBinding) error {
	workspaceName := workspaceRoleBinding.Labels[constants.WorkspaceLabelKey]
	if workspaceName == "" {
		return nil
	}
	workspace := &tenantv1beta1.WorkspaceTemplate{}
	if err := r.Get(ctx, types.NamespacedName{Name: workspaceName}, workspace); err != nil {
		// skip if workspace not found
		return client.IgnoreNotFound(err)
	}
	// owner reference not match workspace label
	if !metav1.IsControlledBy(workspaceRoleBinding, workspace) {
		workspaceRoleBinding.OwnerReferences = k8sutil.RemoveWorkspaceOwnerReference(workspaceRoleBinding.OwnerReferences)
		if err := controllerutil.SetControllerReference(workspace, workspaceRoleBinding, r.Scheme()); err != nil {
			return err
		}
		if err := r.Update(ctx, workspaceRoleBinding); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) CompletelySync(ctx context.Context, hostCluster, cluster clusterv1alpha1.Cluster) error {
	searchMap := map[string]iamv1beta1.WorkspaceRoleBinding{}
	// list resources at host cluster
	hostList := &iamv1beta1.WorkspaceRoleBindingList{}
	hostClusterClient, err := r.ClusterClient.GetRuntimeClient(hostCluster.Name)
	if err != nil {
		return err
	}
	err = hostClusterClient.List(ctx, hostList)
	if err != nil {
		return err
	}

	// list resources at member cluster
	clusterClient, err := r.ClusterClient.GetRuntimeClient(cluster.Name)
	if err != nil {
		return err
	}
	memberList := &iamv1beta1.WorkspaceRoleBindingList{}
	err = clusterClient.List(ctx, memberList)
	if err != nil {
		return err
	}

	for _, item := range memberList.Items {
		searchMap[item.GetName()] = item
	}

	var errList []error
	// check and update
	for _, item := range hostList.Items {
		workspaceTemplate := &tenantv1beta1.WorkspaceTemplate{}
		if err := r.Get(ctx, types.NamespacedName{Name: item.Labels[tenantv1beta1.WorkspaceLabel]}, workspaceTemplate); err != nil {
			return client.IgnoreNotFound(err)
		}
		if !utils.WorkspaceTemplateMatchTargetCluster(workspaceTemplate, &cluster) {
			continue
		}

		memObj, exist := searchMap[item.GetName()]
		if !exist {
			if err := clusterClient.Create(ctx, &item); err != nil {
				err = fmt.Errorf("create worspaceRoleBinding %s at cluster %s failed", item.Name, cluster.Name)
				errList = append(errList, err)
			}
			continue
		}
		if !bindingEqual(item, memObj) {
			memObj.Labels = item.Labels
			memObj.Annotations = item.Annotations
			memObj.RoleRef = item.RoleRef
			memObj.Subjects = item.Subjects
			if err := clusterClient.Update(ctx, &memObj); err != nil {
				err = fmt.Errorf("update worspaceRoleBinding %s at cluster %s failed", item.Name, cluster.Name)
				errList = append(errList, err)
			}
		}
		delete(searchMap, memObj.GetName())
	}

	for _, obj := range searchMap {
		err := clusterClient.Delete(ctx, &obj)
		if err != nil {
			err = fmt.Errorf("delete worspaceRoleBinding %s at cluster %s failed", obj.Name, cluster.Name)
			errList = append(errList, err)
		}
	}
	return errorutils.NewAggregate(errList)
}

func bindingEqual(a, b iamv1beta1.WorkspaceRoleBinding) bool {
	if a.Name != b.Name || len(a.Subjects) != len(b.Subjects) || !reflect.DeepEqual(a.RoleRef, b.RoleRef) {
		return false
	}

	sort.Sort(SortableSubjectSlice(a.Subjects))
	sort.Sort(SortableSubjectSlice(b.Subjects))

	for i, subject := range a.Subjects {
		if !reflect.DeepEqual(subject, b.Subjects[i]) {
			return false
		}
	}

	return true
}

type SortableSubjectSlice []v1.Subject

func (s SortableSubjectSlice) Len() int {
	return len(s)
}

func (s SortableSubjectSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s SortableSubjectSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
