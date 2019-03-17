package checkers

import (
	"github.com/kiali/kiali/business/checkers/virtual_services"
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
)

const VirtualCheckerType = "virtualservice"

type VirtualServiceChecker struct {
	Namespace        string
	DestinationRules []kubernetes.IstioObject
	VirtualServices  []kubernetes.IstioObject
}

// An Object Checker runs all checkers for an specific object type (i.e.: pod, route rule,...)
// It run two kinds of checkers:
// 1. Individual checks: validating individual objects.
// 2. Group checks: validating behaviour between configurations.
func (in VirtualServiceChecker) Check() models.IstioValidations {
	validations := models.IstioValidations{}

	validations = validations.MergeValidations(in.runIndividualChecks())
	validations = validations.MergeValidations(in.runGroupChecks())

	return validations
}

// Runs individual checks for each virtual service
func (in VirtualServiceChecker) runIndividualChecks() models.IstioValidations {
	validations := models.IstioValidations{}

	for _, virtualService := range in.VirtualServices {
		validations.MergeValidations(in.runChecks(virtualService))
	}

	return validations
}

// runGroupChecks runs group checks for all virtual services
func (in VirtualServiceChecker) runGroupChecks() models.IstioValidations {
	validations := models.IstioValidations{}

	enabledCheckers := []GroupChecker{
		virtual_services.SingleHostChecker{Namespace: in.Namespace, VirtualServices: in.VirtualServices},
	}

	for _, checker := range enabledCheckers {
		validations = validations.MergeValidations(checker.Check())
	}

	return validations
}

// runChecks runs all the individual checks for a single virtual service and appends the result into validations.
func (in VirtualServiceChecker) runChecks(virtualService kubernetes.IstioObject) models.IstioValidations {
	virtualServiceName := virtualService.GetObjectMeta().Name
	key := models.IstioValidationKey{Name: virtualServiceName, ObjectType: VirtualCheckerType}
	rrValidation := &models.IstioValidation{
		Name:       virtualServiceName,
		ObjectType: VirtualCheckerType,
		Valid:      true,
		// Explicitly create an empty array as 0-values do not appear in json
		Checks: []*models.IstioCheck{},
	}

	enabledCheckers := []Checker{
		virtual_services.RouteChecker{Route: virtualService},
		virtual_services.SubsetPresenceChecker{Namespace: in.Namespace, DestinationRules: in.DestinationRules, VirtualService: virtualService},
	}

	for _, checker := range enabledCheckers {
		checks, validChecker := checker.Check()
		rrValidation.Checks = append(rrValidation.Checks, checks...)
		rrValidation.Valid = rrValidation.Valid && validChecker
	}

	return models.IstioValidations{key: rrValidation}
}
