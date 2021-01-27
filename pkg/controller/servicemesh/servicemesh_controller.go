/*
Copyright 2021 The KubeSphere Authors.

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

package servicemesh

import (
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	controllerutils "kubesphere.io/kubesphere/pkg/controller/utils/controller"
	"kubesphere.io/kubesphere/pkg/controller/utils/servicemesh"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	controllerName = "servicemesh-controller"
)

// When the service Labels and Annotations of ServiceMesh rules changed, its deployments will be reconciled.
// So that, the deployments have the same ServiceMesh Labels and Annotations.
// Reconciler reconciles a Service object.
type Reconciler struct {
	client.Client
	Logger                  logr.Logger
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.Logger == nil {
		r.Logger = ctrl.Log.WithName("servicemesh-controllers").WithName(controllerName)
	}
	if r.Recorder == nil {
		r.Recorder = mgr.GetEventRecorderFor(controllerName)
	}
	if r.MaxConcurrentReconciles <= 0 {
		r.MaxConcurrentReconciles = 1
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		For(&corev1.Service{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				return isApplication(e.MetaNew, e.MetaOld)
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return isApplication(e.Meta)
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return isApplication(e.Meta)
			},
		})).
		Complete(r)
}

// +kubebuilder:rbac:groups="",resources=service,verbs=get;list;watch;create;update;patch;delete
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	if len(req.Name) == 0 {
		return reconcile.Result{}, nil
	}
	_ = r.Logger.WithValues("servicemesh-controller", req.NamespacedName)
	ctx := context.Background()

	service := &corev1.Service{}
	if err := r.Get(ctx, req.NamespacedName, service); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	deploys := servicemesh.GetDeploymentsFromService(service, r.Client)

	if len(deploys) == 0 {
		return reconcile.Result{}, nil
	}
	for i := range deploys {
		deploy := deploys[i]
		if servicemesh.UpdateDeploymentLabelAndAnnotation(deploy, service) {
			err := r.Update(ctx, deploy)
			if err != nil {
				r.Logger.Error(err, "update deployment failed")
			}
			r.Recorder.Event(deploy, corev1.EventTypeNormal, controllerutils.SuccessSynced, controllerutils.MessageResourceSynced)
		}
	}

	return ctrl.Result{}, nil
}

func isApplication(obs ...metav1.Object) bool {
	for _, o := range obs {
		if o.GetLabels() != nil && servicemesh.IsApplicationComponent(o.GetLabels(), servicemesh.ApplicationLabels) {
			return true
		}
	}
	return false
}
