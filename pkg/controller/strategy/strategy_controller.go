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
	"fmt"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("strategy-controller")

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
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	return r.reconcileStrategy(strategy)
}

func (r *ReconcileStrategy) reconcileStrategy(strategy *servicemeshv1alpha2.Strategy) (reconcile.Result, error) {

	appName := getAppNameByStrategy(strategy)
	service := &v1.Service{}

	err := r.Get(context.TODO(), types.NamespacedName{Namespace: strategy.Namespace, Name: appName}, service)
	if err != nil {
		log.Error(err, "couldn't find service %s/%s,", strategy.Namespace, appName)
		return reconcile.Result{}, errors.NewBadRequest(fmt.Sprintf("service %s not found", appName))
	}

	vs, err := r.generateVirtualService(strategy, service)

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
	if !reflect.DeepEqual(vs.Spec, found.Spec) || len(vs.OwnerReferences) == 0 {
		found.Spec = vs.Spec
		found.OwnerReferences = vs.OwnerReferences
		log.Info("Updating VirtualService", "namespace", vs.Namespace, "name", vs.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileStrategy) generateVirtualService(strategy *servicemeshv1alpha2.Strategy, service *v1.Service) (*v1alpha3.VirtualService, error) {

	// Define VirtualService to be created
	vs := &v1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getAppNameByStrategy(strategy),
			Namespace: strategy.Namespace,
			Labels:    strategy.Spec.Selector.MatchLabels,
		},
		Spec: strategy.Spec.Template.Spec,
	}

	// one version rules them all
	if len(strategy.Spec.GovernorVersion) > 0 {

		governorDestinationWeight := v1alpha3.DestinationWeight{
			Destination: v1alpha3.Destination{
				Host:   getAppNameByStrategy(strategy),
				Subset: strategy.Spec.GovernorVersion,
			},
			Weight: 100,
		}

		if len(strategy.Spec.Template.Spec.Http) > 0 {
			governorRoute := v1alpha3.HTTPRoute{
				Route: []v1alpha3.DestinationWeight{governorDestinationWeight},
			}

			vs.Spec.Http = []v1alpha3.HTTPRoute{governorRoute}
		} else if len(strategy.Spec.Template.Spec.Tcp) > 0 {
			governorRoute := v1alpha3.TCPRoute{
				Route: []v1alpha3.DestinationWeight{governorDestinationWeight},
			}
			vs.Spec.Tcp = []v1alpha3.TCPRoute{governorRoute}
		}

	}

	if err := fillDestinationPort(vs, service); err != nil {
		return nil, err
	}

	return vs, nil
}
