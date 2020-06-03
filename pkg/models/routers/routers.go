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
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog"
	"sort"
	"strings"
)

// choose router node ip by labels, currently select master node
var routerNodeIPLabelSelector = map[string]string{
	"node-role.kubernetes.io/master": "",
}

const (
	servicemeshEnabled         = "servicemesh.kubesphere.io/enabled"
	sidecarInject              = "sidecar.istio.io/inject"
	ingressControllerFolder    = "/etc/kubesphere/ingress-controller"
	ingressControllerPrefix    = "kubesphere-router-"
	ingressControllerNamespace = "kubesphere-controls-system"
	configMapSuffix            = "-nginx"
)

type RouterOperator interface {
	GetRouter(namespace string) (*corev1.Service, error)
	CreateRouter(namespace string, serviceType corev1.ServiceType, annotations map[string]string) (*corev1.Service, error)
	DeleteRouter(namespace string) (*corev1.Service, error)
	UpdateRouter(namespace string, serviceType corev1.ServiceType, annotations map[string]string) (*corev1.Service, error)
}

type routerOperator struct {
	routerTemplates map[string]runtime.Object
	client          kubernetes.Interface
	informers       informers.SharedInformerFactory
}

func NewRouterOperator(client kubernetes.Interface, informers informers.SharedInformerFactory) RouterOperator {
	yamls, err := loadYamls()
	routerTemplates := make(map[string]runtime.Object, 2)

	if err != nil {
		klog.Errorf("error happened during loading external yamls, %v", err)
	}

	for _, f := range yamls {
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(f), nil, nil)

		if err != nil {
			klog.Error(err)
			continue
		}

		switch obj.(type) {
		case *corev1.Service:
			routerTemplates["SERVICE"] = obj
		case *v1.Deployment:
			routerTemplates["DEPLOYMENT"] = obj
		}
	}

	return &routerOperator{
		client:          client,
		informers:       informers,
		routerTemplates: routerTemplates,
	}
}

// get master node ip, if there are multiple master nodes,
// choose first one according by their names alphabetically
func (c *routerOperator) getMasterNodeIp() string {

	nodeLister := c.informers.Core().V1().Nodes().Lister()
	selector := labels.SelectorFromSet(routerNodeIPLabelSelector)

	masters, err := nodeLister.List(selector)
	sort.Slice(masters, func(i, j int) bool {
		return strings.Compare(masters[i].Name, masters[j].Name) > 0
	})
	if err != nil {
		klog.Error(err)
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

func (c *routerOperator) addLoadBalancerIp(service *corev1.Service) {

	if service == nil {
		return
	}

	// append selected node ip as loadbalancer ingress ip
	if service.Spec.Type != corev1.ServiceTypeLoadBalancer && len(service.Status.LoadBalancer.Ingress) == 0 {
		rip := c.getMasterNodeIp()
		if len(rip) == 0 {
			klog.Info("can not get node ip")
			return
		}

		gIngress := corev1.LoadBalancerIngress{
			IP: rip,
		}

		service.Status.LoadBalancer.Ingress = append(service.Status.LoadBalancer.Ingress, gIngress)
	}
}

// Get router from a namespace
func (c *routerOperator) GetRouter(namespace string) (*corev1.Service, error) {
	service, err := c.getRouterService(namespace)
	c.addLoadBalancerIp(service)
	return service, err
}

func (c *routerOperator) getRouterService(namespace string) (*corev1.Service, error) {
	serviceName := ingressControllerPrefix + namespace
	serviceLister := c.informers.Core().V1().Services().Lister()
	service, err := serviceLister.Services(ingressControllerNamespace).Get(serviceName)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.NewNotFound(corev1.Resource("service"), serviceName)
		}
		klog.Error(err)
		return nil, err
	}
	return service, nil
}

// Load all resource yamls
func loadYamls() ([]string, error) {
	var yamls []string
	files, err := ioutil.ReadDir(ingressControllerFolder)
	if err != nil {
		klog.Warning(err)
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}
		content, err := ioutil.ReadFile(ingressControllerFolder + "/" + file.Name())

		if err != nil {
			klog.Error(err)
			return nil, err
		} else {
			yamls = append(yamls, string(content))
		}
	}

	return yamls, nil
}

// Create a ingress controller in a namespace
func (c *routerOperator) CreateRouter(namespace string, routerType corev1.ServiceType, annotations map[string]string) (*corev1.Service, error) {

	injectSidecar := false
	if enabled, ok := annotations[servicemeshEnabled]; ok {
		if enabled == "true" {
			injectSidecar = true
		}
	}

	err := c.createOrUpdateRouterWorkload(namespace, routerType == corev1.ServiceTypeLoadBalancer, injectSidecar)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	router, err := c.createRouterService(namespace, routerType, annotations)
	if err != nil {
		klog.Error(err)
		_ = c.deleteRouterWorkload(namespace)
		return nil, err
	}

	c.addLoadBalancerIp(router)
	return router, nil
}

// DeleteRouter is used to delete ingress controller related resources in namespace
// It will not delete ClusterRole resource cause it maybe used by other controllers
func (c *routerOperator) DeleteRouter(namespace string) (*corev1.Service, error) {
	err := c.deleteRouterWorkload(namespace)
	if err != nil {
		klog.Error(err)
	}

	router, err := c.deleteRouterService(namespace)

	if err != nil {
		klog.Error(err)
		return router, err
	}
	return router, nil
}

func (c *routerOperator) createRouterService(namespace string, routerType corev1.ServiceType, annotations map[string]string) (*corev1.Service, error) {

	obj, ok := c.routerTemplates["SERVICE"]
	if !ok {
		klog.Error("service template not loaded")
		return nil, fmt.Errorf("service template not loaded")
	}

	service := obj.(*corev1.Service)
	service.SetAnnotations(annotations)
	service.Spec.Type = routerType
	service.Name = ingressControllerPrefix + namespace

	// Add project selector
	service.Labels["project"] = namespace
	service.Spec.Selector["project"] = namespace
	service, err := c.client.CoreV1().Services(ingressControllerNamespace).Create(service)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return service, nil
}

func (c *routerOperator) updateRouterService(namespace string, routerType corev1.ServiceType, annotations map[string]string) (*corev1.Service, error) {
	service, err := c.getRouterService(namespace)
	if err != nil {
		klog.Error(err, "get router failed")
		return service, err
	}

	service.Spec.Type = routerType
	service.SetAnnotations(annotations)
	service, err = c.client.CoreV1().Services(ingressControllerNamespace).Update(service)
	return service, err
}

func (c *routerOperator) deleteRouterService(namespace string) (*corev1.Service, error) {

	service, err := c.getRouterService(namespace)
	if err != nil {
		klog.Error(err)
		return service, err
	}

	// delete controller service
	serviceName := ingressControllerPrefix + namespace
	deleteOptions := metav1.DeleteOptions{}

	err = c.client.CoreV1().Services(ingressControllerNamespace).Delete(serviceName, &deleteOptions)
	if err != nil {
		klog.Error(err)
		return service, err
	}

	return service, nil
}

func (c *routerOperator) createOrUpdateRouterWorkload(namespace string, publishService bool, servicemeshEnabled bool) error {
	obj, ok := c.routerTemplates["DEPLOYMENT"]
	if !ok {
		klog.Error("Deployment template file not loaded")
		return fmt.Errorf("deployment template file not loaded")
	}

	deployName := ingressControllerPrefix + namespace

	deployment, err := c.client.AppsV1().Deployments(ingressControllerNamespace).Get(deployName, metav1.GetOptions{})

	createDeployment := true

	if err != nil {
		if errors.IsNotFound(err) {
			deployment = obj.(*v1.Deployment)

			deployment.Name = ingressControllerPrefix + namespace

			// Add project label
			deployment.Spec.Selector.MatchLabels["project"] = namespace
			deployment.Spec.Template.Labels["project"] = namespace

			// Add configmap
			deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--configmap=$(POD_NAMESPACE)/"+deployment.Name+configMapSuffix)

			// Isolate namespace
			deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--watch-namespace="+namespace)

			// Choose self as master
			deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--election-id="+deployment.Name)

		}
	} else {
		createDeployment = false

		for i := range deployment.Spec.Template.Spec.Containers {
			if deployment.Spec.Template.Spec.Containers[i].Name == "nginx-ingress-controller" {
				var args []string
				for j := range deployment.Spec.Template.Spec.Containers[i].Args {
					argument := deployment.Spec.Template.Spec.Containers[i].Args[j]
					if strings.HasPrefix("--publish-service", argument) ||
						strings.HasPrefix("--configmap", argument) ||
						strings.HasPrefix("--report-node-internal-ip-address", argument) {
						continue
					}
					args = append(args, deployment.Spec.Template.Spec.Containers[i].Args[j])
				}
				deployment.Spec.Template.Spec.Containers[i].Args = args
			}
		}
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string, 0)
	}
	if servicemeshEnabled {
		deployment.Spec.Template.Annotations[sidecarInject] = "true"
	} else {
		deployment.Spec.Template.Annotations[sidecarInject] = "false"
	}

	if publishService {
		deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--publish-service="+ingressControllerNamespace+"/"+ingressControllerPrefix+namespace)
	} else {
		deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--report-node-internal-ip-address")
	}

	if createDeployment {
		deployment, err = c.client.AppsV1().Deployments(ingressControllerNamespace).Create(deployment)
	} else {
		deployment, err = c.client.AppsV1().Deployments(ingressControllerNamespace).Update(deployment)
	}

	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

func (c *routerOperator) deleteRouterWorkload(namespace string) error {
	deleteOptions := metav1.DeleteOptions{}
	// delete controller deployment
	deploymentName := ingressControllerPrefix + namespace
	err := c.client.AppsV1().Deployments(ingressControllerNamespace).Delete(deploymentName, &deleteOptions)
	if err != nil {
		klog.Error(err)
	}

	// delete replicaset if there are any
	selector := labels.SelectorFromSet(
		labels.Set{
			"app":       "kubesphere",
			"component": "ks-router",
			"tier":      "backend",
			"project":   namespace,
		})
	replicaSetLister := c.informers.Apps().V1().ReplicaSets().Lister()
	replicaSets, err := replicaSetLister.ReplicaSets(ingressControllerNamespace).List(selector)
	if err != nil {
		klog.Error(err)
	}

	for i := range replicaSets {
		err = c.client.AppsV1().ReplicaSets(ingressControllerNamespace).Delete(replicaSets[i].Name, &deleteOptions)
		if err != nil {
			klog.Error(err)
		}
	}

	return nil
}

// Update Ingress Controller Service, change type from NodePort to loadbalancer or vice versa.
func (c *routerOperator) UpdateRouter(namespace string, routerType corev1.ServiceType, annotations map[string]string) (*corev1.Service, error) {
	var router *corev1.Service

	router, err := c.getRouterService(namespace)

	if err != nil {
		klog.Error(err)
		return router, err
	}

	enableServicemesh := annotations[servicemeshEnabled] == "true"

	err = c.createOrUpdateRouterWorkload(namespace, routerType == corev1.ServiceTypeLoadBalancer, enableServicemesh)
	if err != nil {
		klog.Error(err)
		return router, err
	}

	newRouter, err := c.updateRouterService(namespace, routerType, annotations)
	if err != nil {
		klog.Error(err)
		return newRouter, err
	}

	return newRouter, nil
}
