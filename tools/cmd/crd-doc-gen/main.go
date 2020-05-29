/*
Copyright 2020 The KubeSphere Authors.

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
	"flag"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/meta"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"kubesphere.io/kubesphere/tools/lib"
	"log"
	"os"
	"path/filepath"

	"github.com/go-openapi/spec"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kube-openapi/pkg/common"
	devopsinstall "kubesphere.io/kubesphere/pkg/apis/devops/crdinstall"
	devopsv1alpha1 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	networkinstall "kubesphere.io/kubesphere/pkg/apis/network/crdinstall"
	networkv1alpha1 "kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	servicemeshinstall "kubesphere.io/kubesphere/pkg/apis/servicemesh/crdinstall"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
	tenantinstall "kubesphere.io/kubesphere/pkg/apis/tenant/crdinstall"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
)

var output string

func init() {
	flag.StringVar(&output, "output", "./api/openapi-spec/swagger.json", "--output=./api/openapi-spec/swagger.json")
}

func main() {

	var (
		Scheme = runtime.NewScheme()
		Codecs = serializer.NewCodecFactory(Scheme)
	)

	servicemeshinstall.Install(Scheme)
	tenantinstall.Install(Scheme)
	networkinstall.Install(Scheme)
	devopsinstall.Install(Scheme)

	urlruntime.Must(clusterv1alpha1.AddToScheme(Scheme))
	urlruntime.Must(Scheme.SetVersionPriority(clusterv1alpha1.SchemeGroupVersion))

	mapper := meta.NewDefaultRESTMapper(nil)

	mapper.AddSpecific(servicemeshv1alpha2.SchemeGroupVersion.WithKind(servicemeshv1alpha2.ResourceKindServicePolicy),
		servicemeshv1alpha2.SchemeGroupVersion.WithResource(servicemeshv1alpha2.ResourcePluralServicePolicy),
		servicemeshv1alpha2.SchemeGroupVersion.WithResource(servicemeshv1alpha2.ResourceSingularServicePolicy), meta.RESTScopeRoot)

	mapper.AddSpecific(servicemeshv1alpha2.SchemeGroupVersion.WithKind(servicemeshv1alpha2.ResourceKindStrategy),
		servicemeshv1alpha2.SchemeGroupVersion.WithResource(servicemeshv1alpha2.ResourcePluralStrategy),
		servicemeshv1alpha2.SchemeGroupVersion.WithResource(servicemeshv1alpha2.ResourceSingularStrategy), meta.RESTScopeRoot)

	mapper.AddSpecific(tenantv1alpha1.SchemeGroupVersion.WithKind(tenantv1alpha1.ResourceKindWorkspace),
		tenantv1alpha1.SchemeGroupVersion.WithResource(tenantv1alpha1.ResourcePluralWorkspace),
		tenantv1alpha1.SchemeGroupVersion.WithResource(tenantv1alpha1.ResourceSingularWorkspace), meta.RESTScopeRoot)

	mapper.AddSpecific(devopsv1alpha1.SchemeGroupVersion.WithKind(devopsv1alpha1.ResourceKindS2iBuilder),
		devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourcePluralS2iBuilder),
		devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourceSingularS2iBuilder), meta.RESTScopeRoot)

	mapper.AddSpecific(devopsv1alpha1.SchemeGroupVersion.WithKind(devopsv1alpha1.ResourceKindS2iBuilderTemplate),
		devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourcePluralS2iBuilderTemplate),
		devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourceSingularS2iBuilderTemplate), meta.RESTScopeRoot)

	mapper.AddSpecific(devopsv1alpha1.SchemeGroupVersion.WithKind(devopsv1alpha1.ResourceKindS2iRun),
		devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourcePluralS2iRun),
		devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourceSingularS2iRun), meta.RESTScopeRoot)
	mapper.AddSpecific(devopsv1alpha1.SchemeGroupVersion.WithKind(devopsv1alpha1.ResourceKindS2iBinary),
		devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourceSingularS2iBinary),
		devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourcePluralS2iBinary), meta.RESTScopeRoot)

	mapper.AddSpecific(networkv1alpha1.SchemeGroupVersion.WithKind(networkv1alpha1.ResourceKindNamespaceNetworkPolicy),
		networkv1alpha1.SchemeGroupVersion.WithResource(networkv1alpha1.ResourcePluralNamespaceNetworkPolicy),
		networkv1alpha1.SchemeGroupVersion.WithResource(networkv1alpha1.ResourceSingularNamespaceNetworkPolicy), meta.RESTScopeRoot)
	mapper.AddSpecific(devopsv1alpha3.SchemeGroupVersion.WithKind(devopsv1alpha3.ResourceKindDevOpsProject),
		devopsv1alpha3.SchemeGroupVersion.WithResource(devopsv1alpha3.ResourcePluralDevOpsProject),
		devopsv1alpha3.SchemeGroupVersion.WithResource(devopsv1alpha3.ResourceSingularDevOpsProject), meta.RESTScopeRoot)
	mapper.AddSpecific(devopsv1alpha3.SchemeGroupVersion.WithKind(devopsv1alpha3.ResourceKindPipeline),
		devopsv1alpha3.SchemeGroupVersion.WithResource(devopsv1alpha3.ResourcePluralPipeline),
		devopsv1alpha3.SchemeGroupVersion.WithResource(devopsv1alpha3.ResourceSingularPipeline), meta.RESTScopeRoot)

	mapper.AddSpecific(clusterv1alpha1.SchemeGroupVersion.WithKind(clusterv1alpha1.ResourceKindCluster),
		clusterv1alpha1.SchemeGroupVersion.WithResource(clusterv1alpha1.ResourcesPluralCluster),
		clusterv1alpha1.SchemeGroupVersion.WithResource(clusterv1alpha1.ResourcesSingularCluster), meta.RESTScopeRoot)

	mapper.AddSpecific(clusterv1alpha1.SchemeGroupVersion.WithKind(clusterv1alpha1.ResourceKindCluster),
		clusterv1alpha1.SchemeGroupVersion.WithResource(clusterv1alpha1.ResourcesPluralCluster),
		clusterv1alpha1.SchemeGroupVersion.WithResource(clusterv1alpha1.ResourcesSingularCluster), meta.RESTScopeRoot)

	spec, err := lib.RenderOpenAPISpec(lib.Config{
		Scheme: Scheme,
		Codecs: Codecs,
		Info: spec.InfoProps{
			Title:   "KubeSphere Advanced",
			Version: "v2.0.0",
			Contact: &spec.ContactInfo{
				Name:  "KubeSphere",
				URL:   "https://kubesphere.io/",
				Email: "kubesphere@yunify.com",
			},
			License: &spec.License{
				Name: "Apache 2.0",
				URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		OpenAPIDefinitions: []common.GetOpenAPIDefinitions{
			servicemeshv1alpha2.GetOpenAPIDefinitions,
			tenantv1alpha1.GetOpenAPIDefinitions,
			networkv1alpha1.GetOpenAPIDefinitions,
			devopsv1alpha1.GetOpenAPIDefinitions,
			devopsv1alpha3.GetOpenAPIDefinitions,
			clusterv1alpha1.GetOpenAPIDefinitions,
		},
		Resources: []schema.GroupVersionResource{
			//TODO（runzexia） At present, the document generation requires the openapi structure of the go language,
			// but there is no +k8s:openapi-gen=true in the repository of https://github.com/knative/pkg,
			// and the api document cannot be generated temporarily.
			//servicemeshv1alpha2.SchemeGroupVersion.WithResource(servicemeshv1alpha2.ResourcePluralStrategy),
			//servicemeshv1alpha2.SchemeGroupVersion.WithResource(servicemeshv1alpha2.ResourcePluralServicePolicy),
			tenantv1alpha1.SchemeGroupVersion.WithResource(tenantv1alpha1.ResourcePluralWorkspace),
			devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourcePluralS2iBinary),
			devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourcePluralS2iRun),
			devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourcePluralS2iBuilderTemplate),
			devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourcePluralS2iBuilder),
			networkv1alpha1.SchemeGroupVersion.WithResource(networkv1alpha1.ResourcePluralNamespaceNetworkPolicy),
			devopsv1alpha3.SchemeGroupVersion.WithResource(devopsv1alpha3.ResourcePluralDevOpsProject),
			devopsv1alpha3.SchemeGroupVersion.WithResource(devopsv1alpha3.ResourcePluralPipeline),
			clusterv1alpha1.SchemeGroupVersion.WithResource(clusterv1alpha1.ResourcesPluralCluster),
		},
		Mapper: mapper,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(filepath.Dir(output), 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(output, []byte(spec), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
