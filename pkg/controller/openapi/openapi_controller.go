/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package openapi

import (
	"context"
	"fmt"

	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"

	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"

	"kubesphere.io/kubesphere/kube/pkg/openapi"
)

var SharedOpenAPIController = NewController()

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) WatchOpenAPIChanges(ctx context.Context, cache runtimecache.Cache, openAPIV2Service openapi.APIServiceManager, openAPIV3Service openapi.APIServiceManager) error {
	informer, err := cache.GetInformer(ctx, &extensionsv1alpha1.APIService{})
	if err != nil {
		return fmt.Errorf("get informer failed: %w", err)
	}
	_, err = informer.AddEventHandler(toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			apiService := obj.(*extensionsv1alpha1.APIService)
			return apiService.Status.State == extensionsv1alpha1.StateAvailable
		},
		Handler: &toolscache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				apiService := obj.(*extensionsv1alpha1.APIService)
				if err := openAPIV2Service.AddUpdateApiService(apiService); err != nil {
					klog.V(4).Infof("add openapi v2 service failed: %v", err)
				}
				if err := openAPIV3Service.AddUpdateApiService(apiService); err != nil {
					klog.V(4).Infof("add openapi v3 service failed: %v", err)
				}
			},
			UpdateFunc: func(old, new interface{}) {
				apiService := new.(*extensionsv1alpha1.APIService)
				if err := openAPIV2Service.AddUpdateApiService(apiService); err != nil {
					klog.V(4).Infof("update openapi v2 service failed: %v", err)
				}
				if err := openAPIV3Service.AddUpdateApiService(apiService); err != nil {
					klog.V(4).Infof("update openapi v3 service failed: %v", err)
				}
			},
			DeleteFunc: func(obj interface{}) {
				apiService := obj.(*extensionsv1alpha1.APIService)
				openAPIV2Service.RemoveApiService(apiService.Name)
				openAPIV3Service.RemoveApiService(apiService.Name)
			},
		},
	})
	if err != nil {
		return fmt.Errorf("add event handler failed: %w", err)
	}
	return nil
}
