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
	"encoding/json"
	"errors"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"k8s.io/utils/strings"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/helmrepoindex"
	"kubesphere.io/kubesphere/pkg/utils/helmwrapper"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
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
	ErrGetRepoFailed         = errors.New("get repo failed")
	ErrGetAppFailed          = errors.New("get app failed")
	ErrAppVersionDataIsEmpty = errors.New("app version data is empty")
	ErrGetAppVersionFailed   = errors.New("get app version failed")
	ErrLoadChartFailed       = errors.New("load chart failed")
)

var _ reconcile.Reconciler = &ReconcileHelmRelease{}

// ReconcileWorkspace reconciles a Workspace object
type ReconcileHelmRelease struct {
	MultiClusterEnabled  bool
	useFederatedResource bool
	client.Client
	Scheme *runtime.Scheme
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
		if api_errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.Status.State == "" {
		instance.Status.State = constants.HelmStatusCreating
		instance.Status.LastUpdate = metav1.Now()
		err = r.Status().Update(context.TODO(), instance)
		return reconcile.Result{}, err
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(instance.ObjectMeta.Finalizers, HelmReleaseFinalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, HelmReleaseFinalizer)
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
	} else {
		// The object is being deleting
		if sliceutil.HasString(instance.ObjectMeta.Finalizers, HelmReleaseFinalizer) {

			klog.V(3).Infof("helm uninstall %s/%s from host cluster", instance.Namespace, instance.Spec.Name)
			err := r.uninstallHelmRelease(instance)
			if err != nil {
				return reconcile.Result{}, err
			}

			klog.V(3).Infof("remove helm release %s finalizer", instance.Name)
			//remove finalizer
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
	if rls.Spec.RepoId != "" && rls.Spec.RepoId != constants.AppStoreRepoId {
		//load chart data from helm repo
		repo := v1alpha1.HelmRepo{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: rls.Spec.RepoId}, &repo)
		if err != nil {
			klog.Errorf("get helm repo [%s] failed, error: %v", rls.Spec.RepoId, err)
			return chartName, chartData, ErrGetRepoFailed
		}

		index := &helmrepoindex.SavedIndex{}
		err = json.Unmarshal([]byte(repo.Status.Data), index)

		if version := index.GetApplicationVersion(rls.Spec.ApplicationId, rls.Spec.ApplicationVersionId); version != nil {
			buf, err := helmrepoindex.LoadChart(context.TODO(), version.Spec.URLs[0], &repo.Spec.Credential)
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
		if r.useFederatedResource {
			fedAppVersion := &v1beta1.FederatedHelmApplicationVersion{}
			err = r.Get(context.TODO(), types.NamespacedName{Name: rls.Spec.ApplicationVersionId}, fedAppVersion)
			if err != nil {
				klog.Errorf("get app version %s failed, error: %v", rls.Spec.ApplicationVersionId, err)
				return chartName, chartData, ErrGetAppVersionFailed
			}
			chartData = []byte(fedAppVersion.Spec.Template.Spec.Data)
			chartName = fedAppVersion.GetTrueName()
		} else {
			//load chart data from helm application version
			appVersion := &v1alpha1.HelmApplicationVersion{}
			err = r.Get(context.TODO(), types.NamespacedName{Name: rls.Spec.ApplicationVersionId}, appVersion)
			if err != nil {
				klog.Errorf("get app version %s failed, error: %v", rls.Spec.ApplicationVersionId, err)
				return chartName, chartData, ErrGetAppVersionFailed
			}
			chartData = []byte(appVersion.Spec.Data)
			chartName = appVersion.GetTrueName()
		}
	}
	return
}

func (r *ReconcileHelmRelease) reconcile(instance *v1alpha1.HelmRelease) (reconcile.Result, error) {

	if instance.Status.State == constants.HelmStatusActive && instance.Status.Version == instance.Spec.Version {
		//check release status
		return reconcile.Result{
			//recheck release status after 10 minutes
			RequeueAfter: 10 * time.Minute,
		}, nil
	}

	ft := failedTimes(instance.Status.DeployStatus)
	if constants.HelmStatusFailed == instance.Status.State && ft > 0 {
		//exponential backoff, max delay 180s
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
	case constants.HelmStatusDeleting:
		//no operation
		return reconcile.Result{}, nil
	case constants.HelmStatusActive:
		if instance.Spec.Version != instance.Status.Version {
			instance.Status.State = constants.HelmStatusUpgrading
			err = r.Status().Update(context.TODO(), instance)
			return reconcile.Result{}, err
		}
	case constants.HelmStatusCreating:
		//create new release
		err = r.createOrUpgradeHelmRelease(instance, false)
	case constants.HelmStatusFailed:
		//check failed times
		err = r.createOrUpgradeHelmRelease(instance, false)
	case constants.HelmStatusUpgrading:
		err = r.createOrUpgradeHelmRelease(instance, true)
	case constants.HelmStatusRollbacking:
		//TODO: rollback helm release
	}

	now := metav1.Now()
	var deployStatus v1alpha1.HelmReleaseDeployStatus
	if err != nil {
		instance.Status.State = constants.HelmStatusFailed
		instance.Status.Message = strings.ShortenString(err.Error(), constants.MsgLen)
		deployStatus.Message = instance.Status.Message
		deployStatus.State = constants.HelmStatusFailed
	} else {
		instance.Status.State = constants.StateActive
		instance.Status.Message = ""
		instance.Status.Version = instance.Spec.Version
		deployStatus.State = constants.HelmStatusSuccessful
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
		if status[i].State == constants.HelmStatusFailed {
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

	//Helm install or helm upgrade in host cluster.
	hw := helmwrapper.NewHelmWrapper("", rls.Namespace, rls.Spec.Name)
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
	if rls.Status.State != constants.HelmStatusDeleting {
		rls.Status.State = constants.HelmStatusDeleting
		rls.Status.LastUpdate = metav1.Now()
		err := r.Status().Update(context.TODO(), rls)
		if err != nil {
			return err
		}
	}

	hw := helmwrapper.NewHelmWrapper("", rls.Namespace, rls.Spec.Name)
	res, err := hw.Uninstall()

	if err != nil {
		return errors.New(res.Message)
	}
	return nil
}

func (r *ReconcileHelmRelease) SetupWithManager(mgr ctrl.Manager) error {
	if !r.MultiClusterEnabled {
		_, err := mgr.GetRESTMapper().ResourceSingularizer(v1beta1.ResourcePluralFederatedHelmApplication)
		if err == nil {
			klog.Info("federated helm application exists, use federated resource")
			r.useFederatedResource = true
		}
	} else {
		klog.Info("multi cluster enabled, use federated resource")
		r.useFederatedResource = true
	}

	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.HelmRelease{}).
		Complete(r)
}
