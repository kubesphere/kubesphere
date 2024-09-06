/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"fmt"
	"net/http"

	"github.com/NYTimes/gziphandler"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/go-openapi/spec"
	"k8s.io/klog/v2"

	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"

	"kubesphere.io/kubesphere/kube/pkg/openapi"
	"kubesphere.io/kubesphere/kube/pkg/openapi/merge"
)

var OpenApiPath = "/openapi/v2"

type OpenApiV2Services struct {
	openApiSpecCache         map[string]*openapi.Cache[*spec.Swagger]
	openApiAggregatorService *openapi.OpenApiAggregatorServices
}

func NewOpenApiV2Services() *OpenApiV2Services {
	return &OpenApiV2Services{
		openApiSpecCache:         make(map[string]*openapi.Cache[*spec.Swagger]),
		openApiAggregatorService: openapi.NewOpenApiAggregatorServices(),
	}
}

func (s *OpenApiV2Services) AddUpdateApiService(apiService *extensionsv1alpha1.APIService) error {
	c := &openapi.Cache[*spec.Swagger]{}
	c.Store(&spec.Swagger{})
	s.openApiSpecCache[apiService.Name] = c
	if err := s.openApiAggregatorService.AddUpdateApiService(apiService); err != nil {
		return err
	}
	return s.UpdateOpenApiSpec(apiService.Name)
}

func (s *OpenApiV2Services) UpdateOpenApiSpec(apiServiceName string) error {
	data, err := s.openApiAggregatorService.GetOpenApiSpecV2(apiServiceName)
	if err != nil {
		return err
	}
	openAPISpec := &spec.Swagger{}
	if err := openAPISpec.UnmarshalJSON(data); err != nil {
		return err
	}

	if cache, ok := s.openApiSpecCache[apiServiceName]; ok {
		cache.Store(openAPISpec)
	} else {
		c := openapi.Cache[*spec.Swagger]{}
		c.Store(openAPISpec)
		s.openApiSpecCache[apiServiceName] = &c
	}
	return nil
}

func (s *OpenApiV2Services) RemoveApiService(apiServiceName string) {
	s.openApiAggregatorService.RemoveApiService(apiServiceName)
	delete(s.openApiSpecCache, apiServiceName)
}

func (s *OpenApiV2Services) MergeSpecCache() (*spec.Swagger, error) {
	var merged *spec.Swagger
	for i := range s.openApiSpecCache {
		if cacheValue, ok := s.openApiSpecCache[i]; ok {
			cacheSpec := cacheValue.Load()
			if merged == nil {
				merged = &spec.Swagger{}
				*merged = *cacheSpec
				merged.Paths = nil
				merged.Definitions = nil
				merged.Parameters = nil
			}
			if err := merge.MergeSpecsIgnorePathConflictRenamingDefinitionsAndParameters(merged, cacheSpec); err != nil {
				return nil, fmt.Errorf("failed to build merge specs: %v", err)
			}
		}
	}
	return merged, nil
}

func (s *OpenApiV2Services) RegisterOpenAPIVersionedService(servePath string, handler openapi.PathHandler) {
	handler.Handle(servePath, gziphandler.GzipHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			result, err := s.MergeSpecCache()
			if err != nil {
				klog.Errorf("Error in OpenAPI handler: %s", err)
				// only return a 503 if we have no older cache data to serve
				if result == nil {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
			}
			data, err := (*result).MarshalJSON()
			if err != nil {
				klog.Errorf("Error in OpenAPI handler: %s", err)
				if data == nil {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
			}
			w.Write(data)
		}),
	))
}

func (s *OpenApiV2Services) AddLocalApiService(name string, val *spec.Swagger) {
	s.openApiAggregatorService.AddLocalApiService(name)
	c := &openapi.Cache[*spec.Swagger]{}
	c.Store(val)
	s.openApiSpecCache[name] = c
}

func BuildAndRegisterAggregator(
	config *restfulspec.Config, pathHandler openapi.PathHandler) (*OpenApiV2Services, error) {

	aggregatorOpenAPISpec := restfulspec.BuildSwagger(*config)
	aggregatorOpenAPISpec.Definitions = merge.PruneDefaults(aggregatorOpenAPISpec.Definitions)

	s := buildAndRegisterOpenApiV2ForLocalServices(OpenApiPath, aggregatorOpenAPISpec, pathHandler)
	return s, nil
}

func buildAndRegisterOpenApiV2ForLocalServices(path string, aggregatorSpec *spec.Swagger, pathHandler openapi.PathHandler) *OpenApiV2Services {
	s := NewOpenApiV2Services()
	s.AddLocalApiService("kubeSphere_internal_local_delegation", aggregatorSpec)
	if path == "" {
		path = OpenApiPath
	}
	s.RegisterOpenAPIVersionedService(path, pathHandler)
	return s
}
