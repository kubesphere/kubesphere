package v1alpha1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/spec"
	promresourcesclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus/prometheus/promql/parser"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/customalerting"
)

func TestTest(t *testing.T) {
	expr := `increase((max by(job) (etcd_server_leader_changes_seen_total{job=~".*etcd.*"}) or 0 * absent(etcd_server_leader_changes_seen_total{job=~".*etcd.*"}))[15m:1m]) >= 3`
	//expr = `absent_over_time(sum(nonexistent{job="myjob"})[1h:])`

	pexpr, err := parser.ParseExpr(expr)
	if err != nil {
		panic(err)
	}
	fmt.Println(pexpr.String())
}

func TestApis(t *testing.T) {
	kubeconfig := "D:/ks/conf/ks3-config"

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	k8sClient := kubernetes.NewForConfigOrDie(config)
	promResourcesClient := promresourcesclient.NewForConfigOrDie(config)

	option := &customalerting.Options{
		PrometheusEndpoint: "http://139.198.112.79:39090/",
		ThanosRulerEndpoint: "http://139.198.112.79:39091/",
		ThanosRuleResourceLabels: "role=thanos-alerting-rules,thanosruler:thanos-ruler",
	}
	ruleClient, err := customalerting.NewRuleClient(option)
	if err != nil {
		panic(err)
	}

	stopCh := make(chan struct{})

	informerFactory := informers.NewInformerFactories(k8sClient, nil, nil, nil, nil, promResourcesClient)
	k8sGVRs := []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "namespaces"},
	}
	for _, gvr := range k8sGVRs {
		_, err = informerFactory.KubernetesSharedInformerFactory().ForResource(gvr)
		if err != nil {
			panic(err)
		}
	}
	prometheusGVRs := []schema.GroupVersionResource{
		{Group: "monitoring.coreos.com", Version: "v1", Resource: "prometheuses"},
		{Group: "monitoring.coreos.com", Version: "v1", Resource: "prometheusrules"},
		{Group: "monitoring.coreos.com", Version: "v1", Resource: "thanosrulers"},
	}
	for _, gvr := range prometheusGVRs {
		_, err = informerFactory.PrometheusSharedInformerFactory().ForResource(gvr)
		if err != nil {
			panic(err)
		}
	}
	informerFactory.Start(stopCh)
	informerFactory.KubernetesSharedInformerFactory().WaitForCacheSync(stopCh)
	informerFactory.PrometheusSharedInformerFactory().WaitForCacheSync(stopCh)
	informerFactory.Start(stopCh)

	container := restful.NewContainer()
	AddToContainer(container, informerFactory, promResourcesClient, ruleClient, option)
	server := &http.Server{}
	server.Handler = container
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func TestGenSwaggerJson(t *testing.T) {
	container := runtime.Container

	informerFactory := informers.NewNullInformerFactory()

	AddToContainer(container, informerFactory, nil, nil, nil)

	swagger := restfulspec.BuildSwagger(restfulspec.Config{
		WebServices:                   container.RegisteredWebServices(),
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	})

	swagger.Info.Extensions = make(spec.Extensions)
	swagger.Info.Extensions.Add("x-tagGroups", []struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}{
		{
			Name: "Custom Alerting",
			Tags: []string{constants.CustomAlertingTag},
		},
	})

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	enc.Encode(swagger)
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "KubeSphere",
			Description: "KubeSphere OpenAPI",
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
			Version: "0.1.0",
		}}

	// setup security definitions
	swo.SecurityDefinitions = map[string]*spec.SecurityScheme{
		"jwt": spec.APIKeyAuth("Authorization", "header"),
	}
	swo.Security = []map[string][]string{{"jwt": []string{}}}
}
