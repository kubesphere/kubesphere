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

package routers

import (
	"fmt"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"

	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/informers"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"strings"

	"kubesphere.io/kubesphere/pkg/constants"
)

func GetAllRouters() ([]*corev1.Service, error) {

	selector := labels.SelectorFromSet(labels.Set{"app": "kubesphere", "component": "ks-router", "tier": "backend"})
	serviceLister := informers.SharedInformerFactory().Core().V1().Services().Lister()
	services, err := serviceLister.Services(constants.IngressControllerNamespace).List(selector)

	if err != nil {
		glog.Error(err)
		return nil, err
	}

	return services, nil
}

// Get router from a namespace
func GetRouter(namespace string) (*corev1.Service, error) {
	serviceName := constants.IngressControllerPrefix + namespace

	selector := labels.SelectorFromSet(labels.Set{"app": "kubesphere", "component": "ks-router", "tier": "backend", "project": namespace})
	serviceLister := informers.SharedInformerFactory().Core().V1().Services().Lister()
	services, err := serviceLister.Services(constants.IngressControllerNamespace).List(selector)

	if err != nil {
		glog.Error(err)
		return nil, err
	}
	for _, s := range services {
		if s.Name == serviceName {
			return s, nil
		}
	}

	return nil, fmt.Errorf("resources not found %s", serviceName)
}

// Load all resource yamls
func LoadYamls() ([]string, error) {

	var yamls []string

	files, err := ioutil.ReadDir(constants.IngressControllerFolder)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}
		content, err := ioutil.ReadFile(constants.IngressControllerFolder + "/" + file.Name())

		if err != nil {
			glog.Error(err)
			return nil, err
		} else {
			yamls = append(yamls, string(content))
		}
	}

	return yamls, nil
}

// Create a ingress controller in a namespace
func CreateRouter(namespace string, routerType corev1.ServiceType, annotations map[string]string) (*corev1.Service, error) {

	k8sClient := k8s.Client()

	var router *corev1.Service

	yamls, err := LoadYamls()

	if err != nil {
		glog.Error(err)
	}

	for _, f := range yamls {
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(f), nil, nil)

		if err != nil {
			glog.Error(err)
			return router, err
		}

		switch obj.(type) {
		case *corev1.Service:
			service := obj.(*corev1.Service)

			service.SetAnnotations(annotations)
			service.Spec.Type = routerType
			service.Name = constants.IngressControllerPrefix + namespace

			// Add project selector
			service.Labels["project"] = namespace

			service.Spec.Selector["project"] = namespace

			service, err := k8sClient.CoreV1().Services(constants.IngressControllerNamespace).Create(service)
			if err != nil {
				glog.Error(err)
				return nil, err
			}

			router = service

		case *extensionsv1beta1.Deployment:
			deployment := obj.(*extensionsv1beta1.Deployment)
			deployment.Name = constants.IngressControllerPrefix + namespace

			// Add project label
			deployment.Spec.Selector.MatchLabels["project"] = namespace
			deployment.Spec.Template.Labels["project"] = namespace

			// Isolate namespace
			deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--watch-namespace="+namespace)

			// Choose self as master
			deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--election-id="+deployment.Name)

			if routerType == corev1.ServiceTypeLoadBalancer {
				deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--publish-service="+constants.IngressControllerNamespace+"/"+constants.IngressControllerPrefix+namespace)
			} else {
				deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--report-node-internal-ip-address")
			}

			deployment, err := k8sClient.ExtensionsV1beta1().Deployments(constants.IngressControllerNamespace).Create(deployment)
			if err != nil {
				glog.Error(err)
			}
		default:
			//glog.Info("Default resource")
		}
	}

	return router, nil
}

// DeleteRouter is used to delete ingress controller related resources in namespace
// It will not delete ClusterRole resource cause it maybe used by other controllers
func DeleteRouter(namespace string) (*corev1.Service, error) {
	k8sClient := k8s.Client()

	var err error
	var router *corev1.Service

	if err != nil {
		glog.Error(err)
	}

	// delete controller service
	serviceName := constants.IngressControllerPrefix + namespace
	deleteOptions := meta_v1.DeleteOptions{}

	listOptions := meta_v1.ListOptions{
		LabelSelector: "app=kubesphere,component=ks-router,tier=backend,project=" + namespace,
		FieldSelector: "metadata.name=" + serviceName}

	serviceList, err := k8sClient.CoreV1().Services(constants.IngressControllerNamespace).List(listOptions)

	if err != nil {
		glog.Error(err)
	}

	if len(serviceList.Items) > 0 {
		router = &serviceList.Items[0]
		err = k8sClient.CoreV1().Services(constants.IngressControllerNamespace).Delete(serviceName, &deleteOptions)
		if err != nil {
			glog.Error(err)
		}
	}

	// delete controller deployment
	deploymentName := constants.IngressControllerPrefix + namespace

	listOptions = meta_v1.ListOptions{
		LabelSelector: "app=kubesphere,component=ks-router,tier=backend,project=" + namespace,
	}
	deployments, err := k8sClient.ExtensionsV1beta1().Deployments(constants.IngressControllerNamespace).List(listOptions)
	if err != nil {
		glog.Error(err)
	}

	if len(deployments.Items) > 0 {
		err = k8sClient.ExtensionsV1beta1().Deployments(constants.IngressControllerNamespace).Delete(deploymentName, &deleteOptions)

		if err != nil {
			glog.Error(err)
		}
	}

	return router, nil
}

// Update Ingress Controller Service, change type from NodePort to Loadbalancer or vice versa.
func UpdateRouter(namespace string, routerType corev1.ServiceType, annotations map[string]string) (*corev1.Service, error) {
	k8sClient := k8s.Client()

	var router *corev1.Service

	router, err := GetRouter(namespace)

	if err != nil {
		glog.Error(err)
		return router, nil
	}

	if router == nil {
		glog.Error("Trying to update a non-existed router")
		return nil, fmt.Errorf("router not created yet")
	}

	// from LoadBalancer to NodePort, or vice-versa
	if router.Spec.Type != routerType {
		router, err = DeleteRouter(namespace)

		if err != nil {
			glog.Error(err)
		}

		router, err = CreateRouter(namespace, routerType, annotations)

		if err != nil {
			glog.Error(err)
		}

		return router, err

	} else {
		router.SetAnnotations(annotations)

		router, err = k8sClient.CoreV1().Services(constants.IngressControllerNamespace).Update(router)

		if err != nil {
			glog.Error(err)
			return router, err
		}
	}

	return router, nil
}
