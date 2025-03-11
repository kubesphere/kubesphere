package validator

import (
	"github.com/open-policy-agent/opa/internal/gqlparser/ast"

	//nolint:revive // Validator rules each use dot imports for convenience.
	. "github.com/open-policy-agent/opa/internal/gqlparser/validator"
)

func init() {
	AddRule("UniqueArgumentNames", func(observers *Events, addError AddErrFunc) {
		observers.OnField(func(_ *Walker, field *ast.Field) {
			checkUniqueArgs(field.Arguments, addError)
		})

		observers.OnDirective(func(_ *Walker, directive *ast.Directive) {
			checkUniqueArgs(directive.Arguments, addError)
		})
	})
}

func checkUniqueArgs(args ast.ArgumentList, addError AddErrFunc) {
	knownArgNames := map[string]int{}

	for _, arg := range args {
		if knownArgNames[arg.Name] == 1 {
			addError(
				Message(`There can be only one argument named "%s".`, arg.Name),
				At(arg.Position),
			)
		}

		knownArgNames[arg.Name]++
	}
}
