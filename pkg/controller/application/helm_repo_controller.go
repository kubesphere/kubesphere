/*
Copyright 2023 The KubeSphere Authors.

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

package application

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/api/constants"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"kubesphere.io/utils/s3"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	helmrepo "helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	appv2 "kubesphere.io/api/application/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/simple/client/application"
)

const helmRepoController = "helmrepo"

var _ reconcile.Reconciler = &RepoReconciler{}
var _ kscontroller.Controller = &RepoReconciler{}

type RepoReconciler struct {
	recorder record.EventRecorder
	client.Client
	ossStore s3.Interface
	cmStore  s3.Interface
}

func (r *RepoReconciler) Name() string {
	return helmRepoController
}

func (r *RepoReconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (r *RepoReconciler) mapper(ctx context.Context, o client.Object) (requests []reconcile.Request) {
	workspace := o.(*tenantv1beta1.WorkspaceTemplate)

	klog.Infof("workspace %s has been deleted", workspace.Name)
	repoList := &appv2.RepoList{}
	opts := &client.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{constants.WorkspaceLabelKey: workspace.Name})}
	if err := r.List(ctx, repoList, opts); err != nil {
		klog.Errorf("failed to list repo: %v", err)
		return requests
	}
	for _, repo := range repoList.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: repo.Name}})
	}
	return requests
}

func (r *RepoReconciler) SetupWithManager(mgr *kscontroller.Manager) (err error) {
	r.Client = mgr.GetClient()
	r.recorder = mgr.GetEventRecorderFor(helmRepoController)

	r.cmStore, r.ossStore, err = application.InitStore(mgr.Options.S3Options, r.Client)
	if err != nil {
		klog.Errorf("failed to init store: %v", err)
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Watches(
			&tenantv1beta1.WorkspaceTemplate{},
			handler.EnqueueRequestsFromMapFunc(r.mapper),
			builder.WithPredicates(DeletePredicate{}),
		).
		For(&appv2.Repo{}).
		Named(helmRepoController).
		Complete(r)
}

func (r *RepoReconciler) UpdateStatus(ctx context.Context, helmRepo *appv2.Repo) error {
	newRepo := &appv2.Repo{}
	newRepo.Name = helmRepo.Name
	newRepo.Status.State = helmRepo.Status.State
	newRepo.Status.LastUpdateTime = metav1.Now()

	patch, _ := json.Marshal(newRepo)
	err := r.Status().Patch(ctx, newRepo, client.RawPatch(client.Merge.Type(), patch))
	if err != nil {
		klog.Errorf("update status failed, error: %s", err)
		return err
	}
	klog.Infof("update status successfully, repo: %s", helmRepo.GetName())
	return nil
}

func (r *RepoReconciler) noNeedSync(ctx context.Context, helmRepo *appv2.Repo) (bool, error) {
	if helmRepo.Spec.SyncPeriod == 0 {
		if helmRepo.Status.State != appv2.StatusNosync {
			helmRepo.Status.State = appv2.StatusNosync
			klog.Infof("no sync when SyncPeriod=0, repo: %s", helmRepo.GetName())
			if err := r.UpdateStatus(ctx, helmRepo); err != nil {
				klog.Errorf("update status failed, error: %s", err)
				return false, err
			}
		}
		klog.Infof("no sync when SyncPeriod=0, repo: %s", helmRepo.GetName())
		return true, nil
	}
	passed := time.Since(helmRepo.Status.LastUpdateTime.Time).Seconds()
	if helmRepo.Status.State == appv2.StatusSuccessful && passed < float64(helmRepo.Spec.SyncPeriod) {
		klog.Infof("last sync time is %s, passed %f, no need to sync, repo: %s", helmRepo.Status.LastUpdateTime, passed, helmRepo.GetName())
		return true, nil
	}
	return false, nil
}

func filterVersions(versions []*helmrepo.ChartVersion) []*helmrepo.ChartVersion {
	versionMap := make(map[string]*helmrepo.ChartVersion)
	for _, v := range versions {
		if existing, found := versionMap[v.Version]; found {
			if v.Created.After(existing.Created) {
				versionMap[v.Version] = v
			}
		} else {
			versionMap[v.Version] = v
		}
	}
	result := make([]*helmrepo.ChartVersion, 0, len(versionMap))
	for _, v := range versionMap {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Created.After(result[j].Created)
	})
	return result
}

func (r *RepoReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {

	helmRepo := &appv2.Repo{}
	if err := r.Client.Get(ctx, request.NamespacedName, helmRepo); err != nil {
		klog.Errorf("get helm repo failed, error: %s", err)
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	workspaceTemplate := &tenantv1beta1.WorkspaceTemplate{}
	workspaceName := helmRepo.Labels[constants.WorkspaceLabelKey]
	if workspaceName != "" {
		err := r.Get(ctx, types.NamespacedName{Name: workspaceName}, workspaceTemplate)
		if apierrors.IsNotFound(err) || (err == nil && !workspaceTemplate.DeletionTimestamp.IsZero()) {
			klog.Infof("workspace not found or deleting %s %s", workspaceName, err)
			err = r.Delete(ctx, helmRepo)
			if err != nil {
				klog.Errorf("delete helm repo failed, error: %s", err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	requeueAfter := time.Duration(helmRepo.Spec.SyncPeriod) * time.Second
	noSync, err := r.noNeedSync(ctx, helmRepo)
	if err != nil {
		return reconcile.Result{}, err
	}
	if noSync {
		return reconcile.Result{RequeueAfter: requeueAfter}, nil
	}

	helmRepo.Status.State = appv2.StatusSyncing
	err = r.UpdateStatus(ctx, helmRepo)

	if err != nil {
		klog.Errorf("update status failed, error: %s", err)
		return reconcile.Result{}, err
	}

	index, err := application.LoadRepoIndex(helmRepo.Spec.Url, helmRepo.Spec.Credential)
	if err != nil {
		klog.Errorf("load index failed, repo: %s, url: %s, err: %s", helmRepo.GetName(), helmRepo.Spec.Url, err)
		return reconcile.Result{}, err
	}

	appList := &appv2.ApplicationList{}
	opts := client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{appv2.RepoIDLabelKey: helmRepo.Name}),
	}
	err = r.Client.List(ctx, appList, &opts)
	if err != nil {
		klog.Errorf("list appversion failed, error: %s", err)
		return reconcile.Result{}, err
	}
	indexMap := make(map[string]struct{})
	for appName := range index.Entries {
		shortName := application.GenerateShortNameMD5Hash(appName)
		key := fmt.Sprintf("%s-%s", helmRepo.Name, shortName)
		indexMap[key] = struct{}{}
	}
	for _, i := range appList.Items {
		if _, exists := indexMap[i.Name]; !exists {
			klog.Infof("app %s has been removed from the repo", i.Name)
			err = r.Client.Delete(ctx, &i)
			if err != nil {
				klog.Errorf("delete app %s failed, error: %s", i.Name, err)
				return reconcile.Result{}, err
			}
		}
	}

	for appName, versions := range index.Entries {
		if len(versions) == 0 {
			klog.Infof("no version found for %s", appName)
			continue
		}

		versions = filterVersions(versions)

		vRequests, err := repoParseRequest(r.Client, versions, helmRepo, appName)
		if err != nil {
			klog.Errorf("parse request failed, error: %s", err)
			return reconcile.Result{}, err
		}
		if len(vRequests) == 0 {
			continue
		}

		klog.Infof("found %d/%d versions for %s need to upgrade", len(vRequests), len(versions), appName)

		own := metav1.OwnerReference{
			APIVersion: appv2.SchemeGroupVersion.String(),
			Kind:       "Repo",
			Name:       helmRepo.Name,
			UID:        helmRepo.UID,
		}
		if err = application.CreateOrUpdateApp(r.Client, vRequests, r.cmStore, r.ossStore, own); err != nil {
			klog.Errorf("create or update app failed, error: %s", err)
			return reconcile.Result{}, err
		}
	}

	helmRepo.Status.State = appv2.StatusSuccessful
	err = r.UpdateStatus(ctx, helmRepo)
	if err != nil {
		klog.Errorf("update status failed, error: %s", err)
		return reconcile.Result{}, err
	}

	r.recorder.Eventf(helmRepo, corev1.EventTypeNormal, "Synced", "HelmRepo %s synced successfully", helmRepo.GetName())

	return reconcile.Result{RequeueAfter: requeueAfter}, nil
}

func repoParseRequest(cli client.Client, versions helmrepo.ChartVersions, helmRepo *appv2.Repo, appName string) (result []application.AppRequest, err error) {
	appVersionList := &appv2.ApplicationVersionList{}

	appID := fmt.Sprintf("%s-%s", helmRepo.Name, application.GenerateShortNameMD5Hash(appName))
	opts := client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{
			appv2.RepoIDLabelKey: helmRepo.Name,
			appv2.AppIDLabelKey:  appID,
		}),
	}
	err = cli.List(context.Background(), appVersionList, &opts)
	if err != nil {
		klog.Errorf("list appversion failed, error: %s", err)
		return nil, err
	}

	appVersionDigestMap := make(map[string]string)
	versionMap := make(map[string]struct{})
	for _, ver := range versions {
		v := application.FormatVersion(ver.Version)
		shortName := application.GenerateShortNameMD5Hash(ver.Name)
		key := fmt.Sprintf("%s-%s-%s", helmRepo.Name, shortName, v)
		versionMap[key] = struct{}{}
	}

	for _, i := range appVersionList.Items {
		LegalVersion := application.FormatVersion(i.Spec.VersionName)
		key := fmt.Sprintf("%s-%s", i.GetLabels()[appv2.AppIDLabelKey], LegalVersion)
		_, exists := versionMap[key]
		if !exists {
			klog.Infof("delete appversion %s", i.GetName())
			err = cli.Delete(context.Background(), &i)
			if err != nil {
				klog.Errorf("delete appversion failed, error: %s", err)
				return nil, err
			}
		} else {
			appVersionDigestMap[key] = i.Spec.Digest
		}
	}

	for _, ver := range versions {

		legalVersion := application.FormatVersion(ver.Version)
		shortName := application.GenerateShortNameMD5Hash(ver.Name)
		key := fmt.Sprintf("%s-%s-%s", helmRepo.Name, shortName, legalVersion)
		dig := appVersionDigestMap[key]
		if dig == ver.Digest {
			continue
		}
		if dig != "" {
			klog.Infof("digest not match, key: %s, digest: %s, ver.Digest: %s", key, dig, ver.Digest)
		}

		vRequest := application.AppRequest{
			RepoName:     helmRepo.Name,
			VersionName:  ver.Version,
			AppName:      fmt.Sprintf("%s-%s", helmRepo.Name, shortName),
			AliasName:    appName,
			OriginalName: appName,
			AppHome:      ver.Home,
			Icon:         ver.Icon,
			Digest:       ver.Digest,
			Description:  ver.Description,
			Abstraction:  ver.Description,
			Maintainers:  application.GetMaintainers(ver.Maintainers),
			AppType:      appv2.AppTypeHelm,
			Workspace:    helmRepo.GetWorkspace(),
			Credential:   helmRepo.Spec.Credential,
			FromRepo:     true,
		}
		url := ver.URLs[0]
		methodList := []string{"https://", "http://", "s3://", "oci://"}
		needContact := true
		for _, method := range methodList {
			if strings.HasPrefix(url, method) {
				needContact = false
				break
			}
		}

		if needContact {
			url = strings.TrimSuffix(helmRepo.Spec.Url, "/") + "/" + url
		}
		vRequest.PullUrl = url
		result = append(result, vRequest)
	}
	return result, nil
}
