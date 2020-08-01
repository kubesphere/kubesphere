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

package apiserver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	unionauth "k8s.io/apiserver/pkg/authentication/request/union"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"
	clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	typesv1beta1 "kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	audit "kubesphere.io/kubesphere/pkg/apiserver/auditing"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/authenticators/basic"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/authenticators/jwttoken"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/request/anonymous"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/request/basictoken"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/request/bearertoken"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizerfactory"
	authorizationoptions "kubesphere.io/kubesphere/pkg/apiserver/authorization/options"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/path"
	unionauthorizer "kubesphere.io/kubesphere/pkg/apiserver/authorization/union"
	apiserverconfig "kubesphere.io/kubesphere/pkg/apiserver/config"
	"kubesphere.io/kubesphere/pkg/apiserver/dispatch"
	"kubesphere.io/kubesphere/pkg/apiserver/filters"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/informers"
	alertingv1 "kubesphere.io/kubesphere/pkg/kapis/alerting/v1"
	clusterkapisv1alpha1 "kubesphere.io/kubesphere/pkg/kapis/cluster/v1alpha1"
	configv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/config/v1alpha2"
	devopsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/devops/v1alpha2"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/devops/v1alpha3"
	iamapi "kubesphere.io/kubesphere/pkg/kapis/iam/v1alpha2"
	monitoringv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/monitoring/v1alpha3"
	notificationv1 "kubesphere.io/kubesphere/pkg/kapis/notification/v1"
	"kubesphere.io/kubesphere/pkg/kapis/oauth"
	openpitrixv1 "kubesphere.io/kubesphere/pkg/kapis/openpitrix/v1"
	operationsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/operations/v1alpha2"
	resourcesv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha2"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha3"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/servicemesh/metrics/v1alpha2"
	tenantv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/tenant/v1alpha2"
	terminalv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/terminal/v1alpha2"
	"kubesphere.io/kubesphere/pkg/kapis/version"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"kubesphere.io/kubesphere/pkg/simple/client/auditing"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
	utilnet "kubesphere.io/kubesphere/pkg/utils/net"
	"net/http"
	rt "runtime"
	"time"
)

const (
	// ApiRootPath defines the root path of all KubeSphere apis.
	ApiRootPath = "/kapis"

	// MimeMergePatchJson is the mime header used in merge request
	MimeMergePatchJson = "application/merge-patch+json"

	//
	MimeJsonPatchJson = "application/json-patch+json"
)

type APIServer struct {

	// number of kubesphere apiserver
	ServerCount int

	//
	Server *http.Server

	Config *apiserverconfig.Config

	// webservice container, where all webservice defines
	container *restful.Container

	// kubeClient is a collection of all kubernetes(include CRDs) objects clientset
	KubernetesClient k8s.Client

	// informerFactory is a collection of all kubernetes(include CRDs) objects informers,
	// mainly for fast query
	InformerFactory informers.InformerFactory

	// cache is used for short lived objects, like session
	CacheClient cache.Interface

	// monitoring client set
	MonitoringClient monitoring.Interface

	//
	OpenpitrixClient openpitrix.Client

	//
	LoggingClient logging.Interface

	//
	DevopsClient devops.Interface

	//
	S3Client s3.Interface

	SonarClient sonarqube.SonarInterface

	EventsClient events.Client

	AuditingClient auditing.Client
}

func (s *APIServer) PrepareRun(stopCh <-chan struct{}) error {

	s.container = restful.NewContainer()
	s.container.Filter(logRequestAndResponse)
	s.container.Router(restful.CurlyRouter{})
	s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})

	s.installKubeSphereAPIs()

	for _, ws := range s.container.RegisteredWebServices() {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	s.Server.Handler = s.container

	s.buildHandlerChain(stopCh)

	return nil
}

// Install all kubesphere api groups
// Installation happens before all informers start to cache objects, so
//   any attempt to list objects using listers will get empty results.
func (s *APIServer) installKubeSphereAPIs() {
	urlruntime.Must(configv1alpha2.AddToContainer(s.container, s.Config))
	urlruntime.Must(resourcev1alpha3.AddToContainer(s.container, s.InformerFactory))
	urlruntime.Must(monitoringv1alpha3.AddToContainer(s.container, s.KubernetesClient.Kubernetes(), s.MonitoringClient, s.InformerFactory, s.OpenpitrixClient))
	urlruntime.Must(openpitrixv1.AddToContainer(s.container, s.InformerFactory, s.OpenpitrixClient))
	urlruntime.Must(operationsv1alpha2.AddToContainer(s.container, s.KubernetesClient.Kubernetes()))
	urlruntime.Must(resourcesv1alpha2.AddToContainer(s.container, s.KubernetesClient.Kubernetes(), s.InformerFactory,
		s.KubernetesClient.Master()))
	urlruntime.Must(tenantv1alpha2.AddToContainer(s.container, s.InformerFactory, s.KubernetesClient.Kubernetes(),
		s.KubernetesClient.KubeSphere(), s.EventsClient, s.LoggingClient, s.AuditingClient))
	urlruntime.Must(terminalv1alpha2.AddToContainer(s.container, s.KubernetesClient.Kubernetes(), s.KubernetesClient.Config()))
	urlruntime.Must(clusterkapisv1alpha1.AddToContainer(s.container,
		s.InformerFactory.KubernetesSharedInformerFactory(),
		s.InformerFactory.KubeSphereSharedInformerFactory(),
		s.Config.MultiClusterOptions.ProxyPublishService,
		s.Config.MultiClusterOptions.ProxyPublishAddress,
		s.Config.MultiClusterOptions.AgentImage))
	imOperator := im.NewOperator(s.KubernetesClient.KubeSphere(), s.InformerFactory, s.Config.AuthenticationOptions)
	urlruntime.Must(iamapi.AddToContainer(s.container, imOperator,
		am.NewOperator(s.InformerFactory, s.KubernetesClient.KubeSphere(), s.KubernetesClient.Kubernetes()),
		s.Config.AuthenticationOptions))

	urlruntime.Must(oauth.AddToContainer(s.container, imOperator,
		im.NewTokenOperator(
			s.CacheClient,
			s.Config.AuthenticationOptions),
		im.NewPasswordAuthenticator(
			s.KubernetesClient.KubeSphere(),
			s.InformerFactory.KubeSphereSharedInformerFactory().Iam().V1alpha2().Users().Lister(),
			s.Config.AuthenticationOptions),
		im.NewLoginRecorder(s.KubernetesClient.KubeSphere()),
		s.Config.AuthenticationOptions))
	urlruntime.Must(servicemeshv1alpha2.AddToContainer(s.container))
	urlruntime.Must(devopsv1alpha2.AddToContainer(s.container,
		s.InformerFactory.KubeSphereSharedInformerFactory(),
		s.DevopsClient,
		s.SonarClient,
		s.KubernetesClient.KubeSphere(),
		s.S3Client,
		s.Config.DevopsOptions.Host))
	urlruntime.Must(devopsv1alpha3.AddToContainer(s.container,
		s.DevopsClient,
		s.KubernetesClient.Kubernetes(),
		s.KubernetesClient.KubeSphere(),
		s.InformerFactory.KubeSphereSharedInformerFactory(),
		s.InformerFactory.KubernetesSharedInformerFactory()))
	urlruntime.Must(notificationv1.AddToContainer(s.container, s.Config.NotificationOptions.Endpoint))
	urlruntime.Must(alertingv1.AddToContainer(s.container, s.Config.AlertingOptions.Endpoint))
	urlruntime.Must(version.AddToContainer(s.container, s.KubernetesClient.Discovery()))
}

func (s *APIServer) Run(stopCh <-chan struct{}) (err error) {

	err = s.waitForResourceSync(stopCh)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-stopCh
		_ = s.Server.Shutdown(ctx)
	}()

	klog.V(0).Infof("Start listening on %s", s.Server.Addr)
	if s.Server.TLSConfig != nil {
		err = s.Server.ListenAndServeTLS("", "")
	} else {
		err = s.Server.ListenAndServe()
	}

	return err
}

func (s *APIServer) buildHandlerChain(stopCh <-chan struct{}) {
	requestInfoResolver := &request.RequestInfoFactory{
		APIPrefixes:          sets.NewString("api", "apis", "kapis", "kapi"),
		GrouplessAPIPrefixes: sets.NewString("api", "kapi"),
		GlobalResources: []schema.GroupResource{
			iamv1alpha2.Resource(iamv1alpha2.ResourcesPluralUser),
			iamv1alpha2.Resource(iamv1alpha2.ResourcesPluralGlobalRole),
			iamv1alpha2.Resource(iamv1alpha2.ResourcesPluralGlobalRoleBinding),
			tenantv1alpha1.Resource(tenantv1alpha1.ResourcePluralWorkspace),
			tenantv1alpha2.Resource(tenantv1alpha1.ResourcePluralWorkspace),
			tenantv1alpha2.Resource(clusterv1alpha1.ResourcesPluralCluster),
			clusterv1alpha1.Resource(clusterv1alpha1.ResourcesPluralCluster),
			resourcev1alpha3.Resource(clusterv1alpha1.ResourcesPluralCluster),
		},
	}

	handler := s.Server.Handler
	handler = filters.WithKubeAPIServer(handler, s.KubernetesClient.Config(), &errorResponder{})

	if s.Config.AuditingOptions.Enable {
		handler = filters.WithAuditing(handler,
			audit.NewAuditing(s.InformerFactory, s.Config.AuditingOptions.WebhookUrl, stopCh))
	}

	var authorizers authorizer.Authorizer

	switch s.Config.AuthorizationOptions.Mode {
	case authorizationoptions.AlwaysAllow:
		authorizers = authorizerfactory.NewAlwaysAllowAuthorizer()
	case authorizationoptions.AlwaysDeny:
		authorizers = authorizerfactory.NewAlwaysDenyAuthorizer()
	default:
		fallthrough
	case authorizationoptions.RBAC:
		excludedPaths := []string{"/oauth/*", "/kapis/config.kubesphere.io/*", "/kapis/version"}
		pathAuthorizer, _ := path.NewAuthorizer(excludedPaths)
		amOperator := am.NewReadOnlyOperator(s.InformerFactory)
		authorizers = unionauthorizer.New(pathAuthorizer, authorizerfactory.NewRBACAuthorizer(amOperator))
	}

	handler = filters.WithAuthorization(handler, authorizers)
	if s.Config.MultiClusterOptions.Enable {
		clusterDispatcher := dispatch.NewClusterDispatch(s.InformerFactory.KubeSphereSharedInformerFactory().Cluster().V1alpha1().Clusters(),
			s.InformerFactory.KubeSphereSharedInformerFactory().Cluster().V1alpha1().Clusters().Lister())
		handler = filters.WithMultipleClusterDispatcher(handler, clusterDispatcher)
	}

	loginRecorder := im.NewLoginRecorder(s.KubernetesClient.KubeSphere())
	// authenticators are unordered
	authn := unionauth.New(anonymous.NewAuthenticator(),
		basictoken.New(basic.NewBasicAuthenticator(im.NewPasswordAuthenticator(s.KubernetesClient.KubeSphere(), s.InformerFactory.KubeSphereSharedInformerFactory().Iam().V1alpha2().Users().Lister(), s.Config.AuthenticationOptions))),
		bearertoken.New(jwttoken.NewTokenAuthenticator(im.NewTokenOperator(s.CacheClient, s.Config.AuthenticationOptions))))
	handler = filters.WithAuthentication(handler, authn, loginRecorder)
	handler = filters.WithRequestInfo(handler, requestInfoResolver)
	s.Server.Handler = handler
}

func (s *APIServer) waitForResourceSync(stopCh <-chan struct{}) error {
	klog.V(0).Info("Start cache objects")

	discoveryClient := s.KubernetesClient.Kubernetes().Discovery()
	_, apiResourcesList, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return err
	}

	isResourceExists := func(resource schema.GroupVersionResource) bool {
		for _, apiResource := range apiResourcesList {
			if apiResource.GroupVersion == resource.GroupVersion().String() {
				for _, rsc := range apiResource.APIResources {
					if rsc.Name == resource.Resource {
						return true
					}
				}
			}
		}
		return false
	}

	// resources we have to create informer first
	k8sGVRs := []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "namespaces"},
		{Group: "", Version: "v1", Resource: "nodes"},
		{Group: "", Version: "v1", Resource: "resourcequotas"},
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		{Group: "", Version: "v1", Resource: "secrets"},
		{Group: "", Version: "v1", Resource: "configmaps"},

		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},

		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps", Version: "v1", Resource: "daemonsets"},
		{Group: "apps", Version: "v1", Resource: "replicasets"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "apps", Version: "v1", Resource: "controllerrevisions"},

		{Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},

		{Group: "batch", Version: "v1", Resource: "jobs"},
		{Group: "batch", Version: "v1beta1", Resource: "cronjobs"},

		{Group: "extensions", Version: "v1beta1", Resource: "ingresses"},

		{Group: "autoscaling", Version: "v2beta2", Resource: "horizontalpodautoscalers"},

		{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
	}

	for _, gvr := range k8sGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err := s.InformerFactory.KubernetesSharedInformerFactory().ForResource(gvr)
			if err != nil {
				klog.Errorf("cannot create informer for %s", gvr)
				return err
			}
		}
	}

	s.InformerFactory.KubernetesSharedInformerFactory().Start(stopCh)
	s.InformerFactory.KubernetesSharedInformerFactory().WaitForCacheSync(stopCh)

	ksInformerFactory := s.InformerFactory.KubeSphereSharedInformerFactory()

	ksGVRs := []schema.GroupVersionResource{
		{Group: "tenant.kubesphere.io", Version: "v1alpha1", Resource: "workspaces"},
		{Group: "tenant.kubesphere.io", Version: "v1alpha2", Resource: "workspacetemplates"},
		{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "users"},
		{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "globalroles"},
		{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "globalrolebindings"},
		{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "workspaceroles"},
		{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "workspacerolebindings"},
		{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "loginrecords"},
		{Group: "cluster.kubesphere.io", Version: "v1alpha1", Resource: "clusters"},
		{Group: "devops.kubesphere.io", Version: "v1alpha3", Resource: "devopsprojects"},
	}

	devopsGVRs := []schema.GroupVersionResource{
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibinaries"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibuildertemplates"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2iruns"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibuilders"},
		{Group: "devops.kubesphere.io", Version: "v1alpha3", Resource: "devopsprojects"},
		{Group: "devops.kubesphere.io", Version: "v1alpha3", Resource: "pipelines"},
	}

	servicemeshGVRs := []schema.GroupVersionResource{
		{Group: "servicemesh.kubesphere.io", Version: "v1alpha2", Resource: "strategies"},
		{Group: "servicemesh.kubesphere.io", Version: "v1alpha2", Resource: "servicepolicies"},
	}

	// federated resources on cached in multi cluster setup
	federatedResourceGVRs := []schema.GroupVersionResource{
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedClusterRole),
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedClusterRoleBindingBinding),
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedNamespace),
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedService),
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedDeployment),
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedSecret),
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedConfigmap),
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedStatefulSet),
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedIngress),
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedResourceQuota),
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedPersistentVolumeClaim),
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedWorkspace),
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedUser),
		typesv1beta1.SchemeGroupVersion.WithResource(typesv1beta1.ResourcePluralFederatedApplication),
	}

	// skip caching devops resources if devops not enabled
	if s.DevopsClient != nil {
		ksGVRs = append(ksGVRs, devopsGVRs...)
	}

	// skip caching servicemesh resources if servicemesh not enabled
	if s.KubernetesClient.Istio() != nil {
		ksGVRs = append(ksGVRs, servicemeshGVRs...)
	}

	if s.Config.MultiClusterOptions.Enable {
		ksGVRs = append(ksGVRs, federatedResourceGVRs...)
	}

	for _, gvr := range ksGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err = ksInformerFactory.ForResource(gvr)
			if err != nil {
				return err
			}
		}
	}

	ksInformerFactory.Start(stopCh)
	ksInformerFactory.WaitForCacheSync(stopCh)

	appInformerFactory := s.InformerFactory.ApplicationSharedInformerFactory()

	appGVRs := []schema.GroupVersionResource{
		{Group: "app.k8s.io", Version: "v1beta1", Resource: "applications"},
	}

	for _, gvr := range appGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err = appInformerFactory.ForResource(gvr)
			if err != nil {
				return err
			}
		}
	}

	appInformerFactory.Start(stopCh)
	appInformerFactory.WaitForCacheSync(stopCh)

	snapshotInformerFactory := s.InformerFactory.SnapshotSharedInformerFactory()
	snapshotGVRs := []schema.GroupVersionResource{
		{Group: "snapshot.storage.k8s.io", Version: "v1beta1", Resource: "volumesnapshotclasses"},
		{Group: "snapshot.storage.k8s.io", Version: "v1beta1", Resource: "volumesnapshots"},
		{Group: "snapshot.storage.k8s.io", Version: "v1beta1", Resource: "volumesnapshotcontents"},
	}
	for _, gvr := range snapshotGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err = snapshotInformerFactory.ForResource(gvr)
			if err != nil {
				return err
			}
		}
	}
	snapshotInformerFactory.Start(stopCh)
	snapshotInformerFactory.WaitForCacheSync(stopCh)

	apiextensionsInformerFactory := s.InformerFactory.ApiExtensionSharedInformerFactory()
	apiextensionsGVRs := []schema.GroupVersionResource{
		{Group: "apiextensions.k8s.io", Version: "v1beta1", Resource: "customresourcedefinitions"},
	}

	for _, gvr := range apiextensionsGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err = apiextensionsInformerFactory.ForResource(gvr)
			if err != nil {
				return err
			}
		}
	}
	apiextensionsInformerFactory.Start(stopCh)
	apiextensionsInformerFactory.WaitForCacheSync(stopCh)

	klog.V(0).Info("Finished caching objects")

	return nil

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

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

func logRequestAndResponse(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, resp)

	// Always log error response
	logWithVerbose := klog.V(4)
	if resp.StatusCode() > http.StatusBadRequest {
		logWithVerbose = klog.V(0)
	}

	logWithVerbose.Infof("%s - \"%s %s %s\" %d %d %dms",
		utilnet.GetRequestIP(req.Request),
		req.Request.Method,
		req.Request.URL,
		req.Request.Proto,
		resp.StatusCode(),
		resp.ContentLength(),
		time.Since(start)/time.Millisecond,
	)
}

type errorResponder struct{}

func (e *errorResponder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Error(err)
	responsewriters.InternalError(w, req, err)
}
