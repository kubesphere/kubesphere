package restfulspec

import (
	"reflect"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
)

func buildDefinitions(ws *restful.WebService, cfg Config) (definitions spec.Definitions) {
	definitions = spec.Definitions{}
	for _, each := range ws.Routes() {
		addDefinitionsFromRouteTo(each, cfg, definitions)
	}
	return
}

func addDefinitionsFromRouteTo(r restful.Route, cfg Config, d spec.Definitions) {
	builder := definitionBuilder{Definitions: d, Config: cfg}
	if r.ReadSample != nil {
		builder.addModel(reflect.TypeOf(r.ReadSample), "")
	}
	if r.WriteSample != nil {
		builder.addModel(reflect.TypeOf(r.WriteSample), "")
	}
	for _, v := range r.ResponseErrors {
		if v.Model == nil {
			continue
		}
		builder.addModel(reflect.TypeOf(v.Model), "")
	}
}
