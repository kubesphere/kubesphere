package restfulspec

import (
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-openapi/spec"

	"github.com/emicklei/go-restful/v3"
)

const (
	// KeyOpenAPITags is a Metadata key for a restful Route
	KeyOpenAPITags = "openapi.tags"

	// ExtensionPrefix is the only prefix accepted for VendorExtensible extension keys
	ExtensionPrefix = "x-"

	arrayType      = "array"
	definitionRoot = "#/definitions/"
)

func buildPaths(ws *restful.WebService, cfg Config) spec.Paths {
	p := spec.Paths{Paths: map[string]spec.PathItem{}}
	for _, each := range ws.Routes() {
		path, patterns := sanitizePath(each.Path)
		existingPathItem, ok := p.Paths[path]
		if !ok {
			existingPathItem = spec.PathItem{}
		}
		p.Paths[path] = buildPathItem(ws, each, existingPathItem, patterns, cfg)
	}
	return p
}

// sanitizePath removes regex expressions from named path params,
// since openapi only supports setting the pattern as a property named "pattern".
// Expressions like "/api/v1/{name:[a-z]}/" are converted to "/api/v1/{name}/".
// The second return value is a map which contains the mapping from the path parameter
// name to the extracted pattern
func sanitizePath(restfulPath string) (string, map[string]string) {
	openapiPath := ""
	patterns := map[string]string{}
	for _, fragment := range strings.Split(restfulPath, "/") {
		if fragment == "" {
			continue
		}
		if strings.HasPrefix(fragment, "{") && strings.Contains(fragment, ":") {
			split := strings.Split(fragment, ":")
			// skip google custom method like `resource/{resource-id}:customVerb`
			if !strings.Contains(split[0], "}") {
				fragment = split[0][1:]
				pattern := split[1][:len(split[1])-1]
				if pattern == "*" { // special case
					pattern = ".*"
				}
				patterns[fragment] = pattern
				fragment = "{" + fragment + "}"
			}
		}
		openapiPath += "/" + fragment
	}
	return openapiPath, patterns
}

func buildPathItem(ws *restful.WebService, r restful.Route, existingPathItem spec.PathItem, patterns map[string]string, cfg Config) spec.PathItem {
	op := buildOperation(ws, r, patterns, cfg)
	switch r.Method {
	case http.MethodGet:
		existingPathItem.Get = op
	case http.MethodPost:
		existingPathItem.Post = op
	case http.MethodPut:
		existingPathItem.Put = op
	case http.MethodDelete:
		existingPathItem.Delete = op
	case http.MethodPatch:
		existingPathItem.Patch = op
	case http.MethodOptions:
		existingPathItem.Options = op
	case http.MethodHead:
		existingPathItem.Head = op
	}
	return existingPathItem
}

func buildOperation(ws *restful.WebService, r restful.Route, patterns map[string]string, cfg Config) *spec.Operation {
	o := spec.NewOperation(r.Operation)
	o.Description = r.Notes
	o.Summary = stripTags(r.Doc)
	o.Consumes = r.Consumes
	o.Produces = r.Produces
	o.Deprecated = r.Deprecated
	if r.Metadata != nil {
		if tags, ok := r.Metadata[KeyOpenAPITags]; ok {
			if tagList, ok := tags.([]string); ok {
				o.Tags = tagList
			}
		}
	}

	extractVendorExtensions(&o.VendorExtensible, r.ExtensionProperties)

	// collect any path parameters
	for _, param := range ws.PathParameters() {
		p := buildParameter(r, param, patterns[param.Data().Name], cfg)
		o.Parameters = append(o.Parameters, p)
	}
	// route specific params
	for _, param := range r.ParameterDocs {
		p := buildParameter(r, param, patterns[param.Data().Name], cfg)
		o.Parameters = append(o.Parameters, p)
	}
	o.Responses = new(spec.Responses)
	props := &o.Responses.ResponsesProps
	props.StatusCodeResponses = make(map[int]spec.Response, len(r.ResponseErrors))
	for k, v := range r.ResponseErrors {
		r := buildResponse(v, cfg)
		props.StatusCodeResponses[k] = r
	}
	if r.DefaultResponse != nil {
		rsp := buildResponse(*r.DefaultResponse, cfg)
		o.Responses.Default = &rsp
	}
	if len(o.Responses.StatusCodeResponses) == 0 {
		o.Responses.StatusCodeResponses[200] = spec.Response{ResponseProps: spec.ResponseProps{Description: http.StatusText(http.StatusOK)}}
	}
	return o
}

// stringAutoType picks the correct type when dataType is set. Otherwise, it automatically picks the correct type from
// an ambiguously typed string. Ex. numbers become int, true/false become bool, etc.
func stringAutoType(dataType, ambiguous string) interface{} {
	if ambiguous == "" {
		return nil
	}

	if dataType == "" || dataType == "integer" {
		if parsedInt, err := strconv.ParseInt(ambiguous, 10, 64); err == nil {
			return parsedInt
		}
	}

	if dataType == "" || dataType == "boolean" {
		if parsedBool, err := strconv.ParseBool(ambiguous); err == nil {
			return parsedBool
		}
	}

	return ambiguous
}

func extractVendorExtensions(extensible *spec.VendorExtensible, extensions restful.ExtensionProperties) {
	if len(extensions.Extensions) > 0 {
		for key := range extensions.Extensions {
			if strings.HasPrefix(key, ExtensionPrefix) {
				extensible.AddExtension(key, extensions.Extensions[key])
			}
		}
	}
}

func buildParameter(r restful.Route, restfulParam *restful.Parameter, pattern string, cfg Config) spec.Parameter {
	p := spec.Parameter{}
	param := restfulParam.Data()
	p.In = asParamType(param.Kind)

	if param.AllowMultiple {
		// If the param is an array apply the validations to the items in it
		p.Type = arrayType
		p.Items = spec.NewItems()
		p.Items.Type = param.DataType
		p.Items.Pattern = param.Pattern
		p.Items.MinLength = param.MinLength
		p.Items.MaxLength = param.MaxLength
		p.CollectionFormat = param.CollectionFormat
		p.MinItems = param.MinItems
		p.MaxItems = param.MaxItems
		p.UniqueItems = param.UniqueItems
	} else {
		// Otherwise, for non-arrays, apply the validations directly to the param
		p.Type = param.DataType
		p.MinLength = param.MinLength
		p.MaxLength = param.MaxLength
		p.Minimum = param.Minimum
		p.Maximum = param.Maximum
	}

	// Prefer PossibleValues over deprecated AllowableValues
	if numPossible := len(param.PossibleValues); numPossible > 0 {
		// init Enum to our known size and populate it
		p.Enum = make([]interface{}, 0, numPossible)
		for _, value := range param.PossibleValues {
			p.Enum = append(p.Enum, stringAutoType(p.Type, value))
		}
	} else {
		if numAllowable := len(param.AllowableValues); numAllowable > 0 {
			// If allowable values are defined, set the enum array to the sorted values
			allowableSortedKeys := make([]string, 0, numAllowable)
			for k := range param.AllowableValues {
				allowableSortedKeys = append(allowableSortedKeys, k)
			}

			// sort away
			sort.Strings(allowableSortedKeys)

			// init Enum to our known size and populate it
			p.Enum = make([]interface{}, 0, numAllowable)
			for _, key := range allowableSortedKeys {
				p.Enum = append(p.Enum, stringAutoType(p.Type, key))
			}
		}
	}

	p.Description = param.Description
	p.Name = param.Name
	p.Required = param.Required
	p.AllowEmptyValue = param.AllowEmptyValue

	if param.Kind == restful.PathParameterKind {
		p.Pattern = pattern
	} else if !param.AllowMultiple {
		p.Pattern = param.Pattern
	}
	st := reflect.TypeOf(r.ReadSample)
	if param.Kind == restful.BodyParameterKind && r.ReadSample != nil && param.DataType == st.String() {
		p.Schema = new(spec.Schema)
		p.SimpleSchema = spec.SimpleSchema{}
		if st.Kind() == reflect.Array || st.Kind() == reflect.Slice {
			dataTypeName := keyFrom(st.Elem(), cfg)
			p.Schema.Type = []string{arrayType}
			p.Schema.Items = &spec.SchemaOrArray{
				Schema: &spec.Schema{},
			}
			isPrimitive := isPrimitiveType(dataTypeName)
			if isPrimitive {
				mapped := jsonSchemaType(dataTypeName)
				p.Schema.Items.Schema.Type = []string{mapped}
			} else {
				p.Schema.Items.Schema.Ref = spec.MustCreateRef(definitionRoot + dataTypeName)
			}
		} else {
			dataTypeName := keyFrom(st, cfg)
			p.Schema.Ref = spec.MustCreateRef(definitionRoot + dataTypeName)
		}

	} else {
		if param.AllowMultiple {
			p.Type = arrayType
			p.Items = spec.NewItems()
			p.Items.Type = param.DataType
			p.CollectionFormat = param.CollectionFormat
		} else {
			p.Type = param.DataType
		}
		p.Default = stringAutoType(param.DataType, param.DefaultValue)
		p.Format = param.DataFormat
	}

	extractVendorExtensions(&p.VendorExtensible, param.ExtensionProperties)

	return p
}

func buildResponse(e restful.ResponseError, cfg Config) (r spec.Response) {
	r.Description = e.Message
	if e.Model != nil {
		st := reflect.TypeOf(e.Model)
		if st.Kind() == reflect.Ptr {
			// For pointer type, use element type as the key; otherwise we'll
			// endup with '#/definitions/*Type' which violates openapi spec.
			st = st.Elem()
		}
		r.Schema = new(spec.Schema)
		if st.Kind() == reflect.Array || st.Kind() == reflect.Slice {
			modelName := keyFrom(st.Elem(), cfg)
			r.Schema.Type = []string{arrayType}
			r.Schema.Items = &spec.SchemaOrArray{
				Schema: &spec.Schema{},
			}
			isPrimitive := isPrimitiveType(modelName)
			if isPrimitive {
				mapped := jsonSchemaType(modelName)
				r.Schema.Items.Schema.Type = []string{mapped}
			} else {
				r.Schema.Items.Schema.Ref = spec.MustCreateRef(definitionRoot + modelName)
			}
		} else {
			modelName := keyFrom(st, cfg)
			if isPrimitiveType(modelName) {
				// If the response is a primitive type, then don't reference any definitions.
				// Instead, set the schema's "type" to the model name.
				r.Schema.AddType(modelName, "")
			} else {
				modelName := keyFrom(st, cfg)
				r.Schema.Ref = spec.MustCreateRef(definitionRoot + modelName)
			}
		}
	}

	if len(e.Headers) > 0 {
		r.Headers = make(map[string]spec.Header, len(e.Headers))
		for k, v := range e.Headers {
			r.Headers[k] = buildHeader(v)
		}
	}

	extractVendorExtensions(&r.VendorExtensible, e.ExtensionProperties)
	return r
}

// buildHeader builds a specification header structure from restful.Header
func buildHeader(header restful.Header) spec.Header {
	responseHeader := spec.Header{}
	responseHeader.Type = header.Type
	responseHeader.Description = header.Description
	responseHeader.Format = header.Format
	responseHeader.Default = header.Default

	// If type is "array" items field is required
	if header.Type == arrayType {
		responseHeader.CollectionFormat = header.CollectionFormat
		responseHeader.Items = buildHeadersItems(header.Items)
	}

	return responseHeader
}

// buildHeadersItems builds
func buildHeadersItems(items *restful.Items) *spec.Items {
	responseItems := spec.NewItems()
	responseItems.Format = items.Format
	responseItems.Type = items.Type
	responseItems.Default = items.Default
	responseItems.CollectionFormat = items.CollectionFormat
	if items.Items != nil {
		responseItems.Items = buildHeadersItems(items.Items)
	}

	return responseItems
}

// stripTags takes a snippet of HTML and returns only the text content.
// For example, `<b>&lt;Hi!&gt;</b> <br>` -> `&lt;Hi!&gt; `.
func stripTags(html string) string {
	re := regexp.MustCompile("<[^>]*>")
	return re.ReplaceAllString(html, "")
}

func isPrimitiveType(modelName string) bool {
	if len(modelName) == 0 {
		return false
	}
	return strings.Contains("uint uint8 uint16 uint32 uint64 int int8 int16 int32 int64 float32 float64 bool string byte rune time.Time time.Duration", modelName)
}

func jsonSchemaType(modelName string) string {
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

		"byte":          "integer",
		"float64":       "number",
		"float32":       "number",
		"bool":          "boolean",
		"time.Time":     "string",
		"time.Duration": "integer",
	}
	mapped, ok := schemaMap[modelName]
	if !ok {
		return modelName // use as is (custom or struct)
	}
	return mapped
}
