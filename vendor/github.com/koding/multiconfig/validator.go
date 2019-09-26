package multiconfig

import (
	"fmt"
	"reflect"

	"github.com/fatih/structs"
)

// Validator validates the config against any predefined rules, those predefined
// rules should be given to this package. The implementer will be responsible
// for the logic.
type Validator interface {
	// Validate validates the config struct
	Validate(s interface{}) error
}

// RequiredValidator validates the struct against zero values.
type RequiredValidator struct {
	//  TagName holds the validator tag name. The default is "required"
	TagName string

	// TagValue holds the expected value of the validator. The default is "true"
	TagValue string
}

// Validate validates the given struct agaist field's zero values. If
// intentionaly, the value of a field is `zero-valued`(e.g false, 0, "")
// required tag should not be set for that field.
func (e *RequiredValidator) Validate(s interface{}) error {
	if e.TagName == "" {
		e.TagName = "required"
	}

	if e.TagValue == "" {
		e.TagValue = "true"
	}

	for _, field := range structs.Fields(s) {
		if err := e.processField("", field); err != nil {
			return err
		}
	}

	return nil
}

func (e *RequiredValidator) processField(fieldName string, field *structs.Field) error {
	fieldName += field.Name()
	switch field.Kind() {
	case reflect.Struct:
		// this is used for error messages below, when we have an error at the
		// child properties add parent properties into the error message as well
		fieldName += "."

		for _, f := range field.Fields() {
			if err := e.processField(fieldName, f); err != nil {
				return err
			}
		}
	default:
		val := field.Tag(e.TagName)
		if val != e.TagValue {
			return nil
		}

		if field.IsZero() {
			return fmt.Errorf("multiconfig: field '%s' is required", fieldName)
		}
	}

	return nil
}
