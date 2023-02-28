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

package v1alpha1

import (
	"testing"

	cnet "github.com/projectcalico/calico/libcalico-go/lib/net"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIPAMBlock(t *testing.T) {
	pool := &IPPool{
		ObjectMeta: v1.ObjectMeta{
			Name: "testippool",
		},
		Spec: IPPoolSpec{
			Type:       VLAN,
			CIDR:       "192.168.0.0/24",
			RangeEnd:   "192.168.0.250",
			RangeStart: "192.168.0.10",
		},
	}

	handleID := "testhandle"

	_, cidr, _ := cnet.ParseCIDR("192.168.0.0/24")
	block := NewBlock(pool, *cidr, &ReservedAttr{
		StartOfBlock: 10,
		EndOfBlock:   10,
		Handle:       ReservedHandle,
		Note:         ReservedNote,
	})

	reserved := block.NumReservedAddresses()
	if reserved != 20 {
		t.Fail()
	}

	total := block.NumAddresses()
	free := block.NumFreeAddresses()
	if free != total-reserved {
		t.Fail()
	}

	t.Log("Allocate 10 addresses from block")
	ips := block.AutoAssign(10, handleID, nil)
	if len(ips) != 10 {
		t.Fail()
	}

	free = block.NumFreeAddresses()
	if free != total-reserved-10 {
		t.Fail()
	}

	t.Log("Allocate 1000 addresses from block")
	ips = block.AutoAssign(1000, handleID, nil)
	if len(ips) != free {
		t.Fail()
	}

	free = block.NumFreeAddresses()
	if free != 0 {
		t.Fail()
	}

	t.Log("Free address from block")
	if block.ReleaseByHandle(handleID) != total-reserved {
		t.Fail()
	}
}
