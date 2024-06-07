/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package group

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/tools/record"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"kubesphere.io/kubesphere/pkg/controller"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const (
	controllerName = "group"
	finalizer      = "finalizers.kubesphere.io/groups"
)

type Reconciler struct {
	client.Client

	recorder record.EventRecorder
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	r.Client = mgr.GetClient()
	return builder.
		ControllerManagedBy(mgr).
		For(
			&iamv1beta1.Group{},
			builder.WithPredicates(
				predicate.ResourceVersionChangedPredicate{},
			),
		).
		Named(controllerName).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	group := &iamv1beta1.Group{}
	if err := r.Get(ctx, req.NamespacedName, group); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if group.ObjectMeta.DeletionTimestamp.IsZero() {
		var g *iamv1beta1.Group
		if !sliceutil.HasString(group.Finalizers, finalizer) {
			g = group.DeepCopy()
			g.ObjectMeta.Finalizers = append(g.ObjectMeta.Finalizers, finalizer)
		}

		if g != nil {
			return ctrl.Result{}, r.Update(ctx, g)
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(group.ObjectMeta.Finalizers, finalizer) {
			if err := r.deleteGroupBindings(ctx, group); err != nil {
				return ctrl.Result{}, err
			}

			if err := r.deleteRoleBindings(ctx, group); err != nil {
				return ctrl.Result{}, err
			}

			group.Finalizers = sliceutil.RemoveString(group.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})

			return ctrl.Result{}, r.Update(ctx, group)
		}
		return ctrl.Result{}, nil
	}

	// TODO: sync logic needs to be updated and no longer relies on KubeFed, it needs to be synchronized manually.

	r.recorder.Event(group, corev1.EventTypeNormal, controller.Synced, controller.MessageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *Reconciler) deleteGroupBindings(ctx context.Context, group *iamv1beta1.Group) error {
	if len(group.Name) > validation.LabelValueMaxLength {
		// ignore invalid label value error
		return nil
	}

	// Group bindings that created by kubesphere will be deleted directly.
	return r.DeleteAllOf(ctx, &iamv1beta1.GroupBinding{}, client.GracePeriodSeconds(0), client.MatchingLabelsSelector{
		Selector: labels.SelectorFromValidatedSet(labels.Set{iamv1beta1.GroupReferenceLabel: group.Name}),
	})
}

// remove all RoleBindings.
func (r *Reconciler) deleteRoleBindings(ctx context.Context, group *iamv1beta1.Group) error {
	if len(group.Name) > validation.LabelValueMaxLength {
		// ignore invalid label value error
		return nil
	}

	selector := labels.SelectorFromValidatedSet(labels.Set{iamv1beta1.GroupReferenceLabel: group.Name})
	deleteOption := client.GracePeriodSeconds(0)

	if err := r.DeleteAllOf(ctx, &iamv1beta1.WorkspaceRoleBinding{}, deleteOption, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return err
	}

	if err := r.DeleteAllOf(ctx, &rbacv1.ClusterRoleBinding{}, deleteOption, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return err
	}

	namespaces := &corev1.NamespaceList{}
	if err := r.List(ctx, namespaces); err != nil {
		return err
	}
	for _, namespace := range namespaces.Items {
		if err := r.DeleteAllOf(ctx, &rbacv1.RoleBinding{}, deleteOption, client.MatchingLabelsSelector{Selector: selector}, client.InNamespace(namespace.Name)); err != nil {
			return err
		}
	}
	return nil
}
