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
	"github.com/go-openapi/strfmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
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
	"math"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strings"
)

type ReleaseInterface interface {
	ListApplications(workspace, clusterName, namespace string, conditions *params.Conditions, limit, offset int, orderBy string, reverse bool) (*models.PageableResponse, error)
	DescribeApplication(workspace, clusterName, namespace, applicationId string) (*Application, error)
	CreateApplication(workspace, clusterName, namespace string, request CreateClusterRequest) error
	ModifyApplication(request ModifyClusterAttributesRequest) error
	DeleteApplication(workspace, clusterName, namespace, id string) error
	UpgradeApplication(request UpgradeClusterRequest) error
}

type releaseOperator struct {
	informers        informers.SharedInformerFactory
	rlsClient        typed_v1alpha1.HelmReleaseInterface
	rlsLister        listers_v1alpha1.HelmReleaseLister
	appVersionLister listers_v1alpha1.HelmApplicationVersionLister
	cachedRepos      reposcache.ReposCache
	clusterClients   clusterclient.ClusterClients
}

func newReleaseOperator(cached reposcache.ReposCache, k8sFactory informers.SharedInformerFactory, ksFactory externalversions.SharedInformerFactory, ksClient versioned.Interface) ReleaseInterface {
	c := &releaseOperator{
		informers:        k8sFactory,
		rlsClient:        ksClient.ApplicationV1alpha1().HelmReleases(),
		rlsLister:        ksFactory.Application().V1alpha1().HelmReleases().Lister(),
		cachedRepos:      cached,
		clusterClients:   clusterclient.NewClusterClient(ksFactory.Cluster().V1alpha1().Clusters()),
		appVersionLister: ksFactory.Application().V1alpha1().HelmApplicationVersions().Lister(),
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

func (c *releaseOperator) UpgradeApplication(request UpgradeClusterRequest) error {
	oldRls, err := c.rlsLister.Get(request.ClusterId)

	// todo check namespace
	if err != nil {
		klog.Errorf("get release %s/%s failed, error: %s", request.Namespace, request.ClusterId, err)
		return err
	}

	if oldRls.Status.State != v1alpha1.HelmStatusActive {
		return errors.New("application is not active now")
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

	result := models.PageableResponse{TotalCount: len(releases)}
	result.Items = make([]interface{}, 0, int(math.Min(float64(limit), float64(len(releases)))))

	for i, j := offset, 0; i < len(releases) && j < limit; i, j = i+1, j+1 {
		app := convertApplication(releases[i], nil)
		result.Items = append(result.Items, app)
	}

	return &result, nil
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
