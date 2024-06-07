/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package core

import (
	"context"
	"strconv"
	"strings"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	categoryController       = "extension-category"
	countOfRelatedExtensions = "kubesphere.io/count"
)

var _ kscontroller.Controller = &CategoryReconciler{}
var _ reconcile.Reconciler = &CategoryReconciler{}

func (r *CategoryReconciler) Name() string {
	return categoryController
}

func (r *CategoryReconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

type CategoryReconciler struct {
	client.Client
	recorder record.EventRecorder
	logger   logr.Logger
}

func (r *CategoryReconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	r.logger = ctrl.Log.WithName("controllers").WithName(categoryController)
	r.recorder = mgr.GetEventRecorderFor(categoryController)
	return ctrl.NewControllerManagedBy(mgr).
		Named(categoryController).
		For(&corev1alpha1.Category{}).
		Watches(
			&corev1alpha1.Extension{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
				var requests []reconcile.Request
				extension := object.(*corev1alpha1.Extension)
				if category := extension.Labels[corev1alpha1.CategoryLabel]; category != "" {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name: category,
						},
					})
				}
				return requests
			}),
			builder.WithPredicates(predicate.LabelChangedPredicate{}),
		).
		Complete(r)
}

func (r *CategoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.logger.WithValues("category", req.String())
	logger.V(4).Info("sync category")
	ctx = klog.NewContext(ctx, logger)

	category := &corev1alpha1.Category{}
	if err := r.Client.Get(ctx, req.NamespacedName, category); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	extensions := &corev1alpha1.ExtensionList{}
	if err := r.List(ctx, extensions, client.MatchingLabels{corev1alpha1.CategoryLabel: category.Name}); err != nil {
		return ctrl.Result{}, err
	}

	total := strconv.Itoa(len(extensions.Items))
	if category.Annotations[countOfRelatedExtensions] != total {
		if category.Annotations == nil {
			category.Annotations = make(map[string]string)
		}
		category.Annotations[countOfRelatedExtensions] = total
		if err := r.Update(ctx, category); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
