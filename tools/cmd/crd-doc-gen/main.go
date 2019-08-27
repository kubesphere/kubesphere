package main

import (
	"flag"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/meta"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
	"kubesphere.io/kubesphere/tools/lib"
	"log"
	"os"
	"path/filepath"

	"github.com/go-openapi/spec"
	s2iv1alpha1 "github.com/kubesphere/s2ioperator/pkg/apis/devops/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/kube-openapi/pkg/common"
	devopsinstall "kubesphere.io/kubesphere/pkg/apis/devops/crdinstall"
	devopsv1alpha1 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	networkinstall "kubesphere.io/kubesphere/pkg/apis/network/crdinstall"
	networkv1alpha1 "kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	servicemeshinstall "kubesphere.io/kubesphere/pkg/apis/servicemesh/crdinstall"
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
	// tmp install s2i api group because
	// cannot use Scheme (type *"kubesphere.io/kubesphere/vendor/k8s.io/apimachinery/pkg/runtime".Scheme) as
	// type *"github.com/kubesphere/s2ioperator/vendor/k8s.io/apimachinery/pkg/runtime".Scheme in argument to "github.com/kubesphere/s2ioperator/pkg/apis/devops/install".Install
	urlruntime.Must(s2iv1alpha1.AddToScheme(Scheme))
	urlruntime.Must(Scheme.SetVersionPriority(s2iv1alpha1.SchemeGroupVersion))

	servicemeshinstall.Install(Scheme)
	tenantinstall.Install(Scheme)
	networkinstall.Install(Scheme)
	devopsinstall.Install(Scheme)

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

	mapper.AddSpecific(s2iv1alpha1.SchemeGroupVersion.WithKind(s2iv1alpha1.ResourceKindS2iBuilder),
		s2iv1alpha1.SchemeGroupVersion.WithResource(s2iv1alpha1.ResourcePluralS2iBuilder),
		s2iv1alpha1.SchemeGroupVersion.WithResource(s2iv1alpha1.ResourceSingularS2iBuilder), meta.RESTScopeRoot)

	mapper.AddSpecific(s2iv1alpha1.SchemeGroupVersion.WithKind(s2iv1alpha1.ResourceKindS2iBuilderTemplate),
		s2iv1alpha1.SchemeGroupVersion.WithResource(s2iv1alpha1.ResourcePluralS2iBuilderTemplate),
		s2iv1alpha1.SchemeGroupVersion.WithResource(s2iv1alpha1.ResourceSingularS2iBuilderTemplate), meta.RESTScopeRoot)

	mapper.AddSpecific(s2iv1alpha1.SchemeGroupVersion.WithKind(s2iv1alpha1.ResourceKindS2iRun),
		s2iv1alpha1.SchemeGroupVersion.WithResource(s2iv1alpha1.ResourcePluralS2iRun),
		s2iv1alpha1.SchemeGroupVersion.WithResource(s2iv1alpha1.ResourceSingularS2iRun), meta.RESTScopeRoot)
	mapper.AddSpecific(devopsv1alpha1.SchemeGroupVersion.WithKind(devopsv1alpha1.ResourceKindS2iBinary),
		devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourceSingularServicePolicy),
		devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourcePluralServicePolicy), meta.RESTScopeRoot)

	mapper.AddSpecific(networkv1alpha1.SchemeGroupVersion.WithKind(networkv1alpha1.ResourceKindWorkspaceNetworkPolicy),
		networkv1alpha1.SchemeGroupVersion.WithResource(networkv1alpha1.ResourcePluralWorkspaceNetworkPolicy),
		networkv1alpha1.SchemeGroupVersion.WithResource(networkv1alpha1.ResourceSingularWorkspaceNetworkPolicy), meta.RESTScopeRoot)

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
			s2iv1alpha1.GetOpenAPIDefinitions,
			networkv1alpha1.GetOpenAPIDefinitions,
			devopsv1alpha1.GetOpenAPIDefinitions,
		},
		Resources: []schema.GroupVersionResource{
			//TODO（runzexia） At present, the document generation requires the openapi structure of the go language,
			// but there is no +k8s:openapi-gen=true in the repository of https://github.com/knative/pkg,
			// and the api document cannot be generated temporarily.
			//servicemeshv1alpha2.SchemeGroupVersion.WithResource(servicemeshv1alpha2.ResourcePluralStrategy),
			//servicemeshv1alpha2.SchemeGroupVersion.WithResource(servicemeshv1alpha2.ResourcePluralServicePolicy),
			tenantv1alpha1.SchemeGroupVersion.WithResource(tenantv1alpha1.ResourcePluralWorkspace),
			s2iv1alpha1.SchemeGroupVersion.WithResource(s2iv1alpha1.ResourcePluralS2iRun),
			s2iv1alpha1.SchemeGroupVersion.WithResource(s2iv1alpha1.ResourcePluralS2iBuilderTemplate),
			s2iv1alpha1.SchemeGroupVersion.WithResource(s2iv1alpha1.ResourcePluralS2iBuilder),
			networkv1alpha1.SchemeGroupVersion.WithResource(networkv1alpha1.ResourcePluralWorkspaceNetworkPolicy),
			devopsv1alpha1.SchemeGroupVersion.WithResource(devopsv1alpha1.ResourceKindS2iBinary),
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
