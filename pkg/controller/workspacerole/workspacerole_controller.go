/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package workspacerole

import (
	"context"
	"fmt"

	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rbachelper "kubesphere.io/kubesphere/pkg/componenthelper/auth/rbac"
	"kubesphere.io/kubesphere/pkg/constants"
	kscontroller "kubesphere.io/kubesphere/pkg/controller"
	"kubesphere.io/kubesphere/pkg/controller/cluster/predicate"
	clusterutils "kubesphere.io/kubesphere/pkg/controller/cluster/utils"
	"kubesphere.io/kubesphere/pkg/controller/workspacetemplate/utils"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
)

const (
	controllerName = "workspacerole"
	finalizer      = "finalizers.kubesphere.io/workspaceroles"
)

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

func (r *Reconciler) Name() string {
	return controllerName
}

// Reconciler reconciles a WorkspaceRole object
type Reconciler struct {
	client.Client
	logger        logr.Logger
	recorder      record.EventRecorder
	helper        *rbachelper.Helper
	ClusterClient clusterclient.Interface
}

func (r *Reconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.ClusterClient = mgr.ClusterClient
	r.Client = mgr.GetClient()
	r.logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	r.helper = rbachelper.NewHelper(r.Client)
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		For(&iamv1beta1.WorkspaceRole{}).
		Watches(
			&clusterv1alpha1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(r.mapper),
			builder.WithPredicates(predicate.ClusterStatusChangedPredicate{}),
		).
		Complete(r)
}

func (r *Reconciler) mapper(ctx context.Context, o client.Object) []reconcile.Request {
	cluster := o.(*clusterv1alpha1.Cluster)
	if !clusterutils.IsClusterReady(cluster) {
		return []reconcile.Request{}
	}
	workspaceRoles := &iamv1beta1.WorkspaceRoleList{}
	if err := r.List(ctx, workspaceRoles); err != nil {
		r.logger.Error(err, "failed to list workspace roles")
		return []reconcile.Request{}
	}
	var result []reconcile.Request
	for _, workspaceRole := range workspaceRoles.Items {
		workspaceTemplate := &tenantv1beta1.WorkspaceTemplate{}
		workspaceName := workspaceRole.Labels[tenantv1beta1.WorkspaceLabel]
		if err := r.Get(ctx, types.NamespacedName{Name: workspaceName}, workspaceTemplate); err != nil {
			klog.Warningf("failed to get workspace template %s: %s", workspaceName, err)
			continue
		}
		if utils.WorkspaceTemplateMatchTargetCluster(workspaceTemplate, cluster) {
			result = append(result, reconcile.Request{NamespacedName: types.NamespacedName{Name: workspaceRole.Name}})
		}
	}
	return result
}

// +kubebuilder:rbac:groups=iam.kubesphere.io,resources=workspaceroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=workspaces,verbs=get;list;watch;

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.logger.WithValues("workspacerole", req.NamespacedName)
	workspaceRole := &iamv1beta1.WorkspaceRole{}
	if err := r.Get(ctx, req.NamespacedName, workspaceRole); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if workspaceRole.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !controllerutil.ContainsFinalizer(workspaceRole, finalizer) {
			expected := workspaceRole.DeepCopy()
			controllerutil.AddFinalizer(expected, finalizer)
			return ctrl.Result{}, r.Patch(ctx, expected, client.MergeFrom(workspaceRole))
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(workspaceRole, finalizer) {
			if err := r.deleteRelatedResources(ctx, workspaceRole); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to delete related resources: %s", err)
			}
			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(workspaceRole, finalizer)
			if err := r.Update(ctx, workspaceRole, &client.UpdateOptions{}); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if err := r.bindWorkspace(ctx, logger, workspaceRole); err != nil {
		return ctrl.Result{}, err
	}
	if workspaceRole.AggregationRoleTemplates != nil {
		if err := r.helper.AggregationRole(ctx, rbachelper.WorkspaceRoleRuleOwner{WorkspaceRole: workspaceRole}, r.recorder); err != nil {
			return ctrl.Result{}, err
		}
	}
	if err := r.multiClusterSync(ctx, workspaceRole); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) deleteRelatedResources(ctx context.Context, workspaceRole *iamv1beta1.WorkspaceRole) error {
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
		if err = clusterClient.Delete(ctx, &iamv1beta1.WorkspaceRole{ObjectMeta: metav1.ObjectMeta{Name: workspaceRole.Name}}); err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return err
		}
	}
	if len(notReadyClusters) > 0 {
		err = fmt.Errorf("cluster not ready: %s", strings.Join(notReadyClusters, ","))
		klog.FromContext(ctx).Error(err, "failed to delete related resources")
		r.recorder.Event(workspaceRole, corev1.EventTypeWarning, kscontroller.SyncFailed, fmt.Sprintf("cluster not ready: %s", strings.Join(notReadyClusters, ",")))
		return err
	}
	return nil
}

func (r *Reconciler) bindWorkspace(ctx context.Context, logger logr.Logger, workspaceRole *iamv1beta1.WorkspaceRole) error {
	workspaceName := workspaceRole.Labels[constants.WorkspaceLabelKey]
	if workspaceName == "" {
		return nil
	}
	var workspace tenantv1beta1.WorkspaceTemplate
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := r.Get(ctx, types.NamespacedName{Name: workspaceName}, &workspace); err != nil {
			return client.IgnoreNotFound(err)
		}
		if !metav1.IsControlledBy(workspaceRole, &workspace) {
			workspaceRole.OwnerReferences = k8sutil.RemoveWorkspaceOwnerReference(workspaceRole.OwnerReferences)
			if err := controllerutil.SetControllerReference(&workspace, workspaceRole, r.Scheme()); err != nil {
				return err
			}
			return r.Update(ctx, workspaceRole)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update workspace role %s: %s", workspaceRole.Name, err)
	}
	return nil
}

func (r *Reconciler) multiClusterSync(ctx context.Context, workspaceRole *iamv1beta1.WorkspaceRole) error {
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
		if err := r.syncWorkspaceRole(ctx, cluster, workspaceRole); err != nil {
			return fmt.Errorf("failed to sync workspace role %s to cluster %s: %s", workspaceRole.Name, cluster.Name, err)
		}
	}
	if len(notReadyClusters) > 0 {
		klog.FromContext(ctx).V(4).Info("cluster not ready", "clusters", strings.Join(notReadyClusters, ","))
		r.recorder.Event(workspaceRole, corev1.EventTypeWarning, kscontroller.SyncFailed, fmt.Sprintf("cluster not ready: %s", strings.Join(notReadyClusters, ",")))
	}
	return nil
}

func (r *Reconciler) syncWorkspaceRole(ctx context.Context, cluster clusterv1alpha1.Cluster, workspaceRole *iamv1beta1.WorkspaceRole) error {
	clusterClient, err := r.ClusterClient.GetRuntimeClient(cluster.Name)
	if err != nil {
		return fmt.Errorf("failed to get cluster client: %s", err)
	}
	workspaceTemplate := &tenantv1beta1.WorkspaceTemplate{}
	if err := r.Get(ctx, types.NamespacedName{Name: workspaceRole.Labels[tenantv1beta1.WorkspaceLabel]}, workspaceTemplate); err != nil {
		return client.IgnoreNotFound(err)
	}
	if utils.WorkspaceTemplateMatchTargetCluster(workspaceTemplate, &cluster) {
		target := &iamv1beta1.WorkspaceRole{ObjectMeta: metav1.ObjectMeta{Name: workspaceRole.Name}}
		op, err := controllerutil.CreateOrUpdate(ctx, clusterClient, target, func() error {
			target.Labels = workspaceRole.Labels
			target.Annotations = workspaceRole.Annotations
			target.Rules = workspaceRole.Rules
			target.AggregationRoleTemplates = workspaceRole.AggregationRoleTemplates
			return nil
		})
		if err != nil {
			return err
		}
		klog.FromContext(ctx).V(4).Info("workspace role successfully synced", "cluster", cluster.Name, "operation", op, "name", workspaceRole.Name)
	} else {
		return client.IgnoreNotFound(clusterClient.DeleteAllOf(ctx, &iamv1beta1.WorkspaceRole{}, client.MatchingLabels{tenantv1beta1.WorkspaceLabel: workspaceTemplate.Name}))
	}
	return nil
}
