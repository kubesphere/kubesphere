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

package destinationrule

import (
	"context"
	apinetworkingv1alpha3 "istio.io/api/networking/v1alpha3"
	destinationrule "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
	"kubesphere.io/kubesphere/pkg/controller/utils/servicemesh"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new destinationrule Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDestinationRule{Client: mgr.GetClient(), scheme: mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor("destinationrule-controller")}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("destinationrule-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	sources := []runtime.Object{
		&appsv1.Deployment{},
		&corev1.Service{},
		&servicemeshv1alpha2.ServicePolicy{},
	}

	for _, s := range sources {
		// Watch for changes to destinationrule
		err = c.Watch(
			&source.Kind{Type: s},
			&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(
				func(h handler.MapObject) []reconcile.Request {
					return []reconcile.Request{{NamespacedName: types.NamespacedName{
						Name:      servicemesh.GetServicemeshName(h.Object, mgr),
						Namespace: h.Meta.GetNamespace()}}}
				})},
			predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					return servicemesh.IsServicemesh(mgr.GetClient(), e.ObjectNew, e.ObjectOld)
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return servicemesh.IsServicemesh(mgr.GetClient(), e.Object)
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return servicemesh.IsServicemesh(mgr.GetClient(), e.Object)
				},
			})
		if err != nil {
			return err
		}
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileDestinationRule{}

// ReconcileDeployment reconciles a DestinationRule object
type ReconcileDestinationRule struct {
	client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=networking.istio.io,resources=destinationrules,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileDestinationRule) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	if len(request.Name) == 0 {
		return reconcile.Result{}, nil
	}
	// Fetch the Application instance
	ctx := context.Background()
	service := &corev1.Service{}
	if err := r.Get(ctx, request.NamespacedName, service); err != nil {
		klog.Errorf("get service %s failed, err %v", request.NamespacedName, err)
		return reconcile.Result{}, nil
	}

	// fetch all deployments that match with service labels
	deployments := servicemesh.GetServicemeshDeploymentsFromService(service, r.Client)
	if len(deployments) == 0 {
		return reconcile.Result{}, nil
	}
	currentDestinationRule := &destinationrule.DestinationRule{}
	createDestinationRule := false
	err := r.Get(ctx, request.NamespacedName, currentDestinationRule)
	if err != nil {
		if errors.IsNotFound(err) {
			createDestinationRule = true
			currentDestinationRule = &destinationrule.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:        service.Name,
					Namespace:   service.Namespace,
					Labels:      service.Labels,
					Annotations: service.Annotations,
				},
				Spec: apinetworkingv1alpha3.DestinationRule{
					Host: request.Name,
				},
			}
		} else {
			klog.Errorf("Couldn't get destinationrule for service %s, err %v", request.Name, err)
			return reconcile.Result{}, err
		}
	}

	// fetch all servicepolicies associated to this service
	servicePolicy := &servicemeshv1alpha2.ServicePolicy{}
	_ = r.Get(ctx, types.NamespacedName{Namespace: request.Namespace, Name: request.Name}, servicePolicy)

	subsets := servicemesh.GetDeploymentSubsets(deployments)

	dr := currentDestinationRule.DeepCopy()
	dr.Spec.TrafficPolicy = nil
	dr.Spec.Subsets = subsets

	if len(servicePolicy.Name) != 0 {
		if servicePolicy.Spec.Template.Spec.TrafficPolicy != nil {
			dr.Spec.TrafficPolicy = servicePolicy.Spec.Template.Spec.TrafficPolicy
		}

		// not supported currently, can not add traffic for subsets from console
		for _, subset := range servicePolicy.Spec.Template.Spec.Subsets {
			for i := range dr.Spec.Subsets {
				if subset.Name == dr.Spec.Subsets[i].Name && subset.TrafficPolicy != nil {
					dr.Spec.Subsets[i].TrafficPolicy = subset.TrafficPolicy
				}
			}
		}
	}

	if !createDestinationRule &&
		reflect.DeepEqual(currentDestinationRule.Spec, dr.Spec) &&
		reflect.DeepEqual(currentDestinationRule.Labels, service.Labels) {
		klog.V(5).Info("destinationrule are equal, skipping update", "key", types.NamespacedName{Namespace: service.Namespace, Name: service.Name}.String())
		return reconcile.Result{}, nil
	}

	newDestinationRule := currentDestinationRule.DeepCopy()
	newDestinationRule.Spec = dr.Spec
	newDestinationRule.Labels = service.Labels
	if newDestinationRule.Annotations == nil {
		newDestinationRule.Annotations = make(map[string]string)
	}

	if createDestinationRule {
		err = r.Create(ctx, newDestinationRule)
	} else {
		err = r.Update(ctx, newDestinationRule)
	}
	if err != nil {
		if createDestinationRule && errors.IsForbidden(err) {
			// A request is forbidden primarily for two reasons:
			// 1. namespace is terminating, endpoint creation is not allowed by default.
			// 2. policy is misconfigured, in which case no service would function anywhere.
			// Given the frequency of 1, we log at a lower level.
			klog.V(5).Infof("Forbidden from creating endpoints, err: %v", err)
		}
		return reconcile.Result{}, err
	}
	klog.V(4).Infof("Successfully Reconciled destinationrule for service %s in namespace %s", request.Namespace, request.Name)
	return reconcile.Result{}, nil
}
