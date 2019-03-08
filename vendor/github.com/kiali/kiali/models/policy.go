package models

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiali/kiali/kubernetes"
)

type Policies []Policy
type Policy struct {
	Metadata meta_v1.ObjectMeta `json:"metadata"`
	Spec     struct {
		Targets          interface{} `json:"targets"`
		Peers            interface{} `json:"peers"`
		PeerIsOptional   interface{} `json:"peerIsOptional"`
		Origins          interface{} `json:"origins"`
		OriginIsOptional interface{} `json:"originIsOptional"`
		PrincipalBinding interface{} `json:"principalBinding"`
	} `json:"spec"`
}

func (ps *Policies) Parse(policies []kubernetes.IstioObject) {
	for _, qs := range policies {
		policy := Policy{}
		policy.Parse(qs)
		*ps = append(*ps, policy)
	}
}

func (p *Policy) Parse(policy kubernetes.IstioObject) {
	p.Metadata = policy.GetObjectMeta()
	p.Spec.Targets = policy.GetSpec()["targets"]
	p.Spec.Peers = policy.GetSpec()["peers"]
	p.Spec.PeerIsOptional = policy.GetSpec()["peersIsOptional"]
	p.Spec.Origins = policy.GetSpec()["origins"]
	p.Spec.OriginIsOptional = policy.GetSpec()["originIsOptinal"]
	p.Spec.PrincipalBinding = policy.GetSpec()["principalBinding"]
}
