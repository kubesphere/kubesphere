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
	"regexp"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/validation"

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

const helmRepoController = "helmrepo-controller"

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

func (r *RepoReconciler) SetupWithManager(mgr *kscontroller.Manager) (err error) {
	r.Client = mgr.GetClient()
	r.recorder = mgr.GetEventRecorderFor(helmRepoController)

	r.cmStore, r.ossStore, err = application.InitStore(mgr.Options.S3Options, r.Client)
	if err != nil {
		klog.Errorf("failed to init store: %v", err)
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appv2.Repo{}).
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
	if helmRepo.Status.State == appv2.StatusSuccessful && time.Since(helmRepo.Status.LastUpdateTime.Time).Seconds() < float64(helmRepo.Spec.SyncPeriod) {
		klog.Infof("last sync time is %s, no need to sync, repo: %s", helmRepo.Status.LastUpdateTime, helmRepo.GetName())
		return true, nil
	}
	if helmRepo.Status.State == appv2.StatusSyncing {
		klog.Infof("repo is syncing, repo: %s, no need to sync", helmRepo.GetName())
		return true, nil
	}
	return false, nil
}

func (r *RepoReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {

	helmRepo := &appv2.Repo{}
	if err := r.Client.Get(ctx, request.NamespacedName, helmRepo); err != nil {
		klog.Errorf("get helm repo failed, error: %s", err)
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	noSync, err := r.noNeedSync(ctx, helmRepo)
	if err != nil {
		return reconcile.Result{}, err
	}
	if noSync {
		return reconcile.Result{}, nil
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
	for appName, versions := range index.Entries {
		if len(versions) == 0 {
			klog.Infof("no version found for %s", appName)
			continue
		}
		if len(versions) > appv2.MaxNumOfVersions {
			versions = versions[:appv2.MaxNumOfVersions]
		}
		klog.Infof("found %d versions for %s begin to sync", len(versions), appName)

		vRequests, err := repoParseRequest(r.Client, versions, helmRepo, appName)
		if err != nil {
			klog.Errorf("parse request failed, error: %s", err)
			return reconcile.Result{}, err
		}
		if len(vRequests) == 0 {
			klog.Infof("no version need to sync for %s", appName)
			continue
		}

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

	requeueAfter := time.Duration(helmRepo.Spec.SyncPeriod)
	return reconcile.Result{RequeueAfter: requeueAfter}, nil
}

func repoParseRequest(cli client.Client, versions helmrepo.ChartVersions, helmRepo *appv2.Repo, appName string) (result []application.AppRequest, err error) {
	appVersionList := &appv2.ApplicationVersionList{}

	opts := client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{appv2.RepoIDLabelKey: helmRepo.Name}),
	}
	err = cli.List(context.Background(), appVersionList, &opts)
	if err != nil {
		klog.Errorf("list appversion failed, error: %s", err)
		return nil, err
	}

	appVersionDigestMap := make(map[string]string)
	for _, i := range appVersionList.Items {
		//name format: repoName-appName-version myrepo-mysql-v3.0.0
		appVersionDigestMap[i.Name] = i.Spec.Digest
	}
	re := regexp.MustCompile(`[^a-z0-9-.]`)
	for _, ver := range versions {

		errs := validation.IsDNS1123Subdomain(ver.Version)
		if len(errs) != 0 {
			klog.Warningf("Version %s does not meet the Kubernetes naming standard, replacing invalid characters with '-'", ver.Version)
			ver.Version = re.ReplaceAllStringFunc(ver.Version, func(s string) string {
				return "-"
			})
		}

		key := fmt.Sprintf("%s-%s-%s", helmRepo.Name, ver.Name, ver.Version)
		dig := appVersionDigestMap[key]
		if dig == ver.Digest {
			klog.Infof("version %s-%s-%s already exists", helmRepo.Name, ver.Name, ver.Version)
			continue
		}
		vRequest := application.AppRequest{
			RepoName:     helmRepo.Name,
			VersionName:  ver.Version,
			AppName:      fmt.Sprintf("%s-%s", helmRepo.Name, appName),
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
		methodList := []string{"https://", "http://", "s3://"}
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
