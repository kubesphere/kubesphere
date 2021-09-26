package regolib

const (
	targetLibSrc = `
package hooks["{{.Target}}"]

violation[response] {
	data.hooks["{{.Target}}"].library.autoreject_review[rejection]
	review := get_default(input, "review", {})
	constraint := get_default(rejection, "constraint", {})
	spec := get_default(constraint, "spec", {})
	enforcementAction := get_default(spec, "enforcementAction", "deny")
	response = {
		"msg": get_default(rejection, "msg", ""),
		"metadata": {"details": get_default(rejection, "details", {})},
		"constraint": constraint,
		"review": review,
		"enforcementAction": enforcementAction,
	}
}

# Finds all violations for a given target
violation[response] {
	data.hooks["{{.Target}}"].library.matching_constraints[constraint]
	review := get_default(input, "review", {})
	inp := {
		"review": review,
		"parameters": get_default(get_default(constraint, "spec", {}), "parameters", {}),
	}
	inventory[inv]
	data.templates["{{.Target}}"][constraint.kind].violation[r] with input as inp with data.inventory as inv
	spec := get_default(constraint, "spec", {})
	enforcementAction := get_default(spec, "enforcementAction", "deny")
	response = {
		"msg": r.msg,
		"metadata": {"details": get_default(r, "details", {})},
		"constraint": constraint,
		"review": review,
		"enforcementAction": enforcementAction,
	}
}


# Finds all violations in the cached state of a given target
audit[response] {
	data.hooks["{{.Target}}"].library.matching_reviews_and_constraints[[review, constraint]]
	inp := {
		"review": review,
		"parameters": get_default(get_default(constraint, "spec", {}), "parameters", {}),
	}
	inventory[inv]
	data.templates["{{.Target}}"][constraint.kind].violation[r] with input as inp with data.inventory as inv
	spec := get_default(constraint, "spec", {})
	enforcementAction := get_default(spec, "enforcementAction", "deny")
	response = {
		"msg": r.msg,
		"metadata": {"details": get_default(r, "details", {})},
		"constraint": constraint,
		"review": review,
		"enforcementAction": enforcementAction,
	}
}

# get_default(data, "external", {}) seems to cause this error:
# "rego_type_error: undefined function data.hooks.<target>.get_default"
inventory[inv] {
	inv = data.external["{{.Target}}"]
}

inventory[{}] {
	not data.external["{{.Target}}"]
}

# get_default returns the value of an object's field or the provided default value.
# It avoids creating an undefined state when trying to access an object attribute that does
# not exist
get_default(object, field, _default) = object[field]

get_default(object, field, _default) = _default {
  not has_field(object, field)
}

has_field(object, field) {
  _ = object[field]
}
`
)
