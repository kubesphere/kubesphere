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
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"kubesphere.io/kubesphere/pkg/apis"
	"log"
)

func main() {
	generateSwaggerJson()
}

func generateSwaggerJson() {

	config := restfulspec.Config{
		WebServices: restful.RegisteredWebServices(),
	}

	swagger := restfulspec.BuildSwagger(config)

	container := restful.NewContainer()

	apis.AddToContainer(container)

	apiTree(container)

	data, _ := json.Marshal(swagger)
	log.Println(string(data))
}

func apiTree(container *restful.Container) {
	buf := bytes.NewBufferString("\n")
	for _, ws := range container.RegisteredWebServices() {
		for _, route := range ws.Routes() {
			buf.WriteString(fmt.Sprintf("%s %s\n", route.Method, route.Path))
		}
	}
	log.Println(buf.String())
}
