package restfulspec

import (
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	restful "github.com/emicklei/go-restful"
	"github.com/go-openapi/spec"
)

// KeyOpenAPITags is a Metadata key for a restful Route
const KeyOpenAPITags = "openapi.tags"

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
// since openapi only supports setting the pattern as a a property named "pattern".
// Expressions like "/api/v1/{name:[a-z]/" are converted to "/api/v1/{name}/".
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
			fragment = split[0][1:]
			pattern := split[1][:len(split[1])-1]
			patterns[fragment] = pattern
			fragment = "{" + fragment + "}"
		}
		openapiPath += "/" + fragment
	}
	return openapiPath, patterns
}

func buildPathItem(ws *restful.WebService, r restful.Route, existingPathItem spec.PathItem, patterns map[string]string, cfg Config) spec.PathItem {
	op := buildOperation(ws, r, patterns, cfg)
	switch r.Method {
	case "GET":
		existingPathItem.Get = op
	case "POST":
		existingPathItem.Post = op
	case "PUT":
		existingPathItem.Put = op
	case "DELETE":
		existingPathItem.Delete = op
	case "PATCH":
		existingPathItem.Patch = op
	case "OPTIONS":
		existingPathItem.Options = op
	case "HEAD":
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
	// collect any path parameters
	for _, param := range ws.PathParameters() {
		o.Parameters = append(o.Parameters, buildParameter(r, param, patterns[param.Data().Name], cfg))
	}
	// route specific params
	for _, each := range r.ParameterDocs {
		o.Parameters = append(o.Parameters, buildParameter(r, each, patterns[each.Data().Name], cfg))
	}
	o.Responses = new(spec.Responses)
	props := &o.Responses.ResponsesProps
	props.StatusCodeResponses = map[int]spec.Response{}
	for k, v := range r.ResponseErrors {
		r := buildResponse(v, cfg)
		props.StatusCodeResponses[k] = r
	}
	if r.DefaultResponse != nil {
		r := buildResponse(*r.DefaultResponse, cfg)
		o.Responses.Default = &r
	}
	if len(o.Responses.StatusCodeResponses) == 0 {
		o.Responses.StatusCodeResponses[200] = spec.Response{ResponseProps: spec.ResponseProps{Description: http.StatusText(http.StatusOK)}}
	}
	return o
}

// stringAutoType automatically picks the correct type from an ambiguously typed
// string. Ex. numbers become int, true/false become bool, etc.
func stringAutoType(ambiguous string) interface{} {
	if ambiguous == "" {
		return nil
	}
	if parsedInt, err := strconv.ParseInt(ambiguous, 10, 64); err == nil {
		return parsedInt
	}
	if parsedBool, err := strconv.ParseBool(ambiguous); err == nil {
		return parsedBool
	}
	return ambiguous
}

func buildParameter(r restful.Route, restfulParam *restful.Parameter, pattern string, cfg Config) spec.Parameter {
	p := spec.Parameter{}
	param := restfulParam.Data()
	p.In = asParamType(param.Kind)
	p.Description = param.Description
	p.Name = param.Name
	p.Required = param.Required

	if len(param.AllowableValues) > 0 {
		p.Enum = make([]interface{}, 0, len(param.AllowableValues))
		for key := range param.AllowableValues {
			p.Enum = append(p.Enum, key)
		}
	}

	if param.Kind == restful.PathParameterKind {
		p.Pattern = pattern
	}
	st := reflect.TypeOf(r.ReadSample)
	if param.Kind == restful.BodyParameterKind && r.ReadSample != nil && param.DataType == st.String() {
		p.Schema = new(spec.Schema)
		p.SimpleSchema = spec.SimpleSchema{}
		if st.Kind() == reflect.Array || st.Kind() == reflect.Slice {
			dataTypeName := keyFrom(st.Elem(), cfg)
			p.Schema.Type = []string{"array"}
			p.Schema.Items = &spec.SchemaOrArray{
				Schema: &spec.Schema{},
			}
			isPrimitive := isPrimitiveType(dataTypeName)
			if isPrimitive {
				mapped := jsonSchemaType(dataTypeName)
				p.Schema.Items.Schema.Type = []string{mapped}
			} else {
				p.Schema.Items.Schema.Ref = spec.MustCreateRef("#/definitions/" + dataTypeName)
			}
		} else {
			dataTypeName := keyFrom(st, cfg)
			p.Schema.Ref = spec.MustCreateRef("#/definitions/" + dataTypeName)
		}

	} else {
		if param.AllowMultiple {
			p.Type = "array"
			p.Items = spec.NewItems()
			p.Items.Type = param.DataType
			p.CollectionFormat = param.CollectionFormat
		} else {
			p.Type = param.DataType
		}
		p.Default = stringAutoType(param.DefaultValue)
		p.Format = param.DataFormat
	}

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
			r.Schema.Type = []string{"array"}
			r.Schema.Items = &spec.SchemaOrArray{
				Schema: &spec.Schema{},
			}
			isPrimitive := isPrimitiveType(modelName)
			if isPrimitive {
				mapped := jsonSchemaType(modelName)
				r.Schema.Items.Schema.Type = []string{mapped}
			} else {
				r.Schema.Items.Schema.Ref = spec.MustCreateRef("#/definitions/" + modelName)
			}
		} else {
			modelName := keyFrom(st, cfg)
			if isPrimitiveType(modelName) {
				// If the response is a primitive type, then don't reference any definitions.
				// Instead, set the schema's "type" to the model name.
				r.Schema.AddType(modelName, "")
			} else {
				modelName := keyFrom(st, cfg)
				r.Schema.Ref = spec.MustCreateRef("#/definitions/" + modelName)
			}
		}
	}
	return r
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

		"byte":      "integer",
		"float64":   "number",
		"float32":   "number",
		"bool":      "boolean",
		"time.Time": "string",
		"time.Duration": "integer",
	}
	mapped, ok := schemaMap[modelName]
	if !ok {
		return modelName // use as is (custom or struct)
	}
	return mapped
}
