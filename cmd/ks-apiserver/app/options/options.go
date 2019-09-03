package options

import (
	"flag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	genericoptions "kubesphere.io/kubesphere/pkg/options"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/simple/client/prometheus"
	"kubesphere.io/kubesphere/pkg/simple/client/servicemesh"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
	"strings"
)

type ServerRunOptions struct {
	GenericServerRunOptions *genericoptions.ServerRunOptions

	KubernetesOptions *k8s.KubernetesOptions

	DevopsOptions *devops.DevopsOptions

	SonarQubeOptions *sonarqube.SonarQubeOptions

	ServiceMeshOptions *servicemesh.ServiceMeshOptions

	MySQLOptions *mysql.MySQLOptions

	MonitoringOptions *prometheus.PrometheusOptions
}

func NewServerRunOptions() *ServerRunOptions {

	s := ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		KubernetesOptions:       k8s.NewKubernetesOptions(),
		DevopsOptions:           devops.NewDevopsOptions(),
		SonarQubeOptions:        sonarqube.NewSonarQubeOptions(),
		ServiceMeshOptions:      servicemesh.NewServiceMeshOptions(),
		MySQLOptions:            mysql.NewMySQLOptions(),
		MonitoringOptions:       prometheus.NewPrometheusOptions(),
	}

	return &s
}

func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {

	s.GenericServerRunOptions.AddFlags(fss.FlagSet("generic"))
	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"))
	s.DevopsOptions.AddFlags(fss.FlagSet("devops"))
	s.SonarQubeOptions.AddFlags(fss.FlagSet("sonarqube"))
	s.ServiceMeshOptions.AddFlags(fss.FlagSet("servicemesh"))
	s.MonitoringOptions.AddFlags(fss.FlagSet("monitoring"))

	fs := fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})

	return fss
}
