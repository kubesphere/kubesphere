/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

const (
	GroupName = "config.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func (h *handler) AddToContainer(c *restful.Container) error {
	webservice := runtime.NewWebService(GroupVersion)

	webservice.Route(webservice.GET("/configs/oauth").
		Doc("OAuth configurations").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagPlatformConfigurations}).
		Notes("Information about the authorization server are published.").
		Operation("oauth-config").
		To(h.getOAuthConfiguration))

	webservice.Route(webservice.GET("/configs/configz").
		Deprecate().
		Doc("Component configurations").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagPlatformConfigurations}).
		Notes("Information about the components configuration").
		Operation("component-config").
		To(func(request *restful.Request, response *restful.Response) {
			_ = response.WriteAsJson(h.config)
		}))

	webservice.Route(webservice.GET("/configs/theme").
		Doc("Retrieve theme configurations").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagPlatformConfigurations}).
		Notes("Retrieve theme configuration settings.").
		Operation("get-theme-config").
		To(h.getThemeConfiguration))

	webservice.Route(webservice.POST("/configs/theme").
		Doc("Update theme configurations").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagPlatformConfigurations}).
		Notes("Update theme configuration settings.").
		Operation("update-theme-config").
		To(h.updateThemeConfiguration))

	c.Add(webservice)
	return nil
}
