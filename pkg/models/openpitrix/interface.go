package openpitrix

import (
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	ks_informers "kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/utils/helmrepoindex"
	"sync"
)

type Interface interface {
	AttachmentInterface
	ApplicationInterface
	RepoInterface
	ReleaseInterface
	CategoryInterface
}

type openpitrixOperator struct {
	AttachmentInterface
	ApplicationInterface
	RepoInterface
	ReleaseInterface
	CategoryInterface
}

var cachedReposData *cachedRepos
var helmRepoesInformer cache.SharedIndexInformer

func init() {
	cachedReposData = &cachedRepos{
		repos:    map[string]*v1alpha1.HelmRepo{},
		apps:     map[string]*v1beta1.FederatedHelmApplication{},
		versions: map[string]*v1beta1.FederatedHelmApplicationVersion{},
	}
}

func NewOpenpitrixOperator(ksInformers ks_informers.InformerFactory, k8sClient k8s.Client, ksClient versioned.Interface, multiClusterEnabled bool) Interface {
	useFederatedResource := false
	if multiClusterEnabled {
		useFederatedResource = true
	} else {
		resourceList, err := k8sClient.ApiExtensions().Discovery().ServerResourcesForGroupVersion("types.kubefed.io/v1beta1")
		if err == nil {
			for _, r := range resourceList.APIResources {
				if r.Name == v1beta1.ResourcePluralFederatedHelmApplication {
					useFederatedResource = true
				}
			}
		}
	}

	if useFederatedResource {
		klog.V(2).Infof("use federated helm application resource")
	}

	if helmRepoesInformer == nil {
		klog.Infof("start helm repo informer")
		helmRepoesInformer := ksInformers.KubeSphereSharedInformerFactory().Application().V1alpha1().HelmRepos().Informer()
		helmRepoesInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				r := obj.(*v1alpha1.HelmRepo)

				if len(r.Status.Data) == 0 {
					return
				}
				cachedReposData.Lock()
				defer cachedReposData.Unlock()
				index := &helmrepoindex.SavedIndex{}
				err := json.Unmarshal([]byte(r.Status.Data), index)
				if err != nil {
					klog.Errorf("json unmarshal failed, error: %s", err)
					return
				}

				cachedReposData.repos[r.GetHelmRepoId()] = r

				for key, app := range index.Applications {
					HelmApp := v1alpha1.HelmApplication{
						ObjectMeta: metav1.ObjectMeta{
							Name:        app.ApplicationId,
							Annotations: map[string]string{constants.CreatorAnnotationKey: r.Spec.Creator},
							Labels: map[string]string{
								constants.ChartRepoIdLabelKey: r.GetHelmRepoId(),
								//constants.ChartApplicationIdLabelKey: app.ApplicationId,
							},
						},
						Spec: v1alpha1.HelmApplicationSpec{
							Name:        key,
							Description: app.Description,
							Icon:        app.Icon,
							Status:      app.Status,
							//UpdateTime: metav1.Time{}
						},
					}
					cachedReposData.apps[app.ApplicationId] = toFederateHelmApplication(&HelmApp)

					for _, ver := range app.Charts {
						version := &v1alpha1.HelmApplicationVersion{
							ObjectMeta: metav1.ObjectMeta{
								Name:        ver.ApplicationVersionId,
								Annotations: map[string]string{constants.CreatorAnnotationKey: r.Spec.Creator},
								Labels: map[string]string{
									constants.ChartApplicationIdLabelKey: app.ApplicationId,
									constants.ChartRepoIdLabelKey:        r.GetHelmRepoId(),
								},
							},
							Spec: v1alpha1.HelmApplicationVersionSpec{
								//Creator: r.Spec.Creator,
								Metadata: &v1alpha1.Metadata{
									Name:       ver.Name,
									AppVersion: ver.AppVersion,
									Version:    ver.Version,
								},
								URLs:   ver.URLs,
								Digest: ver.Digest,
							},
						}
						cachedReposData.versions[ver.ApplicationVersionId] = toFederateHelmApplicationVersion(version)
					}
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				r := newObj.(*v1alpha1.HelmRepo)

				if len(r.Status.Data) == 0 {
					return
				}

				cachedReposData.Lock()
				defer cachedReposData.Unlock()
				index := &helmrepoindex.SavedIndex{}
				err := json.Unmarshal([]byte(r.Status.Data), index)
				if err != nil {
					klog.Errorf("json unmarshal failed, error: %s", err)
					return
				}

				cachedReposData.repos[r.GetHelmRepoId()] = r

				for key, app := range index.Applications {
					HelmApp := v1alpha1.HelmApplication{
						ObjectMeta: metav1.ObjectMeta{
							Name: app.ApplicationId,
							Annotations: map[string]string{
								constants.CreatorAnnotationKey: r.Spec.Creator,
							},
							Labels: map[string]string{
								constants.ChartRepoIdLabelKey:        r.GetHelmRepoId(),
								constants.ChartApplicationIdLabelKey: app.ApplicationId,
							},
						},
						Spec: v1alpha1.HelmApplicationSpec{
							Name:        key,
							Description: app.Description,
							Icon:        app.Icon,
							Status:      app.Status,
						},
						Status: v1alpha1.HelmApplicationStatus{
							State: constants.StateActive,
						},
					}
					cachedReposData.apps[app.ApplicationId] = toFederateHelmApplication(&HelmApp)

					for _, ver := range app.Charts {
						version := &v1alpha1.HelmApplicationVersion{
							ObjectMeta: metav1.ObjectMeta{
								Name:        ver.ApplicationVersionId,
								Annotations: map[string]string{constants.CreatorAnnotationKey: r.Spec.Creator},
								Labels: map[string]string{
									constants.ChartApplicationIdLabelKey: app.ApplicationId,
									constants.ChartRepoIdLabelKey:        r.GetHelmRepoId(),
								},
								CreationTimestamp: metav1.Time{Time: ver.Created},
							},
							Spec: v1alpha1.HelmApplicationVersionSpec{
								Metadata: &v1alpha1.Metadata{
									Name:       ver.Name,
									AppVersion: ver.AppVersion,
									Version:    ver.Version,
								},
								//Creator: r.Spec.Creator,
								URLs:   ver.URLs,
								Digest: ver.Digest,
							},
							Status: v1alpha1.HelmApplicationVersionStatus{
								State: constants.StateActive,
							},
						}
						cachedReposData.versions[ver.ApplicationVersionId] = toFederateHelmApplicationVersion(version)
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				r := obj.(*v1alpha1.HelmRepo)

				if len(r.Status.Data) == 0 {
					return
				}
				cachedReposData.Lock()
				defer cachedReposData.Unlock()
				index := &helmrepoindex.SavedIndex{}
				err := json.Unmarshal([]byte(r.Status.Data), index)
				if err != nil {
					klog.Errorf("json unmarshal failed, error: %s", err)
					return
				}
				delete(cachedReposData.repos, r.GetHelmRepoId())

				for _, app := range index.Applications {
					delete(cachedReposData.apps, app.ApplicationId)
					for _, ver := range app.Charts {
						delete(cachedReposData.versions, ver.ApplicationVersionId)
					}
				}
			},
		})
		go helmRepoesInformer.Run(wait.NeverStop)

		//ksInformers.KubeSphereSharedInformerFactory().
	}

	return &openpitrixOperator{
		AttachmentInterface:  newAttachmentOperator(ksInformers.KubernetesSharedInformerFactory()),
		ApplicationInterface: newApplicationOperator(cachedReposData, ksInformers.KubeSphereSharedInformerFactory(), ksClient, useFederatedResource),
		RepoInterface:        newRepoOperator(ksInformers.KubeSphereSharedInformerFactory(), ksClient),
		ReleaseInterface:     newReleaseOperator(cachedReposData, ksInformers.KubernetesSharedInformerFactory(), ksInformers.KubeSphereSharedInformerFactory(), ksClient, useFederatedResource),
		CategoryInterface:    newCategoryOperator(ksInformers.KubeSphereSharedInformerFactory(), ksClient),
	}
}

type cachedRepos struct {
	sync.RWMutex
	repos    map[string]*v1alpha1.HelmRepo
	apps     map[string]*v1beta1.FederatedHelmApplication
	versions map[string]*v1beta1.FederatedHelmApplicationVersion
}

func (c *cachedRepos) listAppVersions(appId string) (ret []*v1beta1.FederatedHelmApplicationVersion) {
	ret = make([]*v1beta1.FederatedHelmApplicationVersion, 0, 10)
	for _, ver := range c.versions {
		if ver.GetHelmApplicationId() == appId {
			ret = append(ret, ver)
		}
	}
	return ret
}
