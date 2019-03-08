package models

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiali/kiali/kubernetes"
)

type ServiceEntries []ServiceEntry
type ServiceEntry struct {
	Metadata meta_v1.ObjectMeta `json:"metadata"`
	Spec     struct {
		Hosts      interface{} `json:"hosts"`
		Addresses  interface{} `json:"addresses"`
		Ports      interface{} `json:"ports"`
		Location   interface{} `json:"location"`
		Resolution interface{} `json:"resolution"`
		Endpoints  interface{} `json:"endpoints"`
	} `json:"spec"`
}

func (ses *ServiceEntries) Parse(serviceEntries []kubernetes.IstioObject) {
	for _, se := range serviceEntries {
		serviceEntry := ServiceEntry{}
		serviceEntry.Parse(se)
		*ses = append(*ses, serviceEntry)
	}
}

func (se *ServiceEntry) Parse(serviceEntry kubernetes.IstioObject) {
	se.Metadata = serviceEntry.GetObjectMeta()
	se.Spec.Hosts = serviceEntry.GetSpec()["hosts"]
	se.Spec.Addresses = serviceEntry.GetSpec()["addresses"]
	se.Spec.Ports = serviceEntry.GetSpec()["ports"]
	se.Spec.Location = serviceEntry.GetSpec()["location"]
	se.Spec.Resolution = serviceEntry.GetSpec()["resolution"]
	se.Spec.Endpoints = serviceEntry.GetSpec()["endpoints"]
}
