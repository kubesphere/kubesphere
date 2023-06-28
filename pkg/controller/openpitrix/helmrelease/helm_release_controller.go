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
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/flowcontrol"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/api/application/v1alpha1"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmwrapper"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
)

const (
	HelmReleaseFinalizer = "helmrelease.application.kubesphere.io"
	MaxBackoffTime       = 15 * time.Minute
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
	// mock helm install && uninstall
	helmMock                  bool
	checkReleaseStatusBackoff *flowcontrol.Backoff

	clusterClients     clusterclient.ClusterClients
	MultiClusterEnable bool

	MaxConcurrent int
	// wait time when check release is ready or not
	WaitTime time.Duration

	StopChan <-chan struct{}
}

//	=========================>
//	^                         |
//	|        <==upgraded<==upgrading================
//	|        \      =========^                     /
//	|         |   /                               |
//
// creating=>created===>active=====>deleting=>deleted       |
//
//	\    ^           /                     |
//	 \   |  /======>                      /
//	  \=>failed<==========================
//
// Reconcile reads that state of the cluster for a helmreleases object and makes changes based on the state read
// and what is in the helmreleases.Spec
// +kubebuilder:rbac:groups=application.kubesphere.io,resources=helmreleases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=application.kubesphere.io,resources=helmreleases/status,verbs=get;update;patch
func (r *ReconcileHelmRelease) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
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
				return item == HelmReleaseFinalizer
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

	var err error
	switch instance.Status.State {
	case v1alpha1.HelmStatusDeleting:
		// no operation
		return reconcile.Result{}, nil
	case v1alpha1.HelmStatusFailed:
		// Release used to be failed, but instance.Status.Version not equal to instance.Spec.Version
		if instance.Status.Version > 0 && instance.Status.Version != instance.Spec.Version {
			return r.createOrUpgradeHelmRelease(instance, true)
		} else {
			return reconcile.Result{}, nil
		}
	case v1alpha1.HelmStatusActive:
		// Release used to be active, but instance.Status.Version not equal to instance.Spec.Version
		if instance.Status.Version != instance.Spec.Version {
			instance.Status.State = v1alpha1.HelmStatusUpgrading
			// Update the state first.
			err = r.Status().Update(context.TODO(), instance)
			return reconcile.Result{}, err
		} else {
			return reconcile.Result{}, nil
		}
	case v1alpha1.HelmStatusCreating:
		// create new release
		return r.createOrUpgradeHelmRelease(instance, false)
	case v1alpha1.HelmStatusUpgrading:
		// We can update the release now.
		return r.createOrUpgradeHelmRelease(instance, true)
	case v1alpha1.HelmStatusCreated, v1alpha1.HelmStatusUpgraded:
		if instance.Status.Version != instance.Spec.Version {
			// Start a new backoff.
			r.checkReleaseStatusBackoff.DeleteEntry(rlsBackoffKey(instance))

			instance.Status.State = v1alpha1.HelmStatusUpgrading
			err = r.Status().Update(context.TODO(), instance)
			return reconcile.Result{}, err
		} else {
			retry, err := r.checkReleaseIsReady(instance)
			return reconcile.Result{RequeueAfter: retry}, err
		}
	case v1alpha1.HelmStatusRollbacking:
		// TODO: rollback helm release
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

func rlsBackoffKey(rls *v1alpha1.HelmRelease) string {
	return rls.Name
}

// doCheck check whether helm release's resources are ready or not.
func (r *ReconcileHelmRelease) doCheck(rls *v1alpha1.HelmRelease) (retryAfter time.Duration, err error) {
	backoffKey := rlsBackoffKey(rls)
	clusterName := rls.GetRlsCluster()

	var clusterConfig string
	if r.MultiClusterEnable && clusterName != "" {
		clusterConfig, err = r.clusterClients.GetClusterKubeconfig(clusterName)
		if err != nil {
			klog.Errorf("get cluster %s config failed", clusterConfig)
			return
		}
	}

	hw := helmwrapper.NewHelmWrapper(clusterConfig, rls.GetRlsNamespace(), rls.Spec.Name,
		helmwrapper.SetMock(r.helmMock))

	ready, err := hw.IsReleaseReady(r.WaitTime)

	if err != nil {
		// release resources not ready
		klog.Errorf("check release %s/%s status failed, error: %s", rls.GetRlsNamespace(), rls.GetTrueName(), err)
		// check status next time
		r.checkReleaseStatusBackoff.Next(backoffKey, r.checkReleaseStatusBackoff.Clock.Now())
		retryAfter = r.checkReleaseStatusBackoff.Get(backoffKey)
		err := r.updateStatus(rls, rls.Status.State, err.Error())
		return retryAfter, err
	} else {
		klog.V(4).Infof("check release %s/%s status success, ready: %v", rls.GetRlsNamespace(), rls.GetTrueName(), ready)
		// install or upgrade success, remove the release from the queue.
		r.checkReleaseStatusBackoff.DeleteEntry(backoffKey)
		// Release resources are ready, it's active now.
		err := r.updateStatus(rls, v1alpha1.HelmStatusActive, "")
		// If update status failed, the controller need update the status next time.
		return 0, err
	}
}

// checkReleaseIsReady check whether helm release's are ready or not.
// If retryAfter > 0 , then the controller will recheck it next time.
func (r *ReconcileHelmRelease) checkReleaseIsReady(rls *v1alpha1.HelmRelease) (retryAfter time.Duration, err error) {
	backoffKey := rlsBackoffKey(rls)
	now := time.Now()
	if now.Sub(rls.Status.LastDeployed.Time) > MaxBackoffTime {
		klog.V(2).Infof("check release %s/%s too much times, ignore it", rls.GetRlsNamespace(), rls.GetTrueName())
		r.checkReleaseStatusBackoff.DeleteEntry(backoffKey)
		return 0, nil
	}

	if !r.checkReleaseStatusBackoff.IsInBackOffSinceUpdate(backoffKey, r.checkReleaseStatusBackoff.Clock.Now()) {
		klog.V(4).Infof("start to check release %s/%s status ", rls.GetRlsNamespace(), rls.GetTrueName())
		return r.doCheck(rls)
	} else {
		// backoff, check next time
		retryAfter := r.checkReleaseStatusBackoff.Get(backoffKey)
		klog.V(4).Infof("check release %s/%s status has been limited by backoff - %v remaining",
			rls.GetRlsNamespace(), rls.GetTrueName(), retryAfter)
		return retryAfter, nil
	}
}

func (r *ReconcileHelmRelease) updateStatus(rls *v1alpha1.HelmRelease, currentState, msg string) error {
	now := metav1.Now()
	var deployStatus v1alpha1.HelmReleaseDeployStatus
	rls.Status.Message = stringutils.ShortenString(msg, v1alpha1.MsgLen)

	deployStatus.Message = stringutils.ShortenString(msg, v1alpha1.MsgLen)
	deployStatus.State = currentState
	deployStatus.Time = now

	if rls.Status.State != currentState &&
		(currentState == v1alpha1.HelmStatusCreated || currentState == v1alpha1.HelmStatusUpgraded) {
		rls.Status.Version = rls.Spec.Version
		rls.Status.LastDeployed = &now
	}

	rls.Status.State = currentState
	// record then new state
	rls.Status.DeployStatus = append([]v1alpha1.HelmReleaseDeployStatus{deployStatus}, rls.Status.DeployStatus...)

	if len(rls.Status.DeployStatus) > 10 {
		rls.Status.DeployStatus = rls.Status.DeployStatus[:10:10]
	}

	rls.Status.LastUpdate = now
	err := r.Status().Update(context.TODO(), rls)

	return err
}

// createOrUpgradeHelmRelease will run helm install to install a new release if upgrade is false,
// run helm upgrade if upgrade is true
func (r *ReconcileHelmRelease) createOrUpgradeHelmRelease(rls *v1alpha1.HelmRelease, upgrade bool) (reconcile.Result, error) {

	// Install or upgrade release
	var chartData []byte
	var err error
	_, chartData, err = r.GetChartData(rls)
	if err != nil {
		return reconcile.Result{}, err
	}

	if len(chartData) == 0 {
		klog.Errorf("empty chart data failed, release name %s, chart name: %s", rls.Name, rls.Spec.ChartName)
		return reconcile.Result{}, ErrAppVersionDataIsEmpty
	}

	clusterName := rls.GetRlsCluster()

	var clusterConfig string
	if r.MultiClusterEnable && clusterName != "" {
		clusterConfig, err = r.clusterClients.GetClusterKubeconfig(clusterName)
		if err != nil {
			klog.Errorf("get cluster %s config failed", clusterConfig)
			return reconcile.Result{}, err
		}
	}

	// If clusterConfig is empty, this application will be installed in current host.
	hw := helmwrapper.NewHelmWrapper(clusterConfig, rls.GetRlsNamespace(), rls.Spec.Name,
		helmwrapper.SetAnnotations(map[string]string{constants.CreatorAnnotationKey: rls.GetCreator()}),
		helmwrapper.SetLabels(map[string]string{
			v1alpha1.ApplicationInstance: rls.GetTrueName(),
		}),
		helmwrapper.SetMock(r.helmMock))

	var currentState string
	if upgrade {
		err = hw.Upgrade(rls.Spec.ChartName, string(chartData), string(rls.Spec.Values))
		currentState = v1alpha1.HelmStatusUpgraded
	} else {
		err = hw.Install(rls.Spec.ChartName, string(chartData), string(rls.Spec.Values))
		currentState = v1alpha1.HelmStatusCreated
	}

	var msg string
	if err != nil {
		// install or upgrade failed
		currentState = v1alpha1.HelmStatusFailed
		msg = err.Error()
	}
	err = r.updateStatus(rls, currentState, msg)

	return reconcile.Result{}, err
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

	err = hw.Uninstall()

	return err
}

func (r *ReconcileHelmRelease) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	if r.KsFactory != nil && r.MultiClusterEnable {
		r.clusterClients = clusterclient.NewClusterClient(r.KsFactory.Cluster().V1alpha1().Clusters())
	}

	// exponential backoff
	r.checkReleaseStatusBackoff = flowcontrol.NewBackOff(2*time.Second, MaxBackoffTime)
	go wait.Until(r.checkReleaseStatusBackoff.GC, 1*time.Minute, r.StopChan)

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: r.MaxConcurrent}).
		For(&v1alpha1.HelmRelease{}).
		Complete(r)
}
