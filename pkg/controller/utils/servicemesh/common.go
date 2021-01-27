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
	"istio.io/api/networking/v1alpha3"
	"k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// normalize version names
// strip [_.-]
func NormalizeVersionName(version string) string {
	for _, char := range TrimChars {
		version = strings.ReplaceAll(version, char, "")
	}
	return version
}

// Get deployment from service by service selector
func GetDeploymentsFromService(service *v12.Service, c client.Client) (deploys []*v1.Deployment) {
	if service == nil {
		return deploys
	}

	selectors := labels.Set(service.Spec.Selector).AsSelectorPreValidated()

	allDeploys := &v1.DeploymentList{}
	err := c.List(context.Background(), allDeploys, &client.ListOptions{Namespace: service.Namespace})
	if err != nil {
		klog.Errorf("list deployments in namespace %s failed", service.Namespace)
		return deploys
	}

	for i := range allDeploys.Items {
		deploy := allDeploys.Items[i]
		if selectors.Matches(labels.Set(deploy.Spec.Selector.MatchLabels)) {
			deploys = append(deploys, &deploy)
		}
	}
	if len(deploys) == 0 {
		klog.Infof("Get deployment from service %s spec failed", service.Name)
		return deploys
	}
	return deploys
}

// 1. Fetch deployments from service by the necessary application labels.
// 2. Deployment and service should satisfy servicemesh rules.
// 3. One service can have multi deployments.
func GetServicemeshDeploymentsFromService(service *v12.Service, c client.Client) (deploys []*v1.Deployment) {

	svc := Service{service}
	if !svc.IsObjServicemesh() {
		klog.V(4).Infof("service %s/%s is not servicemesh", service.Name, service.Namespace)
		return deploys
	}

	appLabels := ExtractApplicationLabels(service.GetLabels())

	ctx := context.TODO()
	deployments := &v1.DeploymentList{}
	if err := c.List(ctx, deployments, client.MatchingLabels(appLabels)); err != nil {
		klog.V(4).Infof("no such deployment with labels %v", appLabels)
		return deploys
	}

	for i := range deployments.Items {
		deploy := &deployments.Items[i]
		d := Deployment{deploy}
		if !d.IsObjServicemesh() {
			err := errors.New("deployment %s/%s is not a servicemesh", d.Name, d.Namespace)
			klog.Error(err)
			continue
		}
		if len(deploy.Name) != 0 {
			deploys = append(deploys, deploy)
		}
	}
	return deploys
}

// 1. Get service from deployment by necessary application lables.
// 2. Service and deployment should satisfy servicemesh rules.
// 3. One deployment can only have one service.
func GetServicemeshServiceFromDeployment(deploy *v1.Deployment, c client.Client) (service *v12.Service) {
	d := Deployment{deploy}
	if !d.IsObjServicemesh() {
		err := errors.New("deployment %s/%s is not a servicemesh", d.Name, d.Namespace)
		klog.V(4).Info(err)
		return service
	}

	appLabels := ExtractApplicationLabels(deploy.GetLabels())
	if appLabels == nil {
		klog.Errorf("deployment %s has invalid application labels", deploy.Name)
		return service
	}

	ctx := context.TODO()
	services := &v12.ServiceList{}
	if err := c.List(ctx, services, client.MatchingLabels(appLabels)); err != nil {
		klog.Errorf("Get no service for labels %v", appLabels)
		return service
	}
	// one deployment can only have one service
	if len(services.Items) > 1 {
		klog.Errorf("deployment %s/%s limit to one service, but has %d actually", deploy.Name, deploy.Namespace, len(services.Items))
		return service
	}
	service = &services.Items[0]
	s := Service{service}
	if !s.IsObjServicemesh() {
		klog.Errorf("service %s/%s is not a servicemesh", s.Name, s.Namespace)
		return
	}
	return service
}

// Get subsets from deployment
func GetDeploymentSubsets(deployments []*v1.Deployment) (subsets []*v1alpha3.Subset) {
	for i := range deployments {
		deployment := Deployment{Deployment: deployments[i]}
		if !deployment.IsObjServicemesh() {
			return subsets
		}

		version := GetComponentVersion(&deployment.ObjectMeta)

		if len(version) == 0 {
			klog.V(4).Infof("Deployment %s doesn't have a version label", types.NamespacedName{Namespace: deployment.Namespace, Name: deployment.Name}.String())
			continue
		}

		subset := &v1alpha3.Subset{
			Name:   NormalizeVersionName(version),
			Labels: map[string]string{VersionLabel: version},
		}

		if len(subset.Name) != 0 {
			subsets = append(subsets, subset)
		}
	}
	return subsets
}

// Update the deployment Labels and Annotations with service
func UpdateDeploymentLabelAndAnnotation(deploy *v1.Deployment, service *v12.Service) bool {
	if len(deploy.Labels) == 0 {
		deploy.Labels = make(map[string]string)
	}
	origin := deploy.DeepCopy()
	for _, l := range ApplicationLabels {
		deploy.Labels[l] = service.Labels[l]
	}
	ret := false
	// Deployment should have `version` label
	if len(deploy.Labels[VersionLabel]) == 0 {
		deploy.Labels[VersionLabel] = DefaultDeploymentVersion
	}

	if !reflect.DeepEqual(origin.Labels, deploy.Labels) {
		// deployment labels had been changed, should be updated
		ret = true
	}

	s := service.Annotations[ServiceMeshEnabledAnnotation]
	if len(deploy.Annotations) == 0 {
		annotations := make(map[string]string)
		deploy.Annotations = annotations
	}

	if deploy.Annotations[ServiceMeshEnabledAnnotation] != s {
		deploy.SetAnnotations(map[string]string{ServiceMeshEnabledAnnotation: s})
		// deployment annotation had been changed, should be updated
		ret = true
	}

	// needn't be updated
	return ret
}
