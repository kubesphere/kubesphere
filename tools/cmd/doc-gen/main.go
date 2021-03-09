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
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/version"
	"log"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/pkg/errors"
	promfake "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/fake"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	alertingv2alpha1 "kubesphere.io/kubesphere/pkg/kapis/alerting/v2alpha1"
	clusterkapisv1alpha1 "kubesphere.io/kubesphere/pkg/kapis/cluster/v1alpha1"
	devopsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/devops/v1alpha2"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/devops/v1alpha3"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/iam/v1alpha2"
	monitoringv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/monitoring/v1alpha3"
	networkv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/network/v1alpha2"
	"kubesphere.io/kubesphere/pkg/kapis/oauth"
	openpitrixv1 "kubesphere.io/kubesphere/pkg/kapis/openpitrix/v1"
	openpitrixv2 "kubesphere.io/kubesphere/pkg/kapis/openpitrix/v2alpha1"
	operationsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/operations/v1alpha2"
	resourcesv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha2"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha3"
	metricsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/servicemesh/metrics/v1alpha2"
	tenantv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/tenant/v1alpha2"
	terminalv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/terminal/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/iam/group"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
	fakedevops "kubesphere.io/kubesphere/pkg/simple/client/devops/fake"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	fakes3 "kubesphere.io/kubesphere/pkg/simple/client/s3/fake"
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

	urlruntime.Must(oauth.AddToContainer(container, nil, nil, nil, nil, nil, nil))
	urlruntime.Must(clusterkapisv1alpha1.AddToContainer(container, informerFactory.KubernetesSharedInformerFactory(),
		informerFactory.KubeSphereSharedInformerFactory(), "", "", ""))
	urlruntime.Must(devopsv1alpha2.AddToContainer(container, informerFactory.KubeSphereSharedInformerFactory(), &fakedevops.Devops{}, nil, clientsets.KubeSphere(), fakes3.NewFakeS3(), "", nil))
	urlruntime.Must(devopsv1alpha3.AddToContainer(container, &fakedevops.Devops{}, clientsets.Kubernetes(), clientsets.KubeSphere(), informerFactory.KubeSphereSharedInformerFactory(), informerFactory.KubernetesSharedInformerFactory()))
	urlruntime.Must(iamv1alpha2.AddToContainer(container, nil, nil, group.New(informerFactory, clientsets.KubeSphere(), clientsets.Kubernetes()), nil))
	urlruntime.Must(monitoringv1alpha3.AddToContainer(container, clientsets.Kubernetes(), nil, nil, informerFactory))
	urlruntime.Must(openpitrixv1.AddToContainer(container, informerFactory, fake.NewSimpleClientset(), nil))
	urlruntime.Must(openpitrixv2.AddToContainer(container, informerFactory, fake.NewSimpleClientset(), nil))
	urlruntime.Must(operationsv1alpha2.AddToContainer(container, clientsets.Kubernetes()))
	urlruntime.Must(resourcesv1alpha2.AddToContainer(container, clientsets.Kubernetes(), informerFactory, ""))
	urlruntime.Must(resourcesv1alpha3.AddToContainer(container, informerFactory, nil))
	urlruntime.Must(tenantv1alpha2.AddToContainer(container, informerFactory, nil, nil, nil, nil, nil, nil, nil, nil, nil))
	urlruntime.Must(terminalv1alpha2.AddToContainer(container, clientsets.Kubernetes(), nil))
	urlruntime.Must(metricsv1alpha2.AddToContainer(container))
	urlruntime.Must(networkv1alpha2.AddToContainer(container, ""))
	alertingOptions := &alerting.Options{}
	alertingClient, _ := alerting.NewRuleClient(alertingOptions)
	urlruntime.Must(alertingv2alpha1.AddToContainer(container, informerFactory, promfake.NewSimpleClientset(), alertingClient, alertingOptions))

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
			Name: "Authentication",
			Tags: []string{constants.AuthenticationTag},
		},
		{
			Name: "Identity Management",
			Tags: []string{
				constants.UserTag,
			},
		},
		{
			Name: "Access Management",
			Tags: []string{
				constants.ClusterMemberTag,
				constants.WorkspaceMemberTag,
				constants.DevOpsProjectMemberTag,
				constants.NamespaceMemberTag,
				constants.GlobalRoleTag,
				constants.ClusterRoleTag,
				constants.WorkspaceRoleTag,
				constants.DevOpsProjectRoleTag,
				constants.NamespaceRoleTag,
			},
		},
		{
			Name: "Multi-tenancy",
			Tags: []string{
				constants.WorkspaceTag,
				constants.NamespaceTag,
				constants.UserResourceTag,
			},
		},
		{
			Name: "Multi-cluster",
			Tags: []string{
				constants.MultiClusterTag,
			},
		},
		{
			Name: "Resources",
			Tags: []string{
				constants.ClusterResourcesTag,
				constants.NamespaceResourcesTag,
			},
		},
		{
			Name: "App Store",
			Tags: []string{
				constants.OpenpitrixAppInstanceTag,
				constants.OpenpitrixAppTemplateTag,
				constants.OpenpitrixCategoryTag,
				constants.OpenpitrixAttachmentTag,
				constants.OpenpitrixRepositoryTag,
				constants.OpenpitrixManagementTag,
			},
		},
		{
			Name: "Other",
			Tags: []string{
				constants.RegistryTag,
				constants.GitTag,
				constants.ToolboxTag,
				constants.TerminalTag,
			},
		},
		{
			Name: "DevOps",
			Tags: []string{
				constants.DevOpsProjectTag,
				constants.DevOpsCredentialTag,
				constants.DevOpsPipelineTag,
				constants.DevOpsProjectMemberTag,
				constants.DevOpsWebhookTag,
				constants.DevOpsJenkinsfileTag,
				constants.DevOpsScmTag,
				constants.DevOpsJenkinsTag,
			},
		},
		{
			Name: "Monitoring",
			Tags: []string{
				constants.ClusterMetricsTag,
				constants.NodeMetricsTag,
				constants.NamespaceMetricsTag,
				constants.WorkloadMetricsTag,
				constants.PodMetricsTag,
				constants.ContainerMetricsTag,
				constants.WorkspaceMetricsTag,
				constants.ComponentMetricsTag,
				constants.ComponentStatusTag,
			},
		},
		{
			Name: "Logging",
			Tags: []string{constants.LogQueryTag},
		},
		{
			Name: "Events",
			Tags: []string{constants.EventsQueryTag},
		},
		{
			Name: "Auditing",
			Tags: []string{constants.AuditingQueryTag},
		},
		{
			Name: "Network",
			Tags: []string{constants.NetworkTopologyTag},
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
			Version:     version.Get().GitVersion,
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "KubeSphere",
					URL:   "https://kubesphere.io/",
					Email: "kubesphere@yunify.com",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "Apache 2.0",
					URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
				},
			},
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
