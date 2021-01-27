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

package sidecar

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	controllerutils "kubesphere.io/kubesphere/pkg/controller/utils/controller"
	"kubesphere.io/kubesphere/pkg/controller/utils/servicemesh"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	controllerName = "sidecar-controller"
)

// Reconciler reconciles a Deployment object
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
		r.Logger = ctrl.Log.WithName("controllers").WithName(controllerName)
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
		For(&v1.Deployment{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=apps,resources=deployment,verbs=get;list;watch;create;update;patch;delete
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = r.Logger.WithValues("sidecar-controller", req.NamespacedName)
	ctx := context.Background()
	deployment := &v1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// If annotation "servicemesh.kubesphere.io/enabled" doesn't exists, no need to reconcile
	value, ok := deployment.Annotations[servicemesh.ServiceMeshEnabledAnnotation]
	if !ok {
		return ctrl.Result{}, nil
	}

	deploy := &v1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, deploy)
	if err != nil {
		r.Logger.Error(err, "get deploy %s failed", req.NamespacedName)
		return ctrl.Result{}, err
	}

	// If the annotations are already the same, no need to reconcile
	if deploy.Spec.Template.Annotations[servicemesh.SidecarInjectAnnotation] ==
		deploy.Annotations[servicemesh.ServiceMeshEnabledAnnotation] {
		return ctrl.Result{}, nil
	}

	deploy.Spec.Template.SetAnnotations(map[string]string{servicemesh.SidecarInjectAnnotation: value})

	err = r.Update(ctx, deploy)
	if err != nil {
		r.Logger.Error(err, "update deploy failed ", "deployment", req.NamespacedName.String())
		return ctrl.Result{}, err
	}

	r.Recorder.Event(deployment, corev1.EventTypeNormal, controllerutils.SuccessSynced, controllerutils.MessageResourceSynced)
	return ctrl.Result{}, nil
}
