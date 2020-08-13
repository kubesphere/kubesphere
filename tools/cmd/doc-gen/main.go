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
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/pkg/errors"
	"io/ioutil"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	devopsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/devops/v1alpha2"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/devops/v1alpha3"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/iam/v1alpha2"
	monitoringv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/monitoring/v1alpha3"
	networkv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/network/v1alpha2"
	openpitrixv1 "kubesphere.io/kubesphere/pkg/kapis/openpitrix/v1"
	operationsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/operations/v1alpha2"
	resourcesv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha2"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha3"
	metricsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/servicemesh/metrics/v1alpha2"
	tenantv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/tenant/v1alpha2"
	terminalv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/terminal/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/fake"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	fakes3 "kubesphere.io/kubesphere/pkg/simple/client/s3/fake"
	"kubesphere.io/kubesphere/pkg/version"
	"log"
)

var output string

func init() {
	flag.StringVar(&output, "output", "./api/ks-openapi-spec/swagger.json", "--output=./api.json")
}

func main() {
	flag.Parse()
	swaggerSpec := generateSwaggerJson()

	err := validateSpec(swaggerSpec)
	if err != nil {
		klog.Warningf("Swagger specification has errors")
	}
}

func validateSpec(apiSpec []byte) error {

	swaggerDoc, err := loads.Analyzed(apiSpec, "")
	if err != nil {
		return err
	}

	// Attempts to report about all errors
	validate.SetContinueOnErrors(true)

	v := validate.NewSpecValidator(swaggerDoc.Schema(), strfmt.Default)
	result, _ := v.Validate(swaggerDoc)

	if result.HasWarnings() {
		log.Printf("See warnings below:\n")
		for _, desc := range result.Warnings {
			log.Printf("- WARNING: %s\n", desc.Error())
		}

	}
	if result.HasErrors() {
		str := fmt.Sprintf("The swagger spec is invalid against swagger specification %s.\nSee errors below:\n", swaggerDoc.Version())
		for _, desc := range result.Errors {
			str += fmt.Sprintf("- %s\n", desc.Error())
		}
		log.Println(str)
		return errors.New(str)
	}

	return nil
}

func generateSwaggerJson() []byte {

	container := runtime.Container
	clientsets := k8s.NewNullClient()

	informerFactory := informers.NewNullInformerFactory()

	urlruntime.Must(devopsv1alpha2.AddToContainer(container, informerFactory.KubeSphereSharedInformerFactory(), &fake.Devops{}, nil, clientsets.KubeSphere(), fakes3.NewFakeS3(), ""))
	urlruntime.Must(devopsv1alpha3.AddToContainer(container, &fake.Devops{}, clientsets.Kubernetes(), clientsets.KubeSphere(), informerFactory.KubeSphereSharedInformerFactory(), informerFactory.KubernetesSharedInformerFactory()))
	urlruntime.Must(iamv1alpha2.AddToContainer(container, im.NewOperator(clientsets.KubeSphere(), informerFactory, nil), am.NewReadOnlyOperator(informerFactory), authoptions.NewAuthenticateOptions()))
	urlruntime.Must(monitoringv1alpha3.AddToContainer(container, clientsets.Kubernetes(), nil, informerFactory, nil))
	urlruntime.Must(openpitrixv1.AddToContainer(container, informerFactory, nil))
	urlruntime.Must(operationsv1alpha2.AddToContainer(container, clientsets.Kubernetes()))
	urlruntime.Must(resourcesv1alpha2.AddToContainer(container, clientsets.Kubernetes(), informerFactory, ""))
	urlruntime.Must(resourcesv1alpha3.AddToContainer(container, informerFactory))
	urlruntime.Must(tenantv1alpha2.AddToContainer(container, informerFactory, nil, nil, nil, nil, nil))
	urlruntime.Must(terminalv1alpha2.AddToContainer(container, clientsets.Kubernetes(), nil))
	urlruntime.Must(metricsv1alpha2.AddToContainer(container))
	urlruntime.Must(networkv1alpha2.AddToContainer(container, ""))

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
			Tags: []string{constants.LogQueryTag},
		},
	})

	data, _ := json.MarshalIndent(swagger, "", "  ")
	err := ioutil.WriteFile(output, data, 420)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully written to %s", output)

	return data
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
			Version: version.Get().GitVersion,
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
