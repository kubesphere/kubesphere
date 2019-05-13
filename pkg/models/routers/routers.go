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
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"sort"

	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/informers"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"strings"

	"kubesphere.io/kubesphere/pkg/constants"
)

// choose router node ip by labels, currently select master node
var RouterNodeIPLabelSelector = map[string]string{
	"node-role.kubernetes.io/master": "",
}

const (
	SERVICEMESH_ENABLED = "servicemesh.kubesphere.io/enabled"
	SIDECAR_INJECT      = "sidecar.istio.io/inject"
)

var routerTemplates map[string]runtime.Object

// Load yamls
func init() {
	yamls, err := LoadYamls()
	routerTemplates = make(map[string]runtime.Object, 2)

	if err != nil {
		glog.Error(err)
		return
	}

	for _, f := range yamls {
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(f), nil, nil)

		if err != nil {
			glog.Error(err)
			continue
		}

		switch obj.(type) {
		case *corev1.Service:
			routerTemplates["SERVICE"] = obj
		case *extensionsv1beta1.Deployment:
			routerTemplates["DEPLOYMENT"] = obj
		}
	}

}

// get master node ip, if there are multiple master nodes,
// choose first one according by their names alphabetically
func getMasterNodeIp() string {

	nodeLister := informers.SharedInformerFactory().Core().V1().Nodes().Lister()
	selector := labels.SelectorFromSet(RouterNodeIPLabelSelector)

	masters, err := nodeLister.List(selector)
	sort.Slice(masters, func(i, j int) bool {
		return strings.Compare(masters[i].Name, masters[j].Name) > 0
	})
	if err != nil {
		glog.Error(err)
		return ""
	}

	if len(masters) == 0 {
		return ""
	} else {
		for _, address := range masters[0].Status.Addresses {
			if address.Type == corev1.NodeInternalIP {
				return address.Address
			}
		}
	}

	return ""
}

func addLoadBalancerIp(service *corev1.Service) {

	// append selected node ip as loadbalancer ingress ip
	if service.Spec.Type != corev1.ServiceTypeLoadBalancer && len(service.Status.LoadBalancer.Ingress) == 0 {
		rip := getMasterNodeIp()
		if len(rip) == 0 {
			glog.Info("can not get node ip")
			return
		}

		gIngress := corev1.LoadBalancerIngress{
			IP: rip,
		}

		service.Status.LoadBalancer.Ingress = append(service.Status.LoadBalancer.Ingress, gIngress)
	}
}

func GetAllRouters() ([]*corev1.Service, error) {

	selector := labels.SelectorFromSet(labels.Set{"app": "kubesphere", "component": "ks-router", "tier": "backend"})
	serviceLister := informers.SharedInformerFactory().Core().V1().Services().Lister()
	services, err := serviceLister.Services(constants.IngressControllerNamespace).List(selector)

	for i := range services {
		addLoadBalancerIp(services[i])
	}

	if err != nil {
		glog.Error(err)
		return nil, err
	}

	return services, nil
}

// Get router from a namespace
func GetRouter(namespace string) (*corev1.Service, error) {
	service, err := getRouterService(namespace)
	addLoadBalancerIp(service)
	return service, err
}

func getRouterService(namespace string) (*corev1.Service, error) {
	serviceName := constants.IngressControllerPrefix + namespace

	serviceLister := informers.SharedInformerFactory().Core().V1().Services().Lister()
	service, err := serviceLister.Services(constants.IngressControllerNamespace).Get(serviceName)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.NewNotFound(corev1.Resource("service"), serviceName)
		}
		glog.Error(err)
		return nil, err
	}
	return service, nil
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

	injectSidecar := false
	if enabled, ok := annotations[SERVICEMESH_ENABLED]; ok {
		if enabled == "true" {
			injectSidecar = true
		}
	}

	err := createOrUpdateRouterWorkload(namespace, routerType == corev1.ServiceTypeLoadBalancer, injectSidecar)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	router, err := createRouterService(namespace, routerType, annotations)
	if err != nil {
		glog.Error(err)
		_ = deleteRouterWorkload(namespace)
		return nil, err
	}

	addLoadBalancerIp(router)
	return router, nil
}

// DeleteRouter is used to delete ingress controller related resources in namespace
// It will not delete ClusterRole resource cause it maybe used by other controllers
func DeleteRouter(namespace string) (*corev1.Service, error) {
	err := deleteRouterWorkload(namespace)
	if err != nil {
		glog.Error(err)
	}

	router, err := deleteRouterService(namespace)

	if err != nil {
		glog.Error(err)
		return router, err
	}
	return router, nil
}

func createRouterService(namespace string, routerType corev1.ServiceType, annotations map[string]string) (*corev1.Service, error) {

	obj, ok := routerTemplates["SERVICE"]
	if !ok {
		glog.Error("service template not loaded")
		return nil, fmt.Errorf("service template not loaded")
	}

	k8sClient := k8s.Client()

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

	return service, nil
}

func updateRouterService(namespace string, routerType corev1.ServiceType, annotations map[string]string) (*corev1.Service, error) {

	k8sClient := k8s.Client()

	service, err := getRouterService(namespace)
	if err != nil {
		glog.Error(err, "get router failed")
		return service, err
	}

	service.Spec.Type = routerType

	originalAnnotations := service.GetAnnotations()

	for key, val := range annotations {
		originalAnnotations[key] = val
	}

	service.SetAnnotations(originalAnnotations)

	service, err = k8sClient.CoreV1().Services(constants.IngressControllerNamespace).Update(service)

	return service, err
}

func deleteRouterService(namespace string) (*corev1.Service, error) {

	service, err := getRouterService(namespace)

	if err != nil {
		glog.Error(err)
		return service, err
	}

	k8sClient := k8s.Client()

	// delete controller service
	serviceName := constants.IngressControllerPrefix + namespace
	deleteOptions := meta_v1.DeleteOptions{}

	err = k8sClient.CoreV1().Services(constants.IngressControllerNamespace).Delete(serviceName, &deleteOptions)
	if err != nil {
		glog.Error(err)
		return service, err
	}

	return service, nil
}

func createOrUpdateRouterWorkload(namespace string, publishService bool, servicemeshEnabled bool) error {
	obj, ok := routerTemplates["DEPLOYMENT"]
	if !ok {
		glog.Error("Deployment template file not loaded")
		return fmt.Errorf("deployment template file not loaded")
	}

	k8sClient := k8s.Client()

	createDeployment := true
	deployment := obj.(*extensionsv1beta1.Deployment)
	deployment.Name = constants.IngressControllerPrefix + namespace

	// Add project label
	deployment.Spec.Selector.MatchLabels["project"] = namespace
	deployment.Spec.Template.Labels["project"] = namespace

	if servicemeshEnabled {
		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = make(map[string]string, 0)
		}
		deployment.Spec.Template.Annotations[SIDECAR_INJECT] = "true"
	}

	// Isolate namespace
	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--watch-namespace="+namespace)

	// Choose self as master
	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--election-id="+deployment.Name)

	if publishService {
		deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--publish-service="+constants.IngressControllerNamespace+"/"+constants.IngressControllerPrefix+namespace)
	} else {
		deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--report-node-internal-ip-address")
	}

	_, err := k8sClient.ExtensionsV1beta1().Deployments(constants.IngressControllerNamespace).Get(deployment.Name, meta_v1.GetOptions{})

	if err == nil {
		createDeployment = false
	}

	if createDeployment {
		deployment, err = k8sClient.ExtensionsV1beta1().Deployments(constants.IngressControllerNamespace).Create(deployment)
	} else {
		deployment, err = k8sClient.ExtensionsV1beta1().Deployments(constants.IngressControllerNamespace).Update(deployment)
	}

	if err != nil {
		glog.Error(err)
		return err
	}

	return nil
}

func deleteRouterWorkload(namespace string) error {
	k8sClient := k8s.Client()

	deleteOptions := meta_v1.DeleteOptions{}
	// delete controller deployment
	deploymentName := constants.IngressControllerPrefix + namespace
	err := k8sClient.ExtensionsV1beta1().Deployments(constants.IngressControllerNamespace).Delete(deploymentName, &deleteOptions)
	if err != nil {
		glog.Error(err)
	}

	// delete replicaset if there are any
	selector := labels.SelectorFromSet(
		labels.Set{
			"app":       "kubesphere",
			"component": "ks-router",
			"tier":      "backend",
			"project":   namespace,
		})
	replicaSetLister := informers.SharedInformerFactory().Apps().V1().ReplicaSets().Lister()
	replicaSets, err := replicaSetLister.ReplicaSets(constants.IngressControllerNamespace).List(selector)

	if err != nil {
		glog.Error(err)
	}

	for i := range replicaSets {
		err = k8sClient.AppsV1().ReplicaSets(constants.IngressControllerNamespace).Delete(replicaSets[i].Name, &deleteOptions)
		if err != nil {
			glog.Error(err)
		}
	}

	return nil
}

// Update Ingress Controller Service, change type from NodePort to loadbalancer or vice versa.
func UpdateRouter(namespace string, routerType corev1.ServiceType, annotations map[string]string) (*corev1.Service, error) {
	var router *corev1.Service

	router, err := getRouterService(namespace)

	if err != nil {
		glog.Error(err)
		return router, err
	}

	enableServicemesh := annotations[SERVICEMESH_ENABLED] == "true"

	err = createOrUpdateRouterWorkload(namespace, routerType == corev1.ServiceTypeLoadBalancer, enableServicemesh)
	if err != nil {
		glog.Error(err)
		return router, err
	}

	newRouter, err := updateRouterService(namespace, routerType, annotations)

	if err != nil {
		glog.Error(err)
		return newRouter, err
	}

	return newRouter, nil
}
