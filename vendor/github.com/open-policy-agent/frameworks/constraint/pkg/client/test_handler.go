package client

import (
	"encoding/json"
	"text/template"

	"github.com/open-policy-agent/frameworks/constraint/pkg/types"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ TargetHandler = &handler{}

type handler struct{}

func (h *handler) GetName() string {
	return "test.target"
}

var libTempl = template.Must(template.New("library").Parse(`
package foo

autoreject_review[rejection] {
	constraint := {{.ConstraintsRoot}}[_][_]
	spec := get_default(constraint, "spec", {})
	match := get_default(spec, "match", {})
	has_field(match, "namespaceSelector")
	not {{.DataRoot}}.cluster["v1"]["Namespace"]
	rejection := {
		"msg": "REJECTION",
		"details": {},
		"constraint": constraint,
	}
}

matching_constraints[constraint] {
	constraint = {{.ConstraintsRoot}}[input.review.ForConstraint][_]
}

matching_reviews_and_constraints[[review, constraint]] {
	matching_constraints[constraint] with input as {"review": review}
	review = {{.DataRoot}}[_]
}

has_field(object, field) = true {
	object[field]
}

has_field(object, field) = true {
object[field] == false
}

has_field(object, field) = false {
not object[field]
not object[field] == false
}

get_default(object, field, _default) = output {
has_field(object, field)
output = object[field]
}

get_default(object, field, _default) = output {
has_field(object, field) == false
output = _default
}

`))

func (h *handler) Library() *template.Template {
	return libTempl
}

func (h *handler) ProcessData(obj interface{}) (bool, string, interface{}, error) {
	switch data := obj.(type) {
	case targetData:
		return true, data.Name, &data, nil
	case *targetData:
		return true, data.Name, data, nil
	}

	return false, "", nil, nil
}

func (h *handler) HandleReview(obj interface{}) (bool, interface{}, error) {
	handled, _, review, err := h.ProcessData(obj)
	return handled, review, err
}

func (h *handler) HandleViolation(result *types.Result) error {
	res, err := json.Marshal(result.Review)
	if err != nil {
		return err
	}
	d := &targetData{}
	if err := json.Unmarshal(res, d); err != nil {
		return err
	}
	result.Resource = d
	return nil
}

func (h *handler) MatchSchema() apiextensions.JSONSchemaProps {
	return apiextensions.JSONSchemaProps{
		Type: "object",
		Properties: map[string]apiextensions.JSONSchemaProps{
			"label": {Type: "string"},
		},
	}
}

func (h *handler) ValidateConstraint(u *unstructured.Unstructured) error {
	return nil
}

type targetData struct {
	Name          string
	ForConstraint string
}
