package apiserver

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/informers"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/server/options"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/db"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
)

const (
	// ApiRootPath defines the root path of all KubeSphere apis.
	ApiRootPath = "/kapis"

	// MimeMergePatchJson is the mime header used in merge request
	MimeMergePatchJson = "application/merge-patch+json"

	//
	MimeJsonPatchJson = "application/json-patch+json"
)

// Dependencies is objects constructed at runtime that are necessary for running apiserver.
type Dependencies struct {

	// Injected Dependencies
	KubeClient k8s.Client
	S3         s3.Interface
	OpenPitrix openpitrix.Client
	Monitoring monitoring.Interface
	Logging    logging.Interface
	Devops     devops.Interface
	DB         db.Interface
}

type APIServer struct {

	// number of kubesphere apiserver
	apiserverCount int

	//
	genericServerOptions *options.ServerRunOptions

	// webservice container, where all webservice defines
	container *restful.Container

	// kubeClient is a collection of all kubernetes(include CRDs) objects clientset
	kubeClient k8s.Client

	// informerFactory is a collection of all kubernetes(include CRDs) objects informers,
	// mainly for fast query
	informerFactory informers.InformerFactory

	// cache is used for short lived objects, like session
	cache cache.Interface

	//

}

func New(deps *Dependencies) *APIServer {

	server := &APIServer{}

	return server
}

func (s *APIServer) InstallKubeSphereAPIs() {

	resourcev1alpha3.AddWebService(s.container, s.kubeClient)
}

func (s *APIServer) Serve() error {
	panic("implement me")
}
