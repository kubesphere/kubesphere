/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha2

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

const (
	GroupName  = "config.kubesphere.io"
	Version    = "v1alpha2"
	APIVersion = GroupName + "/" + Version
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

func (h *handler) AddToContainer(c *restful.Container) error {
	webservice := runtime.NewWebService(GroupVersion)
	webservice.Route(webservice.GET("/configs/configz").
		Deprecate().
		Doc("Retrieve multicluster configuration").
		Notes("Provides information about the multicluster configuration.").
		Operation("getMulticlusterConfiguration").
		To(h.getConfigz))

	if h.config.MultiClusterOptions.ClusterRole == string(clusterv1alpha1.ClusterRoleHost) {
		webservice.Route(webservice.GET("/configs/oauth").
			Doc("Retrieve OAuth configuration").
			Notes("Provides information about the authorization server.").
			Operation("getOAuthConfiguration").
			To(h.getOAuthConfiguration))

		webservice.Route(webservice.GET("/configs/theme").
			Doc("Retrieve the current theme configuration").
			Notes("Provides the current theme configuration details.").
			Operation("getThemeConfiguration").
			To(h.getThemeConfiguration))

		webservice.Route(webservice.PUT("/configs/theme").
			Doc("Update the theme configuration settings").
			Notes("Allows the user to update the theme configuration settings.").
			Operation("updateThemeConfiguration").
			To(h.updateThemeConfiguration))

		webservice.Route(webservice.GET("/clusterconnectionconfigurations").
			Doc("Retrieve all configurations for cluster connection").
			Notes("Provides information about all cluster connection plugins").
			Operation("listClusterConnectionConfiguration").
			To(h.listClusterConnectionConfiguration))

		webservice.Route(webservice.GET("/clusterconnectionconfigurations/{config}").
			Doc("Retrieve the configuration for cluster connection").
			Notes("Provides information about the cluster connection plugin").
			Operation("getClusterConnectionConfiguration").
			To(h.getClusterConnectionConfiguration))

		webservice.Route(webservice.POST("/platformconfigs").
			Doc("Create a new platform configuration").
			Notes("Allows the user to create a new configuration for the specified platform.").
			Operation("createPlatformConfiguration").
			To(h.createPlatformConfiguration))

		webservice.Route(webservice.DELETE("/platformconfigs/{config}").
			Doc("Delete the specified platform configuration").
			Notes("Allows the user to delete the configuration settings of the specified platform.").
			Operation("deletePlatformConfiguration").
			To(h.deletePlatformConfiguration))

		webservice.Route(webservice.PUT("/platformconfigs/{config}").
			Doc("Update the specified platform configuration settings").
			Notes("Allows the user to modify the configuration settings of the specified platform").
			Operation("updatePlatformConfiguration").
			To(h.updatePlatformConfiguration))

		webservice.Route(webservice.PATCH("/platformconfigs/{config}").
			Doc("Patch the specified platform configuration settings").
			Consumes(runtime.MimeMergePatchJson).
			Notes("Allows the user to apply partial modifications to the configuration settings of the specified platform").
			Operation("patchPlatformConfiguration").
			To(h.patchPlatformConfiguration))

		webservice.Route(webservice.GET("/platformconfigs/{config}").
			Doc("Retrieve the specified platform configuration").
			Notes("Provides details of the specified platform configuration.").
			Operation("getPlatformConfiguration").
			To(h.getPlatformConfiguration))
	}

	for _, route := range webservice.Routes() {
		route.Metadata = map[string]interface{}{
			restfulspec.KeyOpenAPITags: []string{api.TagPlatformConfigurations},
		}
	}

	c.Add(webservice)
	return nil
}
