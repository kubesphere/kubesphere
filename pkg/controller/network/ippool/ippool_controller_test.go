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

package ippool

import (
	"context"
	"flag"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	ksfake "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/controller/network/utils"
	"kubesphere.io/kubesphere/pkg/simple/client/network/ippool"
	"kubesphere.io/kubesphere/pkg/simple/client/network/ippool/ipam"
)

func TestIPPoolSuit(t *testing.T) {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Set("v", "4")
	flag.Parse()
	klog.SetOutput(GinkgoWriter)
	RegisterFailHandler(Fail)
	RunSpecs(t, "IPPool Suite")
}

var (
	alwaysReady = func() bool { return true }
)

var _ = Describe("test ippool", func() {
	pool := &v1alpha1.IPPool{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name: "testippool",
		},
		Spec: v1alpha1.IPPoolSpec{
			Type: v1alpha1.VLAN,
			CIDR: "192.168.0.0/24",
		},
		Status: v1alpha1.IPPoolStatus{},
	}

	ksclient := ksfake.NewSimpleClientset()
	k8sclinet := k8sfake.NewSimpleClientset()
	ksInformer := ksinformers.NewSharedInformerFactory(ksclient, 0)
	k8sInformer := k8sinformers.NewSharedInformerFactory(k8sclinet, 0)

	p := ippool.NewProvider(k8sInformer, ksclient, k8sclinet, v1alpha1.IPPoolTypeLocal, nil)
	ipamClient := ipam.NewIPAMClient(ksclient, v1alpha1.VLAN)
	c := NewIPPoolController(ksInformer, k8sInformer, k8sclinet, ksclient, p)

	stopCh := make(chan struct{})
	go ksInformer.Start(stopCh)
	go k8sInformer.Start(stopCh)
	go c.Start(stopCh)

	It("test create ippool", func() {
		clone := pool.DeepCopy()
		clone.Spec.CIDR = "testxxx"
		Expect(c.ValidateCreate(clone)).Should(HaveOccurred())

		clone = pool.DeepCopy()
		clone.Spec.CIDR = "192.168.0.0/24"
		clone.Spec.RangeStart = "192.168.0.100"
		clone.Spec.RangeEnd = "192.168.0.99"
		Expect(c.ValidateCreate(clone)).Should(HaveOccurred())

		clone = pool.DeepCopy()
		clone.Spec.CIDR = "192.168.0.0/24"
		clone.Spec.RangeStart = "192.168.3.100"
		clone.Spec.RangeEnd = "192.168.3.111"
		Expect(c.ValidateCreate(clone)).Should(HaveOccurred())

		clone = pool.DeepCopy()
		clone.Spec.CIDR = "192.168.0.0/24"
		clone.Spec.BlockSize = 23
		Expect(c.ValidateCreate(clone)).Should(HaveOccurred())

		clone = pool.DeepCopy()
		_, err := ksclient.NetworkV1alpha1().IPPools().Create(context.TODO(), clone, v1.CreateOptions{})
		Expect(err).ShouldNot(HaveOccurred())

		Eventually(func() bool {
			result, _ := ksclient.NetworkV1alpha1().IPPools().Get(context.TODO(), pool.Name, v1.GetOptions{})
			if len(result.Labels) != 3 {
				return false
			}

			if utils.NeedToAddFinalizer(result, v1alpha1.IPPoolFinalizer) {
				return false
			}

			return true
		}, 3*time.Second).Should(Equal(true))

		clone = pool.DeepCopy()
		Expect(c.ValidateCreate(clone)).Should(HaveOccurred())
	})

	It("test update ippool", func() {
		old, _ := ksclient.NetworkV1alpha1().IPPools().Get(context.TODO(), pool.Name, v1.GetOptions{})
		new := old.DeepCopy()
		new.Spec.CIDR = "192.168.1.0/24"
		Expect(c.ValidateUpdate(old, new)).Should(HaveOccurred())
	})

	It("test ippool stats", func() {
		ipamClient.AutoAssign(ipam.AutoAssignArgs{
			HandleID: "testhandle",
			Attrs:    nil,
			Pool:     "testippool",
		})

		Eventually(func() bool {
			result, _ := ksclient.NetworkV1alpha1().IPPools().Get(context.TODO(), pool.Name, v1.GetOptions{})
			if result.Status.Allocations != 1 {
				return false
			}

			return true
		}, 3*time.Second).Should(Equal(true))
	})

	It("test delete pool", func() {
		result, _ := ksclient.NetworkV1alpha1().IPPools().Get(context.TODO(), pool.Name, v1.GetOptions{})
		Expect(c.ValidateDelete(result)).Should(HaveOccurred())

		ipamClient.ReleaseByHandle("testhandle")
		Eventually(func() bool {
			result, _ := ksclient.NetworkV1alpha1().IPPools().Get(context.TODO(), pool.Name, v1.GetOptions{})
			if result.Status.Allocations != 0 {
				return false
			}

			return true
		}, 3*time.Second).Should(Equal(true))

		err := ksclient.NetworkV1alpha1().IPPools().Delete(context.TODO(), pool.Name, v1.DeleteOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		blocks, _ := ksclient.NetworkV1alpha1().IPAMBlocks().List(context.TODO(), v1.ListOptions{})
		Expect(len(blocks.Items)).Should(Equal(0))
	})
})
