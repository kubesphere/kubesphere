/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package core

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"
)

const (
	extensionProtection = "kubesphere.io/extension-protection"
	extensionController = "extension"
)

var _ kscontroller.Controller = &ExtensionReconciler{}
var _ reconcile.Reconciler = &ExtensionReconciler{}

func (r *ExtensionReconciler) Name() string {
	return extensionController
}

func (r *ExtensionReconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

type ExtensionReconciler struct {
	client.Client
	k8sVersion *semver.Version
	logger     logr.Logger
}

func (r *ExtensionReconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	r.k8sVersion = mgr.K8sVersion
	r.logger = ctrl.Log.WithName("controllers").WithName(extensionController)
	return ctrl.NewControllerManagedBy(mgr).
		Named(extensionController).
		For(&corev1alpha1.Extension{}).
		Watches(
			&corev1alpha1.ExtensionVersion{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
				var requests []reconcile.Request
				extensionVersion := object.(*corev1alpha1.ExtensionVersion)
				extensionName := extensionVersion.Labels[corev1alpha1.ExtensionReferenceLabel]
				if extensionName != "" {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name: extensionName,
						},
					})
				}
				return requests
			}),
			builder.WithPredicates(predicate.Funcs{
				GenericFunc: func(event event.GenericEvent) bool {
					return false
				},
				UpdateFunc: func(updateEvent event.UpdateEvent) bool {
					return false
				},
			}),
		).
		Complete(r)
}

func (r *ExtensionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	extension := &corev1alpha1.Extension{}
	if err := r.Client.Get(ctx, req.NamespacedName, extension); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.logger.V(4).Info("reconcile", "extension", extension.Name)

	if extension.ObjectMeta.DeletionTimestamp != nil {
		return r.reconcileDelete(ctx, extension)
	}

	if !controllerutil.ContainsFinalizer(extension, extensionProtection) {
		expected := extension.DeepCopy()
		controllerutil.AddFinalizer(expected, extensionProtection)
		return ctrl.Result{}, r.Patch(ctx, expected, client.MergeFrom(extension))
	}

	if err := r.syncExtensionStatus(ctx, extension); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to sync extension status: %s", err)
	}

	r.logger.V(4).Info("synced", "extension", extension.Name)

	return ctrl.Result{}, nil
}

// reconcileDelete delete the extension.
func (r *ExtensionReconciler) reconcileDelete(ctx context.Context, extension *corev1alpha1.Extension) (ctrl.Result, error) {
	deletePolicy := metav1.DeletePropagationBackground
	if err := r.DeleteAllOf(ctx, &corev1alpha1.ExtensionVersion{}, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			LabelSelector: labels.SelectorFromSet(labels.Set{corev1alpha1.ExtensionReferenceLabel: extension.Name}),
		},
		DeleteOptions: client.DeleteOptions{PropagationPolicy: &deletePolicy},
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to delete related ExtensionVersion: %s", err)
	}

	// Remove the finalizer from the extension
	controllerutil.RemoveFinalizer(extension, extensionProtection)
	if err := r.Update(ctx, extension); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ExtensionReconciler) syncExtensionStatus(ctx context.Context, extension *corev1alpha1.Extension) error {
	versionList := corev1alpha1.ExtensionVersionList{}
	if err := r.List(ctx, &versionList, client.MatchingLabels{
		corev1alpha1.ExtensionReferenceLabel: extension.Name,
	}); err != nil {
		return err
	}

	versions := make([]corev1alpha1.ExtensionVersionInfo, 0, len(versionList.Items))
	for i := range versionList.Items {
		if versionList.Items[i].DeletionTimestamp.IsZero() {
			versions = append(versions, corev1alpha1.ExtensionVersionInfo{
				Version:           versionList.Items[i].Spec.Version,
				CreationTimestamp: versionList.Items[i].CreationTimestamp,
			})
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version < versions[j].Version
	})

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := r.Get(ctx, types.NamespacedName{Name: extension.Name}, extension); err != nil {
			return err
		}
		expected := extension.DeepCopy()
		if recommended, err := getRecommendedExtensionVersion(versionList.Items, r.k8sVersion); err == nil {
			expected.Status.RecommendedVersion = recommended
		} else {
			r.logger.Error(err, "failed to get recommended extension version")
		}
		expected.Status.Versions = versions
		if expected.Status.RecommendedVersion != extension.Status.RecommendedVersion ||
			!reflect.DeepEqual(expected.Status.Versions, extension.Status.Versions) {
			return r.Update(ctx, expected)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update extension status: %s", err)
	}
	return nil
}
