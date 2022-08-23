/*
Copyright 2022 KubeSphere Authors

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

package core

import (
	"context"
	"fmt"
	"reflect"
	"time"

	helmrepo "helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/api/application/v1alpha1"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmrepoindex"
)

const (
	RepositoryFinalizer = "extensions.kubesphere.io"
	ContainerName       = "catalog"

	RepoStateInstalled = "installed"
	RepoStateUpdating  = "updating"
	RepoStateReady     = "ready"
	RepoStateNotReady  = "notReady"
	RepoStateDeleting  = "deleting"
	DefaultInterval    = 5 * time.Minute
)

var _ reconcile.Reconciler = &RepositoryReconciler{}

type RepositoryReconciler struct {
	client.Client
}

// reconcileDelete delete the repository and pod.
func (r *RepositoryReconciler) reconcileDelete(ctx context.Context, repo *corev1alpha1.Repository) (ctrl.Result, error) {
	pod := &corev1.Pod{}
	podName := generatePodName(repo.Name)
	if err := r.Get(ctx, types.NamespacedName{Namespace: constants.KubeSphereNamespace, Name: podName}, pod); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	if pod.Name != "" {
		if pod.DeletionTimestamp == nil {
			if err := r.Delete(ctx, pod); err != nil && !apierrors.IsNotFound(err) {
				klog.Errorf("delete po %s/%s failed, error: %s", pod.GetNamespace(), pod.GetName(), err)
				return ctrl.Result{}, err
			}
			repo.Status.State = RepoStateDeleting
			return ctrl.Result{Requeue: true}, r.Update(ctx, repo)
		} else {
			// Wait for the pod to be deleted.
			return ctrl.Result{Requeue: true}, nil
		}
	}

	klog.V(4).Infof("remove the finalizer for repository %s", repo.Name)
	// Remove the finalizer from the subscription and update it.
	controllerutil.RemoveFinalizer(repo, RepositoryFinalizer)
	if err := r.Update(ctx, repo); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RepositoryReconciler) reconcile(ctx context.Context, repo *corev1alpha1.Repository) (ctrl.Result, error) {
	pod := &corev1.Pod{}
	podName := generatePodName(repo.Name)
	if err := r.Get(ctx, types.NamespacedName{Namespace: constants.KubeSphereNamespace, Name: podName}, pod); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	repoCopy := repo.DeepCopy()
	if pod.Name == "" {
		// Create the pod
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: constants.KubeSphereNamespace,
				Labels:    map[string]string{},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image: repo.Spec.Image,
						Name:  ContainerName,
					},
				},
			},
		}
		err := r.Create(ctx, pod)
		if err != nil {
			return ctrl.Result{}, err
		}
		repoCopy.Status.State = RepoStateInstalled
	} else {
		// Update image of pod.
		var container *corev1.Container
		for ind := range pod.Spec.Containers {
			if pod.Spec.Containers[ind].Name == ContainerName {
				container = &pod.Spec.Containers[ind]
				break
			}
		}
		// TODO: parse the image tag of repo.Spec.Image
		if container.Image != repo.Spec.Image {
			container.Image = repo.Spec.Image
			if err := r.Update(ctx, pod); err != nil {
				return ctrl.Result{}, err
			}
			repoCopy.Status.State = RepoStateUpdating
		} else {
			if pod.Status.Phase == corev1.PodRunning {
				if err := r.syncExtensions(ctx, repo, pod); err != nil {
					return ctrl.Result{}, err
				}
				repoCopy.Status.State = RepoStateReady
			} else {
				repoCopy.Status.State = RepoStateNotReady
			}
		}
	}

	if !reflect.DeepEqual(repo, repoCopy) {
		return ctrl.Result{}, r.Update(ctx, repoCopy)
	} else {
		duration := DefaultInterval
		if repoCopy.Status.State != RepoStateReady {
			duration = 15 * time.Second
		} else if repo.Spec.UpdateStrategy.RegistryPoll != nil && repo.Spec.UpdateStrategy.Interval != nil {
			duration = repo.Spec.UpdateStrategy.Interval.Duration
		}
		return ctrl.Result{Requeue: true, RequeueAfter: duration}, nil
	}
}

// syncExtensions fetch the index.yaml from pod and create extensions that belong to the repo.
func (r *RepositoryReconciler) syncExtensions(ctx context.Context, repo *corev1alpha1.Repository, pod *corev1.Pod) error {
	newCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	index, err := helmrepoindex.LoadRepoIndex(newCtx, fmt.Sprintf("http://%s:8080", pod.Status.PodIP), &v1alpha1.HelmRepoCredential{})
	if err != nil {
		return err
	} else {
		return r.createExtensions(ctx, index, repo)
	}
}

// updateOrCreateExtension create a new extension if the extension does not exist.
// Or it will update info of the extension.
func (r *RepositoryReconciler) updateOrCreateExtension(ctx context.Context, repo *corev1alpha1.Repository, extensionName string, latestExtensionVersion *corev1alpha1.ExtensionVersion) (*corev1alpha1.Extension, error) {
	extension := &corev1alpha1.Extension{}
	var err error
	klog.V(2).Infof("update or create extension: %s/%s ", repo.Name, extensionName)
	if err = r.Get(ctx, types.NamespacedName{Name: extensionName}, extension); err == nil {
		if extension.ObjectMeta.Labels[corev1alpha1.RepositoryLabel] != repo.Name {
			err = fmt.Errorf("extension: %s/%s already exists", repo.Name, extensionName)
			klog.Error(err)
			return nil, err
		}
		extensionCopy := extension.DeepCopy()
		if latestExtensionVersion != nil {
			extensionCopy.Spec = corev1alpha1.ExtensionSpec{
				ExtensionInfo: &corev1alpha1.ExtensionInfo{
					Description: latestExtensionVersion.Spec.Description,
					Icon:        latestExtensionVersion.Spec.Icon,
					Maintainers: latestExtensionVersion.Spec.Maintainers,
					Version:     latestExtensionVersion.Spec.Version,
				},
			}
			if !reflect.DeepEqual(extension.Spec, extensionCopy.Spec) {
				klog.Infof("update extension: %s/%s ", repo.Name, extensionName)
				return extensionCopy, r.Update(ctx, extensionCopy)
			}
		}
		return extension, nil
	} else if apierrors.IsNotFound(err) {
		extension = &corev1alpha1.Extension{
			ObjectMeta: metav1.ObjectMeta{
				Name: extensionName,
				Labels: map[string]string{
					corev1alpha1.RepositoryLabel: repo.Name,
				},
			},
		}
		if latestExtensionVersion != nil {
			extension.Spec = corev1alpha1.ExtensionSpec{
				ExtensionInfo: &corev1alpha1.ExtensionInfo{
					Description: latestExtensionVersion.Spec.Description,
					Icon:        latestExtensionVersion.Spec.Icon,
					Maintainers: latestExtensionVersion.Spec.Maintainers,
					Version:     latestExtensionVersion.Spec.Version,
				},
			}
		}
		if err := controllerutil.SetControllerReference(repo, extension, r.Scheme()); err != nil {
			return nil, err
		}
		klog.V(2).Infof("create new extension: %s/%s", repo.Name, extensionName)
		if err := r.Create(ctx, extension, &client.CreateOptions{}); err != nil {
			klog.Errorf("failed to create extension: %s/%s", repo.Name, extensionName)
			return nil, err
		}
		return extension, nil
	}

	return nil, err
}

func (r *RepositoryReconciler) updateOrCreateExtensionVersion(ctx context.Context, repo *corev1alpha1.Repository, extension *corev1alpha1.Extension, extensionVersion *corev1alpha1.ExtensionVersion) error {
	version := &corev1alpha1.ExtensionVersion{}
	klog.V(2).Infof("update or create extension version: %s/%s ", repo.Name, extensionVersion.Name)
	if err := r.Get(ctx, types.NamespacedName{Name: extensionVersion.Name}, version); err == nil {
		if !reflect.DeepEqual(version.Spec, extensionVersion.Spec) {
			version.Spec = extensionVersion.Spec
			klog.V(2).Infof("update extension version: %s in repo: %s", repo.Name, extensionVersion.Name)
			if err := r.Update(ctx, version); err != nil {
				klog.Errorf("failed to update extension version: %s/%s, error: %s", repo.Name, extensionVersion.Name, err)
				return err
			}
		}
		return nil
	} else if apierrors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(extension, extensionVersion, r.Scheme()); err != nil {
			return err
		}
		klog.V(2).Infof("create new extension version: %s in repo: ", extensionVersion.Name)
		if err := r.Create(ctx, extensionVersion, &client.CreateOptions{}); err != nil {
			klog.Errorf("failed to create extension version: %s/%s, error: %s", repo.Name, extensionVersion.Name, err)
			return err
		}
		return nil
	} else {
		return err
	}
}

// createExtensions create all the extensions that belong to the repo.
func (r *RepositoryReconciler) createExtensions(ctx context.Context, index *helmrepo.IndexFile, repo *corev1alpha1.Repository) error {
	for key, versions := range index.Entries {
		extensionVersions := make([]corev1alpha1.ExtensionVersion, 0, len(versions))
		for _, version := range versions {
			if version.Metadata == nil {
				klog.Warningf("version metadata is empty in repo: %s", repo.Name)
				continue
			}
			klog.V(2).Infof("find version: %s/%s", repo.Name, version.Name)
			maintainers := make([]corev1alpha1.Maintainer, 0, len(version.Maintainers))
			for _, m := range version.Maintainers {
				maintainers = append(maintainers, corev1alpha1.Maintainer{Name: m.Name, Email: m.Email, URL: m.URL})
			}
			extensionVersions = append(extensionVersions, corev1alpha1.ExtensionVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%s", version.Name, version.Version),
					Labels: map[string]string{
						corev1alpha1.RepositoryLabel: repo.Name,
						corev1alpha1.ExtensionLabel:  key,
					},
				},
				Spec: corev1alpha1.ExtensionVersionSpec{
					Keywords:       version.Keywords,
					Repo:           repo.Name,
					MinKubeVersion: version.KubeVersion,
					Home:           version.Home,
					Digest:         version.Digest,
					Sources:        version.Sources,
					ExtensionInfo: &corev1alpha1.ExtensionInfo{
						Description: version.Description,
						Icon:        version.Icon,
						Maintainers: maintainers,
						Version:     version.Version,
					},
					URLs: version.URLs,
				},
			})
		}

		latestExtensionVersion := getLatestExtensionVersion(extensionVersions)
		if extension, err := r.updateOrCreateExtension(ctx, repo, key, latestExtensionVersion); err != nil {
			return err
		} else {
			for _, extensionVersion := range extensionVersions {
				if err := r.updateOrCreateExtensionVersion(ctx, repo, extension, &extensionVersion); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(4).Infof("sync repository: %s ", req.String())

	repo := &corev1alpha1.Repository{}
	if err := r.Client.Get(ctx, req.NamespacedName, repo); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !controllerutil.ContainsFinalizer(repo, RepositoryFinalizer) {
		patch := client.MergeFrom(repo.DeepCopy())
		controllerutil.AddFinalizer(repo, RepositoryFinalizer)
		if err := r.Patch(ctx, repo, patch); err != nil {
			klog.Errorf("unable to register finalizer for repository %s, error: %s", repo.Name, err)
			return ctrl.Result{}, err
		}
	}

	// Delete this repo
	if repo.ObjectMeta.DeletionTimestamp != nil {
		return r.reconcileDelete(ctx, repo)
	} else {
		return r.reconcile(ctx, repo)
	}
}

func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewControllerManagedBy(mgr).
		Named("repository-controller").
		For(&corev1alpha1.Repository{}).Complete(r)
}
