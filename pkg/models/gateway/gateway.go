/*
Copyright 2021 The KubeSphere Authors.

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

package gateway

import (
	"context"
	"fmt"
	"io"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	jsonpatch "github.com/evanphx/json-patch"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/api/gateway/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/pod"
	"kubesphere.io/kubesphere/pkg/simple/client/gateway"
)

const (
	MasterLabel             = "node-role.kubernetes.io/master"
	SidecarInject           = "sidecar.istio.io/inject"
	gatewayPrefix           = "kubesphere-router-"
	workingNamespace        = "kubesphere-controls-system"
	globalGatewayNameSuffix = "kubesphere-system"
	globalGatewayName       = gatewayPrefix + globalGatewayNameSuffix
	helmPatch               = `{"metadata":{"annotations":{"meta.helm.sh/release-name":"%s-ingress","meta.helm.sh/release-namespace":"%s"},"labels":{"helm.sh/chart":"ingress-nginx-3.35.0","app.kubernetes.io/managed-by":"Helm","app":null,"component":null,"tier":null}},"spec":{"selector":null}}`
)

type GatewayOperator interface {
	GetGateways(namespace string) ([]*v1alpha1.Gateway, error)
	CreateGateway(namespace string, obj *v1alpha1.Gateway) (*v1alpha1.Gateway, error)
	DeleteGateway(namespace string) error
	UpdateGateway(namespace string, obj *v1alpha1.Gateway) (*v1alpha1.Gateway, error)
	UpgradeGateway(namespace string) (*v1alpha1.Gateway, error)
	ListGateways(query *query.Query) (*api.ListResult, error)
	GetPods(namespace string, query *query.Query) (*api.ListResult, error)
	GetPodLogs(ctx context.Context, namespace string, podName string, logOptions *corev1.PodLogOptions, responseWriter io.Writer) error
}

type gatewayOperator struct {
	k8sclient kubernetes.Interface
	factory   informers.InformerFactory
	client    client.Client
	cache     cache.Cache
	options   *gateway.Options
}

func NewGatewayOperator(client client.Client, cache cache.Cache, options *gateway.Options, factory informers.InformerFactory, k8sclient kubernetes.Interface) GatewayOperator {
	return &gatewayOperator{
		client:    client,
		cache:     cache,
		options:   options,
		k8sclient: k8sclient,
		factory:   factory,
	}
}

func (c *gatewayOperator) getWorkingNamespace(namespace string) string {
	ns := c.options.Namespace
	// Set the working namespace to watching namespace when the Gateway's Namespace Option is empty
	if ns == "" {
		ns = namespace
	}
	// Convert the global gateway query parameter
	if namespace == globalGatewayNameSuffix {
		ns = workingNamespace
	}
	return ns
}

// override user's setting when create/update a project gateway.
func (c *gatewayOperator) overrideDefaultValue(gateway *v1alpha1.Gateway, namespace string) *v1alpha1.Gateway {
	// override default name
	gateway.Name = fmt.Sprint(gatewayPrefix, namespace)
	if gateway.Name != globalGatewayName {
		gateway.Spec.Controller.Scope = v1alpha1.Scope{Enabled: true, Namespace: namespace}
	}
	gateway.Namespace = c.getWorkingNamespace(namespace)
	return gateway
}

// getGlobalGateway returns the global gateway
func (c *gatewayOperator) getGlobalGateway() *v1alpha1.Gateway {
	globalkey := types.NamespacedName{
		Namespace: workingNamespace,
		Name:      globalGatewayName,
	}

	global := &v1alpha1.Gateway{}
	if err := c.client.Get(context.TODO(), globalkey, global); err != nil {
		return nil
	}
	return global
}

// getLegacyGateway returns gateway created by the router api.
// Should always prompt user to upgrade the gateway.
func (c *gatewayOperator) getLegacyGateway(namespace string) *v1alpha1.Gateway {
	s := &corev1.ServiceList{}

	// filter legacy service by labels
	_ = c.client.List(context.TODO(), s, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(
			labels.Set{
				"app":       "kubesphere",
				"component": "ks-router",
				"tier":      "backend",
				"project":   namespace,
			}),
	})

	// create a fake Gateway object when legacy service exists
	if len(s.Items) > 0 {
		d := &appsv1.Deployment{}
		c.client.Get(context.TODO(), client.ObjectKeyFromObject(&s.Items[0]), d)

		return c.convert(namespace, &s.Items[0], d)
	}
	return nil
}

func (c *gatewayOperator) convert(namespace string, svc *corev1.Service, deploy *appsv1.Deployment) *v1alpha1.Gateway {
	legacy := v1alpha1.Gateway{
		TypeMeta: v1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
		},
		Spec: v1alpha1.GatewaySpec{
			Controller: v1alpha1.ControllerSpec{
				Scope: v1alpha1.Scope{
					Enabled:   true,
					Namespace: namespace,
				},
			},
			Service: v1alpha1.ServiceSpec{
				Annotations: svc.Annotations,
				Type:        svc.Spec.Type,
			},
			Deployment: v1alpha1.DeploymentSpec{
				Replicas: deploy.Spec.Replicas,
			},
		},
	}
	if an, ok := deploy.Annotations[SidecarInject]; ok {
		legacy.Spec.Deployment.Annotations = make(map[string]string)
		legacy.Spec.Deployment.Annotations[SidecarInject] = an
	}
	if len(deploy.Spec.Template.Spec.Containers) > 0 {
		legacy.Spec.Deployment.Resources = deploy.Spec.Template.Spec.Containers[0].Resources
	}
	return &legacy
}

func (c *gatewayOperator) getMasterNodeIp() []string {
	internalIps := []string{}
	masters := &corev1.NodeList{}
	err := c.cache.List(context.TODO(), masters, &client.ListOptions{LabelSelector: labels.SelectorFromSet(
		labels.Set{
			MasterLabel: "",
		})})

	if err != nil {
		klog.Info(err)
		return internalIps
	}

	for _, node := range masters.Items {
		for _, address := range node.Status.Addresses {
			if address.Type == corev1.NodeInternalIP {
				internalIps = append(internalIps, address.Address)
			}
		}
	}
	return internalIps
}

func (c *gatewayOperator) updateStatus(gateway *v1alpha1.Gateway, svc *corev1.Service) (*v1alpha1.Gateway, error) {
	// append selected node ip as loadBalancer ingress ip
	if svc.Spec.Type != corev1.ServiceTypeLoadBalancer && len(svc.Status.LoadBalancer.Ingress) == 0 {
		rips := c.getMasterNodeIp()
		for _, rip := range rips {
			gIngress := corev1.LoadBalancerIngress{
				IP: rip,
			}
			svc.Status.LoadBalancer.Ingress = append(svc.Status.LoadBalancer.Ingress, gIngress)
		}
	}

	status := unstructured.Unstructured{
		Object: map[string]interface{}{
			"loadBalancer": svc.Status.LoadBalancer,
			"service":      svc.Spec.Ports,
		},
	}

	target, err := status.MarshalJSON()
	if err != nil {
		return gateway, err
	}
	if gateway.Status.Raw != nil {
		//merge with origin status
		patch, err := jsonpatch.CreateMergePatch([]byte(`{}`), target)
		if err != nil {
			return gateway, err
		}
		modified, err := jsonpatch.MergePatch(gateway.Status.Raw, patch)
		if err != nil {
			return gateway, err
		}
		gateway.Status.Raw = modified
		return gateway, err
	}
	gateway.Status.Raw = target
	return gateway, nil
}

// GetGateways returns all Gateways from the project. There are at most 2 gateways exists in a project,
// a Global Gateway and a Project Gateway or a Legacy Project Gateway.
func (c *gatewayOperator) GetGateways(namespace string) ([]*v1alpha1.Gateway, error) {

	var gateways []*v1alpha1.Gateway

	if g := c.getGlobalGateway(); g != nil {
		gateways = append(gateways, g)
	}
	if g := c.getLegacyGateway(namespace); g != nil {
		gateways = append(gateways, g)
	}

	key := types.NamespacedName{
		Namespace: c.getWorkingNamespace(namespace),
		Name:      fmt.Sprint(gatewayPrefix, namespace),
	}
	obj := &v1alpha1.Gateway{}
	err := c.client.Get(context.TODO(), key, obj)

	if err == nil {
		gateways = append(gateways, obj)
	} else if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	for _, g := range gateways {
		s := &corev1.Service{}
		// We supports the Service name always as same as gateway name.
		// TODO: We need a mapping relation between the service and the gateway. Label Selector should be a good option.
		err := c.client.Get(context.TODO(), client.ObjectKeyFromObject(g), s)
		if err != nil {
			klog.Info(err)
			continue
		}
		_, err = c.updateStatus(g, s)
		if err != nil {
			klog.Info(err)
		}
	}

	return gateways, nil
}

// Create a Gateway in a namespace
func (c *gatewayOperator) CreateGateway(namespace string, obj *v1alpha1.Gateway) (*v1alpha1.Gateway, error) {

	if g := c.getGlobalGateway(); g != nil {
		return nil, fmt.Errorf("can't create project gateway if global gateway enabled")
	}

	if g := c.getLegacyGateway(namespace); g != nil {
		return nil, fmt.Errorf("can't create project gateway if legacy gateway exists, please upgrade the gateway firstly")
	}

	c.overrideDefaultValue(obj, namespace)
	err := c.client.Create(context.TODO(), obj)
	return obj, err
}

// DeleteGateway is used to delete Gateway related resources in the namespace
func (c *gatewayOperator) DeleteGateway(namespace string) error {
	obj := &v1alpha1.Gateway{
		ObjectMeta: v1.ObjectMeta{
			Namespace: c.getWorkingNamespace(namespace),
			Name:      fmt.Sprint(gatewayPrefix, namespace),
		},
	}
	return c.client.Delete(context.TODO(), obj)
}

// Update Gateway
func (c *gatewayOperator) UpdateGateway(namespace string, obj *v1alpha1.Gateway) (*v1alpha1.Gateway, error) {
	if c.options.Namespace == "" && obj.Namespace != namespace || c.options.Namespace != "" && c.options.Namespace != obj.Namespace {
		return nil, fmt.Errorf("namespace doesn't match with origin namespace")
	}
	c.overrideDefaultValue(obj, namespace)
	err := c.client.Update(context.TODO(), obj)
	return obj, err
}

// UpgradeGateway upgrade the legacy Project Gateway to a Gateway CRD.
// No rolling upgrade guaranteed, Service would be interrupted when deleting old deployment.
func (c *gatewayOperator) UpgradeGateway(namespace string) (*v1alpha1.Gateway, error) {
	l := c.getLegacyGateway(namespace)
	if l == nil {
		return nil, fmt.Errorf("invalid operation, no legacy gateway was found")
	}
	if l.Namespace != c.getWorkingNamespace(namespace) {
		return nil, fmt.Errorf("invalid operation, can't upgrade legacy gateway when working namespace changed")
	}

	// Get legacy gateway's config from configmap
	cm := &corev1.ConfigMap{}
	err := c.client.Get(context.TODO(), client.ObjectKey{Namespace: l.Namespace, Name: fmt.Sprintf("%s-nginx", l.Name)}, cm)
	if err == nil {
		l.Spec.Controller.Config = cm.Data
		defer func() {
			c.client.Delete(context.TODO(), cm)
		}()
	}

	// Delete old deployment, because it's not compatible with the deployment in the helm chart.
	// We can't defer here, there's a potential race condition causing gateway operator fails.
	d := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Namespace: l.Namespace,
			Name:      l.Name,
		},
	}
	err = c.client.Delete(context.TODO(), d)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	// Patch the legacy Service with helm annotations, So that it can be managed by the helm release.
	patch := []byte(fmt.Sprintf(helmPatch, l.Name, l.Namespace))
	err = c.client.Patch(context.Background(), &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Namespace: l.Namespace,
			Name:      l.Name,
		},
	}, client.RawPatch(types.StrategicMergePatchType, patch))

	if err != nil {
		return nil, err
	}

	c.overrideDefaultValue(l, namespace)
	err = c.client.Create(context.TODO(), l)
	return l, err
}

func (c *gatewayOperator) ListGateways(query *query.Query) (*api.ListResult, error) {
	gateways := v1alpha1.GatewayList{}
	err := c.cache.List(context.TODO(), &gateways, &client.ListOptions{LabelSelector: query.Selector()})
	if err != nil {
		return nil, err
	}
	var result []runtime.Object
	for i := range gateways.Items {
		result = append(result, &gateways.Items[i])
	}

	services := &corev1.ServiceList{}

	// filter legacy service by labels
	_ = c.client.List(context.TODO(), services, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(
			labels.Set{
				"app":       "kubesphere",
				"component": "ks-router",
				"tier":      "backend",
			}),
	})

	for i := range services.Items {
		result = append(result, &services.Items[i])
	}

	return v1alpha3.DefaultList(result, query, c.compare, c.filter, c.transform), nil
}

func (c *gatewayOperator) transform(obj runtime.Object) runtime.Object {
	if g, ok := obj.(*v1alpha1.Gateway); ok {
		svc := &corev1.Service{}
		// We supports the Service name always same as gateway name.
		err := c.client.Get(context.TODO(), client.ObjectKeyFromObject(g), svc)
		if err != nil {
			klog.Info(err)
			return g
		}
		g, err := c.updateStatus(g, svc)
		if err != nil {
			klog.Info(err)
		}
		return g

	}
	if svc, ok := obj.(*corev1.Service); ok {
		d := &appsv1.Deployment{}
		c.client.Get(context.TODO(), client.ObjectKeyFromObject(svc), d)
		g, err := c.updateStatus(c.convert(svc.Labels["project"], svc, d), svc)
		if err != nil {
			klog.Info(err)
		}
		return g
	}
	return nil
}

func (c *gatewayOperator) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftGateway, ok := left.(*v1alpha1.Gateway)
	if !ok {
		return false
	}

	rightGateway, ok := right.(*v1alpha1.Gateway)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftGateway.ObjectMeta, rightGateway.ObjectMeta, field)
}

func (c *gatewayOperator) filter(object runtime.Object, filter query.Filter) bool {
	var objMeta v1.ObjectMeta
	var namespace string

	gateway, ok := object.(*v1alpha1.Gateway)
	if !ok {
		svc, ok := object.(*corev1.Service)
		if !ok {
			return false
		}
		namespace = svc.Labels["project"]
		objMeta = svc.ObjectMeta
	} else {
		namespace = gateway.Spec.Controller.Scope.Namespace
		objMeta = gateway.ObjectMeta
	}

	switch filter.Field {
	case query.FieldNamespace:
		return strings.Compare(namespace, string(filter.Value)) == 0
	default:
		return v1alpha3.DefaultObjectMetaFilter(objMeta, filter)
	}
}

func (c *gatewayOperator) GetPods(namespace string, query *query.Query) (*api.ListResult, error) {
	podGetter := pod.New(c.factory.KubernetesSharedInformerFactory())

	//TODO: move the selector string to options
	selector, err := labels.Parse(fmt.Sprintf("app.kubernetes.io/name=ingress-nginx,app.kubernetes.io/instance=kubesphere-router-%s-ingress", namespace))
	if err != nil {
		return nil, fmt.Errorf("invaild selector config")
	}
	query.LabelSelector = selector.String()
	return podGetter.List(c.getWorkingNamespace(namespace), query)
}

func (c *gatewayOperator) GetPodLogs(ctx context.Context, namespace string, podName string, logOptions *corev1.PodLogOptions, responseWriter io.Writer) error {
	workingNamespace := c.getWorkingNamespace(namespace)

	pods, err := c.GetPods(namespace, query.New())
	if err != nil {
		return err
	}
	if !c.hasPod(pods.Items, types.NamespacedName{Namespace: workingNamespace, Name: podName}) {
		return fmt.Errorf("pod does not exist")
	}

	podLogRequest := c.k8sclient.CoreV1().
		Pods(workingNamespace).
		GetLogs(podName, logOptions)
	reader, err := podLogRequest.Stream(context.TODO())
	if err != nil {
		return err
	}
	_, err = io.Copy(responseWriter, reader)
	if err != nil {
		return err
	}
	return nil
}

func (c *gatewayOperator) hasPod(slice []interface{}, key types.NamespacedName) bool {
	for _, s := range slice {
		pod, ok := s.(*corev1.Pod)
		if ok && client.ObjectKeyFromObject(pod) == key {
			return true
		}
	}
	return false
}
