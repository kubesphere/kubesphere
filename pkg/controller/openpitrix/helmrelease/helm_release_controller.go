/*
Copyright 2019 The KubeSphere Authors.

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

package helmrelease

import (
	"context"
	"errors"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmrepoindex"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmwrapper"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"math"
	"path"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"
)

const (
	HelmReleaseFinalizer = "helmrelease.application.kubesphere.io"
	IndexerName          = "clusterNamespace"
)

var (
	ErrGetRepoFailed              = errors.New("get repo failed")
	ErrGetAppFailed               = errors.New("get app failed")
	ErrAppVersionDataIsEmpty      = errors.New("app version data is empty")
	ErrGetAppVersionFailed        = errors.New("get app version failed")
	ErrLoadChartFailed            = errors.New("load chart failed")
	ErrLoadChartFromStorageFailed = errors.New("load chart from storage failed")
)

var _ reconcile.Reconciler = &ReconcileHelmRelease{}

// ReconcileWorkspace reconciles a Workspace object
type ReconcileHelmRelease struct {
	StorageClient  s3.Interface
	KsFactory      externalversions.SharedInformerFactory
	clusterClients clusterclient.ClusterClients
	client.Client
	recorder record.EventRecorder
	// mock helm install && uninstall
	helmMock bool
	informer cache.SharedIndexInformer
}

//
//                 <==>upgrading===================
//               |                                 \
// creating===>active=====>deleting=>deleted       |
//          \    ^           /                     |
//           \   |  /======>                      /
//            \=>failed<==========================
// Reconcile reads that state of the cluster for a helmreleases object and makes changes based on the state read
// and what is in the helmreleases.Spec
// +kubebuilder:rbac:groups=application.kubesphere.io,resources=helmreleases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=application.kubesphere.io,resources=helmreleases/status,verbs=get;update;patch
func (r *ReconcileHelmRelease) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the helmReleases instance
	instance := &v1alpha1.HelmRelease{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.Status.State == "" {
		instance.Status.State = v1alpha1.HelmStatusCreating
		instance.Status.LastUpdate = metav1.Now()
		err = r.Status().Update(context.TODO(), instance)
		return reconcile.Result{}, err
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(instance.ObjectMeta.Finalizers, HelmReleaseFinalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, HelmReleaseFinalizer)
			// add owner References
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
	} else {
		// The object is being deleting
		if sliceutil.HasString(instance.ObjectMeta.Finalizers, HelmReleaseFinalizer) {

			klog.V(3).Infof("helm uninstall %s/%s from host cluster", instance.GetRlsNamespace(), instance.Spec.Name)
			err := r.uninstallHelmRelease(instance)
			if err != nil {
				return reconcile.Result{}, err
			}

			klog.V(3).Infof("remove helm release %s finalizer", instance.Name)
			// remove finalizer
			instance.ObjectMeta.Finalizers = sliceutil.RemoveString(instance.ObjectMeta.Finalizers, func(item string) bool {
				if item == HelmReleaseFinalizer {
					return true
				}
				return false
			})
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	return r.reconcile(instance)
}

func (r *ReconcileHelmRelease) GetChartData(rls *v1alpha1.HelmRelease) (chartName string, chartData []byte, err error) {
	if rls.Spec.RepoId != "" && rls.Spec.RepoId != v1alpha1.AppStoreRepoId {
		// load chart data from helm repo
		repo := v1alpha1.HelmRepo{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: rls.Spec.RepoId}, &repo)
		if err != nil {
			klog.Errorf("get helm repo %s failed, error: %v", rls.Spec.RepoId, err)
			return chartName, chartData, ErrGetRepoFailed
		}

		index, err := helmrepoindex.ByteArrayToSavedIndex([]byte(repo.Status.Data))

		if version := index.GetApplicationVersion(rls.Spec.ApplicationId, rls.Spec.ApplicationVersionId); version != nil {
			url := version.Spec.URLs[0]
			if !(strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "s3://")) {
				url = repo.Spec.Url + "/" + url
			}
			buf, err := helmrepoindex.LoadChart(context.TODO(), url, &repo.Spec.Credential)
			if err != nil {
				klog.Infof("load chart failed, error: %s", err)
				return chartName, chartData, ErrLoadChartFailed
			}
			chartData = buf.Bytes()
			chartName = version.Name
		} else {
			klog.Errorf("get app version: %s failed", rls.Spec.ApplicationVersionId)
			return chartName, chartData, ErrGetAppVersionFailed
		}
	} else {
		// load chart data from helm application version
		appVersion := &v1alpha1.HelmApplicationVersion{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: rls.Spec.ApplicationVersionId}, appVersion)
		if err != nil {
			klog.Errorf("get app version %s failed, error: %v", rls.Spec.ApplicationVersionId, err)
			return chartName, chartData, ErrGetAppVersionFailed
		}

		chartData, err = r.StorageClient.Read(path.Join(appVersion.GetWorkspace(), appVersion.Name))
		if err != nil {
			klog.Errorf("load chart from storage failed, error: %s", err)
			return chartName, chartData, ErrLoadChartFromStorageFailed
		}

		chartName = appVersion.GetTrueName()
	}
	return
}

func (r *ReconcileHelmRelease) reconcile(instance *v1alpha1.HelmRelease) (reconcile.Result, error) {

	if instance.Status.State == v1alpha1.HelmStatusActive && instance.Status.Version == instance.Spec.Version {
		// check release status
		return reconcile.Result{
			// recheck release status after 10 minutes
			RequeueAfter: 10 * time.Minute,
		}, nil
	}

	ft := failedTimes(instance.Status.DeployStatus)
	if v1alpha1.HelmStatusFailed == instance.Status.State && ft > 0 {
		// exponential backoff, max delay 180s
		retryAfter := time.Duration(math.Min(math.Exp2(float64(ft)), 180)) * time.Second
		var lastDeploy time.Time

		if instance.Status.LastDeployed != nil {
			lastDeploy = instance.Status.LastDeployed.Time
		} else {
			lastDeploy = instance.Status.LastUpdate.Time
		}
		if time.Now().Before(lastDeploy.Add(retryAfter)) {
			return reconcile.Result{RequeueAfter: retryAfter}, nil
		}
	}

	var err error
	switch instance.Status.State {
	case v1alpha1.HelmStatusDeleting:
		// no operation
		return reconcile.Result{}, nil
	case v1alpha1.HelmStatusActive:
		instance.Status.State = v1alpha1.HelmStatusUpgrading
		err = r.Status().Update(context.TODO(), instance)
		return reconcile.Result{}, err
	case v1alpha1.HelmStatusCreating:
		// create new release
		err = r.createOrUpgradeHelmRelease(instance, false)
	case v1alpha1.HelmStatusFailed:
		// check failed times
		err = r.createOrUpgradeHelmRelease(instance, false)
	case v1alpha1.HelmStatusUpgrading:
		err = r.createOrUpgradeHelmRelease(instance, true)
	case v1alpha1.HelmStatusRollbacking:
		// TODO: rollback helm release
	}

	now := metav1.Now()
	var deployStatus v1alpha1.HelmReleaseDeployStatus
	if err != nil {
		instance.Status.State = v1alpha1.HelmStatusFailed
		instance.Status.Message = stringutils.ShortenString(err.Error(), v1alpha1.MsgLen)
		deployStatus.Message = instance.Status.Message
		deployStatus.State = v1alpha1.HelmStatusFailed
	} else {
		instance.Status.State = v1alpha1.StateActive
		instance.Status.Message = ""
		instance.Status.Version = instance.Spec.Version
		deployStatus.State = v1alpha1.HelmStatusSuccessful
	}

	deployStatus.Time = now
	instance.Status.LastUpdate = now
	instance.Status.LastDeployed = &now
	if len(instance.Status.DeployStatus) > 0 {
		instance.Status.DeployStatus = append([]v1alpha1.HelmReleaseDeployStatus{deployStatus}, instance.Status.DeployStatus...)
		if len(instance.Status.DeployStatus) >= 10 {
			instance.Status.DeployStatus = instance.Status.DeployStatus[:10:10]
		}
	} else {
		instance.Status.DeployStatus = append([]v1alpha1.HelmReleaseDeployStatus{deployStatus})
	}

	err = r.Status().Update(context.TODO(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func failedTimes(status []v1alpha1.HelmReleaseDeployStatus) int {
	count := 0
	for i := range status {
		if status[i].State == v1alpha1.HelmStatusFailed {
			count += 1
		}
	}
	return count
}

func (r *ReconcileHelmRelease) createOrUpgradeHelmRelease(rls *v1alpha1.HelmRelease, upgrade bool) error {
	var chartData []byte
	var err error
	_, chartData, err = r.GetChartData(rls)
	if err != nil {
		return err
	}

	if len(chartData) == 0 {
		klog.Errorf("empty chart data failed, release name %s, chart name: %s", rls.Name, rls.Spec.ChartName)
		return ErrAppVersionDataIsEmpty
	}

	clusterName := rls.GetRlsCluster()

	var clusterConfig string
	if clusterName != "" && r.KsFactory != nil {
		clusterConfig, err = r.clusterClients.GetClusterKubeconfig(clusterName)
		if err != nil {
			klog.Errorf("get cluster %s config failed", clusterConfig)
			return err
		}
	}

	// If clusterConfig is empty, this application will be installed in current host.
	hw := helmwrapper.NewHelmWrapper(clusterConfig, rls.GetRlsNamespace(), rls.Spec.Name, helmwrapper.SetMock(r.helmMock))
	var res helmwrapper.HelmRes
	if upgrade {
		res, err = hw.Upgrade(rls.Spec.ChartName, string(chartData), string(rls.Spec.Values))
	} else {
		res, err = hw.Install(rls.Spec.ChartName, string(chartData), string(rls.Spec.Values))
	}
	if err != nil {
		return errors.New(res.Message)
	}
	return nil
}

func (r *ReconcileHelmRelease) uninstallHelmRelease(rls *v1alpha1.HelmRelease) error {
	if rls.Status.State != v1alpha1.HelmStatusDeleting {
		rls.Status.State = v1alpha1.HelmStatusDeleting
		rls.Status.LastUpdate = metav1.Now()
		err := r.Status().Update(context.TODO(), rls)
		if err != nil {
			return err
		}
	}

	clusterName := rls.GetRlsCluster()
	var clusterConfig string
	var err error
	if clusterName != "" && r.KsFactory != nil {
		clusterConfig, err = r.clusterClients.GetClusterKubeconfig(clusterName)
		if err != nil {
			klog.Errorf("get cluster %s config failed", clusterConfig)
			return err
		}
	}

	hw := helmwrapper.NewHelmWrapper(clusterConfig, rls.GetRlsNamespace(), rls.Spec.Name, helmwrapper.SetMock(r.helmMock))

	res, err := hw.Uninstall()

	if err != nil {
		return errors.New(res.Message)
	}
	return nil
}

func (r *ReconcileHelmRelease) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	if r.KsFactory != nil {
		r.clusterClients = clusterclient.NewClusterClient(r.KsFactory.Cluster().V1alpha1().Clusters())

		r.informer = r.KsFactory.Application().V1alpha1().HelmReleases().Informer()
		err := r.informer.AddIndexers(map[string]cache.IndexFunc{
			IndexerName: func(obj interface{}) ([]string, error) {
				rls := obj.(*v1alpha1.HelmRelease)
				return []string{fmt.Sprintf("%s/%s", rls.GetRlsCluster(), rls.GetRlsNamespace())}, nil
			},
		})
		if err != nil {
			return err
		}

		go func() {
			<-mgr.Elected()
			go r.cleanHelmReleaseWhenNamespaceDeleted()
		}()
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.HelmRelease{}).
		Complete(r)
}

func (r *ReconcileHelmRelease) getClusterConfig(cluster string) (string, error) {
	if cluster == "" {
		return "", nil
	}

	clusterConfig, err := r.clusterClients.GetClusterKubeconfig(cluster)
	if err != nil {
		klog.Errorf("get cluster %s config failed", clusterConfig)
		return "", err
	}

	return clusterConfig, nil
}

// When namespace have been removed from member cluster, we need clean all
// the helmRelease from the host cluster.
func (r *ReconcileHelmRelease) cleanHelmReleaseWhenNamespaceDeleted() {

	ticker := time.NewTicker(2 * time.Minute)
	for _ = range ticker.C {
		keys := r.informer.GetIndexer().ListIndexFuncValues(IndexerName)
		for _, clusterNs := range keys {
			klog.V(4).Infof("clean resource in %s", clusterNs)
			parts := stringutils.Split(clusterNs, "/")
			if len(parts) == 2 {
				cluster, ns := parts[0], parts[1]
				items, err := r.informer.GetIndexer().ByIndex(IndexerName, clusterNs)
				if err != nil {
					klog.Errorf("get items from index failed, error: %s", err)
					continue
				}

				kubeconfig, err := r.getClusterConfig(cluster)
				if err != nil {
					klog.Errorf("get cluster %s config failed, error: %s", cluster, err)
					continue
				}

				// connect to member or host cluster
				var restConfig *restclient.Config
				if kubeconfig == "" {
					restConfig, err = restclient.InClusterConfig()
				} else {
					cc, err := clientcmd.NewClientConfigFromBytes([]byte(kubeconfig))
					if err != nil {
						klog.Errorf("get client config for cluster %s failed, error: %s", cluster, err)
						continue
					}
					restConfig, err = cc.ClientConfig()
				}

				if err != nil {
					klog.Errorf("build rest config for cluster %s failed, error: %s", cluster, err)
					continue
				}

				clientSet, err := kubernetes.NewForConfig(restConfig)
				if err != nil {
					klog.Errorf("create client set failed, error: %s", err)
					continue
				}
				// check namespace exists or not
				namespace, err := clientSet.CoreV1().Namespaces().Get(context.TODO(), ns, metav1.GetOptions{})
				if err != nil {
					if apierrors.IsNotFound(err) {
						klog.V(2).Infof("delete all helm release in %s", clusterNs)
						for ind := range items {
							rls := items[ind].(*v1alpha1.HelmRelease)
							err := r.Client.Delete(context.TODO(), rls)
							if err != nil && !apierrors.IsNotFound(err) {
								klog.Errorf("delete release %s failed", rls.Name)
							}
						}
					} else {
						klog.Errorf("get namespace %s from cluster %s failed, error: %s", ns, cluster, err)
						continue
					}
				} else {
					for ind := range items {
						rls := items[ind].(*v1alpha1.HelmRelease)
						if namespace.CreationTimestamp.After(rls.CreationTimestamp.Time) {
							klog.V(2).Infof("delete helm release %s in %s", rls.Namespace, clusterNs)
							// todo, namespace is newer than helmRelease, should we delete the helmRelease
						}
					}
				}
			}
		}
	}
}
