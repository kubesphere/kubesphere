package options

import (
	"crypto/tls"
	"flag"
	"fmt"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/iam"
	"kubesphere.io/kubesphere/pkg/apiserver"
	"kubesphere.io/kubesphere/pkg/informers"
	genericoptions "kubesphere.io/kubesphere/pkg/server/options"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	esclient "kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring/prometheus"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	fakeS3 "kubesphere.io/kubesphere/pkg/simple/client/s3/fake"
	"kubesphere.io/kubesphere/pkg/simple/client/servicemesh"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
	"net/http"
	"strings"
)

type ServerRunOptions struct {
	ConfigFile              string
	GenericServerRunOptions *genericoptions.ServerRunOptions
	KubernetesOptions       *k8s.KubernetesOptions
	DevopsOptions           *jenkins.Options
	SonarQubeOptions        *sonarqube.Options
	ServiceMeshOptions      *servicemesh.Options
	MySQLOptions            *mysql.Options
	MonitoringOptions       *prometheus.Options
	S3Options               *s3.Options
	OpenPitrixOptions       *openpitrix.Options
	LoggingOptions          *esclient.Options
	LdapOptions             *ldap.Options
	CacheOptions            *cache.Options
	AuthenticateOptions     *iam.AuthenticationOptions

	//
	DebugMode bool
}

func NewServerRunOptions() *ServerRunOptions {

	s := ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		KubernetesOptions:       k8s.NewKubernetesOptions(),
		DevopsOptions:           jenkins.NewDevopsOptions(),
		SonarQubeOptions:        sonarqube.NewSonarQubeOptions(),
		ServiceMeshOptions:      servicemesh.NewServiceMeshOptions(),
		MySQLOptions:            mysql.NewMySQLOptions(),
		MonitoringOptions:       prometheus.NewPrometheusOptions(),
		S3Options:               s3.NewS3Options(),
		OpenPitrixOptions:       openpitrix.NewOptions(),
		LoggingOptions:          esclient.NewElasticSearchOptions(),
		LdapOptions:             ldap.NewOptions(),
		CacheOptions:            cache.NewRedisOptions(),
		AuthenticateOptions:     iam.NewAuthenticateOptions(),
	}

	return &s
}

func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("generic")
	fs.BoolVar(&s.DebugMode, "debug", false, "Don't enable this if you don't know what it means.")
	s.GenericServerRunOptions.AddFlags(fs, s.GenericServerRunOptions)
	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)
	s.AuthenticateOptions.AddFlags(fss.FlagSet("authenticate"), s.AuthenticateOptions)
	s.MySQLOptions.AddFlags(fss.FlagSet("mysql"), s.MySQLOptions)
	s.DevopsOptions.AddFlags(fss.FlagSet("devops"), s.DevopsOptions)
	s.SonarQubeOptions.AddFlags(fss.FlagSet("sonarqube"), s.SonarQubeOptions)
	s.LdapOptions.AddFlags(fss.FlagSet("ldap"), s.LdapOptions)
	s.CacheOptions.AddFlags(fss.FlagSet("cache"), s.CacheOptions)
	s.S3Options.AddFlags(fss.FlagSet("s3"), s.S3Options)
	s.OpenPitrixOptions.AddFlags(fss.FlagSet("openpitrix"), s.OpenPitrixOptions)
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
	apiServer := &apiserver.APIServer{}

	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		return nil, err
	}
	apiServer.KubernetesClient = kubernetesClient

	informerFactory := informers.NewInformerFactories(kubernetesClient.Kubernetes(), kubernetesClient.KubeSphere(), kubernetesClient.Istio(), kubernetesClient.Application())
	apiServer.InformerFactory = informerFactory

	monitoringClient := prometheus.NewPrometheus(s.MonitoringOptions)
	apiServer.MonitoringClient = monitoringClient

	if s.LoggingOptions.Host != "" {
		loggingClient, err := esclient.NewElasticsearch(s.LoggingOptions)
		if err != nil {
			return nil, err
		}
		apiServer.LoggingClient = loggingClient
	}

	if s.S3Options.Endpoint != "" {
		if s.S3Options.Endpoint == fakeInterface && s.DebugMode {
			apiServer.S3Client = fakeS3.NewFakeS3()
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
	if s.CacheOptions.RedisURL != "" {
		if s.CacheOptions.RedisURL == fakeInterface && s.DebugMode {
			apiServer.CacheClient = cache.NewSimpleCache()
		} else {
			cacheClient, err = cache.NewRedisClient(s.CacheOptions, stopCh)
			if err != nil {
				return nil, err
			}
			apiServer.CacheClient = cacheClient
		}
	}

	if s.MySQLOptions.Host != "" {
		dbClient, err := mysql.NewMySQLClient(s.MySQLOptions, stopCh)
		if err != nil {
			return nil, err
		}
		apiServer.DBClient = dbClient
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
