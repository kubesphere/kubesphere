/*
Copyright 2022 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	helmrelease "helm.sh/helm/v3/pkg/release"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/kustomize/api/types"
)

// Executor is used to manage a helm release, you can install/uninstall and upgrade a chart
// or get the status and manifest data of the release, etc.
type Executor interface {
	// Install installs the specified chart and returns the name of the Job that executed the task.
	Install(ctx context.Context, chartName string, chartData, values []byte) (string, error)
	// Upgrade upgrades the specified chart and returns the name of the Job that executed the task.
	Upgrade(ctx context.Context, chartName string, chartData, values []byte) (string, error)
	// Uninstall is used to uninstall the specified chart and returns the name of the Job that executed the task.
	Uninstall(ctx context.Context) (string, error)
	// Manifest returns the manifest data for this release.
	Manifest() (string, error)
	// IsReleaseReady checks if the helm release is ready.
	IsReleaseReady(timeout time.Duration) (bool, error)
}

const (
	workspaceBaseSource = "/tmp/helm-executor-source"
	workspaceBase       = "/tmp/helm-executor"

	statusNotFoundFormat = "release: not found"
	releaseExists        = "release exists"

	kustomizationFile  = "kustomization.yaml"
	postRenderExecFile = "helm-post-render.sh"
	// kustomize cannot read stdio now, so we save helm stdout to file, then kustomize reads that file and build the resources
	kustomizeBuild = `#!/bin/sh
# save helm stdout to file, then kustomize read this file
cat > ./.local-helm-output.yaml
kustomize build
`
)

var (
	errorTimedOutToWaitResource = errors.New("timed out waiting for resources to be ready")
)

type executor struct {
	// target cluster client
	client     kubernetes.Interface
	kubeConfig string
	namespace  string
	// helm release name
	releaseName string
	// helm action Config
	helmConf  *action.Configuration
	helmImage string
	// add labels to helm chart
	labels map[string]string
	// add annotations to helm chart
	annotations     map[string]string
	createNamespace bool
	dryRun          bool
}

type Option func(*executor)

// SetDryRun sets the dryRun option.
func SetDryRun(dryRun bool) Option {
	return func(e *executor) {
		e.dryRun = dryRun
	}
}

// SetAnnotations sets extra annotations added to all resources in chart.
func SetAnnotations(annotations map[string]string) Option {
	return func(e *executor) {
		e.annotations = annotations
	}
}

// SetLabels sets extra labels added to all resources in chart.
func SetLabels(labels map[string]string) Option {
	return func(e *executor) {
		e.labels = labels
	}
}

// SetHelmImage sets the helmImage option.
func SetHelmImage(helmImage string) Option {
	return func(e *executor) {
		e.helmImage = helmImage
	}
}

// SetKubeConfig sets the kube config data of the target cluster.
func SetKubeConfig(kubeConfig string) Option {
	return func(e *executor) {
		e.kubeConfig = kubeConfig
	}
}

// SetCreateNamespace sets the createNamespace option.
func SetCreateNamespace(createNamespace bool) Option {
	return func(e *executor) {
		e.createNamespace = createNamespace
	}
}

// NewExecutor generates a new Executor instance with the following parameters:
//   - namespace: the namespace of the helm release
//   - releaseName: the helm release name
//   - options: functions to set optional parameters
func NewExecutor(namespace, releaseName string, options ...Option) (Executor, error) {
	e := &executor{
		namespace:   namespace,
		releaseName: releaseName,
		helmImage:   "kubesphere/helm:latest",
	}
	for _, option := range options {
		option(e)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(e.kubeConfig))
	if err != nil {
		return nil, err
	}
	clusterClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	e.client = clusterClient

	klog.V(8).Infof("namespace: %s, release name: %s, kube config:%s", e.namespace, e.releaseName, e.kubeConfig)

	getter := NewClusterRESTClientGetter(e.kubeConfig, e.namespace)
	e.helmConf = new(action.Configuration)
	if err = e.helmConf.Init(getter, e.namespace, "", klog.Infof); err != nil {
		return nil, err
	}

	if e.createNamespace {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: e.namespace,
			},
		}
		if _, err = e.client.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			return e, err
		}
	}
	return e, nil
}

// Install installs the specified chart, returns the name of the Job that executed the task.
func (e *executor) Install(ctx context.Context, chartName string, chartData, values []byte) (string, error) {
	sts, err := e.status()
	if err == nil {
		// helm release has been installed
		if sts.Info != nil && sts.Info.Status == "deployed" {
			return "", nil
		}
		return "", errors.New(releaseExists)
	} else {
		if err.Error() == statusNotFoundFormat {
			// continue to install
			return e.createInstallJob(ctx, chartName, chartData, values, false)
		}
		return "", err
	}
}

// Upgrade upgrades the specified chart, returns the name of the Job that executed the task.
func (e *executor) Upgrade(ctx context.Context, chartName string, chartData, values []byte) (string, error) {
	sts, err := e.status()
	if err != nil {
		return "", err
	}

	if sts.Info.Status == "deployed" {
		return e.createInstallJob(ctx, chartName, chartData, values, true)
	}
	return "", fmt.Errorf("cannot upgrade release %s/%s, current state is %s", e.namespace, e.releaseName, sts.Info.Status)
}

func (e *executor) kubeConfigPath() string {
	if len(e.kubeConfig) == 0 {
		return ""
	}
	return "kube.config"
}

func (e *executor) chartPath(chartName string) string {
	return fmt.Sprintf("%s.tgz", chartName)
}

func (e *executor) setupChartData(chartName string, chartData, values []byte) (map[string][]byte, error) {
	if len(e.labels) == 0 && len(e.annotations) == 0 {
		return nil, nil
	}

	kustomizationConfig := types.Kustomization{
		Resources:         []string{"./.local-helm-output.yaml"},
		CommonAnnotations: e.annotations,                    // add extra annotations to output
		Labels:            []types.Label{{Pairs: e.labels}}, // Labels to add to all objects but not selectors.
	}
	kustomizationData, err := yaml.Marshal(kustomizationConfig)
	if err != nil {
		return nil, err
	}

	data := map[string][]byte{
		postRenderExecFile:     []byte(kustomizeBuild),
		kustomizationFile:      kustomizationData,
		e.chartPath(chartName): chartData,
		"values.yaml":          values,
	}
	if e.kubeConfigPath() != "" {
		data[e.kubeConfigPath()] = []byte(e.kubeConfig)
	}
	return data, nil
}

func generateName(name string) string {
	return fmt.Sprintf("helm-executor-%s-%s", name, rand.String(6))
}

func (e *executor) createConfigMap(ctx context.Context, chartName string, chartData, values []byte) (string, error) {
	data, err := e.setupChartData(chartName, chartData, values)
	if err != nil {
		return "", err
	}

	name := generateName(chartName)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: e.namespace,
		},
		// we can't use `Data` here because creating it with client-go will cause our compressed file to be in the
		// wrong format (application/octet-stream)
		BinaryData: data,
	}
	if _, err = e.client.CoreV1().ConfigMaps(e.namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
		return "", err
	}
	return name, nil
}

func (e *executor) createInstallJob(ctx context.Context, chartName string, chartData, values []byte, upgrade bool) (string, error) {
	args := make([]string, 0, 10)
	if upgrade {
		args = append(args, "upgrade")
	} else {
		args = append(args, "install")
	}

	args = append(args, "--wait", e.releaseName, e.chartPath(chartName), "--namespace", e.namespace)

	if len(values) > 0 {
		args = append(args, "--values", "values.yaml")
	}

	if e.dryRun {
		args = append(args, "--dry-run")
	}

	if e.kubeConfigPath() != "" {
		args = append(args, "--kubeconfig", e.kubeConfigPath())
	}

	// Post render, add annotations or labels to resources
	if len(e.labels) > 0 || len(e.annotations) > 0 {
		args = append(args, "--post-renderer", filepath.Join(workspaceBase, postRenderExecFile))
	}

	if klog.V(8).Enabled() {
		// output debug info
		args = append(args, "--debug")
	}

	name, err := e.createConfigMap(ctx, chartName, chartData, values)
	if err != nil {
		return "", err
	}
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: e.namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: pointer.Int32(1),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "helm",
							Image:           e.helmImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"helm"},
							Args:            args,
							WorkingDir:      workspaceBase,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "source",
									MountPath: workspaceBaseSource,
								},
								{
									Name:      "data",
									MountPath: workspaceBase,
								},
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/sh", "-c", fmt.Sprintf("cp -r %s/. %s", workspaceBaseSource, workspaceBase)},
									},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "source",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: name,
									},
									DefaultMode: pointer.Int32(0755),
								},
							},
						},
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					RestartPolicy:                 corev1.RestartPolicyNever,
					TerminationGracePeriodSeconds: new(int64),
				},
			},
		},
	}

	if _, err = e.client.BatchV1().Jobs(e.namespace).Create(ctx, job, metav1.CreateOptions{}); err != nil {
		return "", err
	}
	return name, nil
}

// Uninstall uninstalls the specified chart, returns the name of the Job that executed the task.
func (e *executor) Uninstall(ctx context.Context) (string, error) {
	if _, err := e.status(); err != nil && err.Error() == statusNotFoundFormat {
		// already uninstalled
		return "", nil
	}

	args := []string{
		"uninstall",
		e.releaseName,
		"--namespace",
		e.namespace,
	}
	if e.dryRun {
		args = append(args, "--dry-run")
	}

	name := generateName(e.releaseName)
	if e.kubeConfigPath() != "" {
		args = append(args, "--kubeconfig", e.kubeConfigPath())

		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: e.namespace,
			},
			Data: map[string]string{
				e.kubeConfigPath(): e.kubeConfig,
			},
		}
		if _, err := e.client.CoreV1().ConfigMaps(e.namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
			return "", err
		}
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: e.namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: pointer.Int32(1),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "helm",
							Image:           e.helmImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"helm"},
							Args:            args,
							WorkingDir:      workspaceBase,
						},
					},
					RestartPolicy:                 corev1.RestartPolicyNever,
					TerminationGracePeriodSeconds: new(int64),
				},
			},
		},
	}
	if e.kubeConfigPath() != "" {
		job.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: name,
						},
						DefaultMode: pointer.Int32(0755),
					},
				},
			},
		}
		job.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "data",
				MountPath: workspaceBase,
			},
		}
	}

	if _, err := e.client.BatchV1().Jobs(e.namespace).Create(ctx, job, metav1.CreateOptions{}); err != nil {
		return "", err
	}
	return name, nil
}

// Manifest returns the manifest data for this release.
func (e *executor) Manifest() (string, error) {
	get := action.NewGet(e.helmConf)
	rel, err := get.Run(e.releaseName)
	if err != nil {
		klog.Errorf("namespace: %s, name: %s, run command failed, error: %v", e.namespace, e.releaseName, err)
		return "", err
	}
	klog.V(2).Infof("namespace: %s, name: %s, run command success", e.namespace, e.releaseName)
	klog.V(8).Infof("namespace: %s, name: %s, run command success, manifest: %s", e.namespace, e.releaseName, rel.Manifest)
	return rel.Manifest, nil
}

// IsReleaseReady checks if the helm release is ready.
func (e *executor) IsReleaseReady(timeout time.Duration) (bool, error) {
	// Get the manifest to build resources
	manifest, err := e.Manifest()
	if err != nil {
		return false, err
	}
	kubeClient := e.helmConf.KubeClient
	resources, _ := kubeClient.Build(bytes.NewBufferString(manifest), true)

	err = kubeClient.Wait(resources, timeout)
	if err == nil {
		return true, nil
	}
	if err == wait.ErrWaitTimeout {
		return false, errorTimedOutToWaitResource
	}
	return false, err
}

func (e *executor) status() (*helmrelease.Release, error) {
	helmStatus := action.NewStatus(e.helmConf)
	rel, err := helmStatus.Run(e.releaseName)
	if err != nil {
		if err.Error() == statusNotFoundFormat {
			klog.V(2).Infof("namespace: %s, name: %s, run command failed, error: %v", e.namespace, e.releaseName, err)
			return nil, err
		}
		klog.Errorf("namespace: %s, name: %s, run command failed, error: %v", e.namespace, e.releaseName, err)
		return nil, err
	}

	klog.V(2).Infof("namespace: %s, name: %s, run command success", e.namespace, e.releaseName)
	klog.V(8).Infof("namespace: %s, name: %s, run command success, manifest: %s", e.namespace, e.releaseName, rel.Manifest)
	return rel, nil
}
