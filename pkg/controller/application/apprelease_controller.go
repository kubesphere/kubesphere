/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	helmrelease "helm.sh/helm/v3/pkg/release"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	appv2 "kubesphere.io/api/application/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	"kubesphere.io/api/constants"
	"kubesphere.io/utils/helm"
	"kubesphere.io/utils/s3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/controller"
	kscontroller "kubesphere.io/kubesphere/pkg/controller/options"
	"kubesphere.io/kubesphere/pkg/simple/client/application"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
)

const (
	helminstallerController = "apprelease-helminstaller"
	HelmReleaseFinalizer    = "helmrelease.application.kubesphere.io"
)

var _ controller.Controller = &AppReleaseReconciler{}
var _ reconcile.Reconciler = &AppReleaseReconciler{}

const (
	verificationAgain        = 5
	timeoutVerificationAgain = 600
	timeoutMaxRecheck        = 4
)

func (r *AppReleaseReconciler) Name() string {
	return helminstallerController
}

func (r *AppReleaseReconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (r *AppReleaseReconciler) SetupWithManager(mgr *controller.Manager) error {
	r.HelmExecutorOptions = mgr.HelmExecutorOptions
	r.Client = mgr.GetClient()
	clusterClientSet, err := clusterclient.NewClusterClientSet(mgr.GetCache())
	if err != nil {
		return fmt.Errorf("failed to create cluster client set")
	}
	r.clusterClientSet = clusterClientSet
	r.logger = ctrl.Log.WithName("controllers").WithName(helminstallerController)

	if r.HelmExecutorOptions == nil || r.HelmExecutorOptions.Image == "" {
		return fmt.Errorf("helm executor options is nil or image is empty")
	}

	r.cmStore, r.ossStore, err = application.InitStore(mgr.Options.S3Options, r.Client)
	if err != nil {
		r.logger.Error(err, "failed to init store")
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).Named(helminstallerController).
		Watches(
			&clusterv1alpha1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(r.mapper),
			builder.WithPredicates(DeletePredicate{}),
		).
		WithEventFilter(IgnoreAnnotationChangePredicate{AnnotationKey: appv2.TimeoutRecheck}).
		For(&appv2.ApplicationRelease{}).
		Named(helminstallerController).
		Complete(r)
}

func (r *AppReleaseReconciler) mapper(ctx context.Context, o client.Object) (requests []reconcile.Request) {
	cluster := o.(*clusterv1alpha1.Cluster)

	r.logger.V(4).Info("cluster has been deleted", "cluster", cluster)
	apprlsList := &appv2.ApplicationReleaseList{}
	opts := &client.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{constants.ClusterNameLabelKey: cluster.Name})}
	if err := r.List(ctx, apprlsList, opts); err != nil {
		r.logger.Error(err, "failed to list application releases")
		return requests
	}
	for _, apprls := range apprlsList.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: apprls.Name}})
	}
	return requests
}

type AppReleaseReconciler struct {
	client.Client
	clusterClientSet    clusterclient.Interface
	HelmExecutorOptions *kscontroller.HelmExecutorOptions
	ossStore            s3.Interface
	cmStore             s3.Interface
	logger              logr.Logger
}

func (r *AppReleaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	apprls := &appv2.ApplicationRelease{}
	if err := r.Client.Get(ctx, req.NamespacedName, apprls); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger := r.logger.WithValues("application release", apprls.Name).WithValues("namespace", apprls.Namespace)
	timeoutRecheck := apprls.Annotations[appv2.TimeoutRecheck]
	var reCheck int
	if timeoutRecheck == "" {
		reCheck = 0
	} else {
		reCheck, _ = strconv.Atoi(timeoutRecheck)
	}

	dstKubeConfig, runClient, err := r.getClusterInfo(apprls.GetRlsCluster())
	if err != nil {
		logger.Error(err, "failed to get cluster info")
		return ctrl.Result{}, err
	}
	executor, err := r.getExecutor(apprls, dstKubeConfig, runClient)
	if err != nil {
		logger.Error(err, "failed to get executor")
		return ctrl.Result{}, err
	}

	cluster, err := r.clusterClientSet.Get(apprls.GetRlsCluster())
	if err != nil {
		klog.Errorf("failed to get cluster: %v", err)
		return ctrl.Result{}, err
	}

	helmKubeConfig, err := application.GetHelmKubeConfig(ctx, cluster, runClient)
	if err != nil {
		logger.Error(err, "failed to get helm kubeconfig")
		return ctrl.Result{}, err
	}

	if apierrors.IsNotFound(err) || (err == nil && !cluster.DeletionTimestamp.IsZero()) {
		logger.Error(err, "cluster not found or deleting", "cluster", apprls.GetRlsCluster())
		apprls.Status.State = appv2.StatusClusterDeleted
		apprls.Status.Message = fmt.Sprintf("cluster %s has been deleted", cluster.Name)
		patch, _ := json.Marshal(apprls)
		err = r.Status().Patch(ctx, apprls, client.RawPatch(client.Merge.Type(), patch))
		if err != nil {
			logger.Error(err, "failed to update application release")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(apprls, HelmReleaseFinalizer) && apprls.ObjectMeta.DeletionTimestamp.IsZero() {
		expected := apprls.DeepCopy()
		controllerutil.AddFinalizer(expected, HelmReleaseFinalizer)
		logger.V(6).Info("add finalizer for application release")
		return ctrl.Result{}, r.Patch(ctx, expected, client.MergeFrom(apprls))
	}

	if !apprls.ObjectMeta.DeletionTimestamp.IsZero() {
		if apprls.Status.State != appv2.StatusDeleting {

			result, err := r.removeAll(ctx, apprls, executor, helmKubeConfig)
			if err != nil {
				return result, err
			}
		}
		wait, err := r.cleanJob(ctx, apprls, runClient)
		if err != nil {
			logger.Error(err, "failed to clean job")
			return ctrl.Result{}, err
		}
		if wait {
			logger.V(6).Info("job wait, job is still active")
			return ctrl.Result{RequeueAfter: verificationAgain * time.Second}, nil
		}
		logger.V(4).WithValues().Info("job has been cleaned")

		if err = r.Client.Get(ctx, req.NamespacedName, apprls); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		apprls.Finalizers = nil
		err = r.Update(ctx, apprls)
		if err != nil {
			logger.Error(err, "failed to remove finalizer for application release")
			return ctrl.Result{}, err
		}
		logger.V(6).Info("remove finalizer for application release")
		return ctrl.Result{}, nil
	}

	if apprls.Status.State == "" {
		apprls.Status.SpecHash = apprls.HashSpec()
		return ctrl.Result{}, r.updateStatus(ctx, apprls, appv2.StatusCreating)
	}

	if apprls.HashSpec() != apprls.Status.SpecHash {
		apprls.Status.SpecHash = apprls.HashSpec()
		return ctrl.Result{}, r.updateStatus(ctx, apprls, appv2.StatusUpgrading)
	}

	if apprls.Status.State == appv2.StatusCreated || apprls.Status.State == appv2.StatusUpgraded || apprls.Status.State == appv2.StatusTimeout {

		options := []helm.HelmOption{
			helm.SetNamespace(apprls.GetRlsNamespace()),
			helm.SetKubeconfig(dstKubeConfig),
		}
		release, err := executor.Get(ctx, apprls.Name, options...)
		if err != nil && err.Error() == "release: not found" {
			ct, _, err := r.checkJob(ctx, apprls, runClient, release)
			return ct, err
		}

		if err != nil {
			msg := fmt.Sprintf("%s helm create job failed err: %v", apprls.Name, err)
			err = r.updateStatus(ctx, apprls, appv2.StatusFailed, msg)
			return ctrl.Result{}, err

		}

		if apprls.Status.State == appv2.StatusUpgraded {
			ct, todo, err := r.checkJob(ctx, apprls, runClient, release)
			if err != nil {
				return ct, err
			}
			if !todo {
				return ct, nil
			}
		}

		switch release.Info.Status {
		case helmrelease.StatusFailed:
			if strings.Contains(release.Info.Description, "context deadline exceeded") && reCheck < timeoutMaxRecheck {

				if apprls.Status.State != appv2.StatusTimeout {
					err = r.updateStatus(ctx, apprls, appv2.StatusTimeout, "Installation timeout")
					if err != nil {
						logger.Error(err, "failed to update application release status")
						return ctrl.Result{}, err
					}
					logger.V(2).Info("installation timeout, will check status again after seconds", "timeout verification again", timeoutVerificationAgain)
					return ctrl.Result{RequeueAfter: timeoutVerificationAgain * time.Second}, nil
				}

				deployed, err := application.UpdateHelmStatus(dstKubeConfig, release)
				if err != nil {
					return ctrl.Result{}, err
				}

				apprls.Annotations[appv2.TimeoutRecheck] = strconv.Itoa(reCheck + 1)
				patch, _ := json.Marshal(apprls)
				err = r.Patch(ctx, apprls, client.RawPatch(client.Merge.Type(), patch))
				if err != nil {
					logger.Error(err, "failed to update application release")
					return ctrl.Result{}, err
				}
				logger.V(2).Info("update recheck times", "recheck times", strconv.Itoa(reCheck+1))

				if deployed {
					err = r.updateStatus(ctx, apprls, appv2.StatusActive, "StatusActive")
					if err != nil {
						logger.Error(err, "failed to update application release")
						return ctrl.Result{}, err
					}
					return ctrl.Result{}, nil
				}
				return ctrl.Result{RequeueAfter: timeoutVerificationAgain * time.Second}, nil
			}
			err = r.updateStatus(ctx, apprls, appv2.StatusFailed, release.Info.Description)
			return ctrl.Result{}, err
		case helmrelease.StatusDeployed:
			err = r.updateStatus(ctx, apprls, appv2.StatusActive, release.Info.Description)
			return ctrl.Result{}, err
		default:
			r.logger.V(5).Info(fmt.Sprintf("helm release %s/%s status %s, check again after %d seconds", apprls.GetRlsNamespace(), apprls.Name, release.Info.Status, verificationAgain))
			return ctrl.Result{RequeueAfter: verificationAgain * time.Second}, nil
		}
	}

	if apprls.Status.State == appv2.StatusCreating || apprls.Status.State == appv2.StatusUpgrading {

		return ctrl.Result{}, r.createOrUpgradeAppRelease(ctx, apprls, executor, helmKubeConfig)
	}

	return ctrl.Result{}, nil
}

func (r *AppReleaseReconciler) checkJob(ctx context.Context, apprls *appv2.ApplicationRelease, runClient client.Client, release *helmrelease.Release) (ct ctrl.Result, todo bool, err error) {
	logger := r.logger.WithValues("application release", apprls).WithValues("namespace", apprls.Namespace)
	logger.V(4).Info("helm release %s/%s ready to create or upgrade yet,check job %s", apprls.GetRlsNamespace(), apprls.Name, apprls.Status.InstallJobName)

	job := &batchv1.Job{}
	if err := runClient.Get(ctx, types.NamespacedName{Namespace: apprls.GetRlsNamespace(), Name: apprls.Status.InstallJobName}, job); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, "job not found", "install job", apprls.Status.InstallJobName)
			msg := "deploy failed, job not found"
			return ctrl.Result{}, false, r.updateStatus(ctx, apprls, appv2.StatusDeployFailed, msg)
		}
		return ctrl.Result{}, false, err
	}
	// ensure that the upgraded job has a successful status, otherwise mark the apprelease status as Failed so that the front-end can view the upgrade failure logs.
	if apprls.Status.State == appv2.StatusUpgraded && job.Status.Succeeded > 0 {
		return ctrl.Result{}, false, r.updateStatus(ctx, apprls, appv2.StatusActive, "Upgrade succeeful")
	}
	if job.Status.Failed > 0 {
		logger.V(2).Info(fmt.Sprintf("install job failed, failed times %d/%d", job.Status.Failed, *job.Spec.BackoffLimit+1), "job", job.Name)
	}
	if job.Spec.BackoffLimit != nil && job.Status.Failed > *job.Spec.BackoffLimit {
		// When in the upgrade state, if job execution fails while the HelmRelease status remains deployed, directly mark the AppRelease as StatusDeployFailed.
		if apprls.Status.State != appv2.StatusUpgraded || (release != nil && release.Info.Status == helmrelease.StatusDeployed) {
			msg := fmt.Sprintf("deploy failed, job %s has failed %d times ", apprls.Status.InstallJobName, job.Status.Failed)
			return ctrl.Result{}, false, r.updateStatus(ctx, apprls, appv2.StatusDeployFailed, msg)
		}
		return ctrl.Result{RequeueAfter: verificationAgain * time.Second}, true, nil
	} else {
		return ctrl.Result{RequeueAfter: verificationAgain * time.Second}, false, nil
	}
}

func (r *AppReleaseReconciler) removeAll(ctx context.Context, apprls *appv2.ApplicationRelease, executor helm.Executor, kubeconfig []byte) (ct ctrl.Result, err error) {
	logger := r.logger.WithValues("application release", apprls).WithValues("namespace", apprls.Namespace)
	err = r.updateStatus(ctx, apprls, appv2.StatusDeleting, "Uninstalling")
	if err != nil {
		logger.Error(err, "failed to update application release status")
		return ctrl.Result{}, err
	}

	uninstallJobName, err := r.uninstall(ctx, apprls, executor, kubeconfig)
	if err != nil {
		logger.Error(err, "failed to uninstall application release")
		return ctrl.Result{}, err
	}

	err = r.cleanStore(ctx, apprls)
	if err != nil {
		logger.Error(err, "failed to clean store")
		return ctrl.Result{}, err
	}
	logger.V(4).Info("remove application release success")

	if uninstallJobName != "" {
		logger.V(4).Info("try to update application release uninstall job", "job", uninstallJobName)
		apprls.Status.UninstallJobName = uninstallJobName
		apprls.Status.LastUpdate = metav1.Now()
		patch, _ := json.Marshal(apprls)
		err = r.Status().Patch(ctx, apprls, client.RawPatch(client.Merge.Type(), patch))
		if err != nil {
			logger.Error(err, "failed to update application release")
			return ctrl.Result{}, err
		}
		logger.V(4).Info("update application release uninstall job success", "job", uninstallJobName)
	}

	return ctrl.Result{}, nil
}

func (r *AppReleaseReconciler) getClusterDynamicClient(clusterName string, apprls *appv2.ApplicationRelease) (*dynamic.DynamicClient, error) {
	logger := r.logger.WithValues("application release", apprls).WithValues("namespace", apprls.Namespace)
	clusterClient, err := r.clusterClientSet.GetClusterClient(clusterName)
	if err != nil {
		logger.Error(err, "failed to get cluster client", "cluster", clusterName)
		return nil, err
	}
	creator := apprls.Annotations[constants.CreatorAnnotationKey]
	conf := *clusterClient.RestConfig
	if creator != "" {
		conf.Impersonate = rest.ImpersonationConfig{
			UserName: creator,
		}
	}
	logger.V(4).Info("DynamicClient impersonate kubeAsUser", "creator", creator)
	dynamicClient, err := dynamic.NewForConfig(&conf)
	return dynamicClient, err
}

func (r *AppReleaseReconciler) getClusterInfo(clusterName string) ([]byte, client.Client, error) {
	cluster, err := r.clusterClientSet.Get(clusterName)
	if err != nil {
		return nil, nil, err
	}
	runtimeClient, err := r.clusterClientSet.GetRuntimeClient(clusterName)
	if err != nil {
		return nil, nil, err
	}

	return cluster.Spec.Connection.KubeConfig, runtimeClient, nil

}

func (r *AppReleaseReconciler) updateStatus(ctx context.Context, apprls *appv2.ApplicationRelease, status string, message ...string) error {
	apprls.Status.State = status
	if message != nil {
		apprls.Status.Message = message[0]
	}
	apprls.Status.LastUpdate = metav1.Now()
	patch, _ := json.Marshal(apprls)
	return r.Status().Patch(ctx, apprls, client.RawPatch(client.Merge.Type(), patch))
}
