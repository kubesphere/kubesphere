package models

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiali/kiali/kubernetes"
)

// VirtualServices virtualServices
//
// This type is used for returning an array of VirtualServices with some permission flags
//
// swagger:model virtualServices
// An array of virtualService
// swagger:allOf
type VirtualServices struct {
	Permissions ResourcePermissions `json:"permissions"`
	Items       []VirtualService    `json:"items"`
}

// VirtualService virtualService
//
// This type is used for returning a VirtualService
//
// swagger:model virtualService
type VirtualService struct {
	Metadata meta_v1.ObjectMeta `json:"metadata"`
	Spec     struct {
		Hosts    interface{} `json:"hosts"`
		Gateways interface{} `json:"gateways"`
		Http     interface{} `json:"http"`
		Tcp      interface{} `json:"tcp"`
		Tls      interface{} `json:"tls"`
	} `json:"spec"`
}

func (vServices *VirtualServices) Parse(virtualServices []kubernetes.IstioObject) {
	vServices.Items = []VirtualService{}
	for _, vs := range virtualServices {
		virtualService := VirtualService{}
		virtualService.Parse(vs)
		vServices.Items = append(vServices.Items, virtualService)
	}
}

func (vService *VirtualService) Parse(virtualService kubernetes.IstioObject) {
	vService.Metadata = virtualService.GetObjectMeta()
	vService.Spec.Hosts = virtualService.GetSpec()["hosts"]
	vService.Spec.Gateways = virtualService.GetSpec()["gateways"]
	vService.Spec.Http = virtualService.GetSpec()["http"]
	vService.Spec.Tcp = virtualService.GetSpec()["tcp"]
	vService.Spec.Tls = virtualService.GetSpec()["tls"]
}

// IsValidHost returns true if VirtualService hosts applies to the service
func (vService *VirtualService) IsValidHost(namespace string, serviceName string) bool {
	if serviceName == "" {
		return false
	}
	if hosts, ok := vService.Spec.Hosts.([]interface{}); ok {
		for _, host := range hosts {
			if kubernetes.FilterByHost(host.(string), serviceName, namespace) {
				return true
			}
		}
	}
	return false
}
