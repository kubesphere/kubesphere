/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package apiserver

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	rt "runtime"

	"github.com/Masterminds/semver/v3"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	k8sversion "k8s.io/apimachinery/pkg/version"
	unionauth "k8s.io/apiserver/pkg/authentication/request/union"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/kube/pkg/openapi"
	openapiv2 "kubesphere.io/kubesphere/kube/pkg/openapi/v2"
	openapiv3 "kubesphere.io/kubesphere/kube/pkg/openapi/v3"
	"kubesphere.io/kubesphere/pkg/apiserver/auditing"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/authenticators/basic"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/authenticators/jwt"
	oauth2 "kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/request/anonymous"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/request/basictoken"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/request/bearertoken"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizerfactory"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/path"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/rbac"
	unionauthorizer "kubesphere.io/kubesphere/pkg/apiserver/authorization/union"
	"kubesphere.io/kubesphere/pkg/apiserver/filters"
	"kubesphere.io/kubesphere/pkg/apiserver/metrics"
	"kubesphere.io/kubesphere/pkg/apiserver/options"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	openapicontroller "kubesphere.io/kubesphere/pkg/controller/openapi"
	appv2 "kubesphere.io/kubesphere/pkg/kapis/application/v2"
	clusterkapisv1alpha1 "kubesphere.io/kubesphere/pkg/kapis/cluster/v1alpha1"
	configv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/config/v1alpha2"
	gatewayv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/gateway/v1alpha2"
	iamapiv1beta1 "kubesphere.io/kubesphere/pkg/kapis/iam/v1beta1"
	"kubesphere.io/kubesphere/pkg/kapis/oauth"
	operationsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/operations/v1alpha2"
	packagev1alpha1 "kubesphere.io/kubesphere/pkg/kapis/package/v1alpha1"
	resourcesv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha2"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/kapis/static"
	tenantapiv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/tenant/v1alpha3"
	tenantapiv1beta1 "kubesphere.io/kubesphere/pkg/kapis/tenant/v1beta1"
	terminalv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/terminal/v1alpha2"
	"kubesphere.io/kubesphere/pkg/kapis/version"
	"kubesphere.io/kubesphere/pkg/models/auth"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	resourcev1beta1 "kubesphere.io/kubesphere/pkg/models/resources/v1beta1"
	"kubesphere.io/kubesphere/pkg/server/healthz"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	overviewclient "kubesphere.io/kubesphere/pkg/simple/client/overview"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
)

type APIServer struct {
	Server *http.Server

	options.Options

	// webservice container, where all webservice defines
	container *restful.Container

	// K8sClient is a collection of all kubernetes(include CRDs) objects clientset
	K8sClient k8s.Client

	// cache is used for short-lived objects, like session
	CacheClient cache.Interface

	// controller-runtime cache
	RuntimeCache runtimecache.Cache

	TokenOperator auth.TokenManagementInterface

	// controller-runtime client with informer cache
	RuntimeClient runtimeclient.Client

	ClusterClient clusterclient.Interface

	ResourceManager resourcev1beta1.ResourceManager

	K8sVersionInfo *k8sversion.Info
	K8sVersion     *semver.Version

	OpenAPIConfig    *restfulspec.Config
	openAPIV2Service openapi.APIServiceManager
	openAPIV3Service openapi.APIServiceManager
}

func (s *APIServer) PrepareRun(stopCh <-chan struct{}) error {
	s.container = restful.NewContainer()
	s.container.Router(restful.CurlyRouter{})
	s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})
	s.installDynamicResourceAPI()
	s.installKubeSphereAPIs()
	s.installMetricsAPI()
	s.installHealthz()
	if err := s.installOpenAPI(); err != nil {
		return err
	}

	for _, ws := range s.container.RegisteredWebServices() {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	combinedHandler, err := s.buildHandlerChain(s.container, stopCh)
	if err != nil {
		return fmt.Errorf("failed to build handler chain: %v", err)
	}
	s.Server.Handler = filters.WithGlobalFilter(combinedHandler)
	return nil
}

func (s *APIServer) installOpenAPI() error {
	s.OpenAPIConfig = &restfulspec.Config{
		WebServices:                   s.container.RegisteredWebServices(),
		PostBuildSwaggerObjectHandler: openapicontroller.EnrichSwaggerObject,
	}

	openapiV2Services, err := openapiv2.BuildAndRegisterAggregator(s.OpenAPIConfig, s.container)
	if err != nil {
		klog.Errorf("failed to install openapi v2 service : %s", err)
	}
	s.openAPIV2Service = openapiV2Services
	openapiV3Services, err := openapiv3.BuildAndRegisterAggregator(s.OpenAPIConfig, s.container)
	if err != nil {
		klog.Errorf("failed to install openapi v3 service : %s", err)
	}
	s.openAPIV3Service = openapiV3Services
	return openapicontroller.SharedOpenAPIController.WatchOpenAPIChanges(context.Background(), s.RuntimeCache, s.openAPIV2Service, s.openAPIV3Service)
}

func (s *APIServer) installMetricsAPI() {
	metrics.Install(s.container)
}

// Install all kubesphere api groups
// Installations happens before all informers start to cache objects,
// so any attempt to list objects using listers will get empty results.
func (s *APIServer) installKubeSphereAPIs() {
	imOperator := im.NewOperator(s.RuntimeClient, s.ResourceManager, s.AuthenticationOptions)
	amOperator := am.NewOperator(s.ResourceManager)
	rbacAuthorizer := rbac.NewRBACAuthorizer(amOperator)
	counter := overviewclient.New(s.RuntimeClient)
	counter.RegisterResource(overviewclient.NewDefaultRegisterOptions(s.K8sVersion)...)

	handlers := []rest.Handler{
		configv1alpha2.NewHandler(&s.Options, s.RuntimeClient),
		resourcev1alpha3.NewHandler(s.RuntimeCache, counter, s.K8sVersion),
		operationsv1alpha2.NewHandler(s.RuntimeClient),
		resourcesv1alpha2.NewHandler(s.RuntimeClient, s.K8sVersion, s.K8sClient.Master(), s.TerminalOptions),
		tenantapiv1alpha3.NewHandler(s.RuntimeClient, s.K8sVersion, s.ClusterClient, amOperator, imOperator, rbacAuthorizer),
		tenantapiv1beta1.NewHandler(s.RuntimeClient, s.K8sVersion, s.ClusterClient, amOperator, imOperator, rbacAuthorizer, counter),
		terminalv1alpha2.NewHandler(s.K8sClient, rbacAuthorizer, s.K8sClient.Config(), s.TerminalOptions),
		clusterkapisv1alpha1.NewHandler(s.RuntimeClient),
		iamapiv1beta1.NewHandler(imOperator, amOperator),
		oauth.NewHandler(imOperator, s.TokenOperator, auth.NewPasswordAuthenticator(s.RuntimeClient, s.AuthenticationOptions),
			auth.NewOAuthAuthenticator(s.RuntimeClient),
			auth.NewLoginRecorder(s.RuntimeClient), s.AuthenticationOptions,
			oauth2.NewOAuthClientGetter(s.RuntimeClient)),
		version.NewHandler(s.K8sVersionInfo),
		packagev1alpha1.NewHandler(s.RuntimeCache),
		gatewayv1alpha2.NewHandler(s.RuntimeCache),
		appv2.NewHandler(s.RuntimeClient, s.ClusterClient, s.S3Options),
		static.NewHandler(s.CacheClient),
	}

	for _, handler := range handlers {
		urlruntime.Must(handler.AddToContainer(s.container))
	}
}

// installHealthz creates the healthz endpoint for this server
func (s *APIServer) installHealthz() {
	urlruntime.Must(healthz.InstallHandler(s.container, []healthz.HealthChecker{}...))
}

func (s *APIServer) Run(ctx context.Context) (err error) {
	go func() {
		if err := s.RuntimeCache.Start(ctx); err != nil {
			klog.Errorf("failed to start runtime cache: %s", err)
		}
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		if err := s.Server.Shutdown(ctx); err != nil {
			klog.Errorf("failed to shutdown server: %s", err)
		}
	}()

	klog.V(0).Infof("Start listening on %s", s.Server.Addr)
	if s.Server.TLSConfig != nil {
		err = s.Server.ListenAndServeTLS("", "")
	} else {
		err = s.Server.ListenAndServe()
	}
	return err
}

func (s *APIServer) buildHandlerChain(handler http.Handler, stopCh <-chan struct{}) (http.Handler, error) {
	requestInfoResolver := &request.RequestInfoFactory{
		APIPrefixes:          sets.New("api", "apis", "kapis", "kapi"),
		GrouplessAPIPrefixes: sets.New("api", "kapi"),
		GlobalResources: []schema.GroupResource{
			iamv1beta1.Resource(iamv1beta1.ResourcesPluralUser),
			iamv1beta1.Resource(iamv1beta1.ResourcesPluralGlobalRole),
			iamv1beta1.Resource(iamv1beta1.ResourcesPluralGlobalRoleBinding),
			tenantv1beta1.Resource(tenantv1beta1.ResourcePluralWorkspace),
			tenantv1beta1.Resource(tenantv1beta1.ResourcePluralWorkspace),
			tenantv1beta1.Resource(clusterv1alpha1.ResourcesPluralCluster),
			clusterv1alpha1.Resource(clusterv1alpha1.ResourcesPluralCluster),
			clusterv1alpha1.Resource(clusterv1alpha1.ResourcesPluralLabel),
			resourcev1alpha3.Resource(clusterv1alpha1.ResourcesPluralCluster),
			resourcev1alpha3.Resource(clusterv1alpha1.ResourcesPluralLabel),
		},
	}

	handler = filters.WithKubeAPIServer(handler, s.K8sClient.Config(), s.ExperimentalOptions)
	handler = filters.WithAPIService(handler, s.RuntimeCache)
	handler = filters.WithReverseProxy(handler, s.RuntimeCache)
	handler = filters.WithJSBundle(handler, s.RuntimeCache)

	if s.AuditingOptions.Enable {
		handler = filters.WithAuditing(handler, auditing.NewAuditing(s.K8sClient, s.AuditingOptions, stopCh))
	}

	var authorizers authorizer.Authorizer
	switch s.AuthorizationOptions.Mode {
	case authorization.AlwaysAllow:
		authorizers = authorizerfactory.NewAlwaysAllowAuthorizer()
	case authorization.AlwaysDeny:
		authorizers = authorizerfactory.NewAlwaysDenyAuthorizer()
	default:
		fallthrough
	case authorization.RBAC:
		excludedPaths := []string{"/oauth/*", "/dist/*", "/.well-known/openid-configuration", "/kapis/version", "/version", "/metrics", "/healthz", "/openapi/v2", "/openapi/v3"}
		pathAuthorizer, _ := path.NewAuthorizer(excludedPaths)
		amOperator := am.NewReadOnlyOperator(s.ResourceManager)
		authorizers = unionauthorizer.New(pathAuthorizer, rbac.NewRBACAuthorizer(amOperator))
	}

	handler = filters.WithAuthorization(handler, authorizers)
	handler = filters.WithMulticluster(handler, s.ClusterClient, s.MultiClusterOptions)

	// authenticators are unordered
	authn := unionauth.New(anonymous.NewAuthenticator(),
		basictoken.New(basic.NewBasicAuthenticator(
			auth.NewPasswordAuthenticator(s.RuntimeClient, s.AuthenticationOptions),
			auth.NewLoginRecorder(s.RuntimeClient))),
		bearertoken.New(jwt.NewTokenAuthenticator(s.RuntimeCache, s.TokenOperator, s.MultiClusterOptions.ClusterRole)))

	handler = filters.WithAuthentication(handler, authn)
	handler = filters.WithRequestInfo(handler, requestInfoResolver)
	return handler, nil
}

func (s *APIServer) installDynamicResourceAPI() {
	dynamicResourceHandler := filters.NewDynamicResourceHandle(func(err restful.ServiceError, req *restful.Request, resp *restful.Response) {
		for header, values := range err.Header {
			for _, value := range values {
				resp.Header().Add(header, value)
			}
		}
		if err := resp.WriteErrorString(err.Code, err.Message); err != nil {
			klog.Errorf("failed to write error string: %s", err)
		}
	}, s.ResourceManager)
	s.container.ServiceErrorHandler(dynamicResourceHandler.HandleServiceError)
}

func logStackOnRecover(panicReason interface{}, w http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", panicReason))
	for i := 2; ; i += 1 {
		_, file, line, ok := rt.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	klog.Errorln(buffer.String())

	headers := http.Header{}
	if ct := w.Header().Get("Content-Type"); len(ct) > 0 {
		headers.Set("Accept", ct)
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
