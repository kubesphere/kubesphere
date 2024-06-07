// Copyright 2022 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"encoding/json"
	"errors"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/internal/gojsonschema"
)

// astValueToJSONSchemaLoader converts a value to JSON Loader.
// Value can be ast.String or ast.Object.
func astValueToJSONSchemaLoader(value ast.Value) (gojsonschema.JSONLoader, error) {
	var loader gojsonschema.JSONLoader
	var err error

	// ast.Value type selector.
	switch x := value.(type) {
	case ast.String:
		// In case of string pass it as is as a raw JSON string.
		// Make pre-check that it's a valid JSON at all because gojsonschema won't do that.
		if !json.Valid([]byte(x)) {
			return nil, errors.New("invalid JSON string")
		}
		loader = gojsonschema.NewStringLoader(string(x))
	case ast.Object:
		// In case of object serialize it to JSON representation.
		var data interface{}
		data, err = ast.JSON(value)
		if err != nil {
			return nil, err
		}
		loader = gojsonschema.NewGoLoader(data)
	default:
		// Any other cases will produce an error.
		return nil, errors.New("wrong type, expected string or object")
	}

	return loader, nil
}

func newResultTerm(valid bool, data *ast.Term) *ast.Term {
	return ast.ArrayTerm(ast.BooleanTerm(valid), data)
}

// builtinJSONSchemaVerify accepts 1 argument which can be string or object and checks if it is valid JSON schema.
// Returns array [false, <string>] with error string at index 1, or [true, ""] with empty string at index 1 otherwise.
func builtinJSONSchemaVerify(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	// Take first argument and make JSON Loader from it.
	loader, err := astValueToJSONSchemaLoader(operands[0].Value)
	if err != nil {
		return iter(newResultTerm(false, ast.StringTerm("jsonschema: "+err.Error())))
	}

	// Check that schema is correct and parses without errors.
	if _, err = gojsonschema.NewSchema(loader); err != nil {
		return iter(newResultTerm(false, ast.StringTerm("jsonschema: "+err.Error())))
	}

	return iter(newResultTerm(true, ast.NullTerm()))
}

// builtinJSONMatchSchema accepts 2 arguments both can be string or object and verifies if the document matches the JSON schema.
// Returns an array where first element is a boolean indicating a successful match, and the second is an array of errors that is empty on success and populated on failure.
// In case of internal error returns empty array.
func builtinJSONMatchSchema(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	// Take first argument and make JSON Loader from it.
	// This is a JSON document made from Rego JSON string or object.
	documentLoader, err := astValueToJSONSchemaLoader(operands[0].Value)
	if err != nil {
		return err
	}

	// Take second argument and make JSON Loader from it.
	// This is a JSON schema made from Rego JSON string or object.
	schemaLoader, err := astValueToJSONSchemaLoader(operands[1].Value)
	if err != nil {
		return err
	}

	// Use schema to validate document.
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	// In case of validation errors produce Rego array of objects to describe the errors.
	arr := ast.NewArray()
	for _, re := range result.Errors() {
		o := ast.NewObject(
			[...]*ast.Term{ast.StringTerm("error"), ast.StringTerm(re.String())},
			[...]*ast.Term{ast.StringTerm("type"), ast.StringTerm(re.Type())},
			[...]*ast.Term{ast.StringTerm("field"), ast.StringTerm(re.Field())},
			[...]*ast.Term{ast.StringTerm("desc"), ast.StringTerm(re.Description())},
		)
		arr = arr.Append(ast.NewTerm(o))
	}

	return iter(newResultTerm(result.Valid(), ast.NewTerm(arr)))
}

func init() {
	RegisterBuiltinFunc(ast.JSONSchemaVerify.Name, builtinJSONSchemaVerify)
	RegisterBuiltinFunc(ast.JSONMatchSchema.Name, builtinJSONMatchSchema)
}
