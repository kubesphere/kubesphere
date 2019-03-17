package checkers

import (
	"github.com/kiali/kiali/business/checkers/gateways"
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
)

const GatewayCheckerType = "gateway"

type GatewayChecker struct {
	GatewaysPerNamespace [][]kubernetes.IstioObject
	Namespace            string
}

// Check runs checks for the all namespaces actions as well as for the single namespace validations
func (g GatewayChecker) Check() models.IstioValidations {
	// Multinamespace checkers
	validations := gateways.MultiMatchChecker{
		GatewaysPerNamespace: g.GatewaysPerNamespace,
	}.Check()

	// Single namespace
	for _, nssGw := range g.GatewaysPerNamespace {
		for _, gw := range nssGw {
			if gw.GetObjectMeta().Namespace == g.Namespace {
				validations.MergeValidations(runSingleChecks(gw))
			}
		}
	}

	return validations
}

func runSingleChecks(gw kubernetes.IstioObject) models.IstioValidations {
	validations := models.IstioValidations{}
	checks, valid := gateways.PortChecker{
		Gateway: gw,
	}.Check()

	if !valid {
		key := models.IstioValidationKey{ObjectType: GatewayCheckerType, Name: gw.GetObjectMeta().Name}
		validations[key] = &models.IstioValidation{
			Name:       key.Name,
			ObjectType: key.ObjectType,
			Checks:     checks,
			Valid:      valid,
		}
	}
	return validations
}
