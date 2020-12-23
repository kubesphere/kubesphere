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
	"fmt"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	HelmAppFinalizer = "helmapplication.application.kubesphere.io"
)

var _ reconcile.Reconciler = &ReconcileAudit{}

// ReconcileHelmApplicationVersion reconciles a helm application version object
type ReconcileAudit struct {
	client.Client
	Scheme *runtime.Scheme
	//recorder           record.EventRecorder
	//config             *rest.Config
	MultiClusterEnable bool
}

func (r *ReconcileAudit) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	start := time.Now()
	klog.Infof("sync helm audit: %s", request.String())
	defer func() {
		klog.Infof("sync helm audit end: %s, elapsed: %v", request.String(), time.Now().Sub(start))
	}()

	instance := &v1alpha1.HelmAudit{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	appVersion := &v1beta1.FederatedHelmApplicationVersion{}
	appVersionErr := r.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Name}, appVersion)

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sliceutil.HasString(instance.ObjectMeta.Finalizers, HelmAppFinalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, HelmAppFinalizer)
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
	} else {
		if appVersionErr != nil && apiErrors.IsNotFound(appVersionErr) {
			return reconcile.Result{}, nil
		}
		// The object is being deleted
		if sliceutil.HasString(instance.ObjectMeta.Finalizers, HelmAppFinalizer) {
			// update related helm application
			//if appVersionErr == nil || apiErrors.IsNotFound(appVersionErr){}
			err = updateHelmApplicationStatus(r.Client, appVersion, false)
			if err != nil {
				return reconcile.Result{}, err
			}

			err = updateHelmApplicationStatus(r.Client, appVersion, true)
			if err != nil {
				return reconcile.Result{}, err
			}

			if sliceutil.HasString(appVersion.ObjectMeta.Finalizers, HelmAppFinalizer) {
				//Delete HelmApplicationVersion
				instance.ObjectMeta.Finalizers = sliceutil.RemoveString(appVersion.ObjectMeta.Finalizers, func(item string) bool {
					if item == HelmAppFinalizer {
						return true
					}
					return false
				})
				if err := r.Update(context.Background(), appVersion); err != nil {
					return reconcile.Result{}, err
				}
			}

			//Delete HelmAudit
			instance.ObjectMeta.Finalizers = sliceutil.RemoveString(instance.ObjectMeta.Finalizers, func(item string) bool {
				if item == HelmAppFinalizer {
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
	err = updateHelmApplicationStatus(r.Client, appVersion, false)
	if err != nil {
		return reconcile.Result{}, err
	}

	if instance.Spec.State == constants.StateSuspended || instance.Spec.State == constants.StateActive {
		err = updateHelmApplicationStatus(r.Client, appVersion, true)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func updateHelmApplicationStatus(c client.Client, version *v1beta1.FederatedHelmApplicationVersion, inAppStore bool) error {
	appId := version.GetHelmApplicationId()
	app := v1beta1.FederatedHelmApplication{}

	var err error
	if inAppStore {
		err = c.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("%s%s", appId, constants.HelmApplicationAppStoreSuffix)}, &app)
	} else {
		err = c.Get(context.TODO(), types.NamespacedName{Name: appId}, &app)
	}

	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	var versions v1beta1.FederatedHelmApplicationVersionList
	err = c.List(context.TODO(), &versions, client.MatchingLabels{
		constants.ChartApplicationIdLabelKey: appId,
	})

	if err != nil {
		return err
	}

	for ind := range versions.Items {
		versionId := versions.Items[ind].Name
		audit := v1alpha1.HelmAudit{}
		_ = c.Get(context.TODO(), types.NamespacedName{Name: versionId}, &audit)
		versions.Items[ind].Spec.Template.AuditSpec = audit.Spec
	}

	state := mergeApplicationVersionState(versions)
	template := &app.Spec.Template
	if state != template.Spec.Status {
		template.Spec.Status = state
		template.Spec.StatusTime = &metav1.Time{Time: time.Now()}
		err := c.Update(context.TODO(), &app)
		if err != nil {
			return err
		}
	}
	return nil
}

func mergeApplicationVersionState(versions v1beta1.FederatedHelmApplicationVersionList) string {
	states := make(map[string]int, len(versions.Items))

	for _, version := range versions.Items {
		if version.DeletionTimestamp == nil {
			state := version.Spec.Template.AuditSpec.State
			states[state] = states[state] + 1
		}
	}

	if states[constants.StateActive] > 0 {
		return constants.StateActive
	}

	if states[constants.StateDraft] == len(versions.Items) {
		return constants.StateDraft
	}

	if states[constants.StateSuspended] > 0 {
		return constants.StateSuspended
	}

	return constants.StateDraft
}

func (r *ReconcileAudit) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.HelmAudit{}).
		Complete(r)
}
