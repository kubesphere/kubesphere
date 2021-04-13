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
	"fmt"
	"strconv"
	"strings"

	"github.com/projectcalico/libcalico-go/lib/names"
	cnet "github.com/projectcalico/libcalico-go/lib/net"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindIPAMHandle     = "IPAMHandle"
	ResourceSingularIPAMHandle = "ipamhandle"
	ResourcePluralIPAMHandle   = "ipamhandles"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Cluster
type IPAMHandle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the IPAMHandle.
	Spec IPAMHandleSpec `json:"spec,omitempty"`
}

// IPAMHandleSpec contains the specification for an IPAMHandle resource.
type IPAMHandleSpec struct {
	HandleID string         `json:"handleID"`
	Block    map[string]int `json:"block"`
	Deleted  bool           `json:"deleted"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced
type IPAMHandleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []IPAMHandle `json:"items"`
}

func (h *IPAMHandle) IncrementBlock(block *IPAMBlock, num int) int {
	newNum := num
	if val, ok := h.Spec.Block[block.String()]; ok {
		// An entry exists for this block, increment the number
		// of allocations.
		newNum = val + num
	}
	h.Spec.Block[block.String()] = newNum
	return newNum
}

func (h *IPAMHandle) Empty() bool {
	return len(h.Spec.Block) == 0
}

func (h *IPAMHandle) MarkDeleted() {
	h.Spec.Deleted = true
}

func (h *IPAMHandle) IsDeleted() bool {
	return h.Spec.Deleted
}

func (h *IPAMHandle) DecrementBlock(block *IPAMBlock, num int) (*int, error) {
	if current, ok := h.Spec.Block[block.String()]; !ok {
		// This entry doesn't exist.
		return nil, fmt.Errorf("Tried to decrement block %s by %v but it isn't linked to handle %s", block.BlockName(), num, h.Spec.HandleID)
	} else {
		newNum := current - num
		if newNum < 0 {
			return nil, fmt.Errorf("Tried to decrement block %s by %v but it only has %v addresses on handle %s", block.BlockName(), num, current, h.Spec.HandleID)
		}

		if newNum == 0 {
			delete(h.Spec.Block, block.String())
		} else {
			h.Spec.Block[block.String()] = newNum
		}
		return &newNum, nil
	}
}

func ConvertToBlockName(k string) string {
	strs := strings.SplitN(k, "-", 2)
	id, _ := strconv.Atoi(strs[0])
	_, blockCIDR, _ := cnet.ParseCIDR(strs[1])

	return fmt.Sprintf("%d-%s", id, names.CIDRToName(*blockCIDR))
}
