/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package application

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	appv2 "kubesphere.io/api/application/v2"
	"kubesphere.io/utils/helm"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/application"
)

func (r *AppReleaseReconciler) uninstall(ctx context.Context, rls *appv2.ApplicationRelease, executor helm.Executor, kubeconfig []byte) (jobName string, err error) {
	logger := r.logger.WithValues("application release", rls).WithValues("namespace", rls.Namespace)
	logger.V(4).Info("uninstall helm release")

	creator := rls.Annotations[constants.CreatorAnnotationKey]
	logger.V(4).Info("helm impersonate kubeAsUser", "creator", creator)
	options := []helm.HelmOption{
		helm.SetNamespace(rls.GetRlsNamespace()),
		helm.SetKubeconfig(kubeconfig),
	}

	if jobName, err = executor.Uninstall(ctx, rls.Name, options...); err != nil {
		logger.Error(err, "failed to force delete helm release")
		return jobName, err
	}
	logger.Info("uninstall helm release success", "job", jobName)

	return jobName, nil
}

func (r *AppReleaseReconciler) jobStatus(job *batchv1.Job) (active, completed, failed bool) {
	active = job.Status.Active > 0
	completed = (job.Spec.Completions != nil && job.Status.Succeeded >= *job.Spec.Completions) || job.Status.Succeeded > 0
	failed = (job.Spec.BackoffLimit != nil && job.Status.Failed > *job.Spec.BackoffLimit) || job.Status.Failed > 0
	return
}

func (r *AppReleaseReconciler) createOrUpgradeAppRelease(ctx context.Context, rls *appv2.ApplicationRelease, executor helm.Executor, kubeconfig []byte) error {
	logger := r.logger.WithValues("application release", rls).WithValues("namespace", rls.Namespace)
	clusterName := rls.GetRlsCluster()
	namespace := rls.GetRlsNamespace()
	logger.V(6).Info("begin to create or upgrade app release", "cluster", clusterName)

	creator := rls.Annotations[constants.CreatorAnnotationKey]
	logger.V(6).Info("helm impersonate kubeAsUser", "creator", creator)
	options := []helm.HelmOption{
		helm.SetInstall(true),
		helm.SetNamespace(namespace),
		helm.SetKubeAsUser(creator),
		helm.SetKubeconfig(kubeconfig),
	}

	state := appv2.StatusCreated
	if rls.Spec.AppType == appv2.AppTypeHelm {
		_, err := executor.Get(ctx, rls.Name, options...)
		if err != nil && err.Error() == "release: not found" {
			logger.V(4).Info("release not found, begin to create")
		}
		if err == nil {
			logger.V(6).Info("release  found, begin to upgrade")
			state = appv2.StatusUpgraded
		}
	}

	data, err := application.FailOverGet(r.cmStore, r.ossStore, rls.Spec.AppVersionID, r.Client, true)
	if err != nil {
		logger.Error(err, "failed to get app version data")
		return err
	}
	options = append(options, helm.SetChartData(data))

	if rls.Status.InstallJobName, err = executor.Upgrade(ctx, rls.Name, "", rls.Spec.Values, options...); err != nil {
		logger.Error(err, "failed to create executor job")
		return r.updateStatus(ctx, rls, appv2.StatusFailed, err.Error())
	}

	return r.updateStatus(ctx, rls, state, "Deploying")
}

func (r *AppReleaseReconciler) getExecutor(apprls *appv2.ApplicationRelease, kubeConfig []byte, runClient client.Client) (executor helm.Executor, err error) {

	if apprls.Spec.AppType == appv2.AppTypeHelm {
		return r.getHelmExecutor(apprls, kubeConfig)
	}

	return r.getYamlInstaller(runClient, apprls)
}

func (r *AppReleaseReconciler) getYamlInstaller(runClient client.Client, apprls *appv2.ApplicationRelease) (executor helm.Executor, err error) {
	logger := r.logger.WithValues("application release", apprls).WithValues("namespace", apprls.Namespace)
	dynamicClient, err := r.getClusterDynamicClient(apprls.GetRlsCluster(), apprls)
	if err != nil {
		logger.Error(err, "failed to get dynamic client")
		return nil, err
	}

	jsonList, err := application.ReadYaml(apprls.Spec.Values)
	if err != nil {
		logger.Error(err, "failed to read yaml")
		return nil, err
	}
	var gvrListInfo []application.InsInfo
	for _, i := range jsonList {
		gvr, utd, err := application.GetInfoFromBytes(i, runClient.RESTMapper())
		if err != nil {
			logger.Error(err, "failed to get info from bytes")
			return nil, err
		}
		ins := application.InsInfo{
			GroupVersionResource: gvr,
			Name:                 utd.GetName(),
			Namespace:            utd.GetNamespace(),
		}
		gvrListInfo = append(gvrListInfo, ins)
	}

	return application.YamlInstaller{
		Mapper:      runClient.RESTMapper(),
		DynamicCli:  dynamicClient,
		GvrListInfo: gvrListInfo,
		Namespace:   apprls.GetRlsNamespace(),
	}, nil
}

func (r *AppReleaseReconciler) getHelmExecutor(apprls *appv2.ApplicationRelease, kubeconfig []byte) (executor helm.Executor, err error) {
	logger := r.logger.WithValues("application release", apprls).WithValues("namespace", apprls.Namespace)
	executorOptions := []helm.ExecutorOption{
		helm.SetExecutorKubeConfig(kubeconfig),
		helm.SetExecutorNamespace(apprls.GetRlsNamespace()),
		helm.SetExecutorImage(r.HelmExecutorOptions.Image),
		helm.SetExecutorBackoffLimit(0),
		helm.SetExecutorLabels(labels.Set{
			appv2.AppReleaseReferenceLabelKey: apprls.Name,
			constants.KubeSphereManagedLabel:  "true",
		}),
		helm.SetTTLSecondsAfterFinished(r.HelmExecutorOptions.JobTTLAfterFinished),
	}

	executor, err = helm.NewExecutor(executorOptions...)
	if err != nil {
		logger.Error(err, "failed to create helm executor")
		return nil, err
	}

	return executor, err
}

func (r *AppReleaseReconciler) cleanJob(ctx context.Context, apprls *appv2.ApplicationRelease, runClient client.Client) (wait bool, err error) {
	logger := r.logger.WithValues("application release", apprls).WithValues("namespace", apprls.Namespace)
	jobs := &batchv1.JobList{}

	opts := []client.ListOption{client.InNamespace(apprls.GetRlsNamespace()), client.MatchingLabels{appv2.AppReleaseReferenceLabelKey: apprls.Name}}
	err = runClient.List(ctx, jobs, opts...)
	if err != nil {
		logger.Error(err, "failed to list job")
		return false, err
	}
	if len(jobs.Items) == 0 {
		logger.V(6).Info("no job found", "cluster", apprls.GetRlsCluster())
		return false, nil
	}
	logger.V(6).Info("found jobs", "job number", len(jobs.Items))
	for _, job := range jobs.Items {
		logger.V(6).Info("begin to clean job", "namespace", job.Namespace, "job", job.Name)
		jobActive, jobCompleted, failed := r.jobStatus(&job)
		if jobActive {
			logger.V(6).Info("job is still active", "job", job.Name)
			return true, nil
		}
		if jobCompleted || failed {
			deletePolicy := metav1.DeletePropagationBackground
			opt := client.DeleteOptions{PropagationPolicy: &deletePolicy}
			err = runClient.Delete(ctx, &job, &opt)
			if err != nil {
				logger.Error(err, "failed to delete job", "job", job.Name)
				return false, err
			}
			logger.V(4).Info("job has been deleted", "job", job.Name)
		} else {
			logger.V(4).Info("job status unknown, wait for next reconcile", "job", job.Name, "status", job.Status)
			return true, nil
		}

	}

	logger.Info("all job has been deleted")
	return false, nil
}

func (r *AppReleaseReconciler) cleanStore(ctx context.Context, apprls *appv2.ApplicationRelease) (err error) {
	name := apprls.Labels[appv2.AppVersionIDLabelKey]
	appVersion := &appv2.ApplicationVersion{}
	err = r.Get(ctx, client.ObjectKey{Name: name}, appVersion)
	if apierrors.IsNotFound(err) {
		r.logger.Info("application version has been deleted, cleanup file in oss", "application version", name)
		err = application.FailOverDelete(r.cmStore, r.ossStore, []string{appVersion.Name})
		if err != nil {
			r.logger.Error(err, "failed to cleanup file in oss")
			return nil
		}
	}
	r.logger.V(6).Info("application version still exists, no need to cleanup file in oss", "application version", name)
	return nil
}
