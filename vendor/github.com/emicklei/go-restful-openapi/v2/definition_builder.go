package restfulspec

import (
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

type PostBuildSwaggerSchema interface {
	PostBuildSwaggerSchemaHandler(sm *spec.Schema)
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
	if b.isSliceOrArrayType(st.Kind()) {
		st = st.Elem()
	}

	modelName := keyFrom(st, b.Config)
	if nameOverride != "" {
		modelName = nameOverride
	}
	// no models needed for primitive types unless it has alias
	if b.isPrimitiveType(modelName, st.Kind()) {
		if nameOverride == "" {
			return nil
		}
	}
	// golang encoding/json packages says array and slice values encode as
	// JSON arrays, except that []byte encodes as a base64-encoded string.
	// If we see a []byte here, treat it at as a primitive type (string)
	// and deal with it in buildArrayTypeProperty.
	if b.isByteArrayType(st) {
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

	if st.Kind() == reflect.Map {
		_, sm = b.buildMapType(st, "value", modelName)
		b.Definitions[modelName] = sm
		return &sm
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

	// Call handler to update sch
	if handler, ok := reflect.New(st).Elem().Interface().(PostBuildSwaggerSchema); ok {
		handler.PostBuildSwaggerSchemaHandler(&sm)
	}

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
	// marshalerType := reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	// if fieldType.Implements(marshalerType) {
	// 	var pType = "string"
	// 	if prop.Type == nil {
	// 		prop.Type = []string{pType}
	// 	}
	// 	if prop.Format == "" {
	// 		prop.Format = b.jsonSchemaFormat(keyFrom(fieldType, b.Config), fieldType.Kind())
	// 	}
	// 	return jsonName, modelDescription, prop
	// }

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

	// check for primitive first
	fieldTypeName := keyFrom(fieldType, b.Config)
	if b.isPrimitiveType(fieldTypeName, fieldKind) {
		mapped := b.jsonSchemaType(fieldTypeName, fieldKind)
		prop.Type = []string{mapped}
		prop.Format = b.jsonSchemaFormat(fieldTypeName, fieldKind)
		return jsonName, modelDescription, prop
	}

	// not a primitive
	// Since the `prop` variable in each of the cases below is a new reference, we need to re-set the `Example` field.
	switch {
	case fieldKind == reflect.Struct:
		jsonName, prop := b.buildStructTypeProperty(field, jsonName, model)
		setExample(&prop, field)

		return jsonName, modelDescription, prop
	case b.isSliceOrArrayType(fieldKind):
		jsonName, prop := b.buildArrayTypeProperty(field, jsonName, modelName)
		setExample(&prop, field)

		return jsonName, modelDescription, prop
	case fieldKind == reflect.Ptr:
		jsonName, prop := b.buildPointerTypeProperty(field, jsonName, modelName)
		setExample(&prop, field)

		return jsonName, modelDescription, prop
	case fieldKind == reflect.Map:
		jsonName, prop := b.buildMapTypeProperty(field, jsonName, modelName)
		setExample(&prop, field)

		return jsonName, modelDescription, prop
	}

	modelType := keyFrom(fieldType, b.Config)
	prop.Ref = spec.MustCreateRef("#/definitions/" + modelType)

	if fieldType.Name() == "" { // override type of anonymous structs
		// FIXME: Still need a way to handle anonymous struct model naming.
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
		// FIXME: Still need a way to handle anonymous struct model naming.
		anonType := model.ID + "." + jsonName
		b.addModel(fieldType, anonType)
		prop.Ref = spec.MustCreateRef("#/definitions/" + anonType)
		return jsonName, prop
	}

	if field.Name == fieldType.Name() && field.Anonymous && !hasNamedJSONTag(field) {
		definitions := make(map[string]spec.Schema, len(b.Definitions))
		for k, v := range b.Definitions {
			definitions[k] = v
		}
		// embedded struct
		sub := definitionBuilder{definitions, b.Config}
		sub.addModel(fieldType, "")
		subKey := keyFrom(fieldType, b.Config)
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
	var pType = keyFrom(fieldType, b.Config)
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
	prop.Items = &spec.SchemaOrArray{
		Schema: &spec.Schema{},
	}
	itemSchema := &prop
	itemType := fieldType
	isArray := b.isSliceOrArrayType(fieldType.Kind())
	for isArray {
		itemType = itemType.Elem()
		isArray = b.isSliceOrArrayType(itemType.Kind())
		if itemType.Kind() == reflect.Uint8 {
			stringt := "string"
			prop.Format = "binary"
			itemSchema.Type = []string{stringt}
			return jsonName, prop
		}
		itemSchema.Items = &spec.SchemaOrArray{
			Schema: &spec.Schema{},
		}
		itemSchema.Type = []string{"array"}
		itemSchema = itemSchema.Items.Schema
	}
	isPrimitive := b.isPrimitiveType(itemType.Name(), itemType.Kind())
	elemTypeName := b.getElementTypeName(modelName, jsonName, itemType)

	// If enum exists, move the enum definition from the `type: "array"` definition to `items`.
	if prop.Enum != nil {
		prop.Items.Schema.Enum = prop.Enum
		prop.Enum = nil
	}

	if isPrimitive {
		mapped := b.jsonSchemaType(elemTypeName, itemType.Kind())
		itemSchema.Type = []string{mapped}
		itemSchema.Format = b.jsonSchemaFormat(elemTypeName, itemType.Kind())
	} else {
		itemSchema.Ref = spec.MustCreateRef("#/definitions/" + elemTypeName)
	}
	// add|overwrite model for element type
	if itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}
	if !isPrimitive {
		b.addModel(itemType, elemTypeName)
	}
	return jsonName, prop
}

func (b definitionBuilder) buildMapTypeProperty(field reflect.StructField, jsonName, modelName string) (nameJson string, prop spec.Schema) {
	nameJson, prop = b.buildMapType(field.Type, jsonName, modelName)
	setPropertyMetadata(&prop, field)
	return nameJson, prop
}

func (b definitionBuilder) buildMapType(mapType reflect.Type, jsonName, modelName string) (nameJson string, prop spec.Schema) {
	var pType = "object"
	prop.Type = []string{pType}

	// As long as the element isn't an interface, we should be able to figure out what the
	// intended type is and represent it in `AdditionalProperties`.
	// See: https://swagger.io/docs/specification/data-models/dictionaries/
	if mapType.Elem().Kind().String() != "interface" {
		isSlice := b.isSliceOrArrayType(mapType.Elem().Kind())
		if isSlice && !b.isByteArrayType(mapType.Elem()) {
			mapType = mapType.Elem()
		}
		isPrimitive := b.isPrimitiveType(mapType.Elem().Name(), mapType.Elem().Kind())
		elemTypeName := b.getElementTypeName(modelName, jsonName, mapType.Elem())
		prop.AdditionalProperties = &spec.SchemaOrBool{
			Schema: &spec.Schema{},
		}
		// golang encoding/json packages says array and slice values encode as
		// JSON arrays, except that []byte encodes as a base64-encoded string.
		// If we see a []byte here, treat it at as a string
		if b.isByteArrayType(mapType.Elem()) {
			prop.AdditionalProperties.Schema.Type = []string{"string"}
		} else {
			if isSlice {
				var item *spec.Schema
				if isPrimitive {
					mapped := b.jsonSchemaType(elemTypeName, mapType.Kind())
					item = &spec.Schema{}
					item.Type = []string{mapped}
					item.Format = b.jsonSchemaFormat(elemTypeName, mapType.Kind())
				} else {
					item = spec.RefProperty("#/definitions/" + elemTypeName)
				}
				prop.AdditionalProperties.Schema = spec.ArrayProperty(item)
			} else if isPrimitive {
				mapped := b.jsonSchemaType(elemTypeName, mapType.Elem().Kind())
				prop.AdditionalProperties.Schema.Type = []string{mapped}
			} else {
				prop.AdditionalProperties.Schema.Ref = spec.MustCreateRef("#/definitions/" + elemTypeName)
			}
			// add|overwrite model for element type
			if mapType.Elem().Kind() == reflect.Ptr {
				mapType = mapType.Elem()
			}
			if !isPrimitive {
				b.addModel(mapType.Elem(), elemTypeName)
			}
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
		isPrimitive := b.isPrimitiveType(fieldType.Elem().Elem().Name(), fieldType.Elem().Elem().Kind())
		elemName := b.getElementTypeName(modelName, jsonName, fieldType.Elem().Elem())
		prop.Items = &spec.SchemaOrArray{
			Schema: &spec.Schema{},
		}
		if isPrimitive {
			primName := b.jsonSchemaType(elemName, fieldType.Elem().Elem().Kind())
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
		fieldTypeName := keyFrom(fieldType.Elem(), b.Config)
		isPrimitive := b.isPrimitiveType(fieldTypeName, fieldType.Elem().Kind())
		var pType = b.jsonSchemaType(fieldTypeName, fieldType.Elem().Kind()) // no star, include pkg path
		if isPrimitive {
			prop.Type = []string{pType}
			prop.Format = b.jsonSchemaFormat(fieldTypeName, fieldType.Elem().Kind())
			return jsonName, prop
		}
		prop.Ref = spec.MustCreateRef("#/definitions/" + pType)
		elemName := ""
		if fieldType.Elem().Name() == "" {
			elemName = modelName + "." + jsonName
			prop.Ref = spec.MustCreateRef("#/definitions/" + elemName)
		}
		if !isPrimitive {
			b.addModel(fieldType.Elem(), elemName)
		}
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
	return keyFrom(t, b.Config)
}

func keyFrom(st reflect.Type, cfg Config) string {
	key := st.String()
	if cfg.ModelTypeNameHandler != nil {
		if name, ok := cfg.ModelTypeNameHandler(st); ok {
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

func (b definitionBuilder) isSliceOrArrayType(t reflect.Kind) bool {
	return t == reflect.Slice || t == reflect.Array
}

// Does the type represent a []byte?
func (b definitionBuilder) isByteArrayType(t reflect.Type) bool {
	return (t.Kind() == reflect.Slice || t.Kind() == reflect.Array) &&
		t.Elem().Kind() == reflect.Uint8
}

// see also https://golang.org/ref/spec#Numeric_types
func (b definitionBuilder) isPrimitiveType(modelName string, modelKind reflect.Kind) bool {
	switch modelKind {
	case reflect.Bool:
		return true
	case reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.String:
		return true
	}

	if len(modelName) == 0 {
		return false
	}

	return strings.Contains("time.Time time.Duration json.Number", modelName)
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

	if b.Config.DefinitionNameHandler == nil {
		b.Config.DefinitionNameHandler = DefaultNameHandler
	}
	return b.Config.DefinitionNameHandler(field.Name)
}

// see also http://json-schema.org/latest/json-schema-core.html#anchor8
func (b definitionBuilder) jsonSchemaType(modelName string, modelKind reflect.Kind) string {
	schemaMap := map[string]string{
		"time.Time":     "string",
		"time.Duration": "integer",
		"json.Number":   "number",
	}

	if mapped, ok := schemaMap[modelName]; ok {
		return mapped
	}

	// check if original type is primitive
	switch modelKind {
	case reflect.Bool:
		return "boolean"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.String:
		return "string"
	}

	return modelName // use as is (custom or struct)
}

func (b definitionBuilder) jsonSchemaFormat(modelName string, modelKind reflect.Kind) string {
	if b.Config.SchemaFormatHandler != nil {
		if mapped := b.Config.SchemaFormatHandler(modelName); mapped != "" {
			return mapped
		}
	}

	schemaMap := map[string]string{
		"time.Time":      "date-time",
		"*time.Time":     "date-time",
		"time.Duration":  "int64",
		"*time.Duration": "int64",
		"json.Number":    "double",
		"*json.Number":   "double",
	}

	if mapped, ok := schemaMap[modelName]; ok {
		return mapped
	}

	// check if original type is primitive
	switch modelKind {
	case reflect.Float32:
		return "float"
	case reflect.Float64:
		return "double"
	case reflect.Int:
		return "int32"
	case reflect.Int8:
		return "byte"
	case reflect.Int16:
		return "integer"
	case reflect.Int32:
		return "int32"
	case reflect.Int64:
		return "int64"
	case reflect.Uint:
		return "integer"
	case reflect.Uint8:
		return "byte"
	case reflect.Uint16:
		return "integer"
	case reflect.Uint32:
		return "integer"
	case reflect.Uint64:
		return "integer"
	}

	return "" // no format
}
