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

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"kubesphere.io/utils/s3"

	"kubesphere.io/kubesphere/pkg/simple/client/application"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	appv2 "kubesphere.io/api/application/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	appVersionController = "appversion"
)

var _ reconcile.Reconciler = &AppVersionReconciler{}
var _ kscontroller.Controller = &AppVersionReconciler{}

type AppVersionReconciler struct {
	client.Client
	ossStore s3.Interface
	cmStore  s3.Interface
	logger   logr.Logger
}

func (r *AppVersionReconciler) Name() string {
	return appVersionController
}

func (r *AppVersionReconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (r *AppVersionReconciler) SetupWithManager(mgr *kscontroller.Manager) (err error) {
	r.Client = mgr.GetClient()
	r.logger = ctrl.Log.WithName("controllers").WithName(appVersionController)
	r.cmStore, r.ossStore, err = application.InitStore(mgr.Options.S3Options, r.Client)
	if err != nil {
		r.logger.Error(err, "failed to init store")
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(appVersionController).
		For(&appv2.ApplicationVersion{}).
		Complete(r)
}

func (r *AppVersionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	appVersion := &appv2.ApplicationVersion{}
	if err := r.Client.Get(ctx, req.NamespacedName, appVersion); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	logger := r.logger.WithValues("application version", appVersion.Name)
	if !controllerutil.ContainsFinalizer(appVersion, appv2.CleanupFinalizer) {
		controllerutil.RemoveFinalizer(appVersion, appv2.StoreCleanFinalizer)
		controllerutil.AddFinalizer(appVersion, appv2.CleanupFinalizer)
		return ctrl.Result{}, r.Update(ctx, appVersion)
	}

	//Delete app files, non-important logic, errors will not affect the main process
	if !appVersion.ObjectMeta.DeletionTimestamp.IsZero() {
		err := r.deleteFile(ctx, appVersion)
		if err != nil {
			logger.Error(err, "Failed to clean file")
		}
	}

	return ctrl.Result{}, nil
}

func (r *AppVersionReconciler) deleteFile(ctx context.Context, appVersion *appv2.ApplicationVersion) error {
	logger := r.logger.WithValues("application version", appVersion.Name)
	defer func() {
		controllerutil.RemoveFinalizer(appVersion, appv2.CleanupFinalizer)
		err := r.Update(ctx, appVersion)
		if err != nil {
			logger.Error(err, "Failed to remove finalizer from application version")
		}
		logger.V(4).Info("Remove finalizer from application version %s successfully")
	}()

	logger.V(4).Info("ApplicationVersion has been deleted, try to clean file")
	id := []string{appVersion.Name}
	apprls := &appv2.ApplicationReleaseList{}
	err := r.Client.List(ctx, apprls, client.MatchingLabels{appv2.AppVersionIDLabelKey: appVersion.Name})
	if err != nil {
		logger.Error(err, "Failed to list ApplicationRelease")
		return err
	}
	if len(apprls.Items) > 0 {
		logger.V(4).Info("ApplicationVersion is still in use, keep file in store")
		return nil
	}
	err = application.FailOverDelete(r.cmStore, r.ossStore, id)
	if err != nil {
		logger.Error(err, "Fail to delete application version from store")
		return err
	}
	logger.V(4).Info("Delete file from store successfully")
	return nil
}
