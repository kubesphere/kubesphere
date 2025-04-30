/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package application

import (
	"context"
	"strings"

	"github.com/go-logr/logr"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	erro "errors"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	appv2 "kubesphere.io/api/application/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	categoryController = "app-category"
	categoryFinalizer  = "categories.application.kubesphere.io"
)

var _ reconcile.Reconciler = &AppCategoryReconciler{}
var _ kscontroller.Controller = &AppCategoryReconciler{}

type AppCategoryReconciler struct {
	client.Client
	logger logr.Logger
}

func (r *AppCategoryReconciler) Name() string {
	return categoryController
}

func (r *AppCategoryReconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (r *AppCategoryReconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	r.logger = ctrl.Log.WithName("controllers").WithName(categoryController)
	return ctrl.NewControllerManagedBy(mgr).
		Named(categoryController).
		For(&appv2.Category{}).
		Watches(
			&appv2.Application{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
				var requests []reconcile.Request
				app := object.(*appv2.Application)
				if categoryID := app.Labels[appv2.AppCategoryNameKey]; categoryID != "" {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{Name: categoryID},
					})
				}
				return requests
			}),
			builder.WithPredicates(predicate.LabelChangedPredicate{}),
		).
		Complete(r)
}

func (r *AppCategoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger.V(4).Info("reconcile app category", "app category", req.String())
	logger := r.logger.WithValues("app category", req.String())
	category := &appv2.Category{}
	if err := r.Client.Get(ctx, req.NamespacedName, category); err != nil {
		if errors.IsNotFound(err) {
			if req.Name == appv2.UncategorizedCategoryID {
				return reconcile.Result{}, r.ensureUncategorizedCategory()
			}
			// ignore exceptions caused by incorrectly adding app labels.
			logger.Error(err, "not found, check if you added the correct app category")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if !controllerutil.ContainsFinalizer(category, categoryFinalizer) {
		category.ObjectMeta.Finalizers = append(category.ObjectMeta.Finalizers, categoryFinalizer)
		return ctrl.Result{}, r.Update(ctx, category)
	}

	if !category.ObjectMeta.DeletionTimestamp.IsZero() {
		// our finalizer is present, so lets handle our external dependency
		// remove our finalizer from the list and update it.
		if category.Status.Total > 0 {
			logger.Error(erro.New("category is using"), "can not delete helm category, in which owns applications")
			return reconcile.Result{}, nil
		}

		controllerutil.RemoveFinalizer(category, categoryFinalizer)
		return reconcile.Result{}, r.Update(ctx, category)
	}

	apps := &appv2.ApplicationList{}
	opts := client.MatchingLabels{
		appv2.AppCategoryNameKey: category.Name,
		appv2.RepoIDLabelKey:     appv2.UploadRepoKey,
	}
	if err := r.List(ctx, apps, opts); err != nil {
		r.logger.Error(err, "failed to list apps")
		return ctrl.Result{}, err
	}
	if category.Status.Total != len(apps.Items) {
		category.Status.Total = len(apps.Items)
		if err := r.Status().Update(ctx, category); err != nil {
			r.logger.Error(err, "failed to update category status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *AppCategoryReconciler) ensureUncategorizedCategory() error {
	ctg := &appv2.Category{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: appv2.UncategorizedCategoryID}, ctg)
	if err != nil && !errors.IsNotFound(err) {
		r.logger.Error(err, "failed to get uncategorized category")
		return err
	}
	ctg.Name = appv2.UncategorizedCategoryID

	return r.Create(context.TODO(), ctg)
}
