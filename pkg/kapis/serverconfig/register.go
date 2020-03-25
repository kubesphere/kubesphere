/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package serverconfig

import (
	"github.com/emicklei/go-restful"
	apiserverconfig "kubesphere.io/kubesphere/pkg/apiserver/config"
)

func AddToContainer(c *restful.Container, config *apiserverconfig.Config) error {
	configs := &restful.WebService{}

	configs.Path("/server/configs").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// information about the authorization server are published.
	configs.Route(configs.GET("/oauth-configz").To(func(request *restful.Request, response *restful.Response) {
		response.WriteEntity(config.AuthenticationOptions.OAuthOptions)
	}))

	// information about the server configuration
	configs.Route(configs.GET("/configz").To(func(request *restful.Request, response *restful.Response) {
		response.WriteAsJson(config.ToMap())
	}))

	c.Add(configs)
	return nil
}
