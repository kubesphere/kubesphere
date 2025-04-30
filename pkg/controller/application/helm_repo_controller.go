/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package application

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-logr/logr"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	appv2 "kubesphere.io/api/application/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	"kubesphere.io/api/constants"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	"kubesphere.io/utils/s3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"
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
	logger   logr.Logger
}

func (r *RepoReconciler) Name() string {
	return helmRepoController
}

func (r *RepoReconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (r *RepoReconciler) mapper(ctx context.Context, o client.Object) (requests []reconcile.Request) {
	workspace := o.(*tenantv1beta1.WorkspaceTemplate)

	r.logger.V(4).Info("workspace has been deleted", "workspace", workspace.Name)
	repoList := &appv2.RepoList{}
	opts := &client.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{constants.WorkspaceLabelKey: workspace.Name})}
	if err := r.List(ctx, repoList, opts); err != nil {
		r.logger.Error(err, "failed to list repo")
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
	r.logger = ctrl.Log.WithName("controllers").WithName(helmRepoController)
	r.cmStore, r.ossStore, err = application.InitStore(mgr.Options.S3Options, r.Client)
	if err != nil {
		r.logger.Error(err, "failed to init store")
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
	logger := r.logger.WithValues("repo", helmRepo.Name)

	patch, _ := json.Marshal(newRepo)
	err := r.Status().Patch(ctx, newRepo, client.RawPatch(client.Merge.Type(), patch))
	if err != nil {
		logger.Error(err, "update status failed")
		return err
	}
	logger.V(4).Info("update repo status", "status", helmRepo.Status.State)
	return nil
}

func (r *RepoReconciler) skipSync(helmRepo *appv2.Repo) (bool, error) {
	logger := r.logger.WithValues("repo", helmRepo.Name)
	if helmRepo.Status.State == appv2.StatusManualTrigger || helmRepo.Status.State == appv2.StatusSyncing {
		logger.V(4).Info(fmt.Sprintf("repo state: %s", helmRepo.Status.State))
		return false, nil
	}

	if helmRepo.Spec.SyncPeriod == nil || *helmRepo.Spec.SyncPeriod == 0 {
		logger.V(4).Info("repo no sync SyncPeriod=0")
		return true, nil
	}
	passed := time.Since(helmRepo.Status.LastUpdateTime.Time).Seconds()
	if helmRepo.Status.State == appv2.StatusSuccessful && passed < float64(*helmRepo.Spec.SyncPeriod) {
		logger.V(4).Info(fmt.Sprintf("last sync time is %s, passed %f, no need to sync", helmRepo.Status.LastUpdateTime, passed))
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
	logger := r.logger.WithValues("repo", request.Name)
	helmRepo := &appv2.Repo{}
	if err := r.Client.Get(ctx, request.NamespacedName, helmRepo); err != nil {
		logger.Error(err, "get helm repo failed")
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	if helmRepo.Status.State == "" {
		helmRepo.Status.State = appv2.StatusCreated
		err := r.UpdateStatus(ctx, helmRepo)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	if helmRepo.Spec.SyncPeriod == nil {
		helmRepo.Spec.SyncPeriod = ptr.To(0)
	}

	workspaceTemplate := &tenantv1beta1.WorkspaceTemplate{}
	workspaceName := helmRepo.Labels[constants.WorkspaceLabelKey]
	if workspaceName != "" {
		err := r.Get(ctx, types.NamespacedName{Name: workspaceName}, workspaceTemplate)
		if apierrors.IsNotFound(err) || (err == nil && !workspaceTemplate.DeletionTimestamp.IsZero()) {
			logger.V(4).Error(err, "workspace not found or deleting", "workspace", workspaceName)
			err = r.Delete(ctx, helmRepo)
			if err != nil {
				logger.Error(err, "delete helm repo failed")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	requeueAfter := time.Duration(*helmRepo.Spec.SyncPeriod) * time.Second
	noSync, err := r.skipSync(helmRepo)
	if err != nil {
		return reconcile.Result{}, err
	}
	if noSync {
		return reconcile.Result{RequeueAfter: requeueAfter}, nil
	}

	helmRepo.Status.State = appv2.StatusSyncing
	err = r.UpdateStatus(ctx, helmRepo)

	if err != nil {
		logger.Error(err, "update status failed")
		return reconcile.Result{}, err
	}

	index, err := application.LoadRepoIndex(helmRepo.Spec.Url, helmRepo.Spec.Credential)
	if err != nil {
		logger.Error(err, "load index failed", "url", helmRepo.Spec.Url)
		return reconcile.Result{}, err
	}

	appList := &appv2.ApplicationList{}
	opts := client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{appv2.RepoIDLabelKey: helmRepo.Name}),
	}
	err = r.Client.List(ctx, appList, &opts)
	if err != nil {
		logger.Error(err, "list application failed")
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
			logger.V(4).Info("application has been removed from the repo", "application", i.Name)
			err = r.Client.Delete(ctx, &i)
			if err != nil {
				logger.Error(err, "delete application failed", "application", i.Name)
				return reconcile.Result{}, err
			}
		}
	}

	for appName, versions := range index.Entries {
		if len(versions) == 0 {
			logger.V(4).Info("no version found for application", "application", appName)
			continue
		}

		versions = filterVersions(versions)

		vRequests, err := r.repoParseRequest(ctx, versions, helmRepo, appName, appList)
		if err != nil {
			logger.Error(err, "parse request failed")
			return reconcile.Result{}, err
		}
		if len(vRequests) == 0 {
			continue
		}

		logger.V(6).Info(fmt.Sprintf("found %d/%d versions for application %s need to upgrade or create", len(vRequests), len(versions), appName))

		own := metav1.OwnerReference{
			APIVersion: appv2.SchemeGroupVersion.String(),
			Kind:       "Repo",
			Name:       helmRepo.Name,
			UID:        helmRepo.UID,
		}
		if err = application.CreateOrUpdateApp(r.Client, vRequests, r.cmStore, r.ossStore, own); err != nil {
			logger.Error(err, "create or update app failed")
			return reconcile.Result{}, err
		}
	}

	helmRepo.Status.State = appv2.StatusSuccessful
	err = r.UpdateStatus(ctx, helmRepo)
	if err != nil {
		logger.Error(err, "update status failed")
		return reconcile.Result{}, err
	}

	r.recorder.Eventf(helmRepo, corev1.EventTypeNormal, "Synced", "HelmRepo %s synced successfully", helmRepo.GetName())

	return reconcile.Result{RequeueAfter: requeueAfter}, nil
}

func (r *RepoReconciler) repoParseRequest(ctx context.Context, versions helmrepo.ChartVersions, helmRepo *appv2.Repo, appName string, appList *appv2.ApplicationList) (createOrUpdateList []application.AppRequest, err error) {
	appVersionList := &appv2.ApplicationVersionList{}

	logger := r.logger.WithValues("repo", helmRepo.Name)
	appID := fmt.Sprintf("%s-%s", helmRepo.Name, application.GenerateShortNameMD5Hash(appName))
	opts := client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{
			appv2.RepoIDLabelKey: helmRepo.Name,
			appv2.AppIDLabelKey:  appID,
		}),
	}
	err = r.Client.List(ctx, appVersionList, &opts)
	if err != nil {
		logger.Error(err, "list application version failed")
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
			logger.V(4).Info("delete application version", "application version", i.GetName())
			err = r.Client.Delete(ctx, &i)
			if err != nil {
				logger.Error(err, "delete application version failed")
				return nil, err
			}
		} else {
			appVersionDigestMap[key] = i.Spec.Digest
		}
	}
	var legalVersion, shortName string
	for _, ver := range versions {
		legalVersion = application.FormatVersion(ver.Version)
		shortName = application.GenerateShortNameMD5Hash(ver.Name)
		key := fmt.Sprintf("%s-%s-%s", helmRepo.Name, shortName, legalVersion)
		dig := appVersionDigestMap[key]
		if dig == ver.Digest {
			continue
		}
		if dig != "" {
			logger.V(4).Info(fmt.Sprintf("digest not match, key: %s, digest: %s, ver.Digest: %s", key, dig, ver.Digest))
		}
		vRequest := generateVRequest(helmRepo, ver, shortName, appName)
		createOrUpdateList = append(createOrUpdateList, vRequest)
	}

	appNotFound := true
	for _, i := range appList.Items {
		if i.Name == appID {
			appNotFound = false
			break
		}
	}

	if len(createOrUpdateList) == 0 && len(versions) > 0 && appNotFound {
		//The repo source has been deleted, but the appversion has not been deleted due to the existence of the instance,
		//and the appversion that is scheduled to be updated is empty
		//so you need to ensure that at least one version is used to create the app
		ver := versions[0]
		v := generateVRequest(helmRepo, ver, shortName, appName)
		createOrUpdateList = append(createOrUpdateList, v)
	}
	return createOrUpdateList, nil
}

func generateVRequest(helmRepo *appv2.Repo, ver *helmrepo.ChartVersion, shortName string, appName string) application.AppRequest {
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
	return vRequest
}
