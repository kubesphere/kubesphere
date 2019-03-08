package models

import (
	"encoding/json"
)

// NamespaceValidations represents a set of IstioValidations grouped by namespace
type NamespaceValidations map[string]IstioValidations

// IstioValidationKey is the key value composed of an Istio ObjectType and Name.
type IstioValidationKey struct {
	ObjectType string
	Name       string
}

// IstioValidations represents a set of IstioValidation grouped by IstioValidationKey.
type IstioValidations map[IstioValidationKey]*IstioValidation

// IstioValidation represents a list of checks associated to an Istio object.
// swagger:model
type IstioValidation struct {
	// Name of the object itself
	// required: true
	// example: reviews
	Name string `json:"name"`

	// Type of the object
	// required: true
	// example: virtualservice
	ObjectType string `json:"objectType"`

	// Represents validity of the object: in case of warning, validity remains as true
	// required: true
	// example: false
	Valid bool `json:"valid"`

	// Array of checks. It might be empty.
	Checks []*IstioCheck `json:"checks"`
}

// IstioCheck represents an individual check.
// swagger:model
type IstioCheck struct {
	// Description of the check
	// required: true
	// example: Weight sum should be 100
	Message string `json:"message"`

	// Indicates the level of importance: error or warning
	// required: true
	// example: error
	Severity SeverityLevel `json:"severity"`

	// String that describes where in the yaml file is the check located
	// example: spec/http[0]/route
	Path string `json:"path"`
}

type SeverityLevel string

const (
	ErrorSeverity   SeverityLevel = "error"
	WarningSeverity SeverityLevel = "warning"
)

var ObjectTypeSingular = map[string]string{
	"gateways":          "gateway",
	"virtualservices":   "virtualservice",
	"destinationrules":  "destinationrule",
	"serviceentries":    "serviceentry",
	"rules":             "rule",
	"quotaspecs":        "quotaspec",
	"quotaspecbindings": "quotaspecbinding",
}

var checkDescriptors = map[string]IstioCheck{
	"destinationrules.multimatch": {
		Message:  "More than one DestinationRules for the same host subset combination",
		Severity: WarningSeverity,
	},
	"destinationrules.nodest.matchingworkload": {
		Message:  "This host has no matching workloads",
		Severity: ErrorSeverity,
	},
	"destinationrules.nodest.subsetlabels": {
		Message:  "This subset's labels are not found in any matching host",
		Severity: ErrorSeverity,
	},
	"destinationrules.trafficpolicy.notlssettings": {
		Message:  "mTLS settings of a non-local Destination Rule are overridden",
		Severity: WarningSeverity,
	},
	"gateways.multimatch": {
		Message:  "More than one Gateway for the same host port combination",
		Severity: WarningSeverity,
	},
	"port.name.mismatch": {
		Message:  "Port name must follow <protocol>[-suffix] form",
		Severity: ErrorSeverity,
	},
	"virtualservices.nogateway": {
		Message:  "VirtualService is pointing to a non-existent gateway",
		Severity: ErrorSeverity,
	},
	"virtualservices.nohost.hostnotfound": {
		Message:  "DestinationWeight on route doesn't have a valid service (host not found)",
		Severity: ErrorSeverity,
	},
	"virtualservices.nohost.invalidprotocol": {
		Message:  "VirtualService doesn't define any valid route protocol",
		Severity: ErrorSeverity,
	},
	"virtualservices.route.numericweight": {
		Message:  "Weight must be a number",
		Severity: ErrorSeverity,
	},
	"virtualservices.route.weightrange": {
		Message:  "Weight should be between 0 and 100",
		Severity: ErrorSeverity,
	},
	"virtualservices.route.weightsum": {
		Message:  "Weight sum should be 100",
		Severity: ErrorSeverity,
	},
	"virtualservices.route.allweightspresent": {
		Message:  "All routes should have weight",
		Severity: WarningSeverity,
	},
	"virtualservices.singlehost": {
		Message:  "More than one Virtual Service for same host",
		Severity: WarningSeverity,
	},
	"virtualservices.subsetpresent.destinationmandatory": {
		Message:  "Destination field is mandatory",
		Severity: ErrorSeverity,
	},
	"virtualservices.subsetpresent.subsetnotfound": {
		Message:  "Subset not found",
		Severity: WarningSeverity,
	},
}

func Build(checkId string, path string) IstioCheck {
	check := checkDescriptors[checkId]
	check.Path = path
	return check
}

func BuildKey(objectType, name string) IstioValidationKey {
	return IstioValidationKey{ObjectType: objectType, Name: name}
}

func CheckMessage(checkId string) string {
	return checkDescriptors[checkId].Message
}

func (iv IstioValidations) FilterByKey(objectType, name string) IstioValidations {
	fiv := IstioValidations{}
	for k, v := range iv {
		if k.Name == name && k.ObjectType == objectType {
			fiv[k] = v
		}
	}

	return fiv
}

// FilterByTypes takes an input as ObjectTypes, transforms to singular types and filters the validations
func (iv IstioValidations) FilterByTypes(objectTypes []string) IstioValidations {
	types := make(map[string]bool, len(objectTypes))
	for _, objectType := range objectTypes {
		types[ObjectTypeSingular[objectType]] = true
	}
	fiv := IstioValidations{}
	for k, v := range iv {
		if _, found := types[k.ObjectType]; found {
			fiv[k] = v
		}
	}

	return fiv
}

func (iv IstioValidations) MergeValidations(validations IstioValidations) IstioValidations {
	for key, validation := range validations {
		v, ok := iv[key]
		if !ok {
			iv[key] = validation
		} else {
			v.Checks = append(v.Checks, validation.Checks...)
			v.Valid = v.Valid && validation.Valid
		}
	}
	return iv
}

// MarshalJSON implements the json.Marshaler interface.
func (iv IstioValidations) MarshalJSON() ([]byte, error) {
	out := make(map[string]map[string]*IstioValidation)
	for k, v := range iv {
		_, ok := out[k.ObjectType]
		if !ok {
			out[k.ObjectType] = make(map[string]*IstioValidation)
		}
		out[k.ObjectType][k.Name] = v
	}
	return json.Marshal(out)
}
