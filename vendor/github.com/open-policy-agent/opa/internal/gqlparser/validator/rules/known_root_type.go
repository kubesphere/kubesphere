package validator

import (
	"fmt"

	"github.com/open-policy-agent/opa/internal/gqlparser/ast"

	//nolint:revive // Validator rules each use dot imports for convenience.
	. "github.com/open-policy-agent/opa/internal/gqlparser/validator"
)

func init() {
	AddRule("KnownRootType", func(observers *Events, addError AddErrFunc) {
		// A query's root must be a valid type.  Surprisingly, this isn't
		// checked anywhere else!
		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			var def *ast.Definition
			switch operation.Operation {
			case ast.Query, "":
				def = walker.Schema.Query
			case ast.Mutation:
				def = walker.Schema.Mutation
			case ast.Subscription:
				def = walker.Schema.Subscription
			default:
				// This shouldn't even parse; if it did we probably need to
				// update this switch block to add the new operation type.
				panic(fmt.Sprintf(`got unknown operation type "%s"`, operation.Operation))
			}
			if def == nil {
				addError(
					Message(`Schema does not support operation type "%s"`, operation.Operation),
					At(operation.Position))
			}
		})
	})
}
