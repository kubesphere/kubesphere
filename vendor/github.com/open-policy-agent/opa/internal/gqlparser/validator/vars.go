package validator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/open-policy-agent/opa/internal/gqlparser/ast"
	"github.com/open-policy-agent/opa/internal/gqlparser/gqlerror"
)

var ErrUnexpectedType = fmt.Errorf("Unexpected Type")

// VariableValues coerces and validates variable values
func VariableValues(schema *ast.Schema, op *ast.OperationDefinition, variables map[string]interface{}) (map[string]interface{}, error) {
	coercedVars := map[string]interface{}{}

	validator := varValidator{
		path:   ast.Path{ast.PathName("variable")},
		schema: schema,
	}

	for _, v := range op.VariableDefinitions {
		validator.path = append(validator.path, ast.PathName(v.Variable))

		if !v.Definition.IsInputType() {
			return nil, gqlerror.ErrorPathf(validator.path, "must an input type")
		}

		val, hasValue := variables[v.Variable]

		if !hasValue {
			if v.DefaultValue != nil {
				var err error
				val, err = v.DefaultValue.Value(nil)
				if err != nil {
					return nil, gqlerror.WrapPath(validator.path, err)
				}
				hasValue = true
			} else if v.Type.NonNull {
				return nil, gqlerror.ErrorPathf(validator.path, "must be defined")
			}
		}

		if hasValue {
			if val == nil {
				if v.Type.NonNull {
					return nil, gqlerror.ErrorPathf(validator.path, "cannot be null")
				}
				coercedVars[v.Variable] = nil
			} else {
				rv := reflect.ValueOf(val)

				jsonNumber, isJSONNumber := val.(json.Number)
				if isJSONNumber {
					if v.Type.NamedType == "Int" {
						n, err := jsonNumber.Int64()
						if err != nil {
							return nil, gqlerror.ErrorPathf(validator.path, "cannot use value %d as %s", n, v.Type.NamedType)
						}
						rv = reflect.ValueOf(n)
					} else if v.Type.NamedType == "Float" {
						f, err := jsonNumber.Float64()
						if err != nil {
							return nil, gqlerror.ErrorPathf(validator.path, "cannot use value %f as %s", f, v.Type.NamedType)
						}
						rv = reflect.ValueOf(f)

					}
				}
				if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
					rv = rv.Elem()
				}

				rval, err := validator.validateVarType(v.Type, rv)
				if err != nil {
					return nil, err
				}
				coercedVars[v.Variable] = rval.Interface()
			}
		}

		validator.path = validator.path[0 : len(validator.path)-1]
	}
	return coercedVars, nil
}

type varValidator struct {
	path   ast.Path
	schema *ast.Schema
}

func (v *varValidator) validateVarType(typ *ast.Type, val reflect.Value) (reflect.Value, *gqlerror.Error) {
	currentPath := v.path
	resetPath := func() {
		v.path = currentPath
	}
	defer resetPath()
	if typ.Elem != nil {
		if val.Kind() != reflect.Slice {
			// GraphQL spec says that non-null values should be coerced to an array when possible.
			// Hence if the value is not a slice, we create a slice and add val to it.
			slc := reflect.MakeSlice(reflect.SliceOf(val.Type()), 0, 0)
			slc = reflect.Append(slc, val)
			val = slc
		}
		for i := 0; i < val.Len(); i++ {
			resetPath()
			v.path = append(v.path, ast.PathIndex(i))
			field := val.Index(i)
			if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
				if typ.Elem.NonNull && field.IsNil() {
					return val, gqlerror.ErrorPathf(v.path, "cannot be null")
				}
				field = field.Elem()
			}
			_, err := v.validateVarType(typ.Elem, field)
			if err != nil {
				return val, err
			}
		}
		return val, nil
	}
	def := v.schema.Types[typ.NamedType]
	if def == nil {
		panic(fmt.Errorf("missing def for %s", typ.NamedType))
	}

	if !typ.NonNull && !val.IsValid() {
		// If the type is not null and we got a invalid value namely null/nil, then it's valid
		return val, nil
	}

	switch def.Kind {
	case ast.Enum:
		kind := val.Type().Kind()
		if kind != reflect.Int && kind != reflect.Int32 && kind != reflect.Int64 && kind != reflect.String {
			return val, gqlerror.ErrorPathf(v.path, "enums must be ints or strings")
		}
		isValidEnum := false
		for _, enumVal := range def.EnumValues {
			if strings.EqualFold(val.String(), enumVal.Name) {
				isValidEnum = true
			}
		}
		if !isValidEnum {
			return val, gqlerror.ErrorPathf(v.path, "%s is not a valid %s", val.String(), def.Name)
		}
		return val, nil
	case ast.Scalar:
		kind := val.Type().Kind()
		switch typ.NamedType {
		case "Int":
			if kind == reflect.Int || kind == reflect.Int32 || kind == reflect.Int64 || kind == reflect.Float32 || kind == reflect.Float64 || IsValidIntString(val, kind) {
				return val, nil
			}
		case "Float":
			if kind == reflect.Float32 || kind == reflect.Float64 || kind == reflect.Int || kind == reflect.Int32 || kind == reflect.Int64 || IsValidFloatString(val, kind) {
				return val, nil
			}
		case "String":
			if kind == reflect.String {
				return val, nil
			}

		case "Boolean":
			if kind == reflect.Bool {
				return val, nil
			}

		case "ID":
			if kind == reflect.Int || kind == reflect.Int32 || kind == reflect.Int64 || kind == reflect.String {
				return val, nil
			}
		default:
			// assume custom scalars are ok
			return val, nil
		}
		return val, gqlerror.ErrorPathf(v.path, "cannot use %s as %s", kind.String(), typ.NamedType)
	case ast.InputObject:
		if val.Kind() != reflect.Map {
			return val, gqlerror.ErrorPathf(v.path, "must be a %s", def.Name)
		}

		// check for unknown fields
		for _, name := range val.MapKeys() {
			val.MapIndex(name)
			fieldDef := def.Fields.ForName(name.String())
			resetPath()
			v.path = append(v.path, ast.PathName(name.String()))

			switch {
			case name.String() == "__typename":
				continue
			case fieldDef == nil:
				return val, gqlerror.ErrorPathf(v.path, "unknown field")
			}
		}

		for _, fieldDef := range def.Fields {
			resetPath()
			v.path = append(v.path, ast.PathName(fieldDef.Name))

			field := val.MapIndex(reflect.ValueOf(fieldDef.Name))
			if !field.IsValid() {
				if fieldDef.Type.NonNull {
					if fieldDef.DefaultValue != nil {
						var err error
						_, err = fieldDef.DefaultValue.Value(nil)
						if err == nil {
							continue
						}
					}
					return val, gqlerror.ErrorPathf(v.path, "must be defined")
				}
				continue
			}

			if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
				if fieldDef.Type.NonNull && field.IsNil() {
					return val, gqlerror.ErrorPathf(v.path, "cannot be null")
				}
				//allow null object field and skip it
				if !fieldDef.Type.NonNull && field.IsNil() {
					continue
				}
				field = field.Elem()
			}
			cval, err := v.validateVarType(fieldDef.Type, field)
			if err != nil {
				return val, err
			}
			val.SetMapIndex(reflect.ValueOf(fieldDef.Name), cval)
		}
	default:
		panic(fmt.Errorf("unsupported type %s", def.Kind))
	}
	return val, nil
}

func IsValidIntString(val reflect.Value, kind reflect.Kind) bool {
	if kind != reflect.String {
		return false
	}
	_, e := strconv.ParseInt(fmt.Sprintf("%v", val.Interface()), 10, 64)

	return e == nil
}

func IsValidFloatString(val reflect.Value, kind reflect.Kind) bool {
	if kind != reflect.String {
		return false
	}
	_, e := strconv.ParseFloat(fmt.Sprintf("%v", val.Interface()), 64)
	return e == nil
}
