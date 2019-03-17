package virtual_services

import (
	"strconv"

	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
)

type NoGatewayChecker struct {
	VirtualService kubernetes.IstioObject
	GatewayNames   map[string]struct{}
}

// Check validates that all the VirtualServices are pointing to an existing Gateway
func (s NoGatewayChecker) Check() ([]*models.IstioCheck, bool) {
	validations := make([]*models.IstioCheck, 0)

	valid, index := kubernetes.ValidateVirtualServiceGateways(s.VirtualService.GetSpec(), s.GatewayNames, s.VirtualService.GetObjectMeta().Namespace, s.VirtualService.GetObjectMeta().ClusterName)
	if !valid {
		path := "spec/gateways[" + strconv.Itoa(index) + "]"
		validation := models.Build("virtualservices.nogateway", path)
		validations = append(validations, &validation)
	}

	return validations, valid
}
