package openpitrix

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/go-openapi/strfmt"
	appsv1 "k8s.io/api/apps/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/klog"
	strings2 "k8s.io/utils/strings"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	typed_v1alpha1 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	listers_v1alpha1 "kubesphere.io/kubesphere/pkg/client/listers/application/v1alpha1"
	v1beta12 "kubesphere.io/kubesphere/pkg/client/listers/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/helpers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/helmwrapper"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/resourceparse"
	"math"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strings"
)

type ReleaseInterface interface {
	ListApplications(workspace, namespace string, conditions *params.Conditions, limit, offset int, orderBy string, reverse bool) (*models.PageableResponse, error)
	DescribeApplication(namespace, applicationId string) (*Application, error)
	CreateApplication(namespace string, request CreateClusterRequest) error
	ModifyApplication(request ModifyClusterAttributesRequest) error
	DeleteApplication(namespace, id string) error
	UpgradeApplication(request UpgradeClusterRequest) error
}

type releaseOperator struct {
	informers             informers.SharedInformerFactory
	rlsClient             typed_v1alpha1.ApplicationV1alpha1Interface
	rlsLister             listers_v1alpha1.HelmReleaseLister
	fedAppVersionLister   v1beta12.FederatedHelmApplicationVersionLister
	appVersionLister      listers_v1alpha1.HelmApplicationVersionLister
	cachedRepos           *cachedRepos
	useFederatedResources bool
}

func newReleaseOperator(cached *cachedRepos, k8sFactory informers.SharedInformerFactory, ksFactory externalversions.SharedInformerFactory, ksClient versioned.Interface, multiClusterEnabled bool) ReleaseInterface {
	c := &releaseOperator{
		informers:   k8sFactory,
		rlsClient:   ksClient.ApplicationV1alpha1(),
		rlsLister:   ksFactory.Application().V1alpha1().HelmReleases().Lister(),
		cachedRepos: cached,
	}

	if c.useFederatedResources {
		c.fedAppVersionLister = ksFactory.Types().V1beta1().FederatedHelmApplicationVersions().Lister()
	} else {
		c.appVersionLister = ksFactory.Application().V1alpha1().HelmApplicationVersions().Lister()
	}

	return c
}

type Application struct {
	Name    string      `json:"name" description:"application name"`
	Cluster *Cluster    `json:"cluster,omitempty" description:"application cluster info"`
	Version *AppVersion `json:"version,omitempty" description:"application template version info"`
	App     *App        `json:"app,omitempty" description:"application template info"`

	ReleaseInfo []runtime.Object `json:"releaseInfo,omitempty" description:"release info"`
	//WorkLoads *workLoads        `json:"workloads,omitempty" description:"application workloads"`
	//Services  []v1.Service      `json:"services,omitempty" description:"application services"`
	//Ingresses []v1beta1.Ingress `json:"ingresses,omitempty" description:"application ingresses"`
}

type workLoads struct {
	Deployments  []appsv1.Deployment  `json:"deployments,omitempty" description:"deployment list"`
	Statefulsets []appsv1.StatefulSet `json:"statefulsets,omitempty" description:"statefulset list"`
	Daemonsets   []appsv1.DaemonSet   `json:"daemonsets,omitempty" description:"daemonset list"`
}

func (c *releaseOperator) UpgradeApplication(request UpgradeClusterRequest) error {
	oldRls, err := c.rlsLister.HelmReleases(request.Namespace).Get(request.ClusterId)

	if err != nil {
		klog.Errorf("get release %s/%s failed, error: %s", request.Namespace, request.ClusterId, err)
		return err
	}

	if oldRls.Status.State != constants.HelmStatusActive {
		return errors.New("application is not active")
	}

	version, err := c.getAppVersion("", request.VersionId)
	if err != nil {
		klog.Infof("get helm application version %s/%s failed, error: %s", request.AppId, request.VersionId, err)
		return err
	}

	newRls := oldRls.DeepCopy()
	newRls.Spec.ApplicationId = request.AppId
	newRls.Spec.ApplicationVersionId = request.VersionId
	newRls.Spec.Workspace = ""

	newRls.Spec.Version += 1
	newRls.Spec.RepoId = version.GetHelmRepoId()
	newRls.Spec.ChartVersion = version.GetChartAppVersion()
	newRls.Spec.ChartAppVersion = version.GetChartAppVersion()
	// Use the new conf if the client has one, or server will just use the old conf.
	if request.Conf != "" {
		newRls.Spec.Values = strfmt.Base64(request.Conf)
	}

	patch := client.MergeFrom(oldRls)
	data, err := patch.Data(newRls)

	newRls, err = c.rlsClient.HelmReleases(request.Namespace).Patch(context.TODO(), request.ClusterId, patch.Type(), data, metav1.PatchOptions{})
	if err != nil {
		klog.Errorf("patch release %s/%s failed, error: %s", request.Namespace, request.ClusterId, err)
		return err
	} else {
		klog.V(2).Infof("patch release %s/%s success", request.Namespace, request.ClusterId)
	}

	return nil
}

//create all helm release in host cluster
func (c *releaseOperator) CreateApplication(namespace string, request CreateClusterRequest) error {
	version, err := c.getAppVersion("", request.VersionId)

	if err != nil {
		klog.Errorf("get helm application version %s failed, error: %v", request.Name, err)
		return err
	}

	exists, err := c.releaseExists(namespace, request.Name)

	if err != nil && !apiErrors.IsNotFound(err) {
		klog.Errorf("get helm release %s failed, error: %v", request.Name, err)
		return err
	}

	if exists {
		err = fmt.Errorf("release %s exists", request.Name)
		klog.Error(err)
		return err
	}

	var chartData strfmt.Base64
	rls := &v1alpha1.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name: idutils.GetUuid36(constants.HelmReleasePrefix),
			Annotations: map[string]string{
				constants.CreatorAnnotationKey: request.Username,
			},
			Labels: map[string]string{
				constants.ChartApplicationVersionIdLabelKey: request.VersionId,
				constants.ChartApplicationIdLabelKey:        strings.TrimSuffix(request.AppId, constants.HelmApplicationAppStoreSuffix),
				//constants.ClusterNameLabelKey:               clusterName, // add cluster name in ks-controller-manager
				constants.WorkspaceLabelKey: request.Workspace,
				constants.NamespaceLabelKey: namespace,
			},
		},
		Spec: v1alpha1.HelmReleaseSpec{
			ChartData:            chartData,
			Workspace:            request.Workspace,
			Name:                 request.Name,
			Version:              1,
			Values:               strfmt.Base64(request.Conf),
			ApplicationId:        request.AppId,
			ApplicationVersionId: request.VersionId,
			ChartName:            version.GetTrueName(),
			RepoId:               version.GetHelmRepoId(),
			ChartVersion:         version.GetChartVersion(),
			ChartAppVersion:      version.GetChartAppVersion(),
		},
	}

	rls, err = c.rlsClient.HelmReleases(namespace).Create(context.TODO(), rls, metav1.CreateOptions{})

	if err != nil {
		klog.Errorln(err)
		return err
	} else {
		klog.Infof("create helm release %s success in %s", request.Name, namespace)
	}

	return nil
}

func (c *releaseOperator) releaseExists(namespace, name string) (bool, error) {

	list, err := c.rlsLister.HelmReleases(namespace).List(labels.SelectorFromSet(map[string]string{}))
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

	rls, err := c.rlsLister.HelmReleases(request.Namespace).Get(request.ClusterID)
	if err != nil {
		klog.Errorf("get release failed, error: %s", err)
		return err
	}

	rlsCopy := rls.DeepCopy()
	rlsCopy.Spec.Description = strings2.ShortenString(strings.TrimSpace(*request.Description), constants.MsgLen)

	pt := client.MergeFrom(rls)

	data, err := pt.Data(rlsCopy)
	if err != nil {
		klog.Errorf("create patch failed, error: %s", err)
		return err
	}

	_, err = c.rlsClient.HelmReleases(request.Namespace).Patch(context.TODO(), request.ClusterID, pt.Type(), data, metav1.PatchOptions{})
	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}

func (c *releaseOperator) ListApplications(workspace, namespace string, conditions *params.Conditions, limit, offset int, orderBy string, reverse bool) (*models.PageableResponse, error) {
	appId := conditions.Match[AppId]
	versionId := conditions.Match[VersionId]
	ls := map[string]string{}
	if appId != "" {
		ls[constants.ChartApplicationIdLabelKey] = strings.TrimSuffix(appId, constants.HelmApplicationAppStoreSuffix)
	}

	if versionId != "" {
		ls[constants.ChartApplicationVersionIdLabelKey] = versionId
	}

	if workspace != "" {
		ls[constants.WorkspaceLabelKey] = workspace
	}
	if namespace != "" {
		ls[constants.NamespaceLabelKey] = namespace
	}

	releases, err := c.rlsLister.List(helpers.MapAsLabelSelector(ls))
	//releases, err := rlsLister.List(helpers.MapAsLabelSelector(ls))
	if err != nil && !apiErrors.IsNotFound(err) {
		klog.Errorf("list helm release failed, error: %v", err)
		return nil, err
	}

	if reverse {
		sort.Sort(sort.Reverse(HelmReleaseList(releases)))
	} else {
		sort.Sort(HelmReleaseList(releases))
	}

	result := models.PageableResponse{TotalCount: len(releases)}
	result.Items = make([]interface{}, 0, int(math.Min(float64(limit), float64(len(releases)))))

	cur := offset
	if !strings.HasSuffix(appId, constants.HelmApplicationAppStoreSuffix) {
		cur = 0
	}

	for i, j := 0, 0; cur < len(releases) && j < limit; {
		rls := releases[cur]
		version, _ := c.getAppVersion(rls.Spec.RepoId, rls.Spec.ApplicationVersionId)
		//Just output active and suspended helmapplicationversion's info
		if version != nil && strings.HasSuffix(appId, constants.HelmApplicationAppStoreSuffix) {
			//if version.Status.State == constants.HelmStatusActive || version.Status.State == constants.StateSuspended {
			//TODO: appversion status
			if true {
				if i >= offset {
					app := convertApplication("", rls, version, nil)
					result.Items = append(result.Items, app)
					j++
				}
				i++
			}
		} else {
			app := convertApplication("", rls, version, nil)
			result.Items = append(result.Items, app)
			j++
		}
		cur++
	}

	return &result, nil
}

func (c *releaseOperator) DescribeApplication(namespace, applicationId string) (*Application, error) {

	rls, err := c.rlsLister.HelmReleases(namespace).Get(applicationId)

	if err != nil && !apiErrors.IsNotFound(err) {
		klog.Errorf("list helm release failed, error: %v", err)
		return nil, err
	}

	app := &Application{}
	if rls != nil {
		version, _ := c.getAppVersion(rls.Spec.RepoId, rls.Spec.ApplicationVersionId)
		hw := helmwrapper.NewHelmWrapper("", namespace, rls.Spec.Name)
		manifest, err := hw.Manifest()
		if err != nil {
			klog.Errorf("get manifest failed, error: %s", err)
		}
		infos, err := resourceparse.Parse(bytes.NewBufferString(manifest), namespace, rls.Spec.Name, true)
		if err != nil {
			klog.Errorf("parse resource failed, error: %s", err)
		}
		app = convertApplication("", rls, version, infos)
	}

	return app, nil
}

func (c *releaseOperator) DeleteApplication(namespace, id string) error {
	//err := c.rlsClient.HelmReleases(namespace).Delete(context.TODO(), id, metav1.DeleteOptions{})
	//Helm release controller will update the status and uninstall release
	//Only delete helmRelease in host cluster, controller-manager will delete it from member cluster.
	err := c.rlsClient.HelmReleases(namespace).Delete(context.TODO(), id, metav1.DeleteOptions{})

	if err != nil {
		klog.Errorf("delete release %s/%s failed, error: %s", namespace, id, err)
		return err
	} else {
		klog.V(2).Infof("delete release %s/%s", namespace, id)
	}

	return nil
}

// get app version from repo and helm application
func (c *releaseOperator) getAppVersion(repoId, id string) (ret *v1beta1.FederatedHelmApplicationVersion, err error) {
	c.cachedRepos.RLock()
	ret = c.cachedRepos.versions[id]
	c.cachedRepos.RUnlock()
	if ret != nil {
		return
	}
	if repoId != "" && repoId != constants.AppStoreRepoId {
		return nil, fmt.Errorf("not found")
	}
	if c.useFederatedResources {
		ret, err = c.fedAppVersionLister.Get(id)
	} else {
		var appVersion *v1alpha1.HelmApplicationVersion
		appVersion, err = c.appVersionLister.Get(id)
		if appVersion != nil {
			ret = toFederateHelmApplicationVersion(appVersion)
		}
	}
	if err != nil && !apiErrors.IsNotFound(err) {
		klog.Error(err)
		return nil, err
	}
	return
}
