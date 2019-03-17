package destinationrules

import (
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
)

const DestinationRulesCheckerType = "destinationrule"

type MultiMatchChecker struct {
	DestinationRules []kubernetes.IstioObject
}

type subset struct {
	Name     string
	RuleName string
}

// Check validates that no two destinationRules target the same host+subset combination
func (m MultiMatchChecker) Check() models.IstioValidations {
	validations := models.IstioValidations{}

	// Equality search is: [fqdn][subset]
	seenHostSubsets := make(map[string]map[string]string)

	for _, dr := range m.DestinationRules {
		if host, ok := dr.GetSpec()["host"]; ok {
			destinationRulesName := dr.GetObjectMeta().Name
			if dHost, ok := host.(string); ok {
				fqdn := kubernetes.ParseHost(dHost, dr.GetObjectMeta().Namespace, dr.GetObjectMeta().ClusterName)

				// Skip DR validation if it enables mTLS either namespace or mesh-wide
				if isNonLocalmTLSForServiceEnabled(dr, fqdn.Service) {
					continue
				}

				foundSubsets := extractSubsets(dr, destinationRulesName)

				if fqdn.Service == "*" {
					// We need to check the matching subsets from all hosts now
					for _, h := range seenHostSubsets {
						checkCollisions(validations, destinationRulesName, foundSubsets, h)
					}
					// We add * later
				}
				// Search "*" first and then exact name
				if previous, found := seenHostSubsets["*"]; found {
					// Need to check subsets of "*"
					checkCollisions(validations, destinationRulesName, foundSubsets, previous)
				}

				if previous, found := seenHostSubsets[fqdn.Service]; found {
					// Host found, need to check underlying subsets
					checkCollisions(validations, destinationRulesName, foundSubsets, previous)
				}
				// Nothing threw an error, so add these
				if _, found := seenHostSubsets[fqdn.Service]; !found {
					seenHostSubsets[fqdn.Service] = make(map[string]string)
				}
				for _, s := range foundSubsets {
					seenHostSubsets[fqdn.Service][s.Name] = destinationRulesName
				}
			}
		}
	}

	return validations
}

func isNonLocalmTLSForServiceEnabled(dr kubernetes.IstioObject, service string) bool {
	return service == "*" && ismTLSEnabled(dr)
}

func ismTLSEnabled(dr kubernetes.IstioObject) bool {
	if trafficPolicy, trafficPresent := dr.GetSpec()["trafficPolicy"]; trafficPresent {
		if trafficCasted, ok := trafficPolicy.(map[string]interface{}); ok {
			if tls, found := trafficCasted["tls"]; found {
				if tlsCasted, ok := tls.(map[string]interface{}); ok {
					if mode, found := tlsCasted["mode"]; found {
						if modeCasted, ok := mode.(string); ok {
							return modeCasted == "ISTIO_MUTUAL"
						}
					}
				}
			}
		}
	}
	return false
}

func extractSubsets(dr kubernetes.IstioObject, destinationRulesName string) []subset {
	if subsets, found := dr.GetSpec()["subsets"]; found {
		if subsetSlice, ok := subsets.([]interface{}); ok {
			foundSubsets := make([]subset, 0, len(subsetSlice))
			for _, se := range subsetSlice {
				if element, ok := se.(map[string]interface{}); ok {
					if name, found := element["name"]; found {
						if n, ok := name.(string); ok {
							foundSubsets = append(foundSubsets, subset{n, destinationRulesName})
						}
					}
				}
			}
			return foundSubsets
		}
	}
	// Matches all the subsets:~
	return []subset{subset{"~", destinationRulesName}}
}

func checkCollisions(validations models.IstioValidations, destinationRulesName string, foundSubsets []subset, existing map[string]string) {
	// If current subset is ~
	if len(foundSubsets) == 1 && foundSubsets[0].Name == "~" {
		// This should match any subset in the same hostname
		for _, v := range existing {
			addError(validations, []string{destinationRulesName, v})
		}
	}

	// If we have existing subset with ~
	if ruleName, found := existing["~"]; found {
		addError(validations, []string{destinationRulesName, ruleName})
	}

	for _, s := range foundSubsets {
		if ruleName, found := existing[s.Name]; found {
			addError(validations, []string{destinationRulesName, ruleName})
		}
	}
}

func addError(validations models.IstioValidations, destinationRuleNames []string) models.IstioValidations {
	for _, destinationRuleName := range destinationRuleNames {
		key := models.IstioValidationKey{Name: destinationRuleName, ObjectType: DestinationRulesCheckerType}
		checks := models.Build("destinationrules.multimatch", "spec/host")
		rrValidation := &models.IstioValidation{
			Name:       destinationRuleName,
			ObjectType: DestinationRulesCheckerType,
			Valid:      true,
			Checks: []*models.IstioCheck{
				&checks,
			},
		}

		if _, exists := validations[key]; !exists {
			validations.MergeValidations(models.IstioValidations{key: rrValidation})
		}
	}
	return validations
}
