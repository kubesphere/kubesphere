/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package core

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	"kubesphere.io/utils/helm"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"
)

const (
	repositoryProtection        = "kubesphere.io/repository-protection"
	repositoryController        = "repository"
	minimumRegistryPollInterval = 15 * time.Minute
	defaultRegistryPollTimeout  = 2 * time.Minute
	extensionFileName           = "extension.yaml"
)

var extensionRepoConflict = errors.New("extension repo mismatch")

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
	logger.V(4).Info("reconciling extension repository")
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
			logger.Error(extensionRepoConflict, "extension repo mismatch", "name", extension.Name, "origin", originRepoName, "current", repo.Name)
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
			return errors.Wrapf(err, "failed to set owner reference")
		}
		return nil
	})

	if err != nil {
		return nil, errors.Wrapf(err, "failed to update extension")
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
		if version.Annotations == nil {
			version.Annotations = make(map[string]string)
		}
		for k, v := range extensionVersion.Annotations {
			version.Annotations[k] = v
		}
		version.Spec = extensionVersion.Spec
		if err := controllerutil.SetOwnerReference(extension, version, r.Scheme()); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return errors.Wrapf(err, "failed to update extension version")
	}

	logger.V(4).Info("extension version successfully updated", "operation", op, "name", extensionVersion.Name)
	return nil
}

func (r *RepositoryReconciler) syncExtensionsFromURL(ctx context.Context, repo *corev1alpha1.Repository, timeout time.Duration) error {
	logger := klog.FromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cred, err := createHelmCredential(repo)
	if err != nil {
		return errors.Wrapf(err, "failed to create helm credential")
	}

	repoURL, err := url.Parse(repo.Spec.URL)
	if err != nil {
		return errors.Wrapf(err, "failed to parse repo URL")
	}

	index, err := helm.LoadRepoIndex(ctx, repo.Spec.URL, cred)
	if err != nil {
		return errors.Wrapf(err, "failed to load repo index")
	}

	for extensionName, versions := range index.Entries {
		// check extensionName
		if errs := isValidExtensionName(extensionName); len(errs) > 0 {
			logger.Info("invalid extension name", "extension", extensionName, "error", errs)
			continue
		}

		extensionVersions := make([]corev1alpha1.ExtensionVersion, 0, len(versions))
		for _, version := range versions {
			if version.Name != extensionName {
				logger.V(4).Info("extension name mismatch", "extension", extensionName, "version", version.Version)
				continue
			}

			chartURL := resolveChartURL(version, repoURL)
			if chartURL == nil {
				logger.V(4).Info("failed to resolve chart URL", "extension", extensionName, "version", version.Version)
				continue
			}

			extensionVersion := corev1alpha1.ExtensionVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%s", extensionName, version.Version),
					Labels: map[string]string{
						corev1alpha1.RepositoryReferenceLabel: repo.Name,
						corev1alpha1.ExtensionReferenceLabel:  extensionName,
					},
					Annotations: version.Metadata.Annotations,
				},
				Spec: corev1alpha1.ExtensionVersionSpec{
					ChartURL:   chartURL.String(),
					Repository: repo.Name,
				},
			}

			extensionVersionSpec, err := r.fetchExtensionVersionSpec(ctx, &extensionVersion)
			if err != nil {
				return errors.Wrapf(err, "failed to load extension version spec")
			}

			if extensionVersionSpec.Name != extensionName {
				logger.V(4).Info("extension version name mismatch", "extension", extensionName, "version", version.Version)
				continue
			}

			extensionVersionSpec.ChartURL = chartURL.String()
			extensionVersionSpec.Created = metav1.NewTime(version.Created)
			extensionVersionSpec.Digest = version.Digest
			extensionVersionSpec.Repository = repo.Name
			if extensionVersionSpec.Category != "" {
				extensionVersion.Labels[corev1alpha1.CategoryLabel] = extensionVersionSpec.Category
			}

			extensionVersion.Spec = extensionVersionSpec
			extensionVersions = append(extensionVersions, extensionVersion)
		}

		filteredVersions := filterExtensionVersions(extensionVersions, repo.Spec.Depth)
		if len(filteredVersions) == 0 {
			continue
		}
		// update extension of latest extensionVersion
		extension, err := r.createOrUpdateExtension(ctx, repo, extensionName, ptr.To(filteredVersions[0]))
		if err != nil {
			if errors.Is(err, extensionRepoConflict) {
				continue
			}
			return errors.Wrapf(err, "failed to create or update extension")
		}
		// create extensionVersions of filteredVersions
		for _, extensionVersion := range filteredVersions {
			if err := r.createOrUpdateExtensionVersion(ctx, extension, &extensionVersion); err != nil {
				return errors.Wrapf(err, "failed to create or update extension version")
			}
		}
		// remove extensionVersions of existVersions
		if err := r.removeSuspendedExtensionVersion(ctx, repo.Name, extension.Name, extensionVersions); err != nil {
			return errors.Wrapf(err, "failed to remove suspended extension version")
		}
	}

	extensions := &corev1alpha1.ExtensionList{}
	if err := r.List(ctx, extensions, client.MatchingLabels{corev1alpha1.RepositoryReferenceLabel: repo.Name}); err != nil {
		return errors.Wrapf(err, "failed to list extensions")
	}

	for _, extension := range extensions.Items {
		if _, ok := index.Entries[extension.Name]; !ok {
			// remove all the extensionVersions if the extension is not in the index
			if err := r.removeSuspendedExtensionVersion(ctx, repo.Name, extension.Name, []corev1alpha1.ExtensionVersion{}); err != nil {
				return errors.Wrapf(err, "failed to remove suspended extension version")
			}
		}
	}

	return nil
}

func resolveChartURL(version *helmrepo.ChartVersion, repoURL *url.URL) *url.URL {
	if len(version.URLs) == 0 {
		return nil
	}

	chartURL, err := url.Parse(version.URLs[0])
	if err != nil {
		return nil
	}

	if chartURL.Host == "" {
		chartURL.Scheme = repoURL.Scheme
		chartURL.Host = repoURL.Host
	}

	return chartURL
}

func (r *RepositoryReconciler) reconcileRepository(ctx context.Context, repo *corev1alpha1.Repository) (ctrl.Result, error) {
	registryPollInterval := minimumRegistryPollInterval
	if repo.Spec.UpdateStrategy != nil && repo.Spec.UpdateStrategy.Interval.Duration > minimumRegistryPollInterval {
		registryPollInterval = repo.Spec.UpdateStrategy.Interval.Duration
	}
	registryPollTimeout := defaultRegistryPollTimeout
	if repo.Spec.UpdateStrategy != nil && repo.Spec.UpdateStrategy.Timeout.Duration > 0 {
		registryPollTimeout = repo.Spec.UpdateStrategy.Timeout.Duration
	}

	repoURL := repo.Spec.URL
	if repoURL == "" {
		return ctrl.Result{}, nil
	}

	logger := klog.FromContext(ctx)

	outOfSync := repo.Status.LastSyncTime == nil || time.Now().After(repo.Status.LastSyncTime.Add(registryPollInterval))
	if outOfSync {
		if err := r.syncExtensionsFromURL(ctx, repo, registryPollTimeout); err != nil {
			r.recorder.Eventf(repo, corev1.EventTypeWarning, kscontroller.SyncFailed, "failed to sync extensions from %s: %s", repoURL, err)
			return ctrl.Result{}, errors.Wrapf(err, "failed to sync extensions from %s", repoURL)
		}
		r.recorder.Eventf(repo, corev1.EventTypeNormal, kscontroller.Synced, "sync extensions from %s successfully", repoURL)
		repo = repo.DeepCopy()
		repo.Status.LastSyncTime = &metav1.Time{Time: time.Now()}
		if err := r.Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to update repository status")
		}
		logger.V(4).Info("repository successfully synced", "name", repo.Name)
	}

	logger.V(4).Info("repository successfully reconciled", "name", repo.Name)
	return ctrl.Result{Requeue: true, RequeueAfter: registryPollInterval}, nil
}

func (r *RepositoryReconciler) fetchExtensionVersionSpec(ctx context.Context, extensionVersion *corev1alpha1.ExtensionVersion) (corev1alpha1.ExtensionVersionSpec, error) {
	var extensionVersionSpec corev1alpha1.ExtensionVersionSpec
	var err error
	err = retry.OnError(retry.DefaultRetry, func(err error) bool {
		return true
	}, func() error {
		extensionVersionSpec, err = fetchExtensionVersionSpec(ctx, r.Client, extensionVersion)

		return nil
	})
	if err != nil {
		return extensionVersionSpec, errors.Wrapf(err, "failed to fetch extension version spec")
	}

	return extensionVersionSpec, nil
}

func (r *RepositoryReconciler) removeSuspendedExtensionVersion(ctx context.Context, repoName, extensionName string, versions []corev1alpha1.ExtensionVersion) error {
	extensionVersions := &corev1alpha1.ExtensionVersionList{}
	if err := r.List(ctx, extensionVersions, client.MatchingLabels{corev1alpha1.ExtensionReferenceLabel: extensionName, corev1alpha1.RepositoryReferenceLabel: repoName}); err != nil {
		return errors.Wrapf(err, "failed to list extension versions")
	}
	for _, version := range extensionVersions.Items {
		if checkIfSuspended(versions, version) {
			r.logger.V(4).Info("delete suspended extension version", "name", version.Name, "version", version.Spec.Version)
			if err := r.Delete(ctx, &version); err != nil {
				if apierrors.IsNotFound(err) {
					return nil
				}
				return errors.Wrapf(err, "failed to delete extension version %s", version.Name)
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
