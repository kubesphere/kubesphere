/*
Copyright 2019 The KubeSphere authors.

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

package strategy

import (
	"context"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

// Add creates a new Strategy Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileStrategy{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("strategy-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Strategy
	err = c.Watch(&source.Kind{Type: &servicemeshv1alpha2.Strategy{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &v1alpha3.VirtualService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Watch a VirtualService created by Strategy
	err = c.Watch(&source.Kind{Type: &v1alpha3.VirtualService{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &servicemeshv1alpha2.Strategy{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileStrategy{}

// ReconcileStrategy reconciles a Strategy object
type ReconcileStrategy struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Strategy object and makes changes based on the state read
// and what is in the Strategy.Spec
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servicemesh.kubesphere.io,resources=strategies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servicemesh.kubesphere.io,resources=strategies/status,verbs=get;update;patch
func (r *ReconcileStrategy) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Strategy instance
	strategy := &servicemeshv1alpha2.Strategy{}
	err := r.Get(context.TODO(), request.NamespacedName, strategy)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define VirtualService to be created
	vs := &v1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strategy.Name + "-virtualservice",
			Namespace: strategy.Namespace,
			Labels:    strategy.Spec.Selector.MatchLabels,
		},
		Spec: strategy.Spec.Template.Spec,
	}

	if err := controllerutil.SetControllerReference(strategy, vs, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if the VirtualService already exists
	found := &v1alpha3.VirtualService{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: vs.Name, Namespace: vs.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating VirtualService", "namespace", vs.Namespace, "name", vs.Name)
		err = r.Create(context.TODO(), vs)
		return reconcile.Result{}, err
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(vs.Spec, found.Spec) {
		found.Spec = vs.Spec
		log.Info("Updating VirtualService", "namespace", vs.Namespace, "name", vs.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}
