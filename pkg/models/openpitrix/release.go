/*
Copyright 2020 The KubeSphere Authors.

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

package openpitrix

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"kubesphere.io/kubesphere/pkg/apiserver/query"

	"github.com/go-openapi/strfmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/api/application/v1alpha1"

	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	typed_v1alpha1 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	listers_v1alpha1 "kubesphere.io/kubesphere/pkg/client/listers/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmwrapper"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/reposcache"
	"kubesphere.io/kubesphere/pkg/utils/resourceparse"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
)

type ReleaseInterface interface {
	ListApplications(workspace, clusterName, namespace string, conditions *params.Conditions, limit, offset int, orderBy string, reverse bool) (*models.PageableResponse, error)
	DescribeApplication(workspace, clusterName, namespace, applicationId string) (*Application, error)
	CreateApplication(workspace, clusterName, namespace string, request CreateClusterRequest) error
	ModifyApplication(request ModifyClusterAttributesRequest) error
	DeleteApplication(workspace, clusterName, namespace, id string) error
	UpgradeApplication(request UpgradeClusterRequest) error

	CreateManifest(workspace, clusterName, namespace string, request CreateManifestRequest) error
	DeleteManifest(workspace, clusterName, namespace, manifestName string) error
	ModifyManifest(request ModifyManifestRequest) error
	ListManifests(workspace, clusterName, namespace string, conditions *params.Conditions, limit, offset int, orderBy string, reverse bool) (*models.PageableResponse, error)
	DescribeManifest(workspace, clusterName, namespace, manifestName string) (*Manifest, error)
}

type releaseOperator struct {
	informers        informers.SharedInformerFactory
	rlsClient        typed_v1alpha1.HelmReleaseInterface
	rlsLister        listers_v1alpha1.HelmReleaseLister
	appVersionLister listers_v1alpha1.HelmApplicationVersionLister
	cachedRepos      reposcache.ReposCache
	clusterClients   clusterclient.ClusterClients

	manifestClient    typed_v1alpha1.ManifestInterface
	manifestLister    listers_v1alpha1.ManifestLister
	operatorVerClient typed_v1alpha1.OperatorApplicationVersionInterface
	operatorClient    typed_v1alpha1.OperatorApplicationInterface
	operatorVerLister listers_v1alpha1.OperatorApplicationVersionLister
	operatorLister    listers_v1alpha1.OperatorApplicationLister
}

func newReleaseOperator(cached reposcache.ReposCache, k8sFactory informers.SharedInformerFactory, ksFactory externalversions.SharedInformerFactory, ksClient versioned.Interface) ReleaseInterface {
	c := &releaseOperator{
		informers:        k8sFactory,
		rlsClient:        ksClient.ApplicationV1alpha1().HelmReleases(),
		rlsLister:        ksFactory.Application().V1alpha1().HelmReleases().Lister(),
		cachedRepos:      cached,
		clusterClients:   clusterclient.NewClusterClient(ksFactory.Cluster().V1alpha1().Clusters()),
		appVersionLister: ksFactory.Application().V1alpha1().HelmApplicationVersions().Lister(),

		manifestClient:    ksClient.ApplicationV1alpha1().Manifests(),
		operatorVerClient: ksClient.ApplicationV1alpha1().OperatorApplicationVersions(),
		operatorClient:    ksClient.ApplicationV1alpha1().OperatorApplications(),
		manifestLister:    ksFactory.Application().V1alpha1().Manifests().Lister(),
		operatorVerLister: ksFactory.Application().V1alpha1().OperatorApplicationVersions().Lister(),
		operatorLister:    ksFactory.Application().V1alpha1().OperatorApplications().Lister(),
	}

	return c
}

type Application struct {
	Name    string      `json:"name" description:"application name"`
	Cluster *Cluster    `json:"cluster,omitempty" description:"application cluster info"`
	Version *AppVersion `json:"version,omitempty" description:"application template version info"`
	App     *App        `json:"app,omitempty" description:"application template info"`

	ReleaseInfo []runtime.Object `json:"releaseInfo,omitempty" description:"release info"`
}

type Manifest struct {
	Cluster        string `json:"cluster"`
	Namespace      string `json:"namespace"`
	Kind           string `json:"kind"`
	Description    string `json:"description,omitempty"`
	AppName        string `json:"app,omitempty"`
	AppVersion     string `json:"appVersion"`
	CustomResource string `json:"customResource" yaml:"customResource"`
	Version        int    `json:"version"`
	Name           string `json:"name" description:"cluster name"`
}

func (c *releaseOperator) UpgradeApplication(request UpgradeClusterRequest) error {
	oldRls, err := c.rlsLister.Get(request.ClusterId)

	// todo check namespace
	if err != nil {
		klog.Errorf("get release %s/%s failed, error: %s", request.Namespace, request.ClusterId, err)
		return err
	}

	switch oldRls.Status.State {
	case v1alpha1.StateActive, v1alpha1.HelmStatusUpgraded, v1alpha1.HelmStatusCreated:
		// no operation
	default:
		return errors.New("can not upgrade application now")
	}

	version, err := c.getAppVersion("", request.VersionId)
	if err != nil {
		klog.Errorf("get helm application version %s/%s failed, error: %s", request.AppId, request.VersionId, err)
		return err
	}

	newRls := oldRls.DeepCopy()
	newRls.Spec.ApplicationId = request.AppId
	newRls.Spec.ApplicationVersionId = request.VersionId

	newRls.Spec.Version += 1
	newRls.Spec.RepoId = version.GetHelmRepoId()
	newRls.Spec.ChartVersion = version.GetChartVersion()
	newRls.Spec.ChartAppVersion = version.GetChartAppVersion()
	// Use the new conf if the client has one, or server will just use the old conf.
	if request.Conf != "" {
		newRls.Spec.Values = strfmt.Base64(request.Conf)
	}

	patch := client.MergeFrom(oldRls)
	data, err := patch.Data(newRls)

	newRls, err = c.rlsClient.Patch(context.TODO(), request.ClusterId, patch.Type(), data, metav1.PatchOptions{})
	if err != nil {
		klog.Errorf("patch release %s/%s failed, error: %s", request.Namespace, request.ClusterId, err)
		return err
	} else {
		klog.V(2).Infof("patch release %s/%s success", request.Namespace, request.ClusterId)
	}

	return nil
}

// create all helm release in host cluster
func (c *releaseOperator) CreateApplication(workspace, clusterName, namespace string, request CreateClusterRequest) error {
	version, err := c.getAppVersion("", request.VersionId)

	if err != nil {
		klog.Errorf("get helm application version %s failed, error: %v", request.Name, err)
		return err
	}

	exists, err := c.releaseExists(workspace, clusterName, namespace, request.Name)

	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("get helm release %s failed, error: %v", request.Name, err)
		return err
	}

	if exists {
		err = fmt.Errorf("release %s exists", request.Name)
		klog.Error(err)
		return err
	}

	rls := &v1alpha1.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name: idutils.GetUuid36(v1alpha1.HelmReleasePrefix),
			Annotations: map[string]string{
				constants.CreatorAnnotationKey: request.Username,
			},
			Labels: map[string]string{
				constants.ChartApplicationVersionIdLabelKey: request.VersionId,
				constants.ChartApplicationIdLabelKey:        strings.TrimSuffix(request.AppId, v1alpha1.HelmApplicationAppStoreSuffix),
				constants.WorkspaceLabelKey:                 request.Workspace,
				constants.NamespaceLabelKey:                 namespace,
			},
		},
		Spec: v1alpha1.HelmReleaseSpec{
			Name:                 request.Name,
			Description:          stringutils.ShortenString(request.Description, v1alpha1.MsgLen),
			Version:              1,
			Values:               strfmt.Base64(request.Conf),
			ApplicationId:        strings.TrimSuffix(request.AppId, v1alpha1.HelmApplicationAppStoreSuffix),
			ApplicationVersionId: request.VersionId,
			ChartName:            version.GetTrueName(),
			RepoId:               version.GetHelmRepoId(),
			ChartVersion:         version.GetChartVersion(),
			ChartAppVersion:      version.GetChartAppVersion(),
		},
	}

	if clusterName != "" {
		rls.Labels[constants.ClusterNameLabelKey] = clusterName
	}

	if repoId := version.GetHelmRepoId(); repoId != "" {
		rls.Labels[constants.ChartRepoIdLabelKey] = repoId
	}

	rls, err = c.rlsClient.Create(context.TODO(), rls, metav1.CreateOptions{})

	if err != nil {
		klog.Errorln(err)
		return err
	} else {
		klog.Infof("create helm release %s success in %s", request.Name, namespace)
	}

	return nil
}

func (c *releaseOperator) releaseExists(workspace, clusterName, namespace, name string) (bool, error) {
	set := map[string]string{
		constants.WorkspaceLabelKey: workspace,
		constants.NamespaceLabelKey: namespace,
	}
	if clusterName != "" {
		set[constants.ClusterNameLabelKey] = clusterName
	}

	list, err := c.rlsLister.List(labels.SelectorFromSet(set))
	if err != nil {
		return false, err
	}
	for _, rls := range list {
		if rls.Spec.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (c *releaseOperator) ModifyApplication(request ModifyClusterAttributesRequest) error {

	if request.Description == nil || len(*request.Description) == 0 {
		return nil
	}

	rls, err := c.rlsLister.Get(request.ClusterID)
	if err != nil {
		klog.Errorf("get release failed, error: %s", err)
		return err
	}

	rlsCopy := rls.DeepCopy()
	rlsCopy.Spec.Description = stringutils.ShortenString(strings.TrimSpace(*request.Description), v1alpha1.MsgLen)

	pt := client.MergeFrom(rls)

	data, err := pt.Data(rlsCopy)
	if err != nil {
		klog.Errorf("create patch failed, error: %s", err)
		return err
	}

	_, err = c.rlsClient.Patch(context.TODO(), request.ClusterID, pt.Type(), data, metav1.PatchOptions{})
	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}

func (c *releaseOperator) ListApplications(workspace, clusterName, namespace string, conditions *params.Conditions, limit, offset int, orderBy string, reverse bool) (*models.PageableResponse, error) {
	appId := conditions.Match[AppId]
	versionId := conditions.Match[VersionId]
	ls := map[string]string{}
	if appId != "" {
		ls[constants.ChartApplicationIdLabelKey] = strings.TrimSuffix(appId, v1alpha1.HelmApplicationAppStoreSuffix)
	}

	if versionId != "" {
		ls[constants.ChartApplicationVersionIdLabelKey] = versionId
	}

	repoId := conditions.Match[RepoId]
	if repoId != "" {
		ls[constants.ChartRepoIdLabelKey] = repoId
	}

	if workspace != "" {
		ls[constants.WorkspaceLabelKey] = workspace
	}
	if namespace != "" {
		ls[constants.NamespaceLabelKey] = namespace
	}
	if clusterName != "" {
		ls[constants.ClusterNameLabelKey] = clusterName
	}

	releases, err := c.rlsLister.List(labels.SelectorFromSet(ls))
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("list app release failed, error: %v", err)
		return nil, err
	}

	releases = filterReleases(releases, conditions)

	// only show release whose app versions are active or suspended
	if versionId == "" && strings.HasSuffix(appId, v1alpha1.HelmApplicationAppStoreSuffix) {
		stripId := strings.TrimSuffix(appId, v1alpha1.HelmApplicationAppStoreSuffix)
		versions, err := c.appVersionLister.List(labels.SelectorFromSet(map[string]string{constants.ChartApplicationIdLabelKey: stripId}))
		if err != nil {
			klog.Errorf("list app version failed, error: %s", err)
			return nil, err
		}
		versions = filterAppVersionByState(versions, []string{v1alpha1.StateActive, v1alpha1.StateSuspended})
		versionMap := make(map[string]*v1alpha1.HelmApplicationVersion)
		for _, version := range versions {
			versionMap[version.Name] = version
		}
		releases = filterReleasesWithAppVersions(releases, versionMap)
	}

	if reverse {
		sort.Sort(sort.Reverse(HelmReleaseList(releases)))
	} else {
		sort.Sort(HelmReleaseList(releases))
	}

	totalCount := len(releases)
	start, end := (&query.Pagination{Limit: limit, Offset: offset}).GetValidPagination(totalCount)
	releases = releases[start:end]
	items := make([]interface{}, 0, len(releases))
	for i := range releases {
		app := convertApplication(releases[i], nil)
		items = append(items, app)
	}

	return &models.PageableResponse{TotalCount: totalCount, Items: items}, nil
}

func (c *releaseOperator) DescribeApplication(workspace, clusterName, namespace, applicationId string) (*Application, error) {

	rls, err := c.rlsLister.Get(applicationId)

	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("list helm release failed, error: %v", err)
		return nil, err
	}

	app := &Application{}

	var clusterConfig string
	if rls != nil {
		// TODO check clusterName, workspace, namespace
		if clusterName != "" {
			cluster, err := c.clusterClients.Get(clusterName)
			if err != nil {
				klog.Errorf("get cluster config failed, error: %s", err)
				return nil, err
			}
			if !c.clusterClients.IsHostCluster(cluster) {
				clusterConfig, err = c.clusterClients.GetClusterKubeconfig(rls.GetRlsCluster())
				if err != nil {
					klog.Errorf("get cluster config failed, error: %s", err)
					return nil, err
				}
			}
		}

		// If clusterConfig is empty, this application will be installed in current host.
		hw := helmwrapper.NewHelmWrapper(clusterConfig, namespace, rls.Spec.Name)
		manifest, err := hw.Manifest()
		if err != nil {
			klog.Errorf("get manifest failed, error: %s", err)
		}
		infos, err := resourceparse.Parse(bytes.NewBufferString(manifest), namespace, rls.Spec.Name, true)
		if err != nil {
			klog.Errorf("parse resource failed, error: %s", err)
		}
		app = convertApplication(rls, infos)
	}

	return app, nil
}

func (c *releaseOperator) DeleteApplication(workspace, clusterName, namespace, id string) error {

	rls, err := c.rlsLister.Get(id)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("get release %s/%s failed, err: %s", namespace, id, err)
		return err
	}

	// TODO, check workspace, cluster and namespace
	if rls.GetWorkspace() != workspace || rls.GetRlsCluster() != clusterName || rls.GetRlsNamespace() != namespace {
	}

	err = c.rlsClient.Delete(context.TODO(), id, metav1.DeleteOptions{})

	if err != nil {
		klog.Errorf("delete release %s/%s failed, error: %s", namespace, id, err)
		return err
	} else {
		klog.V(2).Infof("delete release %s/%s", namespace, id)
	}

	return nil
}

// get app version from repo and helm application
func (c *releaseOperator) getAppVersion(repoId, id string) (ret *v1alpha1.HelmApplicationVersion, err error) {
	if ver, exists, _ := c.cachedRepos.GetAppVersion(id); exists {
		return ver, nil
	}

	if repoId != "" && repoId != v1alpha1.AppStoreRepoId {
		return nil, fmt.Errorf("app version not found")
	}
	ret, err = c.appVersionLister.Get(id)

	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error(err)
		return nil, err
	}
	return
}

// get app version from repo and helm application
func (c *releaseOperator) getAppVersionWithData(repoId, id string) (ret *v1alpha1.HelmApplicationVersion, err error) {
	if ver, exists, _ := c.cachedRepos.GetAppVersionWithData(id); exists {
		return ver, nil
	}

	if repoId != "" && repoId != v1alpha1.AppStoreRepoId {
		return nil, fmt.Errorf("not found")
	}
	ret, err = c.appVersionLister.Get(id)

	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error(err)
		return nil, err
	}
	return
}

// manifest
func (c *releaseOperator) CreateManifest(workspace, clusterName, namespace string, request CreateManifestRequest) error {
	version, err := c.getOperatorAppVersion(request.OperatorVersion)

	if err != nil {
		klog.Errorf("get operator version %s failed, error: %v", request.OperatorVersion, err)
		return err
	}

	exists, err := c.manifestExists(workspace, clusterName, namespace, clusterName)

	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("get manifest %s failed, error: %v", clusterName, err)
		return err
	}

	if exists {
		err = fmt.Errorf("manifest %s exists", request.Name)
		klog.Error(err)
		return err
	}

	manifest := &v1alpha1.Manifest{
		ObjectMeta: metav1.ObjectMeta{
			Name: request.Name,
			Annotations: map[string]string{
				constants.CreatorAnnotationKey: request.Username,
			},
			Labels: map[string]string{
				constants.WorkspaceLabelKey: request.Workspace,
				constants.NamespaceLabelKey: namespace,
			},
		},
		Spec: v1alpha1.ManifestSpec{
			Cluster:        clusterName,
			Namespace:      namespace,
			Kind:           request.Kind,
			Description:    request.Description,
			AppVersion:     request.OperatorVersion,
			Version:        request.Version,
			CustomResource: request.CustomResource,
		},
		Status: v1alpha1.ManifestStatus{},
	}

	if clusterName != "" {
		manifest.Labels[constants.ClusterNameLabelKey] = clusterName
	}

	// dbType: mysql, postgresql, clickhouse
	if dbType := version.GetOperatorVersionType(); dbType != "" {
		manifest.Labels[constants.OperatorAppLabelKey] = dbType
	}

	manifest, err = c.manifestClient.Create(context.TODO(), manifest, metav1.CreateOptions{})

	if err != nil {
		klog.Errorln(err)
		return err
	} else {
		klog.Infof("create manifest %s success in %s", request.Name, clusterName)
	}

	return nil
}

func (c *releaseOperator) DeleteManifest(workspace, clusterName, namespace, manifestName string) error {

	_, err := c.manifestLister.Get(manifestName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("get manifest %s failed, err: %s", manifestName, err)
		return err
	}

	err = c.manifestClient.Delete(context.TODO(), manifestName, metav1.DeleteOptions{})

	if err != nil {
		klog.Errorf("delete manifest %s failed, error: %s", manifestName, err)
		return err
	} else {
		klog.V(2).Infof("delete release %s", manifestName)
	}

	return nil
}

func (c *releaseOperator) ModifyManifest(request ModifyManifestRequest) error {

	if request.Description == "" {
		return nil
	}

	manifest, err := c.manifestLister.Get(request.Name)
	if err != nil {
		klog.Errorf("get release failed, error: %s", err)
		return err
	}

	manifestCopy := manifest.DeepCopy()
	manifestCopy.Spec.Description = stringutils.ShortenString(strings.TrimSpace(request.Description), v1alpha1.MsgLen)

	pt := client.MergeFrom(manifest)

	data, err := pt.Data(manifestCopy)
	if err != nil {
		klog.Errorf("create patch failed, error: %s", err)
		return err
	}

	_, err = c.rlsClient.Patch(context.TODO(), request.Name, pt.Type(), data, metav1.PatchOptions{})
	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}

func (c *releaseOperator) ListManifests(workspace, clusterName, namespace string, conditions *params.Conditions, limit, offset int, orderBy string, reverse bool) (*models.PageableResponse, error) {
	ls := map[string]string{}
	if workspace != "" {
		ls[constants.WorkspaceLabelKey] = workspace
	}
	if namespace != "" {
		ls[constants.NamespaceLabelKey] = namespace
	}
	if clusterName != "" {
		ls[constants.ClusterNameLabelKey] = clusterName
	}

	manifests, err := c.manifestLister.List(labels.SelectorFromSet(ls))
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("list app release failed, error: %v", err)
		return nil, err
	}

	manifests = filterManifests(manifests, conditions)

	if reverse {
		sort.Sort(sort.Reverse(ManifestList(manifests)))
	} else {
		sort.Sort(ManifestList(manifests))
	}

	totalCount := len(manifests)
	start, end := (&query.Pagination{Limit: limit, Offset: offset}).GetValidPagination(totalCount)
	manifests = manifests[start:end]
	items := make([]interface{}, 0, len(manifests))
	for i := range manifests {
		mft := convertManifest(manifests[i])
		items = append(items, mft)
	}

	return &models.PageableResponse{TotalCount: totalCount, Items: items}, nil
}

func (c *releaseOperator) DescribeManifest(workspace, clusterName, namespace, manifestName string) (*Manifest, error) {
	mft, err := c.manifestLister.Get(manifestName)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("list manifest failed, error: %v", err)
		return nil, err
	}

	app := &Manifest{}
	app = convertManifest(mft)
	return app, nil
}

func (c *releaseOperator) manifestExists(workspace, clusterName, namespace, name string) (bool, error) {
	set := map[string]string{
		constants.WorkspaceLabelKey: workspace,
		constants.NamespaceLabelKey: namespace,
	}
	if clusterName != "" {
		set[constants.ClusterNameLabelKey] = clusterName
	}

	list, err := c.manifestLister.List(labels.SelectorFromSet(set))
	if err != nil {
		return false, err
	}
	for _, manifest := range list {
		if manifest.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// get operator application version
func (c *releaseOperator) getOperatorAppVersion(versionName string) (ret *v1alpha1.OperatorApplicationVersion, err error) {
	ret, err = c.operatorVerLister.Get(versionName)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error(err)
		return nil, err
	}
	return
}
