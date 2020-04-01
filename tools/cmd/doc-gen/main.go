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
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	fakeksClient "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/constants"
	devopsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/devops/v1alpha2"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/iam/v1alpha2"
	loggingv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/logging/v1alpha2"
	monitoringv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/monitoring/v1alpha2"
	openpitrixv1 "kubesphere.io/kubesphere/pkg/kapis/openpitrix/v1"
	operationsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/operations/v1alpha2"
	resourcesv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha2"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha3"
	metricsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/servicemesh/metrics/v1alpha2"
	tenantv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/tenant/v1alpha2"
	terminalv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/terminal/v1alpha2"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/fake"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	fakes3 "kubesphere.io/kubesphere/pkg/simple/client/s3/fake"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
	"log"
	"time"
)

var output string

func init() {
	flag.StringVar(&output, "output", "./api/ks-openapi-spec/swagger.json", "--output=./api.json")
}

func main() {
	flag.Parse()
	generateSwaggerJson()
}

func generateSwaggerJson() {

	container := runtime.Container
	devopsv1alpha2Service := runtime.NewWebService(devopsv1alpha2.GroupVersion)
	urlruntime.Must(devopsv1alpha2.AddPipelineToWebService(devopsv1alpha2Service, &fake.Devops{}, &mysql.Database{}))
	urlruntime.Must(devopsv1alpha2.AddS2IToWebService(devopsv1alpha2Service, fakeksClient.NewSimpleClientset(), externalversions.NewSharedInformerFactory(
		fakeksClient.NewSimpleClientset(), time.Duration(0)), fakes3.NewFakeS3()))
	urlruntime.Must(devopsv1alpha2.AddSonarToWebService(devopsv1alpha2Service, &fake.Devops{}, &mysql.Database{}, sonarqube.NewSonar(nil)))
	container.Add(devopsv1alpha2Service)
	urlruntime.Must(iamv1alpha2.AddToContainer(container, nil, nil, nil, nil, nil))
	urlruntime.Must(loggingv1alpha2.AddToContainer(container, nil, nil))
	urlruntime.Must(monitoringv1alpha2.AddToContainer(container, nil, nil))
	urlruntime.Must(openpitrixv1.AddToContainer(container, nil, nil))
	urlruntime.Must(operationsv1alpha2.AddToContainer(container, nil))
	urlruntime.Must(resourcesv1alpha2.AddToContainer(container, nil, nil))
	urlruntime.Must(resourcesv1alpha3.AddToContainer(container, nil))
	urlruntime.Must(tenantv1alpha2.AddToContainer(container, nil, nil, nil))
	urlruntime.Must(terminalv1alpha2.AddToContainer(container, nil, nil))
	urlruntime.Must(metricsv1alpha2.AddToContainer(container))

	config := restfulspec.Config{
		WebServices:                   container.RegisteredWebServices(),
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}

	swagger := restfulspec.BuildSwagger(config)

	swagger.Info.Extensions = make(spec.Extensions)
	swagger.Info.Extensions.Add("x-tagGroups", []struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}{
		{
			Name: "IAM",
			Tags: []string{constants.IdentityManagementTag, constants.AccessManagementTag},
		},
		{
			Name: "Resources",
			Tags: []string{constants.ClusterResourcesTag, constants.NamespaceResourcesTag, constants.UserResourcesTag},
		},
		{
			Name: "Monitoring",
			Tags: []string{constants.ComponentStatusTag},
		},
		{
			Name: "Tenant",
			Tags: []string{constants.TenantResourcesTag},
		},
		{
			Name: "Other",
			Tags: []string{constants.VerificationTag, constants.RegistryTag},
		},
		{
			Name: "DevOps",
			Tags: []string{constants.DevOpsProjectTag, constants.DevOpsProjectCredentialTag,
				constants.DevOpsPipelineTag, constants.DevOpsProjectMemberTag,
				constants.DevOpsWebhookTag, constants.DevOpsJenkinsfileTag, constants.DevOpsScmTag},
		},
		{
			Name: "Monitoring",
			Tags: []string{constants.ClusterMetricsTag, constants.NodeMetricsTag, constants.NamespaceMetricsTag, constants.WorkloadMetricsTag,
				constants.PodMetricsTag, constants.ContainerMetricsTag, constants.WorkspaceMetricsTag, constants.ComponentMetricsTag},
		},
		{
			Name: "Logging",
			Tags: []string{constants.LogQueryTag, constants.FluentBitSetting},
		},
	})

	data, _ := json.MarshalIndent(swagger, "", "  ")
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
				URL:   "https://kubesphere.io",
			},
			License: &spec.License{
				Name: "Apache",
				URL:  "http://www.apache.org/licenses/",
			},
			Version: "2.0.2",
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
