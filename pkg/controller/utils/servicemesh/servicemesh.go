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

package servicemesh

import (
	"context"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// When reconcile destinationrule or virtualservice, the rules below should be satisfied.
// Service and deployment should both have such labels and annotations:
//
// labels:
// ```bash
// app=xxx  # usually equal to the service name
// app.kubernetes.io/version=v1
// app.kubernetes.io/name=sample
// ```
//
// annotations: `ServiceDeployment.kubesphere.io/enabled: true`
//
type Interface interface {
	IsObjServicemesh() bool
}

type Service struct{ *corev1.Service }

func (service *Service) IsObjServicemesh() bool {
	if len(service.Labels) < len(ApplicationLabels) ||
		!IsApplicationComponent(service.Labels, ApplicationLabels) ||
		!IsServicemeshEnabled(service.Annotations) ||
		service.Spec.Selector == nil ||
		len(service.Spec.Ports) == 0 {
		// services don't have enough labels
		// or they don't have necessary labels
		// or they don't have ServiceDeployment enabled
		// or they don't have any ports defined
		return false
	}
	return true
}

type Deployment struct {
	*appsv1.Deployment
}

func (deployment *Deployment) IsObjServicemesh() bool {
	if !IsApplicationComponent(deployment.ObjectMeta.Labels, DeploymentLabels) ||
		!IsServicemeshEnabled(deployment.ObjectMeta.Annotations) {
		return false
	}
	return true
}

type ServiceDeployment struct {
	Deployments []*Deployment
	Service     *Service
}

func (svcDeploy *ServiceDeployment) IsObjServicemesh() bool {
	service := svcDeploy.Service
	if !service.IsObjServicemesh() {
		return false
	}

	deployments := svcDeploy.Deployments
	for i := range deployments {
		if !deployments[i].IsObjServicemesh() {
			return false
		}
	}

	return true
}

func IsServicemeshEnabled(annotations map[string]string) bool {
	if enabled, ok := annotations[ServiceMeshEnabledAnnotation]; ok {
		if enabled == "true" {
			return true
		}
	}
	return false
}

// Name of virtualservice and destinationrule is the same with its service name
func GetServicemeshName(obj runtime.Object, mgr manager.Manager) string {
	if deploy, ok := obj.(*appsv1.Deployment); ok {
		svc := GetServicemeshServiceFromDeployment(deploy, mgr.GetClient())
		if svc == nil {
			klog.Errorf("get service from deployment %s/%s failed", deploy.Namespace, deploy.Name)
			return ""
		}
		return svc.Name
	}

	if obj, ok := obj.(*corev1.Service); ok {
		return obj.Name
	}

	if obj, ok := obj.(*v1alpha2.ServicePolicy); ok {
		return obj.Name
	}

	if obj, ok := obj.(*v1alpha2.Strategy); ok {
		return obj.Labels[AppLabel]
	}

	klog.Errorf("get servicemesh name failed, unsupported object %v", obj)
	return ""
}

// Remove vitualservice/Destinationrule/servicepolicy/strategy when its service was not existing
func PurgeServicemeshObj(c client.Client, service *corev1.Service) (bool, error) {
	ctx := context.TODO()
	name := service.Name
	namespace := service.Namespace

	// delete strategy
	strategies := &v1alpha2.StrategyList{}
	err := c.List(ctx, strategies, client.MatchingLabels{AppLabel: service.Name}, client.InNamespace(service.Namespace))
	if strategies.Items != nil && len(strategies.Items) > 0 {
		for i := range strategies.Items {
			s := &v1alpha2.Strategy{
				ObjectMeta: v1.ObjectMeta{Namespace: service.Namespace, Name: strategies.Items[i].Name},
			}
			err = c.Delete(ctx, s)
			if err != nil {
				klog.V(6).Info(err)
			}
		}
	}

	// delete servicepolicy
	sp := &v1alpha2.ServicePolicy{
		ObjectMeta: v1.ObjectMeta{Namespace: namespace, Name: name},
	}
	err = c.Delete(ctx, sp)
	if err != nil {
		klog.V(6).Info(err)
	}

	// delete destinationrule
	dr := &v1alpha3.DestinationRule{
		ObjectMeta: v1.ObjectMeta{Namespace: namespace, Name: name},
	}
	err = c.Delete(ctx, dr)
	if err != nil {
		klog.V(6).Info(err)
	}

	// delete virtualservice
	vs := &v1alpha3.VirtualService{
		ObjectMeta: v1.ObjectMeta{Namespace: namespace, Name: name},
	}
	err = c.Delete(ctx, vs)
	if err != nil {
		klog.V(6).Info(err)
	}

	if err != nil {
		return false, err
	}
	return true, nil
}

// Whether the object satisfies the rules of servicemesh
func IsServicemesh(c client.Client, objs ...runtime.Object) bool {
	for _, obj := range objs {
		// whether deployment satisfied servicemesh rules
		if deploy, ok := obj.(*appsv1.Deployment); ok {
			serviceDeploy := Deployment{Deployment: deploy}
			if serviceDeploy.IsObjServicemesh() {
				return true
			}
		}
		// whether service satisfied servicemesh rules
		if service, ok := obj.(*corev1.Service); ok {
			s := Service{Service: service}
			if s.IsObjServicemesh() {
				return true
			}
			_, _ = PurgeServicemeshObj(c, service)
			return false
		}
		// whether servicepolicy satisfied servicemesh rules
		if _, ok := obj.(*v1alpha2.ServicePolicy); ok {
			return true
		}

		// whether strategy satisfied servicemesh rules
		if _, ok := obj.(*v1alpha2.Strategy); ok {
			return true
		}
	}
	return false
}
