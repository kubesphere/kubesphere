package virtual_services

import (
	"fmt"
	"strings"

	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
)

type NoHostChecker struct {
	Namespace         string
	ServiceNames      []string
	VirtualService    kubernetes.IstioObject
	ServiceEntryHosts map[string][]string
}

func (n NoHostChecker) Check() ([]*models.IstioCheck, bool) {
	validations := make([]*models.IstioCheck, 0)

	routeProtocols := []string{"http", "tcp", "tls"}
	countOfDefinedProtocols := 0
	for _, protocol := range routeProtocols {
		if prot, ok := n.VirtualService.GetSpec()[protocol]; ok {
			countOfDefinedProtocols++
			if aHttp, ok := prot.([]interface{}); ok {
				for k, httpRoute := range aHttp {
					if mHttpRoute, ok := httpRoute.(map[string]interface{}); ok {
						if route, ok := mHttpRoute["route"]; ok {
							if aDestinationWeight, ok := route.([]interface{}); ok {
								for i, destination := range aDestinationWeight {
									if !n.checkDestination(destination, protocol) {
										validation := models.Build("virtualservices.nohost.hostnotfound",
											fmt.Sprintf("spec/%s[%d]/route[%d]/destination/host", protocol, k, i))
										validations = append(validations, &validation)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if countOfDefinedProtocols < 1 {
		validation := models.Build("virtualservices.nohost.invalidprotocol", "")
		validations = append(validations, &validation)
	}

	return validations, len(validations) == 0
}

func (n NoHostChecker) checkDestination(destination interface{}, protocol string) bool {
	if mDestination, ok := destination.(map[string]interface{}); ok {
		if destinationW, ok := mDestination["destination"]; ok {
			if mDestinationW, ok := destinationW.(map[string]interface{}); ok {
				if host, ok := mDestinationW["host"]; ok {
					if sHost, ok := host.(string); ok {
						for _, service := range n.ServiceNames {
							if kubernetes.FilterByHost(sHost, service, n.Namespace) {
								return true
							}
						}
						if protocols, found := n.ServiceEntryHosts[sHost]; found {
							// We have ServiceEntry to check
							for _, prot := range protocols {
								if prot == strings.ToLower(protocol) {
									return true
								}
							}
						}
					}
				}
			}
		}
	}
	return false
}
