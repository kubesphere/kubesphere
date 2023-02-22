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

func TestIPPool(t *testing.T) {
	pool := &IPPool{
		ObjectMeta: v1.ObjectMeta{
			Name: "testippool",
		},
		Spec: IPPoolSpec{
			Type:       VLAN,
			CIDR:       "192.168.0.1/24",
			RangeEnd:   "192.168.0.250",
			RangeStart: "192.168.0.10",
		},
	}
	input := cnet.ParseIP("192.168.0.1")
	offset, _ := pool.IPToOrdinal(*input)
	if offset != 1 {
		t.Fail()
	}

	input = cnet.ParseIP("192.168.1.1")
	_, err := pool.IPToOrdinal(*input)
	if err == nil {
		t.Fail()
	}

	if pool.NumAddresses() != 256 {
		t.Fail()
	}

	if pool.StartReservedAddressed() != 10 {
		t.Fail()
	}

	if pool.EndReservedAddressed() != 5 {
		t.Fail()
	}

	if pool.NumReservedAddresses() != 15 {
		t.Fail()
	}
}
