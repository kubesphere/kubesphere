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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmwrapper"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"math"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

const (
	HelmReleaseFinalizer = "helmrelease.application.kubesphere.io"
)

var (
	ErrGetRepoFailed              = errors.New("get repo failed")
	ErrGetAppFailed               = errors.New("get app failed")
	ErrAppVersionDataIsEmpty      = errors.New("app version data is empty")
	ErrGetAppVersionFailed        = errors.New("get app version failed")
	ErrLoadChartFailed            = errors.New("load chart failed")
	ErrS3Config                   = errors.New("invalid s3 config")
	ErrLoadChartFromStorageFailed = errors.New("load chart from storage failed")
)

var _ reconcile.Reconciler = &ReconcileHelmRelease{}

// ReconcileWorkspace reconciles a Workspace object
type ReconcileHelmRelease struct {
	StorageClient s3.Interface
	KsFactory     externalversions.SharedInformerFactory
	client.Client
	recorder record.EventRecorder
	// mock helm install && uninstall
	helmMock bool
	informer cache.SharedIndexInformer

	clusterClients     clusterclient.ClusterClients
	MultiClusterEnable bool
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
			clusterName := instance.GetRlsCluster()
			if r.MultiClusterEnable && clusterName != "" {
				clusterInfo, err := r.clusterClients.Get(clusterName)
				if err != nil {
					// cluster not exists, delete the crd
					klog.Warningf("cluster %s not found, delete the helm release %s/%s",
						clusterName, instance.GetRlsNamespace(), instance.GetTrueName())
					return reconcile.Result{}, r.Delete(context.TODO(), instance)
				}

				// Host cluster will self-healing, delete host cluster won't cause deletion of  helm release
				if !r.clusterClients.IsHostCluster(clusterInfo) {
					// add owner References
					instance.OwnerReferences = append(instance.OwnerReferences, metav1.OwnerReference{
						APIVersion: clusterv1alpha1.SchemeGroupVersion.String(),
						Kind:       clusterv1alpha1.ResourceKindCluster,
						Name:       clusterInfo.Name,
						UID:        clusterInfo.UID,
					})
				}
			}

			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, HelmReleaseFinalizer)
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

// Check the state of the instance then decide what to do.
func (r *ReconcileHelmRelease) reconcile(instance *v1alpha1.HelmRelease) (reconcile.Result, error) {

	if instance.Status.State == v1alpha1.HelmStatusActive && instance.Status.Version == instance.Spec.Version {
		// todo check release status
		return reconcile.Result{}, nil
	}

	ft := failedTimes(instance.Status.DeployStatus)
	if v1alpha1.HelmStatusFailed == instance.Status.State && ft > 0 {
		// failed too much times, exponential backoff, max delay 180s
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
		// Release used to be active, but instance.Status.Version not equal to instance.Spec.Version
		instance.Status.State = v1alpha1.HelmStatusUpgrading
		// Update the state first.
		err = r.Status().Update(context.TODO(), instance)
		return reconcile.Result{}, err
	case v1alpha1.HelmStatusCreating:
		// create new release
		err = r.createOrUpgradeHelmRelease(instance, false)
	case v1alpha1.HelmStatusFailed:
		err = r.createOrUpgradeHelmRelease(instance, false)
	case v1alpha1.HelmStatusUpgrading:
		// We can update the release now.
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
		// At most ten records will be saved.
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
	if r.MultiClusterEnable && clusterName != "" {
		clusterConfig, err = r.clusterClients.GetClusterKubeconfig(clusterName)
		if err != nil {
			klog.Errorf("get cluster %s config failed", clusterConfig)
			return err
		}
	}

	// If clusterConfig is empty, this application will be installed in current host.
	hw := helmwrapper.NewHelmWrapper(clusterConfig, rls.GetRlsNamespace(), rls.Spec.Name,
		// We just add kubesphere.io/creator annotation now.
		helmwrapper.SetAnnotations(map[string]string{constants.CreatorAnnotationKey: rls.GetCreator()}),
		helmwrapper.SetMock(r.helmMock))
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
	if r.MultiClusterEnable && clusterName != "" {
		clusterInfo, err := r.clusterClients.Get(clusterName)
		if err != nil {
			klog.V(2).Infof("cluster %s was deleted, skip helm release uninstall", clusterName)
			return nil
		}

		// If user deletes helmRelease first and then delete cluster immediately, this may cause helm resources leak.
		if clusterInfo.DeletionTimestamp != nil {
			klog.V(2).Infof("cluster %s is deleting, skip helm release uninstall", clusterName)
			return nil
		}

		clusterConfig = string(clusterInfo.Spec.Connection.KubeConfig)
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
	if r.KsFactory != nil && r.MultiClusterEnable {
		r.clusterClients = clusterclient.NewClusterClient(r.KsFactory.Cluster().V1alpha1().Clusters())
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.HelmRelease{}).
		Complete(r)
}
