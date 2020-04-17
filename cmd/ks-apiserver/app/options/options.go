package options

import (
	"crypto/tls"
	"flag"
	"fmt"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	apiserverconfig "kubesphere.io/kubesphere/pkg/apiserver/config"
	"kubesphere.io/kubesphere/pkg/informers"
	genericoptions "kubesphere.io/kubesphere/pkg/server/options"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	esclient "kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring/prometheus"
	"kubesphere.io/kubesphere/pkg/simple/client/network"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	fakes3 "kubesphere.io/kubesphere/pkg/simple/client/s3/fake"
	"kubesphere.io/kubesphere/pkg/simple/client/servicemesh"
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
		Config: &apiserverconfig.Config{
			KubernetesOptions:     k8s.NewKubernetesOptions(),
			DevopsOptions:         jenkins.NewDevopsOptions(),
			SonarQubeOptions:      sonarqube.NewSonarQubeOptions(),
			ServiceMeshOptions:    servicemesh.NewServiceMeshOptions(),
			NetworkOptions:        network.NewNetworkOptions(),
			MonitoringOptions:     prometheus.NewPrometheusOptions(),
			S3Options:             s3.NewS3Options(),
			OpenPitrixOptions:     openpitrix.NewOptions(),
			LoggingOptions:        esclient.NewElasticSearchOptions(),
			LdapOptions:           ldap.NewOptions(),
			RedisOptions:          cache.NewRedisOptions(),
			AuthenticationOptions: authoptions.NewAuthenticateOptions(),
		},
	}

	return s
}

func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("generic")
	fs.BoolVar(&s.DebugMode, "debug", false, "Don't enable this if you don't know what it means.")
	s.GenericServerRunOptions.AddFlags(fs, s.GenericServerRunOptions)
	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)
	s.AuthenticationOptions.AddFlags(fss.FlagSet("authentication"), s.AuthenticationOptions)
	s.DevopsOptions.AddFlags(fss.FlagSet("devops"), s.DevopsOptions)
	s.SonarQubeOptions.AddFlags(fss.FlagSet("sonarqube"), s.SonarQubeOptions)
	s.LdapOptions.AddFlags(fss.FlagSet("ldap"), s.LdapOptions)
	s.RedisOptions.AddFlags(fss.FlagSet("redis"), s.RedisOptions)
	s.S3Options.AddFlags(fss.FlagSet("s3"), s.S3Options)
	s.OpenPitrixOptions.AddFlags(fss.FlagSet("openpitrix"), s.OpenPitrixOptions)
	s.NetworkOptions.AddFlags(fss.FlagSet("network"), s.NetworkOptions)
	s.ServiceMeshOptions.AddFlags(fss.FlagSet("servicemesh"), s.ServiceMeshOptions)
	s.MonitoringOptions.AddFlags(fss.FlagSet("monitoring"), s.MonitoringOptions)
	s.LoggingOptions.AddFlags(fss.FlagSet("logging"), s.LoggingOptions)

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

	informerFactory := informers.NewInformerFactories(kubernetesClient.Kubernetes(), kubernetesClient.KubeSphere(), kubernetesClient.Istio(), kubernetesClient.Application())
	apiServer.InformerFactory = informerFactory

	if s.MonitoringOptions.Endpoint != "" {
		monitoringClient, err := prometheus.NewPrometheus(s.MonitoringOptions)
		if err != nil {
			return nil, err
		}
		apiServer.MonitoringClient = monitoringClient
	}

	if s.LoggingOptions.Host != "" {
		loggingClient, err := esclient.NewElasticsearch(s.LoggingOptions)
		if err != nil {
			return nil, err
		}
		apiServer.LoggingClient = loggingClient
	}

	if s.S3Options.Endpoint != "" {
		if s.S3Options.Endpoint == fakeInterface && s.DebugMode {
			apiServer.S3Client = fakes3.NewFakeS3()
		} else {
			s3Client, err := s3.NewS3Client(s.S3Options)
			if err != nil {
				return nil, err
			}
			apiServer.S3Client = s3Client
		}
	}

	if s.DevopsOptions.Host != "" {
		devopsClient, err := jenkins.NewDevopsClient(s.DevopsOptions)
		if err != nil {
			return nil, err
		}
		apiServer.DevopsClient = devopsClient
	}

	if s.SonarQubeOptions.Host != "" {
		sonarClient, err := sonarqube.NewSonarQubeClient(s.SonarQubeOptions)
		if err != nil {
			return nil, err
		}
		apiServer.SonarClient = sonarqube.NewSonar(sonarClient.SonarQube())
	}

	if s.LdapOptions.Host != "" {
		if s.LdapOptions.Host == fakeInterface && s.DebugMode {
			apiServer.LdapClient = ldap.NewSimpleLdap()
		} else {
			ldapClient, err := ldap.NewLdapClient(s.LdapOptions, stopCh)
			if err != nil {
				return nil, err
			}
			apiServer.LdapClient = ldapClient
		}
	}

	var cacheClient cache.Interface
	if s.RedisOptions.Host != "" {
		if s.RedisOptions.Host == fakeInterface && s.DebugMode {
			apiServer.CacheClient = cache.NewSimpleCache()
		} else {
			cacheClient, err = cache.NewRedisClient(s.RedisOptions, stopCh)
			if err != nil {
				return nil, err
			}
			apiServer.CacheClient = cacheClient
		}
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
