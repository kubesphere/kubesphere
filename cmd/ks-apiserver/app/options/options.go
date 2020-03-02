package options

import (
	"flag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	genericoptions "kubesphere.io/kubesphere/pkg/server/options"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	esclient "kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/prometheus"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/simple/client/servicemesh"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
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
	}

	return &s
}

func (s *ServerRunOptions) Flags(c *ServerRunOptions) (fss cliflag.NamedFlagSets) {
	s.GenericServerRunOptions.AddFlags(fss.FlagSet("generic"), c.GenericServerRunOptions)
	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), c.KubernetesOptions)
	s.MySQLOptions.AddFlags(fss.FlagSet("mysql"), c.MySQLOptions)
	s.DevopsOptions.AddFlags(fss.FlagSet("devops"), c.DevopsOptions)
	s.SonarQubeOptions.AddFlags(fss.FlagSet("sonarqube"), c.SonarQubeOptions)
	s.S3Options.AddFlags(fss.FlagSet("s3"), c.S3Options)
	s.OpenPitrixOptions.AddFlags(fss.FlagSet("openpitrix"), c.OpenPitrixOptions)
	s.ServiceMeshOptions.AddFlags(fss.FlagSet("servicemesh"), c.ServiceMeshOptions)
	s.MonitoringOptions.AddFlags(fss.FlagSet("monitoring"), c.MonitoringOptions)
	s.LoggingOptions.AddFlags(fss.FlagSet("logging"), c.LoggingOptions)

	fs := fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})

	return fss
}
