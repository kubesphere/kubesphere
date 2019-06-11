package models

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiali/kiali/kubernetes"
)

type QuotaSpecs []QuotaSpec
type QuotaSpec struct {
	Metadata meta_v1.ObjectMeta `json:"metadata"`
	Spec     struct {
		Rules interface{} `json:"rules"`
	} `json:"spec"`
}

func (qss *QuotaSpecs) Parse(quotaSpecs []kubernetes.IstioObject) {
	for _, qs := range quotaSpecs {
		quotaSpec := QuotaSpec{}
		quotaSpec.Parse(qs)
		*qss = append(*qss, quotaSpec)
	}
}

func (qs *QuotaSpec) Parse(quotaSpec kubernetes.IstioObject) {
	qs.Metadata = quotaSpec.GetObjectMeta()
	qs.Spec.Rules = quotaSpec.GetSpec()["rules"]
}
