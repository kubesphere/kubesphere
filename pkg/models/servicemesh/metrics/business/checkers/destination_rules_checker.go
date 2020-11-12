package checkers

import (
	"kubesphere.io/kubesphere/pkg/models/servicemesh/metrics/business/checkers/destinationrules"
	"kubesphere.io/kubesphere/pkg/models/servicemesh/metrics/kubernetes"
	"kubesphere.io/kubesphere/pkg/models/servicemesh/metrics/models"
)

type DestinationRulesChecker struct {
	DestinationRules []kubernetes.IstioObject
	MTLSDetails      kubernetes.MTLSDetails
}

func (in DestinationRulesChecker) Check() models.IstioValidations {
	validations := models.IstioValidations{}

	enabledDRCheckers := []GroupChecker{
		destinationrules.MultiMatchChecker{DestinationRules: in.DestinationRules},
		destinationrules.TrafficPolicyChecker{DestinationRules: in.DestinationRules, MTLSDetails: in.MTLSDetails},
	}

	for _, checker := range enabledDRCheckers {
		validations = validations.MergeValidations(checker.Check())
	}

	return validations
}
