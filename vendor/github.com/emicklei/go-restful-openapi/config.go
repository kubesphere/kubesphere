package restfulspec

import (
	"reflect"

	restful "github.com/emicklei/go-restful"
	"github.com/go-openapi/spec"
)

// MapSchemaFormatFunc can be used to modify typeName at definition time.
// To use it set the SchemaFormatHandler in the config.
type MapSchemaFormatFunc func(typeName string) string

// MapModelTypeNameFunc can be used to return the desired typeName for a given
// type. It will return false if the default name should be used.
// To use it set the ModelTypeNameHandler in the config.
type MapModelTypeNameFunc func(t reflect.Type) (string, bool)

// PostBuildSwaggerObjectFunc can be used to change the creates Swagger Object
// before serving it. To use it set the PostBuildSwaggerObjectHandler in the config.
type PostBuildSwaggerObjectFunc func(s *spec.Swagger)

// Config holds service api metadata.
type Config struct {
	// WebServicesURL is a DEPRECATED field; it never had any effect in this package.
	WebServicesURL string
	// APIPath is the path where the JSON api is avaiable , e.g. /apidocs.json
	APIPath string
	// api listing is constructed from this list of restful WebServices.
	WebServices []*restful.WebService
	// [optional] on default CORS (Cross-Origin-Resource-Sharing) is enabled.
	DisableCORS bool
	// Top-level API version. Is reflected in the resource listing.
	APIVersion string
	// [optional] If set, model builder should call this handler to get addition typename-to-swagger-format-field conversion.
	SchemaFormatHandler MapSchemaFormatFunc
	// [optional] If set, model builder should call this handler to retrieve the name for a given type.
	ModelTypeNameHandler MapModelTypeNameFunc
	// [optional] If set then call this function with the generated Swagger Object
	PostBuildSwaggerObjectHandler PostBuildSwaggerObjectFunc
}
