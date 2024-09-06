/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package application

import (
	"context"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/klog/v2"
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
	appVersionController = "appversion-controller"
)

var _ reconcile.Reconciler = &AppVersionReconciler{}
var _ kscontroller.Controller = &AppVersionReconciler{}

type AppVersionReconciler struct {
	client.Client
	ossStore s3.Interface
	cmStore  s3.Interface
}

func (r *AppVersionReconciler) Name() string {
	return appVersionController
}

func (r *AppVersionReconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (r *AppVersionReconciler) SetupWithManager(mgr *kscontroller.Manager) (err error) {
	r.Client = mgr.GetClient()
	r.cmStore, r.ossStore, err = application.InitStore(mgr.Options.S3Options, r.Client)
	if err != nil {
		klog.Errorf("failed to init store: %v", err)
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

	//Delete app files, non-important logic, errors will not affect the main process
	if !appVersion.ObjectMeta.DeletionTimestamp.IsZero() {
		err := r.deleteFile(ctx, appVersion)
		if err != nil {
			klog.Errorf("Failed to clean file for appversion %s: %v", appVersion.Name, err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *AppVersionReconciler) deleteFile(ctx context.Context, appVersion *appv2.ApplicationVersion) error {
	defer func() {
		controllerutil.RemoveFinalizer(appVersion, appv2.StoreCleanFinalizer)
		err := r.Update(ctx, appVersion)
		if err != nil {
			klog.Errorf("Failed to remove finalizer from appversion %s: %v", appVersion.Name, err)
		}
		klog.Infof("Remove finalizer from appversion %s successfully", appVersion.Name)
	}()

	klog.Infof("ApplicationVersion  %s has been deleted, try to clean file", appVersion.Name)
	id := []string{appVersion.Name}
	apprls := &appv2.ApplicationReleaseList{}
	err := r.Client.List(ctx, apprls, client.MatchingLabels{appv2.AppVersionIDLabelKey: appVersion.Name})
	if err != nil {
		klog.Errorf("Failed to list ApplicationRelease: %v", err)
		return err
	}
	if len(apprls.Items) > 0 {
		klog.Infof("ApplicationVersion %s is still in use, keep file in store", appVersion.Name)
		return nil
	}
	err = application.FailOverDelete(r.cmStore, r.ossStore, id)
	if err != nil {
		klog.Errorf("Fail to delete appversion %s from store: %v", appVersion.Name, err)
		return err
	}
	klog.Infof("Delete file %s from store successfully", appVersion.Name)
	return nil
}
