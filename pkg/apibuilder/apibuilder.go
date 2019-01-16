/*

 Copyright 2019 The KubeSphere Authors.

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
package apibuilder

import "github.com/emicklei/go-restful"

type GroupVersion struct {
	Group   string
	Version string
}

type WebServiceBuilder struct {
	GroupVersion GroupVersion
	Routes       Routes
}

type APIBuilder []func(container *restful.Container)

type Routes []Route

type Route func(ws *restful.WebService)

func (builder *WebServiceBuilder) AddToContainer(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/apis/" + builder.GroupVersion.Group + "/" + builder.GroupVersion.Version).
		Produces(restful.MIME_JSON).Consumes(restful.MIME_JSON)
	for _, route := range builder.Routes {
		route(ws)
	}
	container.Add(ws)
}

func (builder APIBuilder) AddToContainer(container *restful.Container) {
	for _, addToContainer := range builder {
		addToContainer(container)
	}
}
