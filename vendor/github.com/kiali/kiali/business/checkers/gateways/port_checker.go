package gateways

import (
	"fmt"

	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
)

type PortChecker struct {
	Gateway kubernetes.IstioObject
}

func (p PortChecker) Check() ([]*models.IstioCheck, bool) {
	validations := make([]*models.IstioCheck, 0)

	if serversSpec, found := p.Gateway.GetSpec()["servers"]; found {
		if servers, ok := serversSpec.([]interface{}); ok {
			for serverIndex, server := range servers {
				if serverDef, ok := server.(map[string]interface{}); ok {
					if portDef, found := serverDef["port"]; found {
						if !kubernetes.ValidatePort(portDef) {
							validation := models.Build("port.name.mismatch",
								fmt.Sprintf("spec/servers[%d]/port/name", serverIndex))
							validations = append(validations, &validation)
						}
					}
				}
			}
		}
	}

	return validations, len(validations) == 0
}
