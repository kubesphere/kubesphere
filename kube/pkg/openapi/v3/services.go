/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v3

import (
	"fmt"
	"net/http"

	"github.com/NYTimes/gziphandler"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	openapibuilder "k8s.io/apiextensions-apiserver/pkg/controller/openapi/builder"
	"k8s.io/klog/v2"
	"k8s.io/kube-openapi/pkg/openapiconv"
	"k8s.io/kube-openapi/pkg/spec3"
	spec2 "k8s.io/kube-openapi/pkg/validation/spec"

	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"

	"kubesphere.io/kubesphere/kube/pkg/openapi"
	"kubesphere.io/kubesphere/kube/pkg/openapi/merge"
)

var OpenApiPath = "/openapi/v3"

type OpenApiV3Services struct {
	openApiSpecCache         map[string]*openapi.Cache[*spec3.OpenAPI]
	openApiAggregatorService *openapi.OpenApiAggregatorServices
}

func NewOpenApiV3Services() *OpenApiV3Services {
	return &OpenApiV3Services{
		openApiSpecCache:         make(map[string]*openapi.Cache[*spec3.OpenAPI]),
		openApiAggregatorService: openapi.NewOpenApiAggregatorServices(),
	}
}

func (s *OpenApiV3Services) AddUpdateApiService(apiService *extensionsv1alpha1.APIService) error {
	c := &openapi.Cache[*spec3.OpenAPI]{}
	c.Store(&spec3.OpenAPI{})
	s.openApiSpecCache[apiService.Name] = c
	if err := s.openApiAggregatorService.AddUpdateApiService(apiService); err != nil {
		return err
	}
	return s.UpdateOpenApiSpec(apiService.Name)
}

func (s *OpenApiV3Services) UpdateOpenApiSpec(apiServiceName string) error {
	data, err := s.openApiAggregatorService.GetOpenApiSpecV3(apiServiceName)
	if err != nil {
		return err
	}
	openAPISpec := &spec3.OpenAPI{}
	if err := openAPISpec.UnmarshalJSON(data); err != nil {
		return err
	}

	if cache, ok := s.openApiSpecCache[apiServiceName]; ok {
		cache.Store(openAPISpec)
		return nil
	} else {
		c := openapi.Cache[*spec3.OpenAPI]{}
		c.Store(openAPISpec)
		s.openApiSpecCache[apiServiceName] = &c
	}
	return nil
}

func (s *OpenApiV3Services) RemoveApiService(apiServiceName string) {
	s.openApiAggregatorService.RemoveApiService(apiServiceName)
	delete(s.openApiSpecCache, apiServiceName)
}

func (s *OpenApiV3Services) MergeSpecCache() (*spec3.OpenAPI, error) {
	var merged *spec3.OpenAPI
	var err error
	for i := range s.openApiSpecCache {
		if cacheValue, ok := s.openApiSpecCache[i]; ok {
			cacheSpec := cacheValue.Load()
			if merged == nil {
				merged = &spec3.OpenAPI{}
				*merged = *cacheSpec
				merged.Paths = nil
			}
			if merged, err = openapibuilder.MergeSpecsV3(merged, cacheSpec); err != nil {
				return nil, fmt.Errorf("failed to build merge specs: %v", err)
			}
		}
	}
	return merged, nil
}

func (s *OpenApiV3Services) RegisterOpenAPIVersionedService(servePath string, handler openapi.PathHandler) {
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

func (s *OpenApiV3Services) AddLocalApiService(name string, val *spec3.OpenAPI) {
	s.openApiAggregatorService.AddLocalApiService(name)
	c := &openapi.Cache[*spec3.OpenAPI]{}
	c.Store(val)
	s.openApiSpecCache[name] = c
}

func BuildAndRegisterAggregator(
	config *restfulspec.Config, pathHandler openapi.PathHandler) (*OpenApiV3Services, error) {

	aggregatorOpenAPISpec := restfulspec.BuildSwagger(*config)
	aggregatorOpenAPISpec.Definitions = merge.PruneDefaults(aggregatorOpenAPISpec.Definitions)
	swaggerData, err := aggregatorOpenAPISpec.MarshalJSON()
	if err != nil {
		return nil, err
	}
	spec2Swagger := spec2.Swagger{}
	if err = spec2Swagger.UnmarshalJSON(swaggerData); err != nil {
		return nil, err
	}
	convertedOpenAPIV3 := openapiconv.ConvertV2ToV3(&spec2Swagger)
	s := buildAndRegisterOpenApiV3ForLocalServices(OpenApiPath, convertedOpenAPIV3, pathHandler)
	return s, nil
}

func buildAndRegisterOpenApiV3ForLocalServices(path string, aggregatorSpec *spec3.OpenAPI, pathHandler openapi.PathHandler) *OpenApiV3Services {
	s := NewOpenApiV3Services()
	s.AddLocalApiService("kubeSphere_internal_local_delegation", aggregatorSpec)
	if path == "" {
		path = OpenApiPath
	}
	s.RegisterOpenAPIVersionedService(path, pathHandler)
	return s
}
