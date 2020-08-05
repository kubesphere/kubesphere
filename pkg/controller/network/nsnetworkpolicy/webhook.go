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

package nsnetworkpolicy

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	k8snet "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	networkv1alpha1 "kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	"net"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type NSNPValidator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (v *NSNPValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	nsnp := &networkv1alpha1.NamespaceNetworkPolicy{}

	err := v.decoder.Decode(req, nsnp)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	allErrs := field.ErrorList{}
	allErrs = append(allErrs, v.ValidateNSNPSpec(&nsnp.Spec, field.NewPath("spec"))...)

	if len(allErrs) != 0 {
		return admission.Denied(allErrs.ToAggregate().Error())
	}

	return admission.Allowed("")
}

func (v *NSNPValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}

// ValidateNetworkPolicyPort validates a NetworkPolicyPort
func (v *NSNPValidator) ValidateNetworkPolicyPort(port *k8snet.NetworkPolicyPort, portPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if port.Protocol != nil && *port.Protocol != corev1.ProtocolTCP && *port.Protocol != corev1.ProtocolUDP && *port.Protocol != corev1.ProtocolSCTP {
		allErrs = append(allErrs, field.NotSupported(portPath.Child("protocol"), *port.Protocol, []string{string(corev1.ProtocolTCP), string(corev1.ProtocolUDP), string(corev1.ProtocolSCTP)}))
	}
	if port.Port != nil {
		if port.Port.Type == intstr.Int {
			for _, msg := range validation.IsValidPortNum(int(port.Port.IntVal)) {
				allErrs = append(allErrs, field.Invalid(portPath.Child("port"), port.Port.IntVal, msg))
			}
		} else {
			for _, msg := range validation.IsValidPortName(port.Port.StrVal) {
				allErrs = append(allErrs, field.Invalid(portPath.Child("port"), port.Port.StrVal, msg))
			}
		}
	}

	return allErrs
}

func (v *NSNPValidator) ValidateServiceSelector(serviceSelector *networkv1alpha1.ServiceSelector, fldPath *field.Path) field.ErrorList {
	service := &corev1.Service{}
	allErrs := field.ErrorList{}

	err := v.Client.Get(context.TODO(), client.ObjectKey{Namespace: serviceSelector.Namespace, Name: serviceSelector.Name}, service)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, serviceSelector, "cannot get service"))
		return allErrs
	}

	if len(service.Spec.Selector) <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, serviceSelector, "service should have selector"))
	}

	return allErrs
}

// ValidateCIDR validates whether a CIDR matches the conventions expected by net.ParseCIDR
func ValidateCIDR(cidr string) (*net.IPNet, error) {
	_, net, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	return net, nil
}

// ValidateIPBlock validates a cidr and the except fields of an IpBlock NetworkPolicyPeer
func (v *NSNPValidator) ValidateIPBlock(ipb *k8snet.IPBlock, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(ipb.CIDR) == 0 || ipb.CIDR == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("cidr"), ""))
		return allErrs
	}
	cidrIPNet, err := ValidateCIDR(ipb.CIDR)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("cidr"), ipb.CIDR, "not a valid CIDR"))
		return allErrs
	}
	exceptCIDR := ipb.Except
	for i, exceptIP := range exceptCIDR {
		exceptPath := fldPath.Child("except").Index(i)
		exceptCIDR, err := ValidateCIDR(exceptIP)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(exceptPath, exceptIP, "not a valid CIDR"))
			return allErrs
		}
		if !cidrIPNet.Contains(exceptCIDR.IP) {
			allErrs = append(allErrs, field.Invalid(exceptPath, exceptCIDR.IP, "not within CIDR range"))
		}
	}
	return allErrs
}

// ValidateNSNPPeer validates a NetworkPolicyPeer
func (v *NSNPValidator) ValidateNSNPPeer(peer *networkv1alpha1.NetworkPolicyPeer, peerPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	numPeers := 0

	if peer.ServiceSelector != nil {
		numPeers++
		allErrs = append(allErrs, v.ValidateServiceSelector(peer.ServiceSelector, peerPath.Child("service"))...)
	}
	if peer.NamespaceSelector != nil {
		numPeers++
	}
	if peer.IPBlock != nil {
		numPeers++
		allErrs = append(allErrs, v.ValidateIPBlock(peer.IPBlock, peerPath.Child("ipBlock"))...)
	}

	if numPeers == 0 {
		allErrs = append(allErrs, field.Required(peerPath, "must specify a peer"))
	} else if numPeers > 1 && peer.IPBlock != nil {
		allErrs = append(allErrs, field.Forbidden(peerPath, "may not specify both ipBlock and another peer"))
	}

	return allErrs
}

func (v *NSNPValidator) ValidateNSNPSpec(spec *networkv1alpha1.NamespaceNetworkPolicySpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// Validate ingress rules.
	for i, ingress := range spec.Ingress {
		ingressPath := fldPath.Child("ingress").Index(i)
		for i, port := range ingress.Ports {
			portPath := ingressPath.Child("ports").Index(i)
			allErrs = append(allErrs, v.ValidateNetworkPolicyPort(&port, portPath)...)
		}
		for i, from := range ingress.From {
			fromPath := ingressPath.Child("from").Index(i)
			allErrs = append(allErrs, v.ValidateNSNPPeer(&from, fromPath)...)
		}
	}
	// Validate egress rules
	for i, egress := range spec.Egress {
		egressPath := fldPath.Child("egress").Index(i)
		for i, port := range egress.Ports {
			portPath := egressPath.Child("ports").Index(i)
			allErrs = append(allErrs, v.ValidateNetworkPolicyPort(&port, portPath)...)
		}
		for i, to := range egress.To {
			toPath := egressPath.Child("to").Index(i)
			allErrs = append(allErrs, v.ValidateNSNPPeer(&to, toPath)...)
		}
	}

	return allErrs
}
