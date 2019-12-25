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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/testing"
	clientset "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	devopsv1alpha1 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/devops/v1alpha1"
	fakedevopsv1alpha1 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/devops/v1alpha1/fake"
	loggingv1alpha2 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/logging/v1alpha2"
	fakeloggingv1alpha2 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/logging/v1alpha2/fake"
	networkv1alpha1 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/network/v1alpha1"
	fakenetworkv1alpha1 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/network/v1alpha1/fake"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/servicemesh/v1alpha2"
	fakeservicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/servicemesh/v1alpha2/fake"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/tenant/v1alpha1"
	faketenantv1alpha1 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/tenant/v1alpha1/fake"
)

// NewSimpleClientset returns a clientset that will respond with the provided objects.
// It's backed by a very simple object tracker that processes creates, updates and deletions as-is,
// without applying any validations and/or defaults. It shouldn't be considered a replacement
// for a real clientset and is mostly useful in simple unit tests.
func NewSimpleClientset(objects ...runtime.Object) *Clientset {
	o := testing.NewObjectTracker(scheme, codecs.UniversalDecoder())
	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	cs := &Clientset{tracker: o}
	cs.discovery = &fakediscovery.FakeDiscovery{Fake: &cs.Fake}
	cs.AddReactor("*", "*", testing.ObjectReaction(o))
	cs.AddWatchReactor("*", func(action testing.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		watch, err := o.Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		return true, watch, nil
	})

	return cs
}

// Clientset implements clientset.Interface. Meant to be embedded into a
// struct to get a default implementation. This makes faking out just the method
// you want to test easier.
type Clientset struct {
	testing.Fake
	discovery *fakediscovery.FakeDiscovery
	tracker   testing.ObjectTracker
}

func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	return c.discovery
}

func (c *Clientset) Tracker() testing.ObjectTracker {
	return c.tracker
}

var _ clientset.Interface = &Clientset{}

// DevopsV1alpha1 retrieves the DevopsV1alpha1Client
func (c *Clientset) DevopsV1alpha1() devopsv1alpha1.DevopsV1alpha1Interface {
	return &fakedevopsv1alpha1.FakeDevopsV1alpha1{Fake: &c.Fake}
}

// LoggingV1alpha2 retrieves the LoggingV1alpha2Client
func (c *Clientset) LoggingV1alpha2() loggingv1alpha2.LoggingV1alpha2Interface {
	return &fakeloggingv1alpha2.FakeLoggingV1alpha2{Fake: &c.Fake}
}

// NetworkV1alpha1 retrieves the NetworkV1alpha1Client
func (c *Clientset) NetworkV1alpha1() networkv1alpha1.NetworkV1alpha1Interface {
	return &fakenetworkv1alpha1.FakeNetworkV1alpha1{Fake: &c.Fake}
}

// ServicemeshV1alpha2 retrieves the ServicemeshV1alpha2Client
func (c *Clientset) ServicemeshV1alpha2() servicemeshv1alpha2.ServicemeshV1alpha2Interface {
	return &fakeservicemeshv1alpha2.FakeServicemeshV1alpha2{Fake: &c.Fake}
}

// TenantV1alpha1 retrieves the TenantV1alpha1Client
func (c *Clientset) TenantV1alpha1() tenantv1alpha1.TenantV1alpha1Interface {
	return &faketenantv1alpha1.FakeTenantV1alpha1{Fake: &c.Fake}
}
