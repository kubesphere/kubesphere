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
	"flag"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/spec"
	"io/ioutil"
	_ "kubesphere.io/kubesphere/pkg/apis/iam/install"
	_ "kubesphere.io/kubesphere/pkg/apis/logging/install"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"log"
	// Install apis
	_ "kubesphere.io/kubesphere/pkg/apis/devops/install"
	_ "kubesphere.io/kubesphere/pkg/apis/metrics/install"
	_ "kubesphere.io/kubesphere/pkg/apis/monitoring/install"
	_ "kubesphere.io/kubesphere/pkg/apis/operations/install"
	_ "kubesphere.io/kubesphere/pkg/apis/resources/install"
	_ "kubesphere.io/kubesphere/pkg/apis/servicemesh/metrics/install"
	_ "kubesphere.io/kubesphere/pkg/apis/tenant/install"
)

var output string

func init() {
	flag.StringVar(&output, "output", "./api.json", "--output=./api.json")
}

func main() {
	flag.Parse()
	generateSwaggerJson()
}

func generateSwaggerJson() {

	container := runtime.Container

	apiTree(container)

	config := restfulspec.Config{
		WebServices:                   container.RegisteredWebServices(),
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}

	swagger := restfulspec.BuildSwagger(config)

	data, _ := json.Marshal(swagger)
	err := ioutil.WriteFile(output, data, 420)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully written to %s", output)
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "KubeSphere",
			Description: "KubeSphere OpenAPI",
			Contact: &spec.ContactInfo{
				Name:  "kubesphere",
				Email: "kubesphere@yunify.com",
				URL:   "kubesphere.io",
			},
			License: &spec.License{
				Name: "Apache",
				URL:  "http://www.apache.org/licenses/",
			},
			Version: "2.0.0",
		},
	}

	// setup security definitions
	swo.SecurityDefinitions = map[string]*spec.SecurityScheme{
		"jwt": spec.APIKeyAuth("Authorization", "header"),
	}
	swo.Security = []map[string][]string{{"jwt": []string{}}}
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
