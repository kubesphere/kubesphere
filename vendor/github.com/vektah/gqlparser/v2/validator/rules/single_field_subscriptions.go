package validator

import (
	"strconv"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
	. "github.com/vektah/gqlparser/v2/validator"
)

func init() {
	AddRule("SingleFieldSubscriptions", func(observers *Events, addError AddErrFunc) {
		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			if walker.Schema.Subscription == nil || operation.Operation != ast.Subscription {
				return
			}

			fields := retrieveTopFieldNames(operation.SelectionSet)

			name := "Anonymous Subscription"
			if operation.Name != "" {
				name = `Subscription ` + strconv.Quote(operation.Name)
			}

			if len(fields) > 1 {
				addError(
					Message(`%s must select only one top level field.`, name),
					At(fields[1].position),
				)
			}

			for _, field := range fields {
				if strings.HasPrefix(field.name, "__") {
					addError(
						Message(`%s must not select an introspection top level field.`, name),
						At(field.position),
					)
				}
			}
		})
	})
}

type topField struct {
	name     string
	position *ast.Position
}

func retrieveTopFieldNames(selectionSet ast.SelectionSet) []*topField {
	fields := []*topField{}
	inFragmentRecursive := map[string]bool{}
	var walk func(selectionSet ast.SelectionSet)
	walk = func(selectionSet ast.SelectionSet) {
		for _, selection := range selectionSet {
			switch selection := selection.(type) {
			case *ast.Field:
				fields = append(fields, &topField{
					name:     selection.Name,
					position: selection.GetPosition(),
				})
			case *ast.InlineFragment:
				walk(selection.SelectionSet)
			case *ast.FragmentSpread:
				if selection.Definition == nil {
					return
				}
				fragment := selection.Definition.Name
				if !inFragmentRecursive[fragment] {
					inFragmentRecursive[fragment] = true
					walk(selection.Definition.SelectionSet)
				}
			}
		}
	}
	walk(selectionSet)

	seen := make(map[string]bool, len(fields))
	uniquedFields := make([]*topField, 0, len(fields))
	for _, field := range fields {
		if !seen[field.name] {
			uniquedFields = append(uniquedFields, field)
		}
		seen[field.name] = true
	}
	return uniquedFields
}
