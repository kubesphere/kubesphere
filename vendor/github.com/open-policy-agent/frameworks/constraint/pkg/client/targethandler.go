package client

import (
	"text/template"

	"github.com/open-policy-agent/frameworks/constraint/pkg/types"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type TargetHandler interface {
	MatchSchemaProvider

	// GetName returns name of the target. Must match `^[a-zA-Z][a-zA-Z0-9.]*$`
	// This will be the exact name of the field in the ConstraintTemplate
	// spec.target object, so if GetName returns validation.xyz.org, the user
	// will populate target specific rego into .spec.targets."validation.xyz.org".
	GetName() string

	// Library returns the pieces of Rego code required to stitch together constraint
	// evaluation for the target.
	//
	// Templated Parameters:
	//	`{{.ConstraintsRoot}}`: All constraints are found here and are placed under
	//		data["{{.ConstraintsRoot}}"][KIND][NAME] where $KIND is the Kind indicated in the
	//		CT .spec.crd.spec.names.kind field and NAME is the .metadata.name of the
	//		Constraint resource.
	//	`{{.DataRoot}}`: The data root, use data["{{.DataRoot}}"] instead of
	//		"data.inventory"
	//
	//
	// Required Rules:
	// `matching_constraints[constraint]`
	//		summary:
	// 			This rule defines constraint as any constraint where the spec.match
	// 			field evaluates to true for input.review.
	//		params:
	//			constraint: the matching constraint, typically you will evaluate this
	//				this against the list of all constraints `{{.ConstraintsRoot}}[_][_]`
	//
	// `matching_reviews_and_constraints[[review, constraint]]`
	//		summary:
	//			This rule evaluates creating all reviews for items in the inventory
	//		params:
	//			review: the value for input.review when evaluating `matching_constraints`
	//			constraint: constraint that satisfies `matching_constraints`
	//
	// Optional Rules:
	// `autoreject_review[rejection]`
	//		summary:
	//			This rule serves to indicate if the match field cannot be evaluated.
	//			It should define rejection as a rejection message if a constraint's
	//			match field will effectively "error out" while being evaluated.
	//		params:
	//			rejection: the rejection message
	//
	//
	// Libraries are currently templates that have the following parameters:
	//   ConstraintsRoot: The root path under which all constraints for the target are stored
	//   DataRoot: The root path under which all data for the target is stored
	Library() *template.Template

	// ProcessData takes inputs to AddData and converts them into the format that
	// will be stored in data.inventory and returns the relative storage path.
	// Args:
	//	data: the object passed to client.Client.AddData
	// Returns:
	//	handle: true if the target handles the data type
	//	relPath: the relative path under which the data should be stored in OPA under data.inventory, for example, an item to be stored at data.inventory.x.y.z would return x.y.z
	//	inventoryFormat: the data as an object that can be cast into JSON and suitable for storage in the inventory
	//	err: any error encountered
	ProcessData(data interface{}) (handle bool, relPath string, inventoryFormat interface{}, err error)

	// HandleReview determines if this target handler will handle an individual
	// resource review and if so, builds the `review` field of the input object.
	// Args:
	//	object: the object passed to client.Client.Review
	// Returns:
	//	handle: true if the target handler will review this input
	//	review: the data for the `review` field
	//	err: any error encountered
	HandleReview(object interface{}) (handle bool, review interface{}, err error)

	// HandleViolation allows for post-processing of the result object. The object
	// can be mutated if desired.
	// Args:
	//	result: the result generated from the violation rule
	HandleViolation(result *types.Result) error

	// ValidateConstraint returns an error if constraint is not valid in any way.
	// This allows for semantic validation beyond OpenAPI validation given by the
	// spec from MatchSchema().
	ValidateConstraint(constraint *unstructured.Unstructured) error
}

type MatchSchemaProvider interface {
	// MatchSchema returns the JSON Schema for the `match` field of a constraint
	MatchSchema() apiextensions.JSONSchemaProps
}
