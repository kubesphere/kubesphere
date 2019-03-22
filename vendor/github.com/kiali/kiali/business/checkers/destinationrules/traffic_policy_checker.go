package destinationrules

import (
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
)

type TrafficPolicyChecker struct {
	DestinationRules []kubernetes.IstioObject
	MTLSDetails      kubernetes.MTLSDetails
}

func (t TrafficPolicyChecker) Check() models.IstioValidations {
	validations := models.IstioValidations{}

	// When mTLS is not enabled, there is no validation to be added.
	if !t.isNonLocalmTLSEnabled() {
		return validations
	}

	// Check whether DRs override mTLS.
	for _, dr := range t.DestinationRules {
		if !hasTrafficPolicy(dr) || !hasTLSSettings(dr) {
			check := models.Build("destinationrules.trafficpolicy.notlssettings", "spec/trafficPolicy")
			key := models.BuildKey(DestinationRulesCheckerType, dr.GetObjectMeta().Name)
			validation := buildDestinationRuleValidation(dr, check, true)

			if _, exists := validations[key]; !exists {
				validations.MergeValidations(models.IstioValidations{key: validation})
			}
		}
	}

	return validations
}

func (t TrafficPolicyChecker) isNonLocalmTLSEnabled() bool {
	for _, dr := range t.MTLSDetails.DestinationRules {
		if host, ok := dr.GetSpec()["host"]; ok {
			if dHost, ok := host.(string); ok {
				fqdn := kubernetes.ParseHost(dHost, dr.GetObjectMeta().Namespace, dr.GetObjectMeta().ClusterName)
				if isNonLocalmTLSForServiceEnabled(dr, fqdn.Service) {
					return true
				}
			}
		}
	}
	return false
}

func hasTrafficPolicy(dr kubernetes.IstioObject) bool {
	_, trafficPresent := dr.GetSpec()["trafficPolicy"]
	return trafficPresent
}

func hasTLSSettings(dr kubernetes.IstioObject) bool {
	return hasTrafficPolicyTLS(dr) || hasPortTLS(dr)
}

// hasPortTLS returns true when there is one port that specifies any TLS settings
func hasPortTLS(dr kubernetes.IstioObject) bool {
	if trafficPolicy, trafficPresent := dr.GetSpec()["trafficPolicy"]; trafficPresent {
		if trafficCasted, ok := trafficPolicy.(map[string]interface{}); ok {
			if portsSettings, found := trafficCasted["portLevelSettings"]; found {
				if portsSettingsCasted, ok := portsSettings.([]interface{}); ok {
					for _, portSettings := range portsSettingsCasted {
						if portSettingsCasted, ok := portSettings.(map[string]interface{}); ok {
							if _, found := portSettingsCasted["tls"]; found {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

// hasTrafficPolicyTLS returns true when there is a trafficPolicy specifying any tls mode
func hasTrafficPolicyTLS(dr kubernetes.IstioObject) bool {
	if trafficPolicy, trafficPresent := dr.GetSpec()["trafficPolicy"]; trafficPresent {
		if trafficCasted, ok := trafficPolicy.(map[string]interface{}); ok {
			if _, found := trafficCasted["tls"]; found {
				return true
			}
		}
	}
	return false
}

func buildDestinationRuleValidation(dr kubernetes.IstioObject, checks models.IstioCheck, valid bool) *models.IstioValidation {
	validation := &models.IstioValidation{
		Name:       dr.GetObjectMeta().Name,
		ObjectType: DestinationRulesCheckerType,
		Valid:      valid,
		Checks: []*models.IstioCheck{
			&checks,
		},
	}

	return validation
}
