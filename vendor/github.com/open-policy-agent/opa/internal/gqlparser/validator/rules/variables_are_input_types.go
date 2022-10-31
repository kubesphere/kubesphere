package validator

import (
	"github.com/open-policy-agent/opa/internal/gqlparser/ast"

	//nolint:revive // Validator rules each use dot imports for convenience.
	. "github.com/open-policy-agent/opa/internal/gqlparser/validator"
)

func init() {
	AddRule("VariablesAreInputTypes", func(observers *Events, addError AddErrFunc) {
		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			for _, def := range operation.VariableDefinitions {
				if def.Definition == nil {
					continue
				}
				if !def.Definition.IsInputType() {
					addError(
						Message(
							`Variable "$%s" cannot be non-input type "%s".`,
							def.Variable,
							def.Type.String(),
						),
						At(def.Position),
					)
				}
			}
		})
	})
}
