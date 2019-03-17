package virtual_services

import (
	"fmt"
	"reflect"

	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/util/intutil"
)

type RouteChecker struct {
	Route kubernetes.IstioObject
}

// Check returns both an array of IstioCheck and a boolean indicating if the current route rule is valid.
// The array of IstioChecks contains the result of running the following validations:
// 1. All weights with a numeric number.
// 2. All weights have value between 0 and 100.
// 3. Sum of all weights are 100 (if only one weight, then it assumes that is 100).
// 4. All the route has to have weight label.
func (route RouteChecker) Check() ([]*models.IstioCheck, bool) {
	checks, valid := make([]*models.IstioCheck, 0), true
	protocols := []string{"http", "tcp", "tls"}

	for _, protocol := range protocols {
		cs, v := route.checkRoutesFor(protocol)
		checks = append(checks, cs...)
		valid = valid && v
	}

	return checks, valid
}

func (route RouteChecker) checkRoutesFor(kind string) ([]*models.IstioCheck, bool) {
	validations := make([]*models.IstioCheck, 0)
	weightSum, weightCount, valid := 0, 0, true

	http := route.Route.GetSpec()[kind]
	if http == nil {
		return validations, valid
	}

	// Getting a []HTTPRoute
	slice := reflect.ValueOf(http)
	if slice.Kind() != reflect.Slice {
		return validations, valid
	}

	for routeIdx := 0; routeIdx < slice.Len(); routeIdx++ {
		route, ok := slice.Index(routeIdx).Interface().(map[string]interface{})
		if !ok || route["route"] == nil {
			continue
		}

		weightCount, weightSum = 0, 0

		// Getting a []DestinationWeight
		destinationWeights := reflect.ValueOf(route["route"])
		if destinationWeights.Kind() != reflect.Slice {
			return validations, valid
		}

		for destWeightIdx := 0; destWeightIdx < destinationWeights.Len(); destWeightIdx++ {
			destinationWeight, ok := destinationWeights.Index(destWeightIdx).Interface().(map[string]interface{})
			if !ok || destinationWeight["weight"] == nil {
				continue
			}

			weightCount = weightCount + 1
			weight, err := intutil.Convert(destinationWeight["weight"])
			if err != nil {
				valid = false
				path := fmt.Sprintf("spec/%s[%d]/route[%d]/weight/%s",
					kind, routeIdx, destWeightIdx, destinationWeight["weight"])
				validation := buildValidation("virtualservices.route.numericweight", path)
				validations = append(validations, &validation)
			}

			if weight > 100 || weight < 0 {
				valid = false
				path := fmt.Sprintf("spec/%s[%d]/route[%d]/weight/%d",
					kind, routeIdx, destWeightIdx, weight)
				validation := buildValidation("virtualservices.route.weightrange", path)
				validations = append(validations, &validation)
			}

			weightSum = weightSum + weight
		}

		if weightCount > 0 && weightSum != 100 {
			valid = false
			path := fmt.Sprintf("spec/%s[%d]/route", kind, routeIdx)
			validation := buildValidation("virtualservices.route.weightsum", path)
			validations = append(validations, &validation)
			if weightCount != destinationWeights.Len() {
				valid = false
				path := fmt.Sprintf("spec/%s[%d]/route", kind, routeIdx)
				validation := buildValidation("virtualservices.route.allweightspresent", path)
				validations = append(validations, &validation)
			}
		}
	}

	return validations, valid
}

func buildValidation(checkId string, path string) models.IstioCheck {
	validation := models.Build(checkId, path)
	log.Infof("%s Galley should be performing this validation but it isn't. "+
		"Make sure Galley is fully working.", checkId)
	return validation
}
