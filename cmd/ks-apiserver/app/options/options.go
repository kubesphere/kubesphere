/*
Copyright 2020 KubeSphere Authors

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

package options

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"strings"

	openpitrixv1 "kubesphere.io/kubesphere/pkg/kapis/openpitrix/v1"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/token"

	"k8s.io/client-go/kubernetes/scheme"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apis"
	"kubesphere.io/kubesphere/pkg/apiserver"
	apiserverconfig "kubesphere.io/kubesphere/pkg/apiserver/config"
	"kubesphere.io/kubesphere/pkg/informers"
	genericoptions "kubesphere.io/kubesphere/pkg/server/options"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
	auditingclient "kubesphere.io/kubesphere/pkg/simple/client/auditing/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"

	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	eventsclient "kubesphere.io/kubesphere/pkg/simple/client/events/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	esclient "kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring/metricsserver"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring/prometheus"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	fakes3 "kubesphere.io/kubesphere/pkg/simple/client/s3/fake"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
)

type ServerRunOptions struct {
	ConfigFile              string
	GenericServerRunOptions *genericoptions.ServerRunOptions
	*apiserverconfig.Config

	//
	DebugMode bool

	// Enable gops or not.
	GOPSEnabled bool
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		Config:                  apiserverconfig.New(),
	}

	return s
}

func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("generic")
	fs.BoolVar(&s.DebugMode, "debug", false, "Don't enable this if you don't know what it means.")
	fs.BoolVar(&s.GOPSEnabled, "gops", false, "Whether to enable gops or not. When enabled this option, "+
		"ks-apiserver will listen on a random port on 127.0.0.1, then you can use the gops tool to list and diagnose the ks-apiserver currently running.")
	s.GenericServerRunOptions.AddFlags(fs, s.GenericServerRunOptions)
	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)
	s.AuthenticationOptions.AddFlags(fss.FlagSet("authentication"), s.AuthenticationOptions)
	s.AuthorizationOptions.AddFlags(fss.FlagSet("authorization"), s.AuthorizationOptions)
	s.DevopsOptions.AddFlags(fss.FlagSet("devops"), s.DevopsOptions)
	s.SonarQubeOptions.AddFlags(fss.FlagSet("sonarqube"), s.SonarQubeOptions)
	s.S3Options.AddFlags(fss.FlagSet("s3"), s.S3Options)
	s.OpenPitrixOptions.AddFlags(fss.FlagSet("openpitrix"), s.OpenPitrixOptions)
	s.NetworkOptions.AddFlags(fss.FlagSet("network"), s.NetworkOptions)
	s.ServiceMeshOptions.AddFlags(fss.FlagSet("servicemesh"), s.ServiceMeshOptions)
	s.MonitoringOptions.AddFlags(fss.FlagSet("monitoring"), s.MonitoringOptions)
	s.LoggingOptions.AddFlags(fss.FlagSet("logging"), s.LoggingOptions)
	s.MultiClusterOptions.AddFlags(fss.FlagSet("multicluster"), s.MultiClusterOptions)
	s.EventsOptions.AddFlags(fss.FlagSet("events"), s.EventsOptions)
	s.AuditingOptions.AddFlags(fss.FlagSet("auditing"), s.AuditingOptions)
	s.AlertingOptions.AddFlags(fss.FlagSet("alerting"), s.AlertingOptions)

	fs = fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})

	return fss
}

const fakeInterface string = "FAKE"

// NewAPIServer creates an APIServer instance using given options
func (s *ServerRunOptions) NewAPIServer(stopCh <-chan struct{}) (*apiserver.APIServer, error) {
	apiServer := &apiserver.APIServer{
		Config: s.Config,
	}

	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		return nil, err
	}
	apiServer.KubernetesClient = kubernetesClient

	informerFactory := informers.NewInformerFactories(kubernetesClient.Kubernetes(), kubernetesClient.KubeSphere(),
		kubernetesClient.Istio(), kubernetesClient.Snapshot(), kubernetesClient.ApiExtensions(), kubernetesClient.Prometheus())
	apiServer.InformerFactory = informerFactory

	if s.MonitoringOptions == nil || len(s.MonitoringOptions.Endpoint) == 0 {
		return nil, fmt.Errorf("moinitoring service address in configuration MUST not be empty, please check configmap/kubesphere-config in kubesphere-system namespace")
	} else {
		monitoringClient, err := prometheus.NewPrometheus(s.MonitoringOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to prometheus, please check prometheus status, error: %v", err)
		}
		apiServer.MonitoringClient = monitoringClient
	}

	apiServer.MetricsClient = metricsserver.NewMetricsClient(kubernetesClient.Kubernetes(), s.KubernetesOptions)

	if s.LoggingOptions.Host != "" {
		loggingClient, err := esclient.NewClient(s.LoggingOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to elasticsearch, please check elasticsearch status, error: %v", err)
		}
		apiServer.LoggingClient = loggingClient
	}

	if s.S3Options.Endpoint != "" {
		if s.S3Options.Endpoint == fakeInterface && s.DebugMode {
			apiServer.S3Client = fakes3.NewFakeS3()
		} else {
			s3Client, err := s3.NewS3Client(s.S3Options)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to s3, please check s3 service status, error: %v", err)
			}
			apiServer.S3Client = s3Client
		}
	}

	if s.DevopsOptions.Host != "" {
		devopsClient, err := jenkins.NewDevopsClient(s.DevopsOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to jenkins, please check jenkins status, error: %v", err)
		}
		apiServer.DevopsClient = devopsClient
	}

	if s.SonarQubeOptions.Host != "" {
		sonarClient, err := sonarqube.NewSonarQubeClient(s.SonarQubeOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connecto to sonarqube, please check sonarqube status, error: %v", err)
		}
		apiServer.SonarClient = sonarqube.NewSonar(sonarClient.SonarQube())
	}

	// If debug mode is on or CacheOptions is nil, will create a fake cache.
	if s.CacheOptions.Type != "" {
		if s.DebugMode {
			s.CacheOptions.Type = cache.DefaultCacheType
		}
		cacheClient, err := cache.New(s.CacheOptions, stopCh)
		if err != nil {
			return nil, fmt.Errorf("failed to create cache, error: %v", err)
		}
		apiServer.CacheClient = cacheClient
	} else {
		s.CacheOptions = &cache.Options{Type: cache.DefaultCacheType}
		// fake cache has no error to return
		cacheClient, _ := cache.New(s.CacheOptions, stopCh)
		apiServer.CacheClient = cacheClient
		klog.Warning("ks-apiserver starts without cache provided, it will use in memory cache. " +
			"This may cause inconsistencies when running ks-apiserver with multiple replicas, and memory leak risk")
	}

	if s.EventsOptions.Host != "" {
		eventsClient, err := eventsclient.NewClient(s.EventsOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to elasticsearch, please check elasticsearch status, error: %v", err)
		}
		apiServer.EventsClient = eventsClient
	}

	if s.AuditingOptions.Host != "" {
		auditingClient, err := auditingclient.NewClient(s.AuditingOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to elasticsearch, please check elasticsearch status, error: %v", err)
		}
		apiServer.AuditingClient = auditingClient
	}

	if s.AlertingOptions != nil && (s.AlertingOptions.PrometheusEndpoint != "" || s.AlertingOptions.ThanosRulerEndpoint != "") {
		alertingClient, err := alerting.NewRuleClient(s.AlertingOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to init alerting client: %v", err)
		}
		apiServer.AlertingClient = alertingClient
	}

	if s.Config.MultiClusterOptions.Enable {
		cc := clusterclient.NewClusterClient(informerFactory.KubeSphereSharedInformerFactory().Cluster().V1alpha1().Clusters())
		apiServer.ClusterClient = cc
	}

	apiServer.OpenpitrixClient = openpitrixv1.NewOpenpitrixClient(informerFactory, apiServer.KubernetesClient.KubeSphere(), s.OpenPitrixOptions, apiServer.ClusterClient)

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.GenericServerRunOptions.InsecurePort),
	}

	if s.GenericServerRunOptions.SecurePort != 0 {
		certificate, err := tls.LoadX509KeyPair(s.GenericServerRunOptions.TlsCertFile, s.GenericServerRunOptions.TlsPrivateKey)
		if err != nil {
			return nil, err
		}

		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{certificate},
		}
		server.Addr = fmt.Sprintf(":%d", s.GenericServerRunOptions.SecurePort)
	}

	sch := scheme.Scheme
	if err := apis.AddToScheme(sch); err != nil {
		klog.Fatalf("unable add APIs to scheme: %v", err)
	}

	apiServer.RuntimeCache, err = runtimecache.New(apiServer.KubernetesClient.Config(), runtimecache.Options{Scheme: sch})
	if err != nil {
		klog.Fatalf("unable to create controller runtime cache: %v", err)
	}

	apiServer.RuntimeClient, err = runtimeclient.New(apiServer.KubernetesClient.Config(), runtimeclient.Options{Scheme: sch})
	if err != nil {
		klog.Fatalf("unable to create controller runtime client: %v", err)
	}

	apiServer.Issuer, err = token.NewIssuer(s.AuthenticationOptions)
	if err != nil {
		klog.Fatalf("unable to create issuer: %v", err)
	}

	apiServer.Server = server

	return apiServer, nil
}
