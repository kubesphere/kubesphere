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

package helmapplication

import (
	"context"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

const (
	HelmAppVersionFinalizer = "helmappversion.application.kubesphere.io"
)

var _ reconcile.Reconciler = &ReconcileHelmApplicationVersion{}

// ReconcileHelmApplicationVersion reconciles a helm application version object
type ReconcileHelmApplicationVersion struct {
	client.Client
	MultiClusterEnabled bool
}

// Reconcile reads that state of the cluster for a helmapplicationversions object and makes changes based on the state read
// and what is in the helmapplicationversions.Spec
func (r *ReconcileHelmApplicationVersion) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	start := time.Now()
	klog.Infof("sync helm application version: %s", request.String())
	defer func() {
		klog.Infof("sync helm application version end: %s, elapsed: %v", request.String(), time.Now().Sub(start))
	}()

	instance := &v1beta1.FederatedHelmApplicationVersion{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sliceutil.HasString(instance.ObjectMeta.Finalizers, HelmAppVersionFinalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, HelmAppVersionFinalizer)
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(instance.ObjectMeta.Finalizers, HelmAppVersionFinalizer) {
			// update related helm application
			err = updateHelmApplicationStatus(r.Client, instance, false)
			if err != nil {
				return reconcile.Result{}, err
			}

			err = updateHelmApplicationStatus(r.Client, instance, true)
			if err != nil {
				return reconcile.Result{}, err
			}

			//Delete HelmApplicationVersion
			instance.ObjectMeta.Finalizers = sliceutil.RemoveString(instance.ObjectMeta.Finalizers, func(item string) bool {
				if item == HelmAppVersionFinalizer {
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

	//update related helm application
	err = updateHelmApplicationStatus(r.Client, instance, false)
	if err != nil {
		return reconcile.Result{}, err
	}

	audit := v1alpha1.HelmAudit{}
	_ = r.Get(context.TODO(), types.NamespacedName{Name: instance.Name}, &audit)

	if audit.Spec.State == constants.StateActive {
		//add labels to helm application version
		if instance.GetHelmRepoId() == "" {
			instanceCopy := instance.DeepCopy()
			instanceCopy.Labels[constants.ChartRepoIdLabelKey] = constants.AppStoreRepoId
			patch := client.MergeFrom(instance)
			err = r.Client.Patch(context.TODO(), instanceCopy, patch)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		app := v1beta1.FederatedHelmApplication{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: instance.GetHelmApplicationId()}, &app)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, updateHelmApplicationStatus(r.Client, instance, true)
	} else if audit.Spec.State == constants.StateSuspended {
		return reconcile.Result{}, updateHelmApplicationStatus(r.Client, instance, true)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileHelmApplicationVersion) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	if !r.MultiClusterEnabled {
		_, err := mgr.GetRESTMapper().ResourceSingularizer(v1beta1.ResourcePluralFederatedHelmApplicationVersion)
		if err != nil {
			klog.Info("federated helm application version not exists, exit the controller")
			return nil
		}
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.FederatedHelmApplicationVersion{}).
		Complete(r)
}
