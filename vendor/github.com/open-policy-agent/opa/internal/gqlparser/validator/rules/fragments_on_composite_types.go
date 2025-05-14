package validator

import (
	"fmt"

	"github.com/open-policy-agent/opa/internal/gqlparser/ast"

	//nolint:revive // Validator rules each use dot imports for convenience.
	. "github.com/open-policy-agent/opa/internal/gqlparser/validator"
)

func init() {
	AddRule("FragmentsOnCompositeTypes", func(observers *Events, addError AddErrFunc) {
		observers.OnInlineFragment(func(walker *Walker, inlineFragment *ast.InlineFragment) {
			fragmentType := walker.Schema.Types[inlineFragment.TypeCondition]
			if fragmentType == nil || fragmentType.IsCompositeType() {
				return
			}

			message := fmt.Sprintf(`Fragment cannot condition on non composite type "%s".`, inlineFragment.TypeCondition)

			addError(
				Message(message), //nolint:govet
				At(inlineFragment.Position),
			)
		})

		observers.OnFragment(func(_ *Walker, fragment *ast.FragmentDefinition) {
			if fragment.Definition == nil || fragment.TypeCondition == "" || fragment.Definition.IsCompositeType() {
				return
			}

			message := fmt.Sprintf(`Fragment "%s" cannot condition on non composite type "%s".`, fragment.Name, fragment.TypeCondition)

			addError(
				Message(message), //nolint:govet
				At(fragment.Position),
			)
		})
	})
}
