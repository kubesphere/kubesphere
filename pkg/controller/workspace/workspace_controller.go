/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package workspace

import (
	"context"
	"fmt"

	"kubesphere.io/kubesphere/pkg/constants"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"
)

const (
	controllerName = "workspace"
)

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

// Reconciler reconciles a Workspace object
type Reconciler struct {
	client.Client
	logger   logr.Logger
	recorder record.EventRecorder
}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	r.logger = mgr.GetLogger().WithName(controllerName)
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		For(&tenantv1beta1.Workspace{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.logger.WithValues("workspace", req.NamespacedName)
	ctx = klog.NewContext(ctx, logger)
	workspace := &tenantv1beta1.Workspace{}
	if err := r.Get(ctx, req.NamespacedName, workspace); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if workspace.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !controllerutil.ContainsFinalizer(workspace, constants.CascadingDeletionFinalizer) {
			expected := workspace.DeepCopy()
			// Remove legacy finalizer
			controllerutil.RemoveFinalizer(expected, "finalizers.tenant.kubesphere.io")
			controllerutil.AddFinalizer(expected, constants.CascadingDeletionFinalizer)
			if err := r.Patch(ctx, expected, client.MergeFrom(workspace)); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to add finalizer: %s", err)
			}
			workspaceOperation.WithLabelValues("create", workspace.Name).Inc()
		}
	} else {
		if controllerutil.ContainsFinalizer(workspace, constants.CascadingDeletionFinalizer) {
			ok, err := r.workspaceCascadingDeletion(ctx, workspace)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to delete workspace: %s", err)
			}
			if ok {
				controllerutil.RemoveFinalizer(workspace, constants.CascadingDeletionFinalizer)
				if err := r.Update(ctx, workspace); err != nil {
					return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %s", err)
				}
				workspaceOperation.WithLabelValues("delete", workspace.Name).Inc()
			}
		}
		// Our finalizer has finished, so the reconciler can do nothing.
		return ctrl.Result{}, nil
	}

	r.recorder.Event(workspace, corev1.EventTypeNormal, "Reconcile", "Reconcile workspace successfully")
	return ctrl.Result{}, nil
}

// workspaceCascadingDeletion handles the cascading deletion of a workspace based on its deletion propagation policy.
// It returns a boolean indicating whether the deletion was successful and an error if any occurred.
func (r *Reconciler) workspaceCascadingDeletion(ctx context.Context, workspace *tenantv1beta1.Workspace) (bool, error) {
	switch workspace.Annotations[constants.DeletionPropagationAnnotation] {
	case string(metav1.DeletePropagationOrphan):
		// If the deletion propagation policy is "Orphan", return true without deleting namespaces.
		return true, nil
	case string(metav1.DeletePropagationForeground), string(metav1.DeletePropagationBackground):
		// If the deletion propagation policy is "Foreground" or "Background", delete the namespaces.
		if err := r.deleteNamespaces(ctx, workspace); err != nil {
			return false, fmt.Errorf("failed to delete namespaces in workspace %s: %s", workspace.Name, err)
		}
		return true, nil
	default:
		// If the deletion propagation policy is invalid, return an error.
		return false, fmt.Errorf("invalid deletion propagation policy: %s", workspace.Annotations[constants.DeletionPropagationAnnotation])
	}
}

// deleteNamespaces deletes all namespaces associated with the given workspace.
// It uses the "Background" deletion propagation policy.
func (r *Reconciler) deleteNamespaces(ctx context.Context, workspace *tenantv1beta1.Workspace) error {
	namespaces := &corev1.NamespaceList{}
	if err := r.List(ctx, namespaces, client.MatchingLabels{tenantv1beta1.WorkspaceLabel: workspace.Name}); err != nil {
		return fmt.Errorf("failed to list namespaces in workspace %s: %s", workspace.Name, err)
	}
	for _, ns := range namespaces.Items {
		if err := r.Delete(ctx, &ns); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("failed to delete namespace %s: %s", ns.Name, err)
		}
	}
	return nil
}
