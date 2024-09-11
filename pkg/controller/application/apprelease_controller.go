/*
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

	"k8s.io/client-go/rest"

	batchv1 "k8s.io/api/batch/v1"

	"kubesphere.io/utils/helm"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"kubesphere.io/api/constants"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"kubesphere.io/utils/s3"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	"kubesphere.io/kubesphere/pkg/controller"
	kscontroller "kubesphere.io/kubesphere/pkg/controller/options"

	helmrelease "helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	appv2 "kubesphere.io/api/application/v2"

	"kubesphere.io/kubesphere/pkg/simple/client/application"

	"kubesphere.io/kubesphere/pkg/utils/clusterclient"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
		return fmt.Errorf("failed to create cluster client set: %v", err)
	}
	r.clusterClientSet = clusterClientSet

	if r.HelmExecutorOptions == nil || r.HelmExecutorOptions.Image == "" {
		return fmt.Errorf("helm executor options is nil or image is empty")
	}

	r.cmStore, r.ossStore, err = application.InitStore(mgr.Options.S3Options, r.Client)
	if err != nil {
		klog.Errorf("failed to init store: %v", err)
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).Named(helminstallerController).
		Watches(
			&clusterv1alpha1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(r.mapper),
			builder.WithPredicates(ClusterDeletePredicate{}),
		).
		WithEventFilter(IgnoreAnnotationChangePredicate{AnnotationKey: appv2.TimeoutRecheck}).
		For(&appv2.ApplicationRelease{}).Complete(r)
}

func (r *AppReleaseReconciler) mapper(ctx context.Context, o client.Object) (requests []reconcile.Request) {
	cluster := o.(*clusterv1alpha1.Cluster)

	klog.Infof("cluster %s has been deleted", cluster.Name)
	apprlsList := &appv2.ApplicationReleaseList{}
	opts := &client.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{constants.ClusterNameLabelKey: cluster.Name})}
	if err := r.List(ctx, apprlsList, opts); err != nil {
		klog.Errorf("failed to list application releases: %v", err)
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
}

func (r *AppReleaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	apprls := &appv2.ApplicationRelease{}
	if err := r.Client.Get(ctx, req.NamespacedName, apprls); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	timeoutRecheck := apprls.Annotations[appv2.TimeoutRecheck]
	var reCheck int
	if timeoutRecheck == "" {
		reCheck = 0
	} else {
		reCheck, _ = strconv.Atoi(timeoutRecheck)
	}

	dstKubeConfig, runClient, err := r.getClusterInfo(apprls.GetRlsCluster())
	if err != nil {
		klog.Errorf("failed to get cluster info: %v", err)
		return ctrl.Result{}, err
	}
	executor, err := r.getExecutor(apprls, dstKubeConfig, runClient)
	if err != nil {
		klog.Errorf("failed to get executor: %v", err)
		return ctrl.Result{}, err
	}

	cluster, err := r.clusterClientSet.Get(apprls.GetRlsCluster())
	if err != nil {
		klog.Errorf("failed to get cluster: %v", err)
		return ctrl.Result{}, err
	}

	helmKubeConfig, err := application.GetHelmKubeConfig(ctx, cluster, runClient)
	if err != nil {
		klog.Errorf("failed to get helm kubeconfig: %v", err)
		return ctrl.Result{}, err
	}

	if apierrors.IsNotFound(err) || (err == nil && !cluster.DeletionTimestamp.IsZero()) {
		klog.Errorf("cluster not found or deleting %s: %v", apprls.GetRlsCluster(), err)
		apprls.Status.State = appv2.StatusClusterDeleted
		apprls.Status.Message = fmt.Sprintf("cluster %s has been deleted", cluster.Name)
		patch, _ := json.Marshal(apprls)
		err = r.Status().Patch(ctx, apprls, client.RawPatch(client.Merge.Type(), patch))
		if err != nil {
			klog.Errorf("failed to update apprelease %s: %v", apprls.Name, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(apprls, HelmReleaseFinalizer) && apprls.ObjectMeta.DeletionTimestamp.IsZero() {
		expected := apprls.DeepCopy()
		controllerutil.AddFinalizer(expected, HelmReleaseFinalizer)
		klog.Infof("add finalizer for apprelease %s", apprls.Name)
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
			klog.Errorf("failed to clean job: %v", err)
			return ctrl.Result{}, err
		}
		if wait {
			klog.Infof("job wait, job for  %s is still active", apprls.Name)
			return ctrl.Result{RequeueAfter: verificationAgain * time.Second}, nil
		}
		klog.Infof("job for %s has been cleaned", apprls.Name)

		if err = r.Client.Get(ctx, req.NamespacedName, apprls); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		apprls.Finalizers = nil
		err = r.Update(ctx, apprls)
		if err != nil {
			klog.Errorf("failed to remove finalizer for apprelease %s: %v", apprls.Name, err)
			return ctrl.Result{}, err
		}
		klog.Infof("remove finalizer for apprelease %s", apprls.Name)
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

	if apprls.Status.State == appv2.StatusCreated || apprls.Status.State == appv2.StatusTimeout {

		options := []helm.HelmOption{
			helm.SetNamespace(apprls.GetRlsNamespace()),
			helm.SetKubeconfig(dstKubeConfig),
		}
		release, err := executor.Get(ctx, apprls.Name, options...)
		if err != nil && err.Error() == "release: not found" {
			klog.Infof("helm release %s/%s not created yet,check job %s", apprls.GetRlsNamespace(), apprls.Name, apprls.Status.InstallJobName)

			job := &batchv1.Job{}
			if err = runClient.Get(ctx, types.NamespacedName{Namespace: apprls.GetRlsNamespace(), Name: apprls.Status.InstallJobName}, job); err != nil {
				if apierrors.IsNotFound(err) {
					klog.Errorf("job %s not found", apprls.Status.InstallJobName)
					msg := "deploy failed, job not found"
					return ctrl.Result{}, r.updateStatus(ctx, apprls, appv2.StatusDeployFailed, msg)
				}
				return ctrl.Result{}, err
			}
			if job.Status.Failed > 0 {
				klog.Infof("install apprls %s job %s , failed times %d/%d", apprls.Name, job.Name, job.Status.Failed, *job.Spec.BackoffLimit+1)
			}
			if job.Spec.BackoffLimit != nil && job.Status.Failed > *job.Spec.BackoffLimit {
				msg := fmt.Sprintf("deploy failed, job %s has failed %d times ", apprls.Status.InstallJobName, job.Status.Failed)
				return ctrl.Result{}, r.updateStatus(ctx, apprls, appv2.StatusDeployFailed, msg)
			}
			return ctrl.Result{RequeueAfter: verificationAgain * time.Second}, nil
		}

		if err != nil {
			msg := fmt.Sprintf("%s helm create job failed err: %v", apprls.Name, err)
			err = r.updateStatus(ctx, apprls, appv2.StatusFailed, msg)
			return ctrl.Result{}, err

		}

		switch release.Info.Status {
		case helmrelease.StatusFailed:
			if strings.Contains(release.Info.Description, "context deadline exceeded") && reCheck < timeoutMaxRecheck {

				if apprls.Status.State != appv2.StatusTimeout {
					err = r.updateStatus(ctx, apprls, appv2.StatusTimeout, "Installation timeout")
					if err != nil {
						klog.Errorf("failed to update apprelease %s status : %v", apprls.Name, err)
						return ctrl.Result{}, err
					}
					klog.Infof("Installation timeout, will check status again after %d second", timeoutVerificationAgain)
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
					klog.Errorf("failed to update apprelease %s: %v", apprls.Name, err)
					return ctrl.Result{}, err
				}
				klog.Infof("update recheck times %s for %s", strconv.Itoa(reCheck+1), apprls.Name)

				if deployed {
					err = r.updateStatus(ctx, apprls, appv2.StatusActive, "StatusActive")
					if err != nil {
						klog.Errorf("failed to update apprelease %s  %v", apprls.Name, err)
						return ctrl.Result{}, err
					}
					return ctrl.Result{}, nil
				}
				return ctrl.Result{RequeueAfter: timeoutVerificationAgain * time.Second}, nil
			}
			err = r.updateStatus(ctx, apprls, appv2.StatusFailed, release.Info.Description)
			return ctrl.Result{}, err
		case helmrelease.StatusDeployed:
			err = r.updateStatus(ctx, apprls, appv2.StatusActive)
			return ctrl.Result{}, err
		default:
			klog.V(5).Infof("helm release %s/%s status %s, check again after %d second", apprls.GetRlsNamespace(), apprls.Name, release.Info.Status, verificationAgain)
			return ctrl.Result{RequeueAfter: verificationAgain * time.Second}, nil
		}
	}

	if apprls.Status.State == appv2.StatusCreating || apprls.Status.State == appv2.StatusUpgrading {

		return ctrl.Result{}, r.createOrUpgradeAppRelease(ctx, apprls, executor, helmKubeConfig)
	}

	return ctrl.Result{}, nil
}

func (r *AppReleaseReconciler) removeAll(ctx context.Context, apprls *appv2.ApplicationRelease, executor helm.Executor, kubeconfig []byte) (ct ctrl.Result, err error) {
	err = r.updateStatus(ctx, apprls, appv2.StatusDeleting)
	if err != nil {
		klog.Errorf("failed to update apprelease %s status : %v", apprls.Name, err)
		return ctrl.Result{}, err
	}

	uninstallJobName, err := r.uninstall(ctx, apprls, executor, kubeconfig)
	if err != nil {
		klog.Errorf("failed to uninstall helm release %s: %v", apprls.Name, err)
		return ctrl.Result{}, err
	}

	err = r.cleanStore(ctx, apprls)
	if err != nil {
		klog.Errorf("failed to clean store: %v", err)
		return ctrl.Result{}, err
	}
	klog.Infof("remove apprelease %s success", apprls.Name)

	if uninstallJobName != "" {
		klog.Infof("try to update uninstall apprls job name %s to apprelease %s", uninstallJobName, apprls.Name)
		apprls.Status.UninstallJobName = uninstallJobName
		apprls.Status.LastUpdate = metav1.Now()
		patch, _ := json.Marshal(apprls)
		err = r.Status().Patch(ctx, apprls, client.RawPatch(client.Merge.Type(), patch))
		if err != nil {
			klog.Errorf("failed to update apprelease %s: %v", apprls.Name, err)
			return ctrl.Result{}, err
		}
		klog.Infof("update uninstall apprls job name %s to apprelease %s success", uninstallJobName, apprls.Name)
	}

	return ctrl.Result{}, nil
}

func (r *AppReleaseReconciler) getClusterDynamicClient(clusterName string, apprls *appv2.ApplicationRelease) (*dynamic.DynamicClient, error) {
	clusterClient, err := r.clusterClientSet.GetClusterClient(clusterName)
	if err != nil {
		klog.Errorf("failed to get cluster client: %v", err)
		return nil, err
	}
	creator := apprls.Annotations[constants.CreatorAnnotationKey]
	conf := *clusterClient.RestConfig
	if creator != "" {
		conf.Impersonate = rest.ImpersonationConfig{
			UserName: creator,
		}
	}
	klog.Infof("DynamicClient impersonate kubeAsUser: %s", creator)
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
