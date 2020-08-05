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
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver"
	apiserverconfig "kubesphere.io/kubesphere/pkg/apiserver/config"
	"kubesphere.io/kubesphere/pkg/informers"
	genericoptions "kubesphere.io/kubesphere/pkg/server/options"
	auditingclient "kubesphere.io/kubesphere/pkg/simple/client/auditing/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	eventsclient "kubesphere.io/kubesphere/pkg/simple/client/events/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	esclient "kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring/prometheus"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	fakes3 "kubesphere.io/kubesphere/pkg/simple/client/s3/fake"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
	"net/http"
	"strings"
)

type ServerRunOptions struct {
	ConfigFile              string
	GenericServerRunOptions *genericoptions.ServerRunOptions
	*apiserverconfig.Config

	//
	DebugMode bool
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
	s.GenericServerRunOptions.AddFlags(fs, s.GenericServerRunOptions)
	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)
	s.AuthenticationOptions.AddFlags(fss.FlagSet("authentication"), s.AuthenticationOptions)
	s.AuthorizationOptions.AddFlags(fss.FlagSet("authorization"), s.AuthorizationOptions)
	s.DevopsOptions.AddFlags(fss.FlagSet("devops"), s.DevopsOptions)
	s.SonarQubeOptions.AddFlags(fss.FlagSet("sonarqube"), s.SonarQubeOptions)
	s.RedisOptions.AddFlags(fss.FlagSet("redis"), s.RedisOptions)
	s.S3Options.AddFlags(fss.FlagSet("s3"), s.S3Options)
	s.OpenPitrixOptions.AddFlags(fss.FlagSet("openpitrix"), s.OpenPitrixOptions)
	s.NetworkOptions.AddFlags(fss.FlagSet("network"), s.NetworkOptions)
	s.ServiceMeshOptions.AddFlags(fss.FlagSet("servicemesh"), s.ServiceMeshOptions)
	s.MonitoringOptions.AddFlags(fss.FlagSet("monitoring"), s.MonitoringOptions)
	s.LoggingOptions.AddFlags(fss.FlagSet("logging"), s.LoggingOptions)
	s.MultiClusterOptions.AddFlags(fss.FlagSet("multicluster"), s.MultiClusterOptions)
	s.EventsOptions.AddFlags(fss.FlagSet("events"), s.EventsOptions)
	s.AuditingOptions.AddFlags(fss.FlagSet("auditing"), s.AuditingOptions)

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
		kubernetesClient.Istio(), kubernetesClient.Application(), kubernetesClient.Snapshot(), kubernetesClient.ApiExtensions())
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

	if s.LoggingOptions.Host != "" {
		loggingClient, err := esclient.NewElasticsearch(s.LoggingOptions)
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

	var cacheClient cache.Interface
	if s.RedisOptions == nil || len(s.RedisOptions.Host) == 0 {
		return nil, fmt.Errorf("redis service address MUST not be empty, please check configmap/kubesphere-config in kubesphere-system namespace")
	} else {
		if s.RedisOptions.Host == fakeInterface && s.DebugMode {
			apiServer.CacheClient = cache.NewSimpleCache()
		} else {
			cacheClient, err = cache.NewRedisClient(s.RedisOptions, stopCh)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to redis service, please check redis status, error: %v", err)
			}
			apiServer.CacheClient = cacheClient
		}
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

	if s.OpenPitrixOptions != nil && !s.OpenPitrixOptions.IsEmpty() {
		opClient, err := openpitrix.NewClient(s.OpenPitrixOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to openpitrix, please check openpitrix status, error: %v", err)
		}
		apiServer.OpenpitrixClient = opClient
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.GenericServerRunOptions.InsecurePort),
	}

	if s.GenericServerRunOptions.SecurePort != 0 {
		certificate, err := tls.LoadX509KeyPair(s.GenericServerRunOptions.TlsCertFile, s.GenericServerRunOptions.TlsPrivateKey)
		if err != nil {
			return nil, err
		}
		server.TLSConfig.Certificates = []tls.Certificate{certificate}
	}

	apiServer.Server = server

	return apiServer, nil
}
