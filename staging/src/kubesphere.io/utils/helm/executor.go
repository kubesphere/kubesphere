package helm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"path/filepath"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/action"
	helmrelease "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

// Executor is used to manage a helm release, you can install/uninstall and upgrade a chart
// or get the status and manifest data of the release, etc.
type Executor interface {
	// Install installs the specified chart and returns the name of the Job that executed the task.
	Install(ctx context.Context, release, chart string, values []byte, options ...HelmOption) (string, error)

	// Upgrade upgrades the specified chart and returns the name of the Job that executed the task.
	Upgrade(ctx context.Context, release, chart string, values []byte, options ...HelmOption) (string, error)

	// helm uninstall RELEASE_NAME --cascade=orphan --no-hooks=true
	ForceDelete(ctx context.Context, release string, options ...HelmOption) error

	// helm uninstall RELEASE_NAME [flags]
	Uninstall(ctx context.Context, release string, options ...HelmOption) (string, error)

	WaitingForResourcesReady(ctx context.Context, release string, timeout time.Duration, options ...HelmOption) (bool, error)

	// helm get all RELEASE_NAME [flags]
	Get(ctx context.Context, releaseName string, options ...HelmOption) (*helmrelease.Release, error)
}

const (
	workspaceBaseSource = "/tmp/helm-executor-source"
	workspaceBase       = "/tmp/helm-executor"

	kustomizationFile  = "kustomization.yaml"
	postRenderExecFile = "helm-post-render.sh"
	// kustomize cannot read stdio now, so we save helm stdout to file, then kustomize reads that file and build the resources
	kustomizeBuild = `#!/bin/sh
# save helm stdout to file, then kustomize read this file
cat > ./.local-helm-output.yaml
kustomize build
`
	kubeConfigPath = "kube.config"

	caFilePath = "ca-helm.crt"

	DefaultKubectlImage = "kubesphere/helm:v3.12.1"
	MinimumTimeout      = 5 * time.Minute

	ExecutorJobActionAnnotation  = "executor.kubesphere.io/action"
	ExecutorConfigHashAnnotation = "executor.kubesphere.io/config-hash"

	ActionInstall   = "install"
	ActionUpgrade   = "upgrade"
	ActionUninstall = "uninstall"

	HookEnvAction      = "HOOK_ACTION"
	HookEnvClusterRole = "CLUSTER_ROLE"
	HookEnvClusterName = "CLUSTER_NAME"
)

var (
	errorTimedOutToWaitResource = errors.New("timed out waiting for resources to be ready")
)

type executor struct {
	// target cluster client
	client    kubernetes.Interface
	helmImage string
	resources corev1.ResourceRequirements
	labels    map[string]string
	owner     *metav1.OwnerReference

	kubeConfig              []byte
	namespace               string
	backoffLimit            int32
	ttlSecondsAfterFinished int32
}

type ExecutorOption func(*executor)

// SetExecutorImage sets the helmImage option.
func SetExecutorImage(helmImage string) ExecutorOption {
	return func(e *executor) {
		e.helmImage = helmImage
	}
}

func SetExecutorResources(resources corev1.ResourceRequirements) ExecutorOption {
	return func(e *executor) {
		e.resources = resources
	}
}

func SetExecutorLabels(labels map[string]string) ExecutorOption {
	return func(o *executor) {
		o.labels = labels
	}
}

func SetExecutorOwner(owner *metav1.OwnerReference) ExecutorOption {
	return func(o *executor) {
		o.owner = owner
	}
}

func SetExecutorNamespace(namespace string) ExecutorOption {
	return func(o *executor) {
		o.namespace = namespace
	}
}

func SetExecutorKubeConfig(kubeConfig []byte) ExecutorOption {
	return func(o *executor) {
		o.kubeConfig = kubeConfig
	}
}

func SetExecutorBackoffLimit(BackoffLimit int32) ExecutorOption {
	return func(o *executor) {
		o.backoffLimit = BackoffLimit
	}
}

func SetTTLSecondsAfterFinished(t time.Duration) ExecutorOption {
	return func(o *executor) {
		o.ttlSecondsAfterFinished = int32(t.Seconds())
	}
}

// NewExecutor generates a new Executor instance with the following parameters:
//   - kubeConfig: this kube config is used to create the necessary Namespace, ConfigMap and Job to perform
//     the installation tasks during the installation. You only need to give the kube config the necessary permissions,
//     if the kube config is not set, we will use the in-cluster config, this may not work.
//   - namespace: the namespace of the helm release
//   - releaseName: the helm release name
//   - options: functions to set optional parameters
func NewExecutor(options ...ExecutorOption) (Executor, error) {
	e := &executor{
		helmImage: DefaultKubectlImage,
		resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("500Mi"),
			},
		},
	}
	for _, option := range options {
		option(e)
	}

	var err error
	var restConfig *rest.Config
	if len(e.kubeConfig) > 0 {
		restConfig, err = clientcmd.RESTConfigFromKubeConfig(e.kubeConfig)
	} else {
		restConfig, err = config.GetConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("get kubeconfig error: %v", err)
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	e.client = client

	return e, nil
}

type helmOption struct {
	namespace   string
	kubeConfig  []byte
	debug       bool
	timeout     time.Duration
	wait        bool
	historyMax  uint
	version     string
	kubeAsUser  string
	kubeAsGroup string
	// add labels to helm chart
	labels map[string]string
	// add annotations to helm chart
	annotations     map[string]string
	overrides       []string
	createNamespace bool
	install         bool
	dryRun          bool
	chartData       []byte
	caBundle        string
	serviceAccount  string
	hookImage       string
	clusterName     string
	clusterRole     string
}

func (e *executor) newHelmOption(options []HelmOption) *helmOption {
	opt := &helmOption{
		wait:    true,
		debug:   true,
		timeout: MinimumTimeout,
		// default to executor namespace
		namespace: e.namespace,
	}
	opt.apply(options)
	return opt
}

func (o *helmOption) apply(options []HelmOption) {
	for _, f := range options {
		f(o)
	}
}

type HelmOption func(*helmOption)

// SetHelmKubeConfig sets the kube config data of the target cluster used by helm installation.
// NOTE: this kube config is used by the helm command to create specific chart resources.
// You only need to give the kube config the permissions it needs in the target namespace,
// if the kube config is not set, we will use the in-cluster config, this may not work.
func SetKubeconfig(kubeConfig []byte) HelmOption {
	return func(o *helmOption) {
		o.kubeConfig = kubeConfig
	}
}

// SetAnnotations sets extra annotations added to all resources in chart.
func SetAnnotations(annotations map[string]string) HelmOption {
	return func(e *helmOption) {
		e.annotations = annotations
	}
}

func SetNamespace(namespace string) HelmOption {
	return func(e *helmOption) {
		e.namespace = namespace
	}
}

func SetVersion(version string) HelmOption {
	return func(e *helmOption) {
		e.version = version
	}
}

func SetChartData(data []byte) HelmOption {
	return func(e *helmOption) {
		e.chartData = data
	}
}

// SetLabels sets extra labels added to all resources in chart.
func SetLabels(labels map[string]string) HelmOption {
	return func(e *helmOption) {
		e.labels = labels
	}
}

func SetTimeout(duration time.Duration) HelmOption {
	return func(e *helmOption) {
		e.timeout = duration
	}
}

func SetHistoryMax(historyMax uint) HelmOption {
	return func(e *helmOption) {
		e.historyMax = historyMax
	}
}

// SetDebug adds `--debug` argument to helm command.
// The default value is true.
func SetDebug(debug bool) HelmOption {
	return func(o *helmOption) {
		o.debug = debug
	}
}

// SetCreateNamespace sets the createNamespace option.
func SetCreateNamespace(createNamespace bool) HelmOption {
	return func(e *helmOption) {
		e.createNamespace = createNamespace
	}
}

// SetInstall adds `--install` argument to helm command.
func SetInstall(install bool) HelmOption {
	return func(e *helmOption) {
		e.install = install
	}
}

// SetDryRun sets the dryRun option.
func SetDryRun(dryRun bool) HelmOption {
	return func(e *helmOption) {
		e.dryRun = dryRun
	}
}

func SetKubeAsUser(user string) HelmOption {
	return func(o *helmOption) {
		o.kubeAsUser = user
	}
}

func SetKubeAsGroup(group string) HelmOption {
	return func(o *helmOption) {
		o.kubeAsGroup = group
	}
}

func SetOverrides(overrides []string) HelmOption {
	return func(o *helmOption) {
		o.overrides = overrides
	}
}

func SetCABundle(caBundle string) HelmOption {
	return func(o *helmOption) {
		o.caBundle = caBundle
	}
}

func SetServiceAccount(serviceAccount string) HelmOption {
	return func(o *helmOption) {
		o.serviceAccount = serviceAccount
	}
}

func SetHookImage(hookImage string) HelmOption {
	return func(o *helmOption) {
		o.hookImage = hookImage
	}
}

func SetClusterRole(clusterRole string) HelmOption {
	return func(o *helmOption) {
		o.clusterRole = clusterRole
	}
}

func SetClusterName(clusterName string) HelmOption {
	return func(o *helmOption) {
		o.clusterName = clusterName
	}
}

func InitHelmConf(kubeconfig []byte, namespace string) (*action.Configuration, error) {
	getter := NewClusterRESTClientGetter(kubeconfig, namespace)
	helmConf := &action.Configuration{}
	if err := helmConf.Init(getter, namespace, "", klog.Infof); err != nil {
		return nil, err
	}
	return helmConf, nil
}

// Install installs the specified chart, returns the name of the Job that executed the task.
// helm install [NAME] [CHART] [flags]
func (e *executor) Install(ctx context.Context, release, chart string, values []byte, options ...HelmOption) (string, error) {
	helmOptions := e.newHelmOption(options)
	helmConf, err := InitHelmConf(helmOptions.kubeConfig, helmOptions.namespace)
	if err != nil {
		return "", err
	}

	sts, err := e.status(helmConf, release)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			// continue to install
			return e.createInstallJob(ctx, release, chart, values, false, helmOptions)
		}
		return "", err
	}

	// helm release has been installed
	if sts.Info != nil && sts.Info.Status == "deployed" {
		return "", nil
	}

	return "", driver.ErrReleaseExists
}

// Upgrade upgrades the specified chart, returns the name of the Job that executed the task.
func (e *executor) Upgrade(ctx context.Context, release, chart string, values []byte, options ...HelmOption) (string, error) {
	helmOptions := e.newHelmOption(options)
	return e.createInstallJob(ctx, release, chart, values, true, helmOptions)
}

func chartPath(release string) string {
	return fmt.Sprintf("%s.tgz", release)
}

func (e *executor) setupChartData(release string, kubeconfig []byte, chartData, values []byte, labels, annotations map[string]string) (map[string][]byte, error) {
	kustomizationConfig := types.Kustomization{
		Resources:         []string{"./.local-helm-output.yaml"},
		CommonAnnotations: annotations,                    // add extra annotations to output
		Labels:            []types.Label{{Pairs: labels}}, // Labels to add to all objects but not selectors.
	}
	kustomizationData, err := yaml.Marshal(kustomizationConfig)
	if err != nil {
		return nil, err
	}
	data := map[string][]byte{
		postRenderExecFile: []byte(kustomizeBuild),
		kustomizationFile:  kustomizationData,
		"values.yaml":      values,
	}

	if len(chartData) > 0 {
		data[chartPath(release)] = chartData
	}

	if len(kubeconfig) > 0 {
		data[kubeConfigPath] = kubeconfig
	}
	return data, nil
}

func generateName(name, action string) string {
	return fmt.Sprintf("helm-executor-%s-%s-%s", action, name, rand.String(6))
}

func (e *executor) createConfigMap(ctx context.Context, kubeconfig []byte, name string, release string, chartData, values []byte, labels, annotations map[string]string, caBundle string) error {
	data, err := e.setupChartData(release, kubeconfig, chartData, values, labels, annotations)
	if err != nil {
		return err
	}

	// add helm cafile
	data[caFilePath], err = base64.StdEncoding.DecodeString(caBundle)
	if err != nil {
		return err
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: e.namespace,
			Labels:    e.labels,
		},
		// we can't use `Data` here because creating it with client-go will cause our compressed file to be in the
		// wrong format (application/octet-stream)
		BinaryData: data,
	}
	if e.owner != nil {
		configMap.OwnerReferences = []metav1.OwnerReference{*e.owner}
	}
	if _, err = e.client.CoreV1().ConfigMaps(e.namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

func (e *executor) createInstallJob(ctx context.Context, release, chart string, values []byte, upgrade bool, helmOptions *helmOption) (string, error) {
	args := make([]string, 0, 10)
	if upgrade {
		args = append(args, "upgrade")
		args = append(args, "--history-max", fmt.Sprintf("%d", helmOptions.historyMax))
	} else {
		args = append(args, "install")
	}

	if helmOptions.install {
		args = append(args, "--install")
	}

	if helmOptions.caBundle != "" {
		args = append(args, "--ca-file", caFilePath)
	} else {
		args = append(args, "--insecure-skip-tls-verify")
	}

	if len(helmOptions.chartData) > 0 {
		chart = chartPath(release)
	}

	if len(helmOptions.kubeConfig) > 0 {
		args = append(args, "--kubeconfig", kubeConfigPath)
	}

	if helmOptions.kubeAsUser != "" {
		args = append(args, "--kube-as-user", helmOptions.kubeAsUser)
	}

	if helmOptions.kubeAsGroup != "" {
		args = append(args, "--kube-as-group", helmOptions.kubeAsGroup)
	}

	args = append(args, release, chart, "--namespace", helmOptions.namespace)

	if helmOptions.createNamespace {
		args = append(args, "--create-namespace")
	}

	if len(values) > 0 {
		args = append(args, "--values", "values.yaml")
	}

	if helmOptions.version != "" {
		args = append(args, "--version", helmOptions.version)
	}

	if len(helmOptions.overrides) > 0 {
		args = append(args, "--set", strings.Join(helmOptions.overrides, ","))
	}

	// Post render, add annotations or labels to resources
	if len(helmOptions.labels) > 0 || len(helmOptions.annotations) > 0 {
		args = append(args, "--post-renderer", filepath.Join(workspaceBase, postRenderExecFile))
	}

	if helmOptions.dryRun {
		args = append(args, "--dry-run")
	}

	if helmOptions.debug {
		// output debug info
		args = append(args, "--debug")
	}

	if helmOptions.wait {
		args = append(args, "--wait")
		args = append(args, "--wait-for-jobs")
	}

	if helmOptions.timeout > MinimumTimeout {
		args = append(args, "--timeout", helmOptions.timeout.String())
	}

	jobAction := ActionUpgrade
	if !upgrade || (upgrade && helmOptions.install) {
		jobAction = ActionInstall
	}

	jobName := generateName(release, jobAction)
	configMapName := jobName
	e.labels["name"] = jobName

	err := e.createConfigMap(ctx, helmOptions.kubeConfig, configMapName, release, helmOptions.chartData, values, helmOptions.labels, helmOptions.annotations, helmOptions.caBundle)
	if err != nil {
		klog.Errorf("failed to create configmap: %v", err)
		return "", err
	}

	annotations := map[string]string{
		ExecutorJobActionAnnotation:  jobAction,
		ExecutorConfigHashAnnotation: fnv64(values),
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Namespace:   e.namespace,
			Labels:      e.labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: pointer.Int32(e.backoffLimit),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "helm",
							Image:           e.helmImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"/bin/sh", "-c",
								fmt.Sprintf("cp -r %s/. %s && helm %s", workspaceBaseSource, workspaceBase, strings.Join(args, " ")),
							},
							Env: []corev1.EnvVar{
								{
									Name:  "HELM_CACHE_HOME",
									Value: workspaceBase,
								},
							},
							WorkingDir: workspaceBase,
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
							Resources: e.resources,
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: pointer.Bool(false),
								Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "source",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMapName,
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
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:    pointer.Int64(65534),
						RunAsNonRoot: pointer.Bool(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
				},
			},
		},
	}
	if e.owner != nil {
		job.OwnerReferences = []metav1.OwnerReference{*e.owner}
	}
	if e.ttlSecondsAfterFinished > 0 {
		job.Spec.TTLSecondsAfterFinished = pointer.Int32(e.ttlSecondsAfterFinished)
	}
	if helmOptions.serviceAccount != "" {
		job.Spec.Template.Spec.ServiceAccountName = helmOptions.serviceAccount
	}
	if helmOptions.hookImage != "" {
		job.Spec.Template.Spec.InitContainers = []corev1.Container{
			{
				Name:            "helm-init",
				Image:           helmOptions.hookImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "HELM_CACHE_HOME",
						Value: workspaceBase,
					},
					{
						Name:  HookEnvAction,
						Value: jobAction,
					},
					{
						Name:  HookEnvClusterRole,
						Value: helmOptions.clusterRole,
					},
					{
						Name:  HookEnvClusterName,
						Value: helmOptions.clusterName,
					},
				},
				WorkingDir: workspaceBase,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "source",
						MountPath: workspaceBase,
					},
				},
				Resources: e.resources,
				SecurityContext: &corev1.SecurityContext{
					AllowPrivilegeEscalation: pointer.Bool(false),
					Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
				},
			},
		}
	}

	if job, err = e.client.BatchV1().Jobs(e.namespace).Create(ctx, job, metav1.CreateOptions{}); err != nil {
		klog.Errorf("failed to create job: %v", err)
		return "", err
	}
	return jobName, nil
}

// ForceDelete forcibly deletes all resources of the chart.
// The current implementation still uses the helm command to force deletion.
func (e *executor) ForceDelete(ctx context.Context, release string, options ...HelmOption) error {
	helmOptions := e.newHelmOption(options)
	helmConf, err := InitHelmConf(helmOptions.kubeConfig, helmOptions.namespace)
	if err != nil {
		return err
	}

	helmStatus := action.NewStatus(helmConf)
	_, err = helmStatus.Run(release)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil
		}
		return fmt.Errorf("get helm release status error: %v", err)
	}

	uninstall := action.NewUninstall(helmConf)
	uninstall.DisableHooks = true
	uninstall.DeletionPropagation = "orphan"
	if helmOptions.timeout > MinimumTimeout {
		uninstall.Timeout = helmOptions.timeout
	}
	if _, err = uninstall.Run(release); err != nil {
		return err
	}
	return nil
}

// Uninstall uninstalls the specified chart, returns the name of the Job that executed the task.
func (e *executor) Uninstall(ctx context.Context, release string, options ...HelmOption) (string, error) {
	helmOptions := e.newHelmOption(options)
	helmConf, err := InitHelmConf(helmOptions.kubeConfig, helmOptions.namespace)
	if err != nil {
		return "", err
	}

	helmStatus := action.NewStatus(helmConf)
	_, err = helmStatus.Run(release)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("get helm release status error: %v", err)
	}

	args := []string{
		"uninstall",
		release,
		"--namespace",
		helmOptions.namespace,
	}

	args = append(args, "--kubeconfig", kubeConfigPath)

	if helmOptions.dryRun {
		args = append(args, "--dry-run")
	}

	if helmOptions.debug {
		args = append(args, "--debug")
	}

	if helmOptions.wait {
		args = append(args, "--wait")
	}

	if helmOptions.timeout > MinimumTimeout {
		args = append(args, "--timeout", helmOptions.timeout.String())
	}

	name := generateName(release, ActionUninstall)
	if len(helmOptions.kubeConfig) > 0 {

		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: e.namespace,
			},
			BinaryData: map[string][]byte{
				kubeConfigPath: helmOptions.kubeConfig,
			},
		}
		if e.owner != nil {
			configMap.OwnerReferences = []metav1.OwnerReference{*e.owner}
		}
		if _, err = e.client.CoreV1().ConfigMaps(e.namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
			return "", err
		}
	}

	annotations := map[string]string{
		ExecutorJobActionAnnotation: ActionUninstall,
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   e.namespace,
			Labels:      e.labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: pointer.Int32(e.backoffLimit),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "helm",
							Image:           e.helmImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"helm"},
							Args:            args,
							Env: []corev1.EnvVar{
								{
									Name:  "HELM_CACHE_HOME",
									Value: workspaceBase,
								},
							},
							WorkingDir: workspaceBase,
							Resources:  e.resources,
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: pointer.Bool(false),
								Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
							},
						},
					},
					RestartPolicy:                 corev1.RestartPolicyNever,
					TerminationGracePeriodSeconds: new(int64),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:    pointer.Int64(65534),
						RunAsNonRoot: pointer.Bool(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
				},
			},
		},
	}
	if e.owner != nil {
		job.OwnerReferences = []metav1.OwnerReference{*e.owner}
	}
	if e.ttlSecondsAfterFinished > 0 {
		job.Spec.TTLSecondsAfterFinished = pointer.Int32(e.ttlSecondsAfterFinished)
	}
	if helmOptions.hookImage != "" {
		job.Spec.Template.Spec.InitContainers = []corev1.Container{
			{
				Name:            "helm-init",
				Image:           helmOptions.hookImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "HELM_CACHE_HOME",
						Value: workspaceBase,
					},
					{
						Name:  HookEnvAction,
						Value: ActionUninstall,
					},
					{
						Name:  HookEnvClusterRole,
						Value: helmOptions.clusterRole,
					},
					{
						Name:  HookEnvClusterName,
						Value: helmOptions.clusterName,
					},
				},
				WorkingDir: workspaceBase,
				Resources:  e.resources,
				SecurityContext: &corev1.SecurityContext{
					AllowPrivilegeEscalation: pointer.Bool(false),
					Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
				},
			},
		}
	}
	if len(helmOptions.kubeConfig) > 0 {
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
		if len(job.Spec.Template.Spec.InitContainers) > 0 {
			job.Spec.Template.Spec.InitContainers[0].VolumeMounts = []corev1.VolumeMount{
				{
					Name:      "data",
					MountPath: workspaceBase,
				},
			}
		}
	}
	if helmOptions.serviceAccount != "" {
		job.Spec.Template.Spec.ServiceAccountName = helmOptions.serviceAccount
	}

	if job, err = e.client.BatchV1().Jobs(e.namespace).Create(ctx, job, metav1.CreateOptions{}); err != nil {
		return "", err
	}
	return name, nil
}

// helm get all RELEASE_NAME [flags]
func (e *executor) Get(ctx context.Context, release string, options ...HelmOption) (*helmrelease.Release, error) {
	helmOptions := e.newHelmOption(options)
	helmConf, err := InitHelmConf(helmOptions.kubeConfig, helmOptions.namespace)
	if err != nil {
		return nil, err
	}
	get := action.NewGet(helmConf)
	result, err := get.Run(release)
	if err != nil {
		return nil, err
	}
	klog.V(2).Infof("namespace: %s, name: %s, run command success", helmOptions.namespace, release)
	return result, nil
}

func (e *executor) WaitingForResourcesReady(ctx context.Context, release string, timeout time.Duration, options ...HelmOption) (bool, error) {
	helmOptions := e.newHelmOption(options)
	helmConf, err := InitHelmConf(helmOptions.kubeConfig, helmOptions.namespace)
	if err != nil {
		return false, err
	}
	get := action.NewStatus(helmConf)
	get.ShowResources = true
	rel, err := get.Run(release)
	if err != nil {
		return false, err
	}

	var buf bytes.Buffer
	for _, resources := range rel.Info.Resources {
		for _, resource := range resources {
			data, _ := yaml.Marshal(resource)
			buf.WriteString("---\n")
			buf.Write(data)
		}
	}

	kubeClient := helmConf.KubeClient

	//rel.Info.Resources
	resources, err := kubeClient.Build(&buf, false)
	if err != nil {
		return false, err
	}

	if err = kubeClient.Wait(resources, timeout); err == nil {
		return true, nil
	}
	if err == wait.ErrWaitTimeout {
		return false, errorTimedOutToWaitResource
	}
	return false, err
}

func (e *executor) status(helmConf *action.Configuration, release string) (*helmrelease.Release, error) {
	helmStatus := action.NewStatus(helmConf)
	rel, err := helmStatus.Run(release)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil, err
		}
		return nil, err
	}
	return rel, nil
}

func fnv64(text []byte) string {
	h := fnv.New64a()
	if _, err := h.Write(text); err != nil {
		klog.Error(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}
