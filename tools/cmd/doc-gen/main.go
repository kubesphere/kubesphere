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
	"encoding/json"
	"flag"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/spec"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/apis"
	"log"
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

	apis.AddToContainer(restful.DefaultContainer)

	config := restfulspec.Config{
		WebServices:                   restful.RegisteredWebServices(),
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}

	swagger := restfulspec.BuildSwagger(config)

	data, _ := json.Marshal(swagger)
	err := ioutil.WriteFile(output, data, 420)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully written to %s",output)
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "KubeSphere",
			Description: "KubeSphere OpenAPI",
			Contact: &spec.ContactInfo{
				Name:  "kubesphere",
				Email: "kubesphere@gmail.com",
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
