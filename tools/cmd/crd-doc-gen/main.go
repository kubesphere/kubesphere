package main

import (
	"go/build"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/meta"
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
	tenantinstall "kubesphere.io/kubesphere/pkg/apis/tenant/crdinstall"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
)

func main() {

	var (
		Scheme = runtime.NewScheme()
		Codecs = serializer.NewCodecFactory(Scheme)
	)
	// tmp install s2i api group because
	// cannot use Scheme (type *"kubesphere.io/kubesphere/vendor/k8s.io/apimachinery/pkg/runtime".Scheme) as type *"github.com/kubesphere/s2ioperator/vendor/k8s.io/apimachinery/pkg/runtime".Scheme in argument to "github.com/kubesphere/s2ioperator/pkg/apis/devops/install".Install
	urlruntime.Must(s2iv1alpha1.AddToScheme(Scheme))
	urlruntime.Must(Scheme.SetVersionPriority(s2iv1alpha1.SchemeGroupVersion))

	tenantinstall.Install(Scheme)

	mapper := meta.NewDefaultRESTMapper(nil)

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
			tenantv1alpha1.GetOpenAPIDefinitions,
			s2iv1alpha1.GetOpenAPIDefinitions,
		},
		Resources: []schema.GroupVersionResource{
			tenantv1alpha1.SchemeGroupVersion.WithResource(tenantv1alpha1.ResourcePluralWorkspace),
			s2iv1alpha1.SchemeGroupVersion.WithResource(s2iv1alpha1.ResourcePluralS2iRun),
			s2iv1alpha1.SchemeGroupVersion.WithResource(s2iv1alpha1.ResourcePluralS2iBuilderTemplate),
			s2iv1alpha1.SchemeGroupVersion.WithResource(s2iv1alpha1.ResourcePluralS2iBuilder),
		},
		Mapper: mapper,
	})
	if err != nil {
		log.Fatal(err)
	}

	filename := build.Default.GOPATH + "/src/kubesphere.io/kubesphere/api/openapi-spec/swagger.json"
	err = os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(filename, []byte(spec), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
