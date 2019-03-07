package restfulspec

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/go-openapi/spec"
)

type definitionBuilder struct {
	Definitions spec.Definitions
	Config      Config
}

// Documented is
type Documented interface {
	SwaggerDoc() map[string]string
}

// Check if this structure has a method with signature func (<theModel>) SwaggerDoc() map[string]string
// If it exists, retrieve the documentation and overwrite all struct tag descriptions
func getDocFromMethodSwaggerDoc2(model reflect.Type) map[string]string {
	if docable, ok := reflect.New(model).Elem().Interface().(Documented); ok {
		return docable.SwaggerDoc()
	}
	return make(map[string]string)
}

// addModelFrom creates and adds a Schema to the builder and detects and calls
// the post build hook for customizations
func (b definitionBuilder) addModelFrom(sample interface{}) {
	b.addModel(reflect.TypeOf(sample), "")
}

func (b definitionBuilder) addModel(st reflect.Type, nameOverride string) *spec.Schema {
	// Turn pointers into simpler types so further checks are
	// correct.
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
	}

	modelName := b.keyFrom(st)
	if nameOverride != "" {
		modelName = nameOverride
	}
	// no models needed for primitive types
	if b.isPrimitiveType(modelName) {
		return nil
	}
	// golang encoding/json packages says array and slice values encode as
	// JSON arrays, except that []byte encodes as a base64-encoded string.
	// If we see a []byte here, treat it at as a primitive type (string)
	// and deal with it in buildArrayTypeProperty.
	if (st.Kind() == reflect.Slice || st.Kind() == reflect.Array) &&
		st.Elem().Kind() == reflect.Uint8 {
		return nil
	}
	// see if we already have visited this model
	if _, ok := b.Definitions[modelName]; ok {
		return nil
	}
	sm := spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required:   []string{},
			Properties: map[string]spec.Schema{},
		},
	}

	// reference the model before further initializing (enables recursive structs)
	b.Definitions[modelName] = sm

	// check for slice or array
	if st.Kind() == reflect.Slice || st.Kind() == reflect.Array {
		st = st.Elem()
	}
	// check for structure or primitive type
	if st.Kind() != reflect.Struct {
		return &sm
	}

	fullDoc := getDocFromMethodSwaggerDoc2(st)
	modelDescriptions := []string{}

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		jsonName, modelDescription, prop := b.buildProperty(field, &sm, modelName)
		if len(modelDescription) > 0 {
			modelDescriptions = append(modelDescriptions, modelDescription)
		}

		// add if not omitted
		if len(jsonName) != 0 {
			// update description
			if fieldDoc, ok := fullDoc[jsonName]; ok {
				prop.Description = fieldDoc
			}
			// update Required
			if b.isPropertyRequired(field) {
				sm.Required = append(sm.Required, jsonName)
			}
			sm.Properties[jsonName] = prop
		}
	}

	// We always overwrite documentation if SwaggerDoc method exists
	// "" is special for documenting the struct itself
	if modelDoc, ok := fullDoc[""]; ok {
		sm.Description = modelDoc
	} else if len(modelDescriptions) != 0 {
		sm.Description = strings.Join(modelDescriptions, "\n")
	}
	// Needed to pass openapi validation. This field exists for json-schema compatibility,
	// but it conflicts with the openapi specification.
	// See https://github.com/go-openapi/spec/issues/23 for more context
	sm.ID = ""

	// update model builder with completed model
	b.Definitions[modelName] = sm

	return &sm
}

func (b definitionBuilder) isPropertyRequired(field reflect.StructField) bool {
	required := true
	if optionalTag := field.Tag.Get("optional"); optionalTag == "true" {
		return false
	}
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		s := strings.Split(jsonTag, ",")
		if len(s) > 1 && s[1] == "omitempty" {
			return false
		}
	}
	return required
}

func (b definitionBuilder) buildProperty(field reflect.StructField, model *spec.Schema, modelName string) (jsonName, modelDescription string, prop spec.Schema) {
	jsonName = b.jsonNameOfField(field)
	if len(jsonName) == 0 {
		// empty name signals skip property
		return "", "", prop
	}

	if field.Name == "XMLName" && field.Type.String() == "xml.Name" {
		// property is metadata for the xml.Name attribute, can be skipped
		return "", "", prop
	}

	if tag := field.Tag.Get("modelDescription"); tag != "" {
		modelDescription = tag
	}

	setPropertyMetadata(&prop, field)
	if prop.Type != nil {
		return jsonName, modelDescription, prop
	}
	fieldType := field.Type

	// check if type is doing its own marshalling
	marshalerType := reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	if fieldType.Implements(marshalerType) {
		var pType = "string"
		if prop.Type == nil {
			prop.Type = []string{pType}
		}
		if prop.Format == "" {
			prop.Format = b.jsonSchemaFormat(b.keyFrom(fieldType))
		}
		return jsonName, modelDescription, prop
	}

	// check if annotation says it is a string
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		s := strings.Split(jsonTag, ",")
		if len(s) > 1 && s[1] == "string" {
			stringt := "string"
			prop.Type = []string{stringt}
			return jsonName, modelDescription, prop
		}
	}

	fieldKind := fieldType.Kind()
	switch {
	case fieldKind == reflect.Struct:
		jsonName, prop := b.buildStructTypeProperty(field, jsonName, model)
		return jsonName, modelDescription, prop
	case fieldKind == reflect.Slice || fieldKind == reflect.Array:
		jsonName, prop := b.buildArrayTypeProperty(field, jsonName, modelName)
		return jsonName, modelDescription, prop
	case fieldKind == reflect.Ptr:
		jsonName, prop := b.buildPointerTypeProperty(field, jsonName, modelName)
		return jsonName, modelDescription, prop
	case fieldKind == reflect.String:
		stringt := "string"
		prop.Type = []string{stringt}
		return jsonName, modelDescription, prop
	case fieldKind == reflect.Map:
		jsonName, prop := b.buildMapTypeProperty(field, jsonName, modelName)
		return jsonName, modelDescription, prop
	}

	fieldTypeName := b.keyFrom(fieldType)
	if b.isPrimitiveType(fieldTypeName) {
		mapped := b.jsonSchemaType(fieldTypeName)
		prop.Type = []string{mapped}
		prop.Format = b.jsonSchemaFormat(fieldTypeName)
		return jsonName, modelDescription, prop
	}
	modelType := b.keyFrom(fieldType)
	prop.Ref = spec.MustCreateRef("#/definitions/" + modelType)

	if fieldType.Name() == "" { // override type of anonymous structs
		nestedTypeName := modelName + "." + jsonName
		prop.Ref = spec.MustCreateRef("#/definitions/" + nestedTypeName)
		b.addModel(fieldType, nestedTypeName)
	}
	return jsonName, modelDescription, prop
}

func hasNamedJSONTag(field reflect.StructField) bool {
	parts := strings.Split(field.Tag.Get("json"), ",")
	if len(parts) == 0 {
		return false
	}
	for _, s := range parts[1:] {
		if s == "inline" {
			return false
		}
	}
	return len(parts[0]) > 0
}

func (b definitionBuilder) buildStructTypeProperty(field reflect.StructField, jsonName string, model *spec.Schema) (nameJson string, prop spec.Schema) {
	setPropertyMetadata(&prop, field)
	fieldType := field.Type
	// check for anonymous
	if len(fieldType.Name()) == 0 {
		// anonymous
		anonType := model.ID + "." + jsonName
		b.addModel(fieldType, anonType)
		prop.Ref = spec.MustCreateRef("#/definitions/" + anonType)
		return jsonName, prop
	}

	if field.Name == fieldType.Name() && field.Anonymous && !hasNamedJSONTag(field) {
		// embedded struct
		sub := definitionBuilder{make(spec.Definitions), b.Config}
		sub.addModel(fieldType, "")
		subKey := sub.keyFrom(fieldType)
		// merge properties from sub
		subModel, _ := sub.Definitions[subKey]
		for k, v := range subModel.Properties {
			model.Properties[k] = v
			// if subModel says this property is required then include it
			required := false
			for _, each := range subModel.Required {
				if k == each {
					required = true
					break
				}
			}
			if required {
				model.Required = append(model.Required, k)
			}
		}
		// add all new referenced models
		for key, sub := range sub.Definitions {
			if key != subKey {
				if _, ok := b.Definitions[key]; !ok {
					b.Definitions[key] = sub
				}
			}
		}
		// empty name signals skip property
		return "", prop
	}
	// simple struct
	b.addModel(fieldType, "")
	var pType = b.keyFrom(fieldType)
	prop.Ref = spec.MustCreateRef("#/definitions/" + pType)
	return jsonName, prop
}

func (b definitionBuilder) buildArrayTypeProperty(field reflect.StructField, jsonName, modelName string) (nameJson string, prop spec.Schema) {
	setPropertyMetadata(&prop, field)
	fieldType := field.Type
	if fieldType.Elem().Kind() == reflect.Uint8 {
		stringt := "string"
		prop.Type = []string{stringt}
		return jsonName, prop
	}
	var pType = "array"
	prop.Type = []string{pType}
	isPrimitive := b.isPrimitiveType(fieldType.Elem().Name())
	elemTypeName := b.getElementTypeName(modelName, jsonName, fieldType.Elem())
	prop.Items = &spec.SchemaOrArray{
		Schema: &spec.Schema{},
	}
	if isPrimitive {
		mapped := b.jsonSchemaType(elemTypeName)
		prop.Items.Schema.Type = []string{mapped}
	} else {
		prop.Items.Schema.Ref = spec.MustCreateRef("#/definitions/" + elemTypeName)
	}
	// add|overwrite model for element type
	if fieldType.Elem().Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}
	if !isPrimitive {
		b.addModel(fieldType.Elem(), elemTypeName)
	}
	return jsonName, prop
}

func (b definitionBuilder) buildMapTypeProperty(field reflect.StructField, jsonName, modelName string) (nameJson string, prop spec.Schema) {
	setPropertyMetadata(&prop, field)
	fieldType := field.Type
	var pType = "object"
	prop.Type = []string{pType}

	// As long as the element isn't an interface, we should be able to figure out what the
	// intended type is and represent it in `AdditionalProperties`.
	// See: https://swagger.io/docs/specification/data-models/dictionaries/
	if fieldType.Elem().Kind().String() != "interface" {
		isPrimitive := b.isPrimitiveType(fieldType.Elem().Name())
		elemTypeName := b.getElementTypeName(modelName, jsonName, fieldType.Elem())
		prop.AdditionalProperties = &spec.SchemaOrBool{
			Schema: &spec.Schema{},
		}
		if isPrimitive {
			mapped := b.jsonSchemaType(elemTypeName)
			prop.AdditionalProperties.Schema.Type = []string{mapped}
		} else {
			prop.AdditionalProperties.Schema.Ref = spec.MustCreateRef("#/definitions/" + elemTypeName)
		}
		// add|overwrite model for element type
		if fieldType.Elem().Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if !isPrimitive {
			b.addModel(fieldType.Elem(), elemTypeName)
		}
	}
	return jsonName, prop
}

func (b definitionBuilder) buildPointerTypeProperty(field reflect.StructField, jsonName, modelName string) (nameJson string, prop spec.Schema) {
	setPropertyMetadata(&prop, field)
	fieldType := field.Type

	// override type of pointer to list-likes
	if fieldType.Elem().Kind() == reflect.Slice || fieldType.Elem().Kind() == reflect.Array {
		var pType = "array"
		prop.Type = []string{pType}
		isPrimitive := b.isPrimitiveType(fieldType.Elem().Elem().Name())
		elemName := b.getElementTypeName(modelName, jsonName, fieldType.Elem().Elem())
		prop.Items = &spec.SchemaOrArray{
			Schema: &spec.Schema{},
		}
		if isPrimitive {
			primName := b.jsonSchemaType(elemName)
			prop.Items.Schema.Type = []string{primName}
		} else {
			prop.Items.Schema.Ref = spec.MustCreateRef("#/definitions/" + elemName)
		}
		if !isPrimitive {
			// add|overwrite model for element type
			b.addModel(fieldType.Elem().Elem(), elemName)
		}
	} else {
		// non-array, pointer type
		fieldTypeName := b.keyFrom(fieldType.Elem())
		var pType = b.jsonSchemaType(fieldTypeName) // no star, include pkg path
		if b.isPrimitiveType(fieldTypeName) {
			prop.Type = []string{pType}
			prop.Format = b.jsonSchemaFormat(fieldTypeName)
			return jsonName, prop
		}
		prop.Ref = spec.MustCreateRef("#/definitions/" + pType)
		elemName := ""
		if fieldType.Elem().Name() == "" {
			elemName = modelName + "." + jsonName
			prop.Ref = spec.MustCreateRef("#/definitions/" + elemName)
		}
		b.addModel(fieldType.Elem(), elemName)
	}
	return jsonName, prop
}

func (b definitionBuilder) getElementTypeName(modelName, jsonName string, t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Name() == "" {
		return modelName + "." + jsonName
	}
	return b.keyFrom(t)
}

func (b definitionBuilder) keyFrom(st reflect.Type) string {
	key := st.String()
	if b.Config.ModelTypeNameHandler != nil {
		if name, ok := b.Config.ModelTypeNameHandler(st); ok {
			key = name
		}
	}
	if len(st.Name()) == 0 { // unnamed type
		// If it is an array, remove the leading []
		key = strings.TrimPrefix(key, "[]")
		// Swagger UI has special meaning for [
		key = strings.Replace(key, "[]", "||", -1)
	}
	return key
}

// see also https://golang.org/ref/spec#Numeric_types
func (b definitionBuilder) isPrimitiveType(modelName string) bool {
	if len(modelName) == 0 {
		return false
	}
	return strings.Contains("uint uint8 uint16 uint32 uint64 int int8 int16 int32 int64 float32 float64 bool string byte rune time.Time", modelName)
}

// jsonNameOfField returns the name of the field as it should appear in JSON format
// An empty string indicates that this field is not part of the JSON representation
func (b definitionBuilder) jsonNameOfField(field reflect.StructField) string {
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		s := strings.Split(jsonTag, ",")
		if s[0] == "-" {
			// empty name signals skip property
			return ""
		} else if s[0] != "" {
			return s[0]
		}
	}
	return field.Name
}

// see also http://json-schema.org/latest/json-schema-core.html#anchor8
func (b definitionBuilder) jsonSchemaType(modelName string) string {
	schemaMap := map[string]string{
		"uint":   "integer",
		"uint8":  "integer",
		"uint16": "integer",
		"uint32": "integer",
		"uint64": "integer",

		"int":   "integer",
		"int8":  "integer",
		"int16": "integer",
		"int32": "integer",
		"int64": "integer",

		"byte":      "integer",
		"float64":   "number",
		"float32":   "number",
		"bool":      "boolean",
		"time.Time": "string",
	}
	mapped, ok := schemaMap[modelName]
	if !ok {
		return modelName // use as is (custom or struct)
	}
	return mapped
}

func (b definitionBuilder) jsonSchemaFormat(modelName string) string {
	if b.Config.SchemaFormatHandler != nil {
		if mapped := b.Config.SchemaFormatHandler(modelName); mapped != "" {
			return mapped
		}
	}
	schemaMap := map[string]string{
		"int":        "int32",
		"int32":      "int32",
		"int64":      "int64",
		"byte":       "byte",
		"uint":       "integer",
		"uint8":      "byte",
		"float64":    "double",
		"float32":    "float",
		"time.Time":  "date-time",
		"*time.Time": "date-time",
	}
	mapped, ok := schemaMap[modelName]
	if !ok {
		return "" // no format
	}
	return mapped
}
