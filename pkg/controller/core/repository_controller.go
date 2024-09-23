/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package core

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/chart/loader"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	"kubesphere.io/utils/helm"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/constants"
	kscontroller "kubesphere.io/kubesphere/pkg/controller"
)

const (
	repositoryProtection        = "kubesphere.io/repository-protection"
	repositoryController        = "repository"
	minimumRegistryPollInterval = 15 * time.Minute
	defaultRequeueInterval      = 15 * time.Second
	generateNameFormat          = "repository-%s"
	extensionFileName           = "extension.yaml"
	// caTemplate store repository.spec.caBound in local dir.
	caTemplate = "{{ .TempDIR }}/repository/{{ .RepositoryName }}/ssl/ca.crt"
)

var extensionRepoConflict = fmt.Errorf("extension repo mismatch")

var _ kscontroller.Controller = &RepositoryReconciler{}
var _ reconcile.Reconciler = &RepositoryReconciler{}

func (r *RepositoryReconciler) Name() string {
	return repositoryController
}

func (r *RepositoryReconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

type RepositoryReconciler struct {
	client.Client
	recorder record.EventRecorder
	logger   logr.Logger
}

func (r *RepositoryReconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	r.logger = ctrl.Log.WithName("controllers").WithName(repositoryController)
	r.recorder = mgr.GetEventRecorderFor(repositoryController)
	return ctrl.NewControllerManagedBy(mgr).
		Named(repositoryController).
		For(&corev1alpha1.Repository{}).
		Complete(r)
}

func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.logger.WithValues("repository", req.String())
	logger.V(4).Info("sync repository")
	ctx = klog.NewContext(ctx, logger)

	repo := &corev1alpha1.Repository{}
	if err := r.Client.Get(ctx, req.NamespacedName, repo); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !repo.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, repo)
	}

	if !controllerutil.ContainsFinalizer(repo, repositoryProtection) {
		expected := repo.DeepCopy()
		controllerutil.AddFinalizer(expected, repositoryProtection)
		return ctrl.Result{}, r.Patch(ctx, expected, client.MergeFrom(repo))
	}
	return r.reconcileRepository(ctx, repo)
}

// reconcileDelete delete the repository and pod.
func (r *RepositoryReconciler) reconcileDelete(ctx context.Context, repo *corev1alpha1.Repository) (ctrl.Result, error) {
	// Remove the finalizer from the subscription and update it.
	controllerutil.RemoveFinalizer(repo, repositoryProtection)
	if err := r.Update(ctx, repo); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// createOrUpdateExtension create a new extension if the extension does not exist.
// Or it will update info of the extension.
func (r *RepositoryReconciler) createOrUpdateExtension(ctx context.Context, repo *corev1alpha1.Repository, extensionName string, extensionVersion *corev1alpha1.ExtensionVersion) (*corev1alpha1.Extension, error) {
	logger := klog.FromContext(ctx)
	extension := &corev1alpha1.Extension{ObjectMeta: metav1.ObjectMeta{Name: extensionName}}
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, extension, func() error {
		originRepoName := extension.Labels[corev1alpha1.RepositoryReferenceLabel]
		if originRepoName != "" && originRepoName != repo.Name {
			logger.Error(extensionRepoConflict, "conflict", "extension", extensionName, "want", originRepoName, "got", repo.Name)
			return extensionRepoConflict
		}

		if extension.Labels == nil {
			extension.Labels = make(map[string]string)
		}
		if extensionVersion.Spec.Category != "" {
			extension.Labels[corev1alpha1.CategoryLabel] = extensionVersion.Spec.Category
		}
		extension.Labels[corev1alpha1.RepositoryReferenceLabel] = repo.Name
		extension.Spec.ExtensionInfo = extensionVersion.Spec.ExtensionInfo
		if err := controllerutil.SetOwnerReference(repo, extension, r.Scheme()); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update extension: %s", err)
	}

	logger.V(4).Info("extension successfully updated", "operation", op, "name", extension.Name)
	return extension, nil
}

func (r *RepositoryReconciler) createOrUpdateExtensionVersion(ctx context.Context, extension *corev1alpha1.Extension, extensionVersion *corev1alpha1.ExtensionVersion) error {
	logger := klog.FromContext(ctx)
	version := &corev1alpha1.ExtensionVersion{ObjectMeta: metav1.ObjectMeta{Name: extensionVersion.Name}}
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, version, func() error {
		if version.Labels == nil {
			version.Labels = make(map[string]string)
		}
		for k, v := range extensionVersion.Labels {
			version.Labels[k] = v
		}
		version.Spec = extensionVersion.Spec
		if err := controllerutil.SetOwnerReference(extension, version, r.Scheme()); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update extension version: %s", err)
	}

	logger.V(4).Info("extension version successfully updated", "operation", op, "name", extensionVersion.Name)
	return nil
}

func (r *RepositoryReconciler) syncExtensionsFromURL(ctx context.Context, repo *corev1alpha1.Repository, repoURL string) error {
	logger := klog.FromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cred, err := newHelmCred(repo)
	if err != nil {
		return err
	}
	index, err := helm.LoadRepoIndex(ctx, repoURL, cred)
	if err != nil {
		return err
	}

	for extensionName, versions := range index.Entries {
		extensionVersions := make([]corev1alpha1.ExtensionVersion, 0, len(versions))
		for _, version := range versions {
			if version.Metadata == nil {
				logger.Info("version metadata is empty", "repo", repo.Name)
				continue
			}

			if version.Name != extensionName {
				logger.Info("invalid extension version found", "want", extensionName, "got", version.Name)
				continue
			}

			var chartURL string
			if len(version.URLs) > 0 {
				versionURL := version.URLs[0]
				u, err := url.Parse(versionURL)
				if err != nil {
					logger.Error(err, "failed to parse chart URL", "url", versionURL)
					continue
				}
				if u.Host == "" {
					chartURL = fmt.Sprintf("%s/%s", repoURL, versionURL)
				} else {
					chartURL = u.String()
				}
			}

			extensionVersionSpec, err := r.loadExtensionVersionSpecFrom(ctx, chartURL, repo, cred)
			if err != nil {
				return fmt.Errorf("failed to load extension version spec: %s", err)
			}

			if extensionVersionSpec == nil {
				logger.V(4).Info("extension version spec not found: %s", chartURL)
				continue
			}
			extensionVersionSpec.Created = metav1.NewTime(version.Created)
			extensionVersionSpec.Digest = version.Digest
			extensionVersionSpec.Repository = repo.Name
			extensionVersionSpec.ChartDataRef = nil
			extensionVersionSpec.ChartURL = chartURL

			extensionVersion := corev1alpha1.ExtensionVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%s", extensionName, extensionVersionSpec.Version),
					Labels: map[string]string{
						corev1alpha1.RepositoryReferenceLabel: repo.Name,
						corev1alpha1.ExtensionReferenceLabel:  extensionName,
					},
					Annotations: version.Metadata.Annotations,
				},
				Spec: *extensionVersionSpec,
			}
			if extensionVersionSpec.Category != "" {
				extensionVersion.Labels[corev1alpha1.CategoryLabel] = extensionVersionSpec.Category
			}
			extensionVersions = append(extensionVersions, extensionVersion)
		}

		latestExtensionVersion := getLatestExtensionVersion(extensionVersions)
		if latestExtensionVersion == nil {
			continue
		}

		extension, err := r.createOrUpdateExtension(ctx, repo, extensionName, latestExtensionVersion)
		if err != nil {
			if errors.Is(err, extensionRepoConflict) {
				continue
			}
			return fmt.Errorf("failed to create or update extension: %s", err)
		}

		for _, extensionVersion := range extensionVersions {
			if err := r.createOrUpdateExtensionVersion(ctx, extension, &extensionVersion); err != nil {
				return fmt.Errorf("failed to create or update extension version: %s", err)
			}
		}

		if err := r.removeSuspendedExtensionVersion(ctx, repo, extension, extensionVersions); err != nil {
			return fmt.Errorf("failed to remove suspended extension version: %s", err)
		}
	}

	extensions := &corev1alpha1.ExtensionList{}
	if err := r.List(ctx, extensions, client.MatchingLabels{corev1alpha1.RepositoryReferenceLabel: repo.Name}); err != nil {
		return fmt.Errorf("failed to list extensions: %s", err)
	}

	for _, extension := range extensions.Items {
		if _, ok := index.Entries[extension.Name]; !ok {
			if err := r.removeSuspendedExtensionVersion(ctx, repo, &extension, []corev1alpha1.ExtensionVersion{}); err != nil {
				return fmt.Errorf("failed to remove suspended extension version: %s", err)
			}
		}
	}

	return nil
}

func (r *RepositoryReconciler) reconcileRepository(ctx context.Context, repo *corev1alpha1.Repository) (ctrl.Result, error) {
	registryPollInterval := minimumRegistryPollInterval
	if repo.Spec.UpdateStrategy != nil && repo.Spec.UpdateStrategy.Interval.Duration > minimumRegistryPollInterval {
		registryPollInterval = repo.Spec.UpdateStrategy.Interval.Duration
	}

	var repoURL string
	// URL and Image are immutable after creation
	if repo.Spec.URL != "" {
		repoURL = repo.Spec.URL
	} else if repo.Spec.Image != "" {
		var deployment appsv1.Deployment
		if err := r.Get(ctx, types.NamespacedName{Namespace: constants.KubeSphereNamespace, Name: fmt.Sprintf(generateNameFormat, repo.Name)}, &deployment); err != nil {
			if apierrors.IsNotFound(err) {
				if err := r.deployRepository(ctx, repo); err != nil {
					r.recorder.Event(repo, corev1.EventTypeWarning, "RepositoryDeployFailed", err.Error())
					return ctrl.Result{}, fmt.Errorf("failed to deploy repository: %s", err)
				}
				r.recorder.Event(repo, corev1.EventTypeNormal, "RepositoryDeployed", "")
				return ctrl.Result{Requeue: true, RequeueAfter: defaultRequeueInterval}, nil
			}
			return ctrl.Result{}, fmt.Errorf("failed to fetch deployment: %s", err)
		}

		restartAt, _ := time.Parse(time.RFC3339, deployment.Spec.Template.Annotations["kubesphere.io/restartedAt"])
		if restartAt.IsZero() {
			restartAt = deployment.ObjectMeta.CreationTimestamp.Time
		}
		// restart and pull the latest docker image
		if time.Now().After(repo.Status.LastSyncTime.Add(registryPollInterval)) && time.Now().After(restartAt.Add(registryPollInterval)) {
			rawData := []byte(fmt.Sprintf("{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"kubesphere.io/restartedAt\":\"%s\"}}}}}", time.Now().Format(time.RFC3339)))
			if err := r.Patch(ctx, &deployment, client.RawPatch(types.StrategicMergePatchType, rawData)); err != nil {
				return ctrl.Result{}, err
			}
			r.recorder.Event(repo, corev1.EventTypeNormal, "RepositoryRestarted", "")
			return ctrl.Result{Requeue: true, RequeueAfter: defaultRequeueInterval}, nil
		}

		if deployment.Status.AvailableReplicas != deployment.Status.Replicas {
			return ctrl.Result{Requeue: true, RequeueAfter: defaultRequeueInterval}, nil
		}

		// ready to sync
		repoURL = fmt.Sprintf("http://%s.%s.svc", deployment.Name, constants.KubeSphereNamespace)
	}

	outOfSync := repo.Status.LastSyncTime == nil || time.Now().After(repo.Status.LastSyncTime.Add(registryPollInterval))
	if repoURL != "" && outOfSync {
		if err := r.syncExtensionsFromURL(ctx, repo, repoURL); err != nil {
			r.recorder.Eventf(repo, corev1.EventTypeWarning, kscontroller.SyncFailed, "failed to sync extensions from %s: %s", repoURL, err)
			return ctrl.Result{}, fmt.Errorf("failed to sync extensions: %s", err)
		}
		r.recorder.Eventf(repo, corev1.EventTypeNormal, kscontroller.Synced, "sync extensions from %s successfully", repoURL)
		repo = repo.DeepCopy()
		repo.Status.LastSyncTime = &metav1.Time{Time: time.Now()}
		if err := r.Update(ctx, repo); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update repository: %s", err)
		}
	}

	return ctrl.Result{Requeue: true, RequeueAfter: registryPollInterval}, nil
}

func (r *RepositoryReconciler) deployRepository(ctx context.Context, repo *corev1alpha1.Repository) error {
	generateName := fmt.Sprintf(generateNameFormat, repo.Name)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      generateName,
			Namespace: constants.KubeSphereNamespace,
			Labels:    map[string]string{corev1alpha1.RepositoryReferenceLabel: repo.Name},
		},

		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{corev1alpha1.RepositoryReferenceLabel: repo.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{corev1alpha1.RepositoryReferenceLabel: repo.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "repository",
							Image:           repo.Spec.Image,
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name:  "CHART_URL",
									Value: fmt.Sprintf("http://%s.%s.svc", generateName, constants.KubeSphereNamespace),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt32(8080),
									},
								},
								PeriodSeconds:       10,
								InitialDelaySeconds: 5,
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetOwnerReference(repo, deployment, r.Scheme()); err != nil {
		return err
	}
	if err := r.Create(ctx, deployment); err != nil {
		return err
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      generateName,
			Namespace: constants.KubeSphereNamespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt32(8080),
				},
			},
			Selector: map[string]string{
				corev1alpha1.RepositoryReferenceLabel: repo.Name,
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	if err := controllerutil.SetOwnerReference(repo, service, r.Scheme()); err != nil {
		return err
	}

	if err := r.Create(ctx, service); err != nil {
		return err
	}

	return nil
}

func (r *RepositoryReconciler) loadExtensionVersionSpecFrom(ctx context.Context, chartURL string, repo *corev1alpha1.Repository, cred helm.RepoCredential) (*corev1alpha1.ExtensionVersionSpec, error) {
	logger := klog.FromContext(ctx)
	var result *corev1alpha1.ExtensionVersionSpec

	err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		return true
	}, func() error {
		data, err := helm.LoadData(ctx, chartURL, cred)
		if err != nil {
			return err
		}

		files, err := loader.LoadArchiveFiles(data)
		if err != nil {
			return err
		}

		for _, file := range files {
			if file.Name == extensionFileName {
				extensionVersionSpec := &corev1alpha1.ExtensionVersionSpec{}
				if err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(file.Data), 1024).Decode(extensionVersionSpec); err != nil {
					logger.V(4).Info("invalid extension version spec: %s", string(file.Data))
					return nil
				}
				result = extensionVersionSpec
				break
			}
		}

		if result == nil {
			logger.V(6).Info("extension.yaml not found", "chart", chartURL)
			return nil
		}

		if strings.HasPrefix(result.Icon, "http://") ||
			strings.HasPrefix(result.Icon, "https://") ||
			strings.HasPrefix(result.Icon, "data:image") {
			return nil
		}

		absPath := strings.TrimPrefix(result.Icon, "./")
		var iconData []byte
		for _, file := range files {
			if file.Name == absPath {
				iconData = file.Data
				break
			}
		}

		if iconData == nil {
			logger.V(4).Info("invalid extension icon path: %s", absPath)
			return nil
		}

		mimeType := mime.TypeByExtension(path.Ext(result.Icon))
		if mimeType == "" {
			mimeType = http.DetectContentType(iconData)
		}

		base64EncodedData := base64.StdEncoding.EncodeToString(iconData)
		result.Icon = fmt.Sprintf("data:%s;base64,%s", mimeType, base64EncodedData)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch chart data from %s: %s", chartURL, err)
	}

	return result, nil
}

func (r *RepositoryReconciler) removeSuspendedExtensionVersion(ctx context.Context, repo *corev1alpha1.Repository, extension *corev1alpha1.Extension, versions []corev1alpha1.ExtensionVersion) error {
	extensionVersions := &corev1alpha1.ExtensionVersionList{}
	if err := r.List(ctx, extensionVersions, client.MatchingLabels{corev1alpha1.ExtensionReferenceLabel: extension.Name, corev1alpha1.RepositoryReferenceLabel: repo.Name}); err != nil {
		return fmt.Errorf("failed to list extension versions: %s", err)
	}
	for _, version := range extensionVersions.Items {
		if checkIfSuspended(versions, version) {
			r.logger.V(4).Info("delete suspended extension version", "name", version.Name, "version", version.Spec.Version)
			if err := r.Delete(ctx, &version); err != nil {
				if apierrors.IsNotFound(err) {
					return nil
				}
				return fmt.Errorf("failed to delete extension version: %s", err)
			}
		}
	}
	return nil
}

func checkIfSuspended(versions []corev1alpha1.ExtensionVersion, version corev1alpha1.ExtensionVersion) bool {
	for _, v := range versions {
		if v.Name == version.Name && v.Spec.Version == version.Spec.Version {
			return false
		}
	}
	return true
}
