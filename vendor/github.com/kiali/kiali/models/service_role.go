package models

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiali/kiali/kubernetes"
)

type ServiceRoles []ServiceRole
type ServiceRole struct {
	Metadata meta_v1.ObjectMeta `json:"metadata"`
	Spec     struct {
		Rules interface{} `json:"rules"`
	} `json:"spec"`
}

func (srs *ServiceRoles) Parse(serviceRoles []kubernetes.IstioObject) {
	for _, sr := range serviceRoles {
		serviceRole := ServiceRole{}
		serviceRole.Parse(sr)
		*srs = append(*srs, serviceRole)
	}
}

func (sr *ServiceRole) Parse(serviceRole kubernetes.IstioObject) {
	sr.Metadata = serviceRole.GetObjectMeta()
	sr.Spec.Rules = serviceRole.GetSpec()["rules"]
}
