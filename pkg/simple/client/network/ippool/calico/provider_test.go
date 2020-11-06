/*
Copyright 2020 The KubeSphere authors.

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

package calico

import (
	"flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	ksfake "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	calicofake "kubesphere.io/kubesphere/pkg/simple/client/network/ippool/calico/client/clientset/versioned/fake"
	"testing"
)

func TestCalicoIPPoolSuit(t *testing.T) {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Set("v", "4")
	flag.Parse()
	klog.SetOutput(GinkgoWriter)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Calico IPPool Suite")
}

var _ = Describe("test calico ippool", func() {
	pool := &v1alpha1.IPPool{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name: "testippool",
		},
		Spec: v1alpha1.IPPoolSpec{
			Type:      v1alpha1.Calico,
			CIDR:      "192.168.0.0/16",
			BlockSize: 24,
		},
		Status: v1alpha1.IPPoolStatus{},
	}

	ksclient := ksfake.NewSimpleClientset(pool)
	client := calicofake.NewSimpleClientset()

	p := provider{
		options: Options{
			IPIPMode:    "Always",
			VXLANMode:   "Never",
			NATOutgoing: true,
		},
		client:   client,
		ksclient: ksclient,
	}

	It("test create calico ippool", func() {
		err := p.CreateIPPool(pool)
		Expect(err).ShouldNot(HaveOccurred())
	})
})
