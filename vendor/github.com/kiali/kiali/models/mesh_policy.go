package models

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiali/kiali/kubernetes"
)

type MeshPolicies []MeshPolicy
type MeshPolicy struct {
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

func (mps *MeshPolicies) Parse(meshPolicies []kubernetes.IstioObject) {
	for _, qs := range meshPolicies {
		meshPolicy := MeshPolicy{}
		meshPolicy.Parse(qs)
		*mps = append(*mps, meshPolicy)
	}
}

func (mp *MeshPolicy) Parse(meshPolicy kubernetes.IstioObject) {
	mp.Metadata = meshPolicy.GetObjectMeta()
	mp.Spec.Targets = meshPolicy.GetSpec()["targets"]
	mp.Spec.Peers = meshPolicy.GetSpec()["peers"]
	mp.Spec.PeerIsOptional = meshPolicy.GetSpec()["peersIsOptional"]
	mp.Spec.Origins = meshPolicy.GetSpec()["origins"]
	mp.Spec.OriginIsOptional = meshPolicy.GetSpec()["originIsOptinal"]
	mp.Spec.PrincipalBinding = meshPolicy.GetSpec()["principalBinding"]
}
