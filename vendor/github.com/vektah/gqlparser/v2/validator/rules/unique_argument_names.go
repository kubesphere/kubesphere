package validator

import (
	"github.com/vektah/gqlparser/v2/ast"
	. "github.com/vektah/gqlparser/v2/validator"
)

func init() {
	AddRule("UniqueArgumentNames", func(observers *Events, addError AddErrFunc) {
		observers.OnField(func(walker *Walker, field *ast.Field) {
			checkUniqueArgs(field.Arguments, addError)
		})

		observers.OnDirective(func(walker *Walker, directive *ast.Directive) {
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
