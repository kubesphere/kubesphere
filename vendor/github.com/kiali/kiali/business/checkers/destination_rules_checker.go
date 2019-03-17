package checkers

import (
	"github.com/kiali/kiali/business/checkers/destinationrules"
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
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
