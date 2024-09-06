/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package openapi

import (
	"fmt"
	"net/http"
	"sync/atomic"

	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
)

type Cache[T any] struct {
	value atomic.Pointer[T]
}

func (c *Cache[T]) Store(v T) {
	c.value.Store(&v)
}

func (c *Cache[T]) Load() T {
	return *c.value.Load()
}

type PathHandler interface {
	Handle(path string, handler http.Handler)
}

type APIServiceManager interface {
	AddUpdateApiService(apiService *extensionsv1alpha1.APIService) error
	UpdateOpenApiSpec(apiServiceName string) error
	RemoveApiService(apiServiceName string)
}

type OpenApiAggregatorServices struct {
	apiService    map[string]ApiService
	downloaderMap map[string]CacheableDownloader
	downloader    *Downloader
}

func NewOpenApiAggregatorServices() *OpenApiAggregatorServices {
	return &OpenApiAggregatorServices{
		apiService:    make(map[string]ApiService),
		downloaderMap: make(map[string]CacheableDownloader),
		downloader:    NewDownloader(),
	}
}

func (o *OpenApiAggregatorServices) AddUpdateApiService(apiService *extensionsv1alpha1.APIService) error {
	openapiService := NewApiService(apiService)
	o.apiService[apiService.Name] = openapiService

	if d, ok := o.downloaderMap[apiService.Name]; ok {
		if err := d.UpdateDownloader(openapiService); err != nil {
			return err
		}
	} else {
		cacheDownloader, err := NewCacheableDownloader(openapiService, o.downloader)
		if err != nil {
			return err
		}
		o.downloaderMap[apiService.Name] = cacheDownloader
	}
	return nil
}

func (o *OpenApiAggregatorServices) GetOpenApiSpecV2(apiServiceName string) ([]byte, error) {
	if d, ok := o.downloaderMap[apiServiceName]; ok {
		data, err := d.GetV2()
		if err != nil {
			return nil, fmt.Errorf("fetch ApiService %s openapi-v2 error: %s", apiServiceName, err)
		}
		return data, nil
	}
	return nil, fmt.Errorf("update OpenApiSpec failed beaseuse of apiService %s not found", apiServiceName)
}

func (o *OpenApiAggregatorServices) GetOpenApiSpecV3(apiServiceName string) ([]byte, error) {
	if d, ok := o.downloaderMap[apiServiceName]; ok {
		data, err := d.GetV3()
		if err != nil {
			return nil, fmt.Errorf("fetch ApiService %s openapi-v3 error: %s", apiServiceName, err)
		}
		return data, nil
	}
	return nil, fmt.Errorf("update OpenApiSpec failed beaseuse of apiService %s not found", apiServiceName)
}

func (o *OpenApiAggregatorServices) AddLocalApiService(name string) {
	apiService := extensionsv1alpha1.APIService{}
	apiService.Name = name

	o.apiService[apiService.Name] = NewApiService(&apiService)
}

func (o *OpenApiAggregatorServices) RemoveApiService(apiServiceName string) {
	delete(o.apiService, apiServiceName)
	delete(o.downloaderMap, apiServiceName)
}
