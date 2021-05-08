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

package ipam

import (
	"context"
	"flag"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"kubesphere.io/api/network/v1alpha1"

	ksfake "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
)

func TestIPAM_blockGenerator(t *testing.T) {
	pool := &v1alpha1.IPPool{
		ObjectMeta: v1.ObjectMeta{
			Name: "testippool",
		},
		Spec: v1alpha1.IPPoolSpec{
			Type:       v1alpha1.VLAN,
			CIDR:       "192.168.0.0/24",
			RangeEnd:   "192.168.0.250",
			RangeStart: "192.168.0.10",
			BlockSize:  25,
		},
	}
	blocks := blockGenerator(pool)
	for subnet := blocks(); subnet != nil; subnet = blocks() {
		if subnet.String() != "192.168.0.0/25" && subnet.String() != "192.168.0.128/25" {
			t.FailNow()
		}
	}
}

func TestIPAMSuit(t *testing.T) {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Set("v", "4")
	flag.Parse()
	klog.SetOutput(GinkgoWriter)
	RegisterFailHandler(Fail)
	RunSpecs(t, "IPAM Test Suite")
}

const (
	vlanIPPoolName = "testippool"
)

var (
	vlanIPPool = &v1alpha1.IPPool{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name: vlanIPPoolName,
		},
		Spec: v1alpha1.IPPoolSpec{
			Type: v1alpha1.VLAN,
			CIDR: "192.168.0.0/24",
		},
		Status: v1alpha1.IPPoolStatus{},
	}
)

func newVLANIPAMClient() IPAMClient {
	vlanIPPool.Labels = map[string]string{
		v1alpha1.IPPoolTypeLabel: vlanIPPool.Spec.Type,
		v1alpha1.IPPoolNameLabel: vlanIPPool.Name,
		v1alpha1.IPPoolIDLabel:   fmt.Sprintf("%d", vlanIPPool.ID()),
	}
	return IPAMClient{
		typeStr: v1alpha1.VLAN,
		client:  ksfake.NewSimpleClientset(vlanIPPool),
	}
}

func TestIpamClient_GetAllPools(t *testing.T) {
	c := newVLANIPAMClient()
	pools, _ := c.getAllPools()
	if len(pools) != 1 {
		t.FailNow()
	}
}

var _ = Describe("test vlan ippool", func() {
	It("test get all vlan ippool", func() {
		c := newVLANIPAMClient()
		pools, _ := c.getAllPools()
		Expect(len(pools)).To(Equal(1))
	})

	It("test get pool utilization", func() {
		c := newVLANIPAMClient()
		stats, _ := c.GetUtilization(GetUtilizationArgs{Pools: []string{vlanIPPoolName}})
		Expect(len(stats)).To(Equal(1))
		Expect(stats[0].Unallocated).To(Equal(256))
	})

	It("test auto assign/unassign ip", func() {
		c := newVLANIPAMClient()
		result, _ := c.AutoAssign(AutoAssignArgs{
			HandleID: "testhandle",
			Pool:     vlanIPPoolName,
		})
		Expect(len(result.IPs)).To(Equal(1))
		stats, _ := c.GetUtilization(GetUtilizationArgs{Pools: []string{vlanIPPoolName}})
		Expect(stats[0].Unallocated).To(Equal(255))

		blocks, _ := c.client.NetworkV1alpha1().IPAMBlocks().List(context.Background(), v1.ListOptions{})
		Expect(len(blocks.Items)).To(Equal(1))
		Expect(blocks.Items[0].BlockName()).To(Equal(fmt.Sprintf("%d-%s", vlanIPPool.ID(), "192-168-0-0-24")))

		handles, _ := c.client.NetworkV1alpha1().IPAMHandles().List(context.Background(), v1.ListOptions{})
		Expect(len(handles.Items)).To(Equal(1))
		Expect(handles.Items[0].Name).To(Equal("testhandle"))

		Expect(c.ReleaseByHandle("testhandle")).ShouldNot(HaveOccurred())
		blocks, _ = c.client.NetworkV1alpha1().IPAMBlocks().List(context.Background(), v1.ListOptions{})
		Expect(len(blocks.Items)).To(Equal(0))

		handles, _ = c.client.NetworkV1alpha1().IPAMHandles().List(context.Background(), v1.ListOptions{})
		Expect(len(handles.Items)).To(Equal(0))
	})
})
