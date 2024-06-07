/*
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
	"k8s.io/klog/v2"
	appv2 "kubesphere.io/api/application/v2"
	"kubesphere.io/utils/helm"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/application"
)

func (r *AppReleaseReconciler) uninstall(ctx context.Context, rls *appv2.ApplicationRelease, executor helm.Executor, kubeconfig []byte) (jobName string, err error) {

	klog.Infof("uninstall helm release %s", rls.Name)

	creator := rls.Annotations[constants.CreatorAnnotationKey]
	klog.Infof("helm impersonate kubeAsUser: %s", creator)
	options := []helm.HelmOption{
		helm.SetNamespace(rls.GetRlsNamespace()),
		helm.SetKubeconfig(kubeconfig),
	}

	if jobName, err = executor.Uninstall(ctx, rls.Name, options...); err != nil {
		klog.Error(err, "failed to force delete helm release")
		return jobName, err
	}
	klog.Infof("uninstall helm release %s success,job name: %s", rls.Name, jobName)

	return jobName, nil
}

func (r *AppReleaseReconciler) jobStatus(job *batchv1.Job) (active, completed, failed bool) {
	active = job.Status.Active > 0
	completed = (job.Spec.Completions != nil && job.Status.Succeeded >= *job.Spec.Completions) || job.Status.Succeeded > 0
	failed = (job.Spec.BackoffLimit != nil && job.Status.Failed > *job.Spec.BackoffLimit) || job.Status.Failed > 0
	return
}

func (r *AppReleaseReconciler) createOrUpgradeAppRelease(ctx context.Context, rls *appv2.ApplicationRelease, executor helm.Executor, kubeconfig []byte) error {
	clusterName := rls.GetRlsCluster()
	namespace := rls.GetRlsNamespace()
	klog.Infof("begin to create or upgrade %s app release %s in cluster %s ns: %s", rls.Spec.AppType, rls.Name, clusterName, namespace)

	creator := rls.Annotations[constants.CreatorAnnotationKey]
	klog.Infof("helm impersonate kubeAsUser: %s", creator)
	options := []helm.HelmOption{
		helm.SetInstall(true),
		helm.SetNamespace(namespace),
		helm.SetKubeAsUser(creator),
		helm.SetKubeconfig(kubeconfig),
	}

	if rls.Spec.AppType == appv2.AppTypeHelm {
		_, err := executor.Get(ctx, rls.Name, options...)
		if err != nil && err.Error() == "release: not found" {
			klog.Infof("release %s not found, begin to create", rls.Name)
		}
		if err == nil {
			klog.Infof("release %s found, begin to upgrade status", rls.Name)
			return r.updateStatus(ctx, rls, appv2.StatusCreated)
		}
	}

	data, err := application.FailOverGet(r.cmStore, r.ossStore, rls.Spec.AppVersionID, r.Client, true)
	if err != nil {
		klog.Errorf("failed to get app version data, err: %v", err)
		return err
	}
	options = append(options, helm.SetChartData(data))

	if rls.Status.InstallJobName, err = executor.Upgrade(ctx, rls.Name, "", rls.Spec.Values, options...); err != nil {
		klog.Errorf("failed to create executor job, err: %v", err)
		return r.updateStatus(ctx, rls, appv2.StatusFailed, err.Error())
	}

	return r.updateStatus(ctx, rls, appv2.StatusCreated)
}

func (r *AppReleaseReconciler) getExecutor(apprls *appv2.ApplicationRelease, kubeConfig []byte, runClient client.Client) (executor helm.Executor, err error) {

	if apprls.Spec.AppType == appv2.AppTypeHelm {
		return r.getHelmExecutor(apprls, kubeConfig)
	}

	return r.getYamlInstaller(runClient, apprls)
}

func (r *AppReleaseReconciler) getYamlInstaller(runClient client.Client, apprls *appv2.ApplicationRelease) (executor helm.Executor, err error) {
	dynamicClient, err := r.getClusterDynamicClient(apprls.GetRlsCluster(), apprls)
	if err != nil {
		klog.Errorf("failed to get dynamic client: %v", err)
		return nil, err
	}

	jsonList, err := application.ReadYaml(apprls.Spec.Values)
	if err != nil {
		klog.Errorf("failed to read yaml: %v", err)
		return nil, err
	}
	var gvrListInfo []application.InsInfo
	for _, i := range jsonList {
		gvr, utd, err := application.GetInfoFromBytes(i, runClient.RESTMapper())
		if err != nil {
			klog.Errorf("failed to get info from bytes: %v", err)
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
		klog.Errorf("failed to create helm executor: %v", err)
		return nil, err
	}

	return executor, err
}

func (r *AppReleaseReconciler) cleanJob(ctx context.Context, apprls *appv2.ApplicationRelease, runClient client.Client) (wait bool, err error) {

	jobs := &batchv1.JobList{}

	opts := []client.ListOption{client.InNamespace(apprls.GetRlsNamespace()), client.MatchingLabels{appv2.AppReleaseReferenceLabelKey: apprls.Name}}
	err = runClient.List(ctx, jobs, opts...)
	if err != nil {
		klog.Errorf("failed to list job for %s: %v", apprls.Name, err)
		return false, err
	}
	if len(jobs.Items) == 0 {
		klog.Infof("cluster: %s namespace: %s no job found for %s", apprls.GetRlsCluster(), apprls.GetRlsNamespace(), apprls.Name)
		return false, nil
	}
	klog.Infof("found %d jobs for %s", len(jobs.Items), apprls.Name)
	for _, job := range jobs.Items {
		klog.Infof("begin to clean job %s/%s", job.Namespace, job.Name)
		jobActive, jobCompleted, failed := r.jobStatus(&job)
		if jobActive {
			klog.Infof("job %s is still active", job.Name)
			return true, nil
		}
		if jobCompleted || failed {
			deletePolicy := metav1.DeletePropagationBackground
			opt := client.DeleteOptions{PropagationPolicy: &deletePolicy}
			err = runClient.Delete(ctx, &job, &opt)
			if err != nil {
				klog.Errorf("failed to delete job %s: %v", job.Name, err)
				return false, err
			}
			klog.Infof("job %s has been deleted", job.Name)
		} else {
			klog.Infof("job:%s status unknown, wait for next reconcile: %v", job.Name, job.Status)
			return true, nil
		}

	}

	klog.Infof("all job has been deleted")
	return false, nil
}

func (r *AppReleaseReconciler) cleanStore(ctx context.Context, apprls *appv2.ApplicationRelease) (err error) {
	name := apprls.Labels[appv2.AppVersionIDLabelKey]
	appVersion := &appv2.ApplicationVersion{}
	err = r.Get(ctx, client.ObjectKey{Name: name}, appVersion)
	if apierrors.IsNotFound(err) {
		klog.Infof("appVersion %s has been deleted, cleanup file in oss", name)
		err = application.FailOverDelete(r.cmStore, r.ossStore, []string{appVersion.Name})
		if err != nil {
			klog.Warningf("failed to cleanup file in oss: %v", err)
			return nil
		}
	}
	klog.Infof("appVersion %s still exists, no need to cleanup file in oss", name)
	return nil
}
