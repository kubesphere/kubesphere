/*
Copyright 2020 KubeSphere Authors

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

package application

import (
	"context"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta12 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
	"kubesphere.io/kubesphere/pkg/controller/virtualservice/util"
	"sigs.k8s.io/application/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

// Add creates a new Workspace Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileApplication{Client: mgr.GetClient(), scheme: mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor("application-controller")}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("application-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	sources := []runtime.Object{
		&v1.Deployment{},
		&corev1.Service{},
		&v1.StatefulSet{},
		&v1beta12.Ingress{},
		&servicemeshv1alpha2.ServicePolicy{},
		&servicemeshv1alpha2.Strategy{},
	}

	for _, s := range sources {
		// Watch for changes to Application
		err = c.Watch(&source.Kind{Type: s},
			&handler.EnqueueRequestForOwner{OwnerType: &v1beta1.Application{}, IsController: false},
			predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					return isApp(e.MetaOld)
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return isApp(e.Meta)
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return isApp(e.Meta)
				},
			})

		if err != nil {
			return err
		}
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileApplication{}

// ReconcileApplication reconciles a Workspace object
type ReconcileApplication struct {
	client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=app.k8s.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileApplication) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Application instance
	ctx := context.Background()
	app := &v1beta1.Application{}
	err := r.Get(ctx, request.NamespacedName, app)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// add specified annotation for app when triggered by sub-resources,
	// so the application in sigs.k8s.io can reconcile to update status
	annotations := app.GetObjectMeta().GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["kubesphere.io/last-updated"] = time.Now().String()
	app.SetAnnotations(annotations)
	err = r.Update(ctx, app)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(4).Info("application has been deleted during update")
			return reconcile.Result{}, nil
		}
	}
	return reconcile.Result{}, nil
}

func isApp(o metav1.Object) bool {
	if o.GetLabels() == nil || !util.IsApplicationComponent(o.GetLabels()) {
		return false
	}
	return true
}
