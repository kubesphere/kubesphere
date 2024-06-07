/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package globalrole

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
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rbachelper "kubesphere.io/kubesphere/pkg/componenthelper/auth/rbac"
	kscontroller "kubesphere.io/kubesphere/pkg/controller"
	"kubesphere.io/kubesphere/pkg/controller/cluster/predicate"
	clusterutils "kubesphere.io/kubesphere/pkg/controller/cluster/utils"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
)

const (
	controllerName = "globalrole"
	finalizer      = "finalizers.kubesphere.io/globalroles"
)

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

type Reconciler struct {
	client.Client
	logger        logr.Logger
	recorder      record.EventRecorder
	helper        *rbachelper.Helper
	clusterClient clusterclient.Interface
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.clusterClient = mgr.ClusterClient
	r.Client = mgr.GetClient()
	r.helper = rbachelper.NewHelper(r.Client)
	r.logger = mgr.GetLogger().WithName(controllerName)
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	return builder.
		ControllerManagedBy(mgr).
		For(&iamv1beta1.GlobalRole{}).
		Watches(
			&clusterv1alpha1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(r.mapper),
			builder.WithPredicates(predicate.ClusterStatusChangedPredicate{}),
		).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Named(controllerName).
		Complete(r)
}

func (r *Reconciler) mapper(ctx context.Context, o client.Object) []reconcile.Request {
	cluster := o.(*clusterv1alpha1.Cluster)
	var requests []reconcile.Request
	if !clusterutils.IsClusterReady(cluster) {
		return requests
	}
	globalRoles := &iamv1beta1.GlobalRoleList{}
	if err := r.List(ctx, globalRoles); err != nil {
		r.logger.Error(err, "failed to list global roles")
		return requests
	}
	for _, globalRole := range globalRoles.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: globalRole.Name}})
	}
	return requests
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	globalRole := &iamv1beta1.GlobalRole{}
	if err := r.Get(ctx, req.NamespacedName, globalRole); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if globalRole.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !controllerutil.ContainsFinalizer(globalRole, finalizer) {
			expected := globalRole.DeepCopy()
			controllerutil.AddFinalizer(expected, finalizer)
			return ctrl.Result{}, r.Patch(ctx, expected, client.MergeFrom(globalRole))
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(globalRole, finalizer) {
			if err := r.deleteRelatedResources(ctx, globalRole); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to delete related resources: %s", err)
			}
			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(globalRole, finalizer)
			if err := r.Update(ctx, globalRole, &client.UpdateOptions{}); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if globalRole.AggregationRoleTemplates != nil {
		if err := r.helper.AggregationRole(ctx, rbachelper.GlobalRoleRuleOwner{GlobalRole: globalRole}, r.recorder); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.multiClusterSync(ctx, globalRole); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) deleteRelatedResources(ctx context.Context, globalRole *iamv1beta1.GlobalRole) error {
	clusters, err := r.clusterClient.ListClusters(ctx)
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
		clusterClient, err := r.clusterClient.GetRuntimeClient(cluster.Name)
		if err != nil {
			return fmt.Errorf("failed to get cluster client: %s", err)
		}
		if err = clusterClient.Delete(ctx, &iamv1beta1.GlobalRole{ObjectMeta: metav1.ObjectMeta{Name: globalRole.Name}}); err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return err
		}
	}
	if len(notReadyClusters) > 0 {
		err = fmt.Errorf("cluster not ready: %s", strings.Join(notReadyClusters, ","))
		klog.FromContext(ctx).Error(err, "failed to delete related resources")
		r.recorder.Event(globalRole, corev1.EventTypeWarning, kscontroller.SyncFailed, fmt.Sprintf("cluster not ready: %s", strings.Join(notReadyClusters, ",")))
		return err
	}
	return nil
}

func (r *Reconciler) multiClusterSync(ctx context.Context, globalRole *iamv1beta1.GlobalRole) error {
	clusters, err := r.clusterClient.ListClusters(ctx)
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
		if err := r.syncGlobalRole(ctx, cluster, globalRole); err != nil {
			return fmt.Errorf("failed to sync global role %s to cluster %s: %s", globalRole.Name, cluster.Name, err)
		}
	}
	if len(notReadyClusters) > 0 {
		klog.FromContext(ctx).V(4).Info("cluster not ready", "clusters", strings.Join(notReadyClusters, ","))
		r.recorder.Event(globalRole, corev1.EventTypeWarning, kscontroller.SyncFailed, fmt.Sprintf("cluster not ready: %s", strings.Join(notReadyClusters, ",")))
	}
	return nil
}

func (r *Reconciler) syncGlobalRole(ctx context.Context, cluster clusterv1alpha1.Cluster, globalRole *iamv1beta1.GlobalRole) error {
	if clusterutils.IsHostCluster(&cluster) {
		return nil
	}
	clusterClient, err := r.clusterClient.GetRuntimeClient(cluster.Name)
	if err != nil {
		return fmt.Errorf("failed to get cluster client: %s", err)
	}
	target := &iamv1beta1.GlobalRole{ObjectMeta: metav1.ObjectMeta{Name: globalRole.Name}}
	op, err := controllerutil.CreateOrUpdate(ctx, clusterClient, target, func() error {
		target.Labels = globalRole.Labels
		target.Annotations = globalRole.Annotations
		target.Rules = globalRole.Rules
		target.AggregationRoleTemplates = globalRole.AggregationRoleTemplates
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update global role: %s", err)
	}

	r.logger.V(4).Info("global role successfully synced", "cluster", cluster.Name, "operation", op, "name", globalRole.Name)
	return nil
}
