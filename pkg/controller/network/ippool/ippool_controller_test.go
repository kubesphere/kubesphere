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
	"flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	ksfake "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/controller/network/utils"
	"kubesphere.io/kubesphere/pkg/simple/client/network/ippool"
	"kubesphere.io/kubesphere/pkg/simple/client/network/ippool/ipam"
	"testing"
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

var _ = Describe("test ippool", func() {
	pool := &v1alpha1.IPPool{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name: "testippool",
		},
		Spec: v1alpha1.IPPoolSpec{
			Type:      v1alpha1.VLAN,
			CIDR:      "192.168.0.0/24",
			BlockSize: 24,
		},
		Status: v1alpha1.IPPoolStatus{},
	}

	ksclient := ksfake.NewSimpleClientset()
	k8sclinet := k8sfake.NewSimpleClientset()
	options := ippool.Options{}
	p := ippool.NewProvider(ksclient, options)
	ipamClient := ipam.NewIPAMClient(ksclient, v1alpha1.VLAN)

	ksInformer := ksinformers.NewSharedInformerFactory(ksclient, 0)
	ippoolInformer := ksInformer.Network().V1alpha1().IPPools()
	ipamblockInformer := ksInformer.Network().V1alpha1().IPAMBlocks()
	c := NewIPPoolController(ippoolInformer, ipamblockInformer, k8sclinet, ksclient, options, p)

	stopCh := make(chan struct{})
	go ksInformer.Start(stopCh)
	go c.Start(stopCh)

	It("test create ippool", func() {
		_, err := ksclient.NetworkV1alpha1().IPPools().Create(pool)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(func() bool {
			result, _ := ksclient.NetworkV1alpha1().IPPools().Get(pool.Name, v1.GetOptions{})
			if len(result.Labels) != 3 {
				return false
			}

			if utils.NeedToAddFinalizer(result, v1alpha1.IPPoolFinalizer) {
				return false
			}

			return true
		}).Should(Equal(true))
	})

	It("test ippool stats", func() {
		ipamClient.AutoAssign(ipam.AutoAssignArgs{
			HandleID: "testhandle",
			Attrs:    nil,
			Pool:     "testippool",
		})

		Eventually(func() bool {
			result, _ := ksclient.NetworkV1alpha1().IPPools().Get(pool.Name, v1.GetOptions{})
			if result.Status.Allocations != 1 {
				return false
			}

			return true
		}).Should(Equal(true))
	})

	It("test delete pool", func() {
		ipamClient.ReleaseByHandle("testhandle")
		Eventually(func() bool {
			result, _ := ksclient.NetworkV1alpha1().IPPools().Get(pool.Name, v1.GetOptions{})
			if result.Status.Allocations != 0 {
				return false
			}

			return true
		}).Should(Equal(true))

		err := ksclient.NetworkV1alpha1().IPPools().Delete(pool.Name, &v1.DeleteOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		blocks, _ := ksclient.NetworkV1alpha1().IPAMBlocks().List(v1.ListOptions{})
		Expect(len(blocks.Items)).Should(Equal(0))
	})
})
