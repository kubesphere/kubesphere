/*

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

package v1alpha2

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"reflect"
	"strings"

	"github.com/kubesphere/porter/pkg/constant"
	"github.com/kubesphere/porter/pkg/manager/client"
	cnet "github.com/projectcalico/libcalico-go/lib/net"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (e Eip) IPToOrdinal(ip net.IP) int {
	base, size, err := e.GetSize()
	if err != nil {
		return -1
	}
	ipAsInt := cnet.IPToBigInt(cnet.IP{IP: ip})
	baseInt := cnet.IPToBigInt(cnet.IP{IP: base})
	ord := big.NewInt(0).Sub(ipAsInt, baseInt).Int64()
	if ord < 0 || ord >= size {
		return -1
	}
	return int(ord)
}

func (e Eip) GetSpeakerName() string {
	if e.Spec.Protocol == constant.PorterProtocolLayer2 {
		return e.Spec.Interface
	}

	return constant.PorterProtocolBGP
}

func (e Eip) GetProtocol() string {
	if e.Spec.Protocol == constant.PorterProtocolLayer2 {
		return constant.PorterProtocolLayer2
	}

	return constant.PorterProtocolBGP
}

func (e Eip) GetSize() (net.IP, int64, error) {
	ip := net.ParseIP(e.Spec.Address)
	if ip != nil {
		return ip, 1, nil
	}

	_, cidr, err := net.ParseCIDR(e.Spec.Address)
	if err == nil {
		ones, size := cidr.Mask.Size()
		num := 1 << uint(size-ones)
		return cidr.IP, int64(num), nil
	}

	strs := strings.SplitN(e.Spec.Address, constant.EipRangeSeparator, 2)
	if len(strs) != 2 {
		return nil, 0, fmt.Errorf("invalid eip address format")
	}
	base := cnet.ParseIP(strs[0])
	last := cnet.ParseIP(strs[1])
	if base == nil || last == nil {
		return nil, 0, fmt.Errorf("invalid eip address format")
	}

	ord := big.NewInt(0).Sub(cnet.IPToBigInt(*last), cnet.IPToBigInt(*base)).Int64()
	if ord < 0 {
		return nil, 0, fmt.Errorf("invalid eip address format")
	}

	return base.IP, ord + 1, nil
}

var _ webhook.Validator = &Eip{}

// +kubebuilder:webhook:path=/validate-network-kubesphere-io-v1alpha2-eip,mutating=false,failurePolicy=fail,groups=network.kubesphere.io,resources=eips,verbs=create;update;delete,versions=v1alpha2,name=validate.eip.network.kubesphere.io

func (e Eip) IsOverlap(eip Eip) bool {
	base, size, _ := e.GetSize()

	tBase, tSize, _ := eip.GetSize()
	if cnet.IPToBigInt(cnet.IP{IP: base}).
		Cmp(big.NewInt(0).
			Add(cnet.IPToBigInt(cnet.IP{IP: tBase}), big.NewInt(tSize-1))) > 0 ||
		cnet.IPToBigInt(cnet.IP{IP: tBase}).Cmp(big.NewInt(0).
			Add(cnet.IPToBigInt(cnet.IP{IP: base}), big.NewInt(size-1))) > 0 {
		return false
	}
	return true
}

func (e Eip) ValidateCreate() error {
	_, _, err := e.GetSize()
	if err != nil {
		return err
	}

	eips := EipList{}
	err = client.Client.List(context.Background(), &eips)
	if err != nil {
		return err
	}

	for _, eip := range eips.Items {
		if e.IsOverlap(eip) {
			return fmt.Errorf("eip address overlap with %s", eip.Name)
		}
	}

	if e.Spec.Protocol == constant.PorterProtocolLayer2 {
		if e.Spec.Interface == "" {
			return fmt.Errorf("field spec.interface should not be empty")
		}
	}

	return nil
}

func (e Eip) ValidateUpdate(old runtime.Object) error {
	oldE := old.(*Eip)
	if !reflect.DeepEqual(e.Spec, oldE.Spec) {
		if e.Spec.Disable == oldE.Spec.Disable {
			return fmt.Errorf("only allow modify field disable")
		}
	}
	return nil
}

func (e Eip) ValidateDelete() error {
	return nil
}

func (e Eip) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&e).
		Complete()
}

// EipSpec defines the desired state of EIP
type EipSpec struct {
	// +kubebuilder:validation:Required
	Address string `json:"address,required"`
	// +kubebuilder:validation:Enum=bgp;layer2
	Protocol      string `json:"protocol,omitempty"`
	Interface     string `json:"interface,omitempty"`
	Disable       bool   `json:"disable,omitempty"`
	UsingKnownIPs bool   `json:"usingKnownIPs,omitempty"`
}

// EipStatus defines the observed state of EIP
type EipStatus struct {
	Occupied bool              `json:"occupied,omitempty"`
	Usage    int               `json:"usage,omitempty"`
	PoolSize int               `json:"poolSize,omitempty"`
	Used     map[string]string `json:"used,omitempty"`
	FirstIP  string            `json:"firstIP,omitempty"`
	LastIP   string            `json:"lastIP,omitempty"`
	Ready    bool              `json:"ready,omitempty"`
	V4       bool              `json:"v4,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="cidr",type=string,JSONPath=`.spec.address`
// +kubebuilder:printcolumn:name="usage",type=integer,JSONPath=`.status.usage`
// +kubebuilder:printcolumn:name="total",type=integer,JSONPath=`.status.poolSize`
// +kubebuilder:resource:scope=Cluster,categories=networking

// Eip is the Schema for the eips API
type Eip struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EipSpec   `json:"spec,omitempty"`
	Status EipStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// EipList contains a list of Eip
type EipList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Eip `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Eip{}, &EipList{})
}
