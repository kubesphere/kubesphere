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
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/api/gateway/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/simple/client/gateway"
)

const (
	gatewayPrefix     = "kubesphere-router-"
	workingNamespace  = "kubesphere-controls-system"
	globalGatewayname = gatewayPrefix + "kubesphere-system"
	helmPatch         = `{"metadata":{"annotations":{"meta.helm.sh/release-name":"%s-ingress","meta.helm.sh/release-namespace":"%s"},"labels":{"helm.sh/chart":"ingress-nginx-3.35.0","app.kubernetes.io/managed-by":"Helm","app":null,"component":null,"tier":null}}}`
)

type GatewayOperator interface {
	GetGateways(namespace string) ([]*v1alpha1.Gateway, error)
	CreateGateway(namespace string, obj *v1alpha1.Gateway) (*v1alpha1.Gateway, error)
	DeleteGateway(namespace string) error
	UpdateGateway(namespace string, obj *v1alpha1.Gateway) (*v1alpha1.Gateway, error)
	UpgradeGateway(namespace string) (*v1alpha1.Gateway, error)
	ListGateways(query *query.Query) (*api.ListResult, error)
}

type gatewayOperator struct {
	client  client.Client
	cache   cache.Cache
	options *gateway.Options
}

func NewGatewayOperator(client client.Client, cache cache.Cache, options *gateway.Options) GatewayOperator {
	return &gatewayOperator{
		client:  client,
		cache:   cache,
		options: options,
	}
}

func (c *gatewayOperator) getWorkingNamespace(namespace string) string {
	ns := c.options.Namespace
	// Set the working namespace to watching namespace when the Gatway's Namsapce Option is empty
	if ns == "" {
		ns = namespace
	}
	return ns
}

// overide user's setting when create/update a project gateway.
func (c *gatewayOperator) overideDefaultValue(gateway *v1alpha1.Gateway, namespace string) *v1alpha1.Gateway {
	// overide default name
	gateway.Name = fmt.Sprint(gatewayPrefix, namespace)
	if gateway.Name != globalGatewayname {
		gateway.Spec.Conroller.Scope = v1alpha1.Scope{Enabled: true, Namespace: namespace}
	}
	gateway.Namespace = c.getWorkingNamespace(namespace)
	return gateway
}

// getGlobalGateway returns the global gateway
func (c *gatewayOperator) getGlobalGateway() *v1alpha1.Gateway {
	globalkey := types.NamespacedName{
		Namespace: workingNamespace,
		Name:      globalGatewayname,
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
		return c.convert(namespace, &s.Items[0])
	}
	return nil
}

func (c *gatewayOperator) convert(namespace string, svc *corev1.Service) *v1alpha1.Gateway {
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
			Conroller: v1alpha1.ControllerSpec{
				Scope: v1alpha1.Scope{
					Enabled:   true,
					Namespace: namespace,
				},
			},
			Service: v1alpha1.ServiceSpec{
				Annotations: svc.Annotations,
				Type:        svc.Spec.Type,
			},
		},
	}
	return &legacy
}

// GetGateways returns all Gateways from the project. There are at most 2 gatways exists in a project,
// a Glabal Gateway and a Project Gateway or a Legacy Project Gateway.
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
	if errors.IsNotFound(err) {
		return gateways, nil
	} else if err != nil {
		return nil, err
	}
	gateways = append(gateways, obj)
	return gateways, err
}

// Create a Gateway in a namespace
func (c *gatewayOperator) CreateGateway(namespace string, obj *v1alpha1.Gateway) (*v1alpha1.Gateway, error) {

	if g := c.getGlobalGateway(); g != nil {
		return nil, fmt.Errorf("can't create project gateway if global gateway enabled")
	}

	if g := c.getLegacyGateway(namespace); g != nil {
		return nil, fmt.Errorf("can't create project gateway if legacy gateway exists, please upgrade the gateway firstly")
	}

	c.overideDefaultValue(obj, namespace)
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
		return nil, fmt.Errorf("namepsace doesn't match with origin namesapce")
	}
	c.overideDefaultValue(obj, namespace)
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
	if l.Namespace != c.options.Namespace {
		return nil, fmt.Errorf("invalid operation, can't upgrade legacy gateway when working namespace changed")
	}

	// Delete old deployment, because it's not compatile with the deployment in the helm chart.
	d := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Namespace: l.Namespace,
			Name:      l.Name,
		},
	}
	err := c.client.Delete(context.TODO(), d)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	// Patch the legacy Serivce with helm annotations, So that it can be mannaged by the helm release.
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

	c.overideDefaultValue(l, namespace)
	err = c.client.Create(context.TODO(), l)
	return l, err
}

func (c *gatewayOperator) ListGateways(query *query.Query) (*api.ListResult, error) {
	applications := v1alpha1.GatewayList{}
	err := c.cache.List(context.TODO(), &applications, &client.ListOptions{LabelSelector: query.Selector()})
	if err != nil {
		return nil, err
	}
	var result []runtime.Object
	for i := range applications.Items {
		result = append(result, &applications.Items[i])
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

	for _, s := range services.Items {
		g := c.convert(s.Labels["project"], &s)
		result = append(result, g)
	}

	return v1alpha3.DefaultList(result, query, c.compare, c.filter), nil
}

func (d *gatewayOperator) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftApplication, ok := left.(*v1alpha1.Gateway)
	if !ok {
		return false
	}

	rightApplication, ok := right.(*v1alpha1.Gateway)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftApplication.ObjectMeta, rightApplication.ObjectMeta, field)
}

func (d *gatewayOperator) filter(object runtime.Object, filter query.Filter) bool {
	gateway, ok := object.(*v1alpha1.Gateway)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldNamespace:
		return strings.Compare(gateway.Spec.Conroller.Scope.Namespace, string(filter.Value)) == 0
	default:
		return v1alpha3.DefaultObjectMetaFilter(gateway.ObjectMeta, filter)
	}
}
