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

package repository

import (
	"context"
	"fmt"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/api/application/v1alpha1"
	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/controller/extension"
	"kubesphere.io/kubesphere/pkg/controller/extension/util"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmrepoindex"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

const (
	RepositoryFinalizer = "repository.extensions.kubesphere.io"
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
func (r *RepositoryReconciler) reconcileDelete(ctx context.Context, repo *extensionsv1alpha1.Repository) (ctrl.Result, error) {
	pod := &corev1.Pod{}
<<<<<<< HEAD
	podName := fmt.Sprintf("%s-%s", extension.RepoPodPrefix, repo.Name)
=======
	podName := util.GeneratePodName(repo.Name)
>>>>>>> 631acceef (repository controller)
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

func (r *RepositoryReconciler) reconcile(ctx context.Context, repo *extensionsv1alpha1.Repository) (ctrl.Result, error) {
	pod := &corev1.Pod{}
	podName := util.GeneratePodName(repo.Name)
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
				if err := r.syncPlugins(ctx, repo, pod); err != nil {
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
<<<<<<< HEAD
		if repo.Spec.UpdateStrategy.Interval != nil {
=======
		if repo.Spec.UpdateStrategy.RegistryPoll != nil && repo.Spec.UpdateStrategy.Interval != nil {
>>>>>>> 631acceef (repository controller)
			duration = repo.Spec.UpdateStrategy.Interval.Duration
		}
		return ctrl.Result{Requeue: true, RequeueAfter: duration}, nil
	}
}

// syncPlugins fetch the index.yaml from pod and create plugins that belong to the repo.
func (r *RepositoryReconciler) syncPlugins(ctx context.Context, repo *extensionsv1alpha1.Repository, pod *corev1.Pod) error {
	newCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	index, err := helmrepoindex.LoadRepoIndex(newCtx, fmt.Sprintf("http://%s:8080", pod.Status.PodIP), &v1alpha1.HelmRepoCredential{})
	if err != nil {
		return err
	} else {
		return r.createPlugins(ctx, index, repo)
	}
}

<<<<<<< HEAD
func (r *RepositoryReconciler) updateOrCreatePlugin(ctx context.Context, repo *extensionsv1alpha1.Repository, pluginName string, latestPluginVersion *extensionsv1alpha1.PluginVersion) (*extensionsv1alpha1.Plugin, error) {
	plugin := &extensionsv1alpha1.Plugin{}
	var err error
	klog.V(2).Infof("update or create plugin %s in repo %s", pluginName, repo.Name)
	if err = r.Get(ctx, types.NamespacedName{Name: pluginName}, plugin); err == nil {
=======
// updateOrCreatePlugin create a new plugin if the plugin does not exist.
// Or it will update info of the plugin.
func (r *RepositoryReconciler) updateOrCreatePlugin(ctx context.Context, repo *extensionsv1alpha1.Repository, pluginName string, latestPluginVersion *extensionsv1alpha1.PluginVersion) (*extensionsv1alpha1.Plugin, error) {
	plugin := &extensionsv1alpha1.Plugin{}
	var err error
	klog.V(2).Infof("update or create plugin: %s/%s ", repo.Name, pluginName)
	if err = r.Get(ctx, types.NamespacedName{Name: pluginName}, plugin); err == nil {
		if plugin.ObjectMeta.Labels[constants.ExtensionRepositoryLabel] != repo.Name {
			err = fmt.Errorf("plugin: %s/%s already exists", repo.Name, pluginName)
			klog.Error(err)
			return nil, err
		}
>>>>>>> 631acceef (repository controller)
		pluginCopy := plugin.DeepCopy()
		if latestPluginVersion != nil {
			pluginCopy.Spec = extensionsv1alpha1.PluginSpec{
				PluginInfo: &extensionsv1alpha1.PluginInfo{
					Description: latestPluginVersion.Spec.Description,
					Icon:        latestPluginVersion.Spec.Icon,
					Maintainers: latestPluginVersion.Spec.Maintainers,
					Version:     latestPluginVersion.Spec.Version,
				},
			}
			if !reflect.DeepEqual(plugin.Spec, pluginCopy.Spec) {
<<<<<<< HEAD
				klog.Infof("update plugin %s in repo %s", plugin.Name, repo.Name)
=======
				klog.Infof("update plugin: %s/%s ", repo.Name, pluginName)
>>>>>>> 631acceef (repository controller)
				return pluginCopy, r.Update(ctx, pluginCopy)
			}
		}
		return plugin, nil
	} else if apierrors.IsNotFound(err) {
		plugin = &extensionsv1alpha1.Plugin{
			ObjectMeta: metav1.ObjectMeta{
				Name: pluginName,
				Labels: map[string]string{
					constants.ExtensionRepositoryLabel: repo.Name,
				},
			},
		}
		if latestPluginVersion != nil {
			plugin.Spec = extensionsv1alpha1.PluginSpec{
				PluginInfo: &extensionsv1alpha1.PluginInfo{
					Description: latestPluginVersion.Spec.Description,
					Icon:        latestPluginVersion.Spec.Icon,
					Maintainers: latestPluginVersion.Spec.Maintainers,
					Version:     latestPluginVersion.Spec.Version,
				},
			}
		}
		if err := controllerutil.SetControllerReference(repo, plugin, r.Scheme()); err != nil {
			return nil, err
		}
<<<<<<< HEAD
		klog.Infof("create new plugin %s in repo %s", plugin.Name, repo.Name)
		if err := r.Create(ctx, plugin, &client.CreateOptions{}); err != nil {
=======
		klog.V(2).Infof("create new plugin: %s/%s", repo.Name, pluginName)
		if err := r.Create(ctx, plugin, &client.CreateOptions{}); err != nil {
			klog.Errorf("failed to create plugin: %s/%s", repo.Name, pluginName)
>>>>>>> 631acceef (repository controller)
			return nil, err
		}
		return plugin, nil
	}

	return nil, err
}

<<<<<<< HEAD
func (r *RepositoryReconciler) updateOrCreatePluginVersion(ctx context.Context, plugin *extensionsv1alpha1.Plugin, pluginVersion *extensionsv1alpha1.PluginVersion) error {
	version := &extensionsv1alpha1.PluginVersion{}
	klog.V(2).Infof("update or create plugin version %s", pluginVersion.Name)
	if err := r.Get(ctx, types.NamespacedName{Name: pluginVersion.Name}, version); err == nil {
		if !reflect.DeepEqual(version.Spec, pluginVersion.Spec) {
			version.Spec = pluginVersion.Spec
			klog.Infof("update plugin version %s", pluginVersion.Name)
			if err := r.Update(ctx, version); err != nil {
=======
func (r *RepositoryReconciler) updateOrCreatePluginVersion(ctx context.Context, repo *extensionsv1alpha1.Repository, plugin *extensionsv1alpha1.Plugin, pluginVersion *extensionsv1alpha1.PluginVersion) error {
	version := &extensionsv1alpha1.PluginVersion{}
	klog.V(2).Infof("update or create plugin version: %s/%s ", repo.Name, pluginVersion.Name)
	if err := r.Get(ctx, types.NamespacedName{Name: pluginVersion.Name}, version); err == nil {
		if !reflect.DeepEqual(version.Spec, pluginVersion.Spec) {
			version.Spec = pluginVersion.Spec
			klog.V(2).Infof("update plugin version: %s in repo: %s", repo.Name, pluginVersion.Name)
			if err := r.Update(ctx, version); err != nil {
				klog.Errorf("failed to update plugin version: %s/%s, error: %s", repo.Name, pluginVersion.Name, err)
>>>>>>> 631acceef (repository controller)
				return err
			}
		}
		return nil
	} else if apierrors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(plugin, pluginVersion, r.Scheme()); err != nil {
			return err
		}
<<<<<<< HEAD
		klog.Infof("create new plugin version %s", pluginVersion.Name)
		if err := r.Create(ctx, pluginVersion, &client.CreateOptions{}); err != nil {
=======
		klog.V(2).Infof("create new plugin version: %s in repo: ", pluginVersion.Name)
		if err := r.Create(ctx, pluginVersion, &client.CreateOptions{}); err != nil {
			klog.Errorf("failed to create plugin version: %s/%s, error: %s", repo.Name, pluginVersion.Name, err)
>>>>>>> 631acceef (repository controller)
			return err
		}
		return nil
	} else {
		return err
	}
}

// createPlugins create all the plugins that belong to the repo.
func (r *RepositoryReconciler) createPlugins(ctx context.Context, index *helmrepo.IndexFile, repo *extensionsv1alpha1.Repository) error {
	for key, versions := range index.Entries {
		pluginVersions := make([]extensionsv1alpha1.PluginVersion, 0, len(versions))
		for _, version := range versions {
			if version.Metadata == nil {
<<<<<<< HEAD
				klog.Warning("version metadata is empty")
				continue
			}
			klog.V(2).Infof("find version %s", version.Name)
=======
				klog.Warningf("version metadata is empty in repo: %s", repo.Name)
				continue
			}
			klog.V(2).Infof("find version: %s/%s", repo.Name, version.Name)
>>>>>>> 631acceef (repository controller)
			maintainers := make([]extensionsv1alpha1.Maintainer, 0, len(version.Maintainers))
			for _, m := range version.Maintainers {
				maintainers = append(maintainers, extensionsv1alpha1.Maintainer{Name: m.Name, Email: m.Email, URL: m.URL})
			}
			pluginVersions = append(pluginVersions, extensionsv1alpha1.PluginVersion{
				ObjectMeta: metav1.ObjectMeta{
<<<<<<< HEAD
					Name: fmt.Sprintf("%s-%s-%s", repo.Name, version.Name, version.Version),
					Labels: map[string]string{
						constants.ExtensionRepositoryLabel: repo.Name,
						constants.ExtensionPluginLabel:     fmt.Sprintf("%s-%s", repo.Name, version.Name),
=======
					Name: fmt.Sprintf("%s-%s", version.Name, version.Version),
					Labels: map[string]string{
						constants.ExtensionRepositoryLabel: repo.Name,
						constants.ExtensionPluginLabel:     key,
>>>>>>> 631acceef (repository controller)
					},
				},
				Spec: extensionsv1alpha1.PluginVersionSpec{
					Keywords:       version.Keywords,
					Repo:           repo.Name,
					MinKubeVersion: version.KubeVersion,
					Home:           version.Home,
					Digest:         version.Digest,
					Sources:        version.Sources,
					PluginInfo: &extensionsv1alpha1.PluginInfo{
						Description: version.Description,
						Icon:        version.Icon,
						Maintainers: maintainers,
						Version:     version.Version,
					},
				},
			})
		}

		latestPluginVersion := util.GetLatestPluginVersion(pluginVersions)
<<<<<<< HEAD
		if plugin, err := r.updateOrCreatePlugin(ctx, repo, fmt.Sprintf("%s-%s", repo.Name, key), latestPluginVersion); err != nil {
			return err
		} else {
			for _, pluginVersion := range pluginVersions {
				if err := r.updateOrCreatePluginVersion(ctx, plugin, &pluginVersion); err != nil {
=======
		if plugin, err := r.updateOrCreatePlugin(ctx, repo, key, latestPluginVersion); err != nil {
			return err
		} else {
			for _, pluginVersion := range pluginVersions {
				if err := r.updateOrCreatePluginVersion(ctx, repo, plugin, &pluginVersion); err != nil {
>>>>>>> 631acceef (repository controller)
					return err
				}
			}
		}
	}
	return nil
}

func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(4).Infof("sync repository: %s ", req.String())

	repo := &extensionsv1alpha1.Repository{}
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
		For(&extensionsv1alpha1.Repository{}).Complete(r)
}
