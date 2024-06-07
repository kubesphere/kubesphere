/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/pkg/errors"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	appv2 "kubesphere.io/kubesphere/pkg/kapis/application/v2"
	clusterkapisv1alpha1 "kubesphere.io/kubesphere/pkg/kapis/cluster/v1alpha1"
	configv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/config/v1alpha2"
	gatewayv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/gateway/v1alpha2"
	iamv1beta1 "kubesphere.io/kubesphere/pkg/kapis/iam/v1beta1"
	"kubesphere.io/kubesphere/pkg/kapis/oauth"
	operationsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/operations/v1alpha2"
	packagev1alpha1 "kubesphere.io/kubesphere/pkg/kapis/package/v1alpha1"
	resourcesv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha2"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/kapis/static"
	tenantv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/tenant/v1alpha3"
	tenantv1beta1 "kubesphere.io/kubesphere/pkg/kapis/tenant/v1beta1"
	terminalv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/terminal/v1alpha2"
	"kubesphere.io/kubesphere/pkg/kapis/version"
)

var output string

func init() {
	flag.StringVar(&output, "output", "./api/ks-openapi-spec/swagger.json", "--output=./api.json")
}

func main() {
	flag.Parse()
	if err := validateSpec(generateSwaggerJson()); err != nil {
		klog.Warningf("Swagger specification has errors")
	}
}

func validateSpec(apiSpec []byte) error {
	swaggerDoc, err := loads.Analyzed(apiSpec, "")
	if err != nil {
		return err
	}

	// Attempts to report about all errors
	validate.SetContinueOnErrors(false)

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

	handlers := []rest.Handler{
		version.NewFakeHandler(),
		oauth.FakeHandler(),
		clusterkapisv1alpha1.NewFakeHandler(),
		iamv1beta1.NewFakeHandler(),
		operationsv1alpha2.NewFakeHandler(),
		packagev1alpha1.NewFakeHandler(),
		gatewayv1alpha2.NewFakeHandler(),
		configv1alpha2.NewFakeHandler(),
		terminalv1alpha2.NewFakeHandler(),
		resourcesv1alpha2.NewFakeHandler(),
		resourcesv1alpha3.NewFakeHandler(),
		tenantv1beta1.NewFakeHandler(),
		tenantv1alpha3.NewFakeHandler(),
		appv2.NewFakeHandler(),
		static.NewFakeHandler(),
	}

	for _, handler := range handlers {
		urlruntime.Must(handler.AddToContainer(container))
	}

	config := restfulspec.Config{
		WebServices:                   container.RegisteredWebServices(),
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	}

	data, _ := json.MarshalIndent(restfulspec.BuildSwagger(config), "", "  ")
	if err := os.WriteFile(output, data, 0644); err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully written to %s", output)
	return data
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "KS API",
			Description: "KubeSphere OpenAPI",
			Version:     gitVersion(),
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "KubeSphere",
					URL:   "https://kubesphere.com.cn",
					Email: "support@kubesphere.cloud",
				},
			},
		},
	}
	// setup security definitions
	swo.SecurityDefinitions = map[string]*spec.SecurityScheme{
		"BearerToken": {SecuritySchemeProps: spec.SecuritySchemeProps{
			Type:        "apiKey",
			Name:        "Authorization",
			In:          "header",
			Description: "Bearer Token Authentication",
		}},
	}
	swo.Security = []map[string][]string{{"BearerToken": []string{}}}
	swo.Tags = []spec.Tag{

		{
			TagProps: spec.TagProps{
				Name: api.TagAuthentication,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagMultiCluster,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagIdentityManagement,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagAccessManagement,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagClusterResources,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagNamespacedResources,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagComponentStatus,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagUserRelatedResources,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagTerminal,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagNonResourceAPI,
			},
		},
	}
}

func gitVersion() string {
	out, err := exec.Command("sh", "-c", "git tag --sort=committerdate | tail -1 | tr -d '\n'").Output()
	if err != nil {
		log.Printf("failed to get git version: %s", err)
		return "v0.0.0"
	}
	return string(out)
}
