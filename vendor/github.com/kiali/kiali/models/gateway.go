package models

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiali/kiali/kubernetes"
)

type Gateways []Gateway
type Gateway struct {
	Metadata meta_v1.ObjectMeta `json:"metadata"`
	Spec     struct {
		Servers  interface{} `json:"servers"`
		Selector interface{} `json:"selector"`
	} `json:"spec"`
}

func (gws *Gateways) Parse(gateways []kubernetes.IstioObject) {
	for _, gw := range gateways {
		gateway := Gateway{}
		gateway.Parse(gw)
		*gws = append(*gws, gateway)
	}
}

func (gw *Gateway) Parse(gateway kubernetes.IstioObject) {
	gw.Metadata = gateway.GetObjectMeta()
	gw.Spec.Servers = gateway.GetSpec()["servers"]
	gw.Spec.Selector = gateway.GetSpec()["selector"]
}
