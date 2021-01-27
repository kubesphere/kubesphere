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

package virtualservice

import (
	"context"
	"fmt"
	apinetworkingv1alpha3 "istio.io/api/networking/v1alpha3"
	apisnetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
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
	"strings"
)

// Add creates a new virtualservice Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileVitualService{Client: mgr.GetClient(), scheme: mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor("virtualservice-controller")}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("virtualservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	sources := []runtime.Object{
		&appsv1.Deployment{},
		&corev1.Service{},
		&servicemeshv1alpha2.Strategy{},
	}

	for _, s := range sources {
		// Watch for changes to virtualservice
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

var _ reconcile.Reconciler = &ReconcileVitualService{}

// ReconcileDeployment reconciles a VirtualService object
type ReconcileVitualService struct {
	client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=networking.istio.io,resources=destinationrules,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileVitualService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
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

	// fetch all deployments that match with service selector
	deployments := servicemesh.GetServicemeshDeploymentsFromService(service, r.Client)
	if len(deployments) == 0 {
		klog.V(4).Infof("service %s is servicemesh, but its deployment is not servicemesh", request.NamespacedName)
		return reconcile.Result{}, nil
	}

	subsets := servicemesh.GetDeploymentSubsets(deployments)
	if len(subsets) == 0 {
		klog.V(4).Info("Get subsets failed")
		return reconcile.Result{}, nil
	}

	currentVirtualService := &apisnetworkingv1alpha3.VirtualService{}
	createVirtualService := false
	err := r.Get(ctx, request.NamespacedName, currentVirtualService)
	if err != nil {
		if errors.IsNotFound(err) {
			createVirtualService = true
			currentVirtualService = &apisnetworkingv1alpha3.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:        service.Name,
					Namespace:   service.Namespace,
					Labels:      servicemesh.ExtractApplicationLabels(service.Labels),
					Annotations: service.Annotations,
				},
			}
		} else {
			klog.Errorf("failed get virtualservice %s in namespace %s, err %v", request.Name, request.Namespace, err)
			return reconcile.Result{}, err
		}
	}

	vs := currentVirtualService.DeepCopy()
	vs.Spec.Hosts = []string{request.Name}

	// check if service has TCP protocol ports
	for _, port := range service.Spec.Ports {
		var route apinetworkingv1alpha3.HTTPRouteDestination
		if port.Protocol == corev1.ProtocolTCP {
			route = apinetworkingv1alpha3.HTTPRouteDestination{
				Destination: &apinetworkingv1alpha3.Destination{
					Host:   request.Name,
					Subset: subsets[0].Name,
					Port: &apinetworkingv1alpha3.PortSelector{
						Number: uint32(port.Port),
					},
				},
				Weight: 100,
			}

			// a http port, add to HTTPRoute
			if len(port.Name) > 0 && (port.Name == "http" || strings.HasPrefix(port.Name, "http-")) {
				vs.Spec.Http = []*apinetworkingv1alpha3.HTTPRoute{{Route: []*apinetworkingv1alpha3.HTTPRouteDestination{&route}}}
				break
			}

			// everything else treated as TCPRoute
			tcpRoute := apinetworkingv1alpha3.TCPRoute{
				Route: []*apinetworkingv1alpha3.RouteDestination{
					{
						Destination: route.Destination,
						Weight:      route.Weight,
					},
				},
			}
			vs.Spec.Tcp = []*apinetworkingv1alpha3.TCPRoute{&tcpRoute}
		}
	}

	// fetch all strategies associated to this service
	strategies := &servicemeshv1alpha2.StrategyList{}
	err = r.List(ctx, strategies, client.MatchingLabels{servicemesh.AppLabel: request.Name}, client.InNamespace(request.Namespace))
	if err != nil {
		klog.Errorf("list strategy for service %s failed", request.NamespacedName)
		return reconcile.Result{}, err
	}

	if len(strategies.Items) > 1 {
		err := fmt.Errorf("service %s can only have one strategy %v", request.NamespacedName, strategies.Items)
		klog.Error(err)
		return reconcile.Result{}, err
	}

	if len(strategies.Items) > 0 {
		strategy := &strategies.Items[0]
		// apply strategy spec to virtualservice
		switch strategy.Spec.StrategyPolicy {
		case servicemeshv1alpha2.PolicyPause:
			break
		case servicemeshv1alpha2.PolicyWaitForWorkloadReady:
			set := servicemesh.GetStrategySubsets(strategy)
			setNames := sets.String{}
			for i := range subsets {
				setNames.Insert(subsets[i].Name)
			}
			nonExist := false
			for k := range set {
				if !setNames.Has(k) {
					nonExist = true
				}
			}
			// strategy has subset that are not ready
			if nonExist {
				break
			} else {
				vs.Spec = servicemesh.GenerateVirtualServiceSpec(strategy, service).Spec
			}
		case servicemeshv1alpha2.PolicyImmediately:
			vs.Spec = servicemesh.GenerateVirtualServiceSpec(strategy, service).Spec
		default:
			vs.Spec = servicemesh.GenerateVirtualServiceSpec(strategy, service).Spec
		}
	}

	if !createVirtualService &&
		reflect.DeepEqual(vs.Spec, currentVirtualService.Spec) &&
		reflect.DeepEqual(service.Labels, currentVirtualService.Labels) {
		klog.V(4).Infof("virtual service %s are equal, skipping update ", request.NamespacedName)
		return reconcile.Result{}, nil
	}
	if len(vs.Spec.Http) == 0 && len(vs.Spec.Tcp) == 0 && len(vs.Spec.Tls) == 0 {
		err := fmt.Errorf("service %s doesn't have a valid port spec", request.NamespacedName)
		klog.Error(err)
		return reconcile.Result{}, err
	}

	if createVirtualService {
		err = r.Create(ctx, vs)
	} else {
		err = r.Update(ctx, vs)
	}

	if err != nil {
		klog.Error(err)
		return reconcile.Result{}, err
	}

	klog.V(4).Infof("Successfully Reconciled Virtualservice %s in namespace %s", request.Namespace, request.Name)
	return reconcile.Result{}, nil
}
