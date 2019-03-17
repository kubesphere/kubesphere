package models

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiali/kiali/kubernetes"
)

type ClusterRbacConfigs []ClusterRbacConfig
type ClusterRbacConfig struct {
	Metadata meta_v1.ObjectMeta `json:"metadata"`
	Spec     struct {
		Mode      interface{} `json:"mode"`
		Inclusion interface{} `json:"inclusion"`
		Exclusion interface{} `json:"exclusion"`
	} `json:"spec"`
}

func (rcs *ClusterRbacConfigs) Parse(clusterRbacConfigs []kubernetes.IstioObject) {
	for _, rc := range clusterRbacConfigs {
		clusterRbacConfig := ClusterRbacConfig{}
		clusterRbacConfig.Parse(rc)
		*rcs = append(*rcs, clusterRbacConfig)
	}
}

func (rc *ClusterRbacConfig) Parse(clusterRbacConfig kubernetes.IstioObject) {
	rc.Metadata = clusterRbacConfig.GetObjectMeta()
	rc.Spec.Mode = clusterRbacConfig.GetSpec()["mode"]
	rc.Spec.Inclusion = clusterRbacConfig.GetSpec()["inclusion"]
	rc.Spec.Exclusion = clusterRbacConfig.GetSpec()["exclusion"]
}
