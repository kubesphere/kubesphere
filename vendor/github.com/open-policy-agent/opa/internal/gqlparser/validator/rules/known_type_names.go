package validator

import (
	"github.com/open-policy-agent/opa/internal/gqlparser/ast"

	//nolint:revive // Validator rules each use dot imports for convenience.
	. "github.com/open-policy-agent/opa/internal/gqlparser/validator"
)

func init() {
	AddRule("KnownTypeNames", func(observers *Events, addError AddErrFunc) {
		observers.OnVariable(func(walker *Walker, variable *ast.VariableDefinition) {
			typeName := variable.Type.Name()
			typdef := walker.Schema.Types[typeName]
			if typdef != nil {
				return
			}

			addError(
				Message(`Unknown type "%s".`, typeName),
				At(variable.Position),
			)
		})

		observers.OnInlineFragment(func(walker *Walker, inlineFragment *ast.InlineFragment) {
			typedName := inlineFragment.TypeCondition
			if typedName == "" {
				return
			}

			def := walker.Schema.Types[typedName]
			if def != nil {
				return
			}

			addError(
				Message(`Unknown type "%s".`, typedName),
				At(inlineFragment.Position),
			)
		})

		observers.OnFragment(func(walker *Walker, fragment *ast.FragmentDefinition) {
			typeName := fragment.TypeCondition
			def := walker.Schema.Types[typeName]
			if def != nil {
				return
			}

			var possibleTypes []string
			for _, t := range walker.Schema.Types {
				possibleTypes = append(possibleTypes, t.Name)
			}

			addError(
				Message(`Unknown type "%s".`, typeName),
				SuggestListQuoted("Did you mean", typeName, possibleTypes),
				At(fragment.Position),
			)
		})
	})
}
