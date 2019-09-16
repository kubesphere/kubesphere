package options

import (
	"flag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	genericoptions "kubesphere.io/kubesphere/pkg/server/options"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/prometheus"
	"kubesphere.io/kubesphere/pkg/simple/client/redis"
	"kubesphere.io/kubesphere/pkg/simple/client/s2is3"
	"kubesphere.io/kubesphere/pkg/simple/client/servicemesh"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
	"strings"
)

type ServerRunOptions struct {
	GenericServerRunOptions *genericoptions.ServerRunOptions

	KubernetesOptions  *k8s.KubernetesOptions
	DevopsOptions      *devops.DevopsOptions
	SonarQubeOptions   *sonarqube.SonarQubeOptions
	ServiceMeshOptions *servicemesh.ServiceMeshOptions
	MySQLOptions       *mysql.MySQLOptions
	MonitoringOptions  *prometheus.PrometheusOptions
	LdapOptions        *ldap.LdapOptions
	S3Options          *s2is3.S3Options
	RedisOptions       *redis.RedisOptions
	OpenPitrixOptions  *openpitrix.OpenPitrixOptions
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
		LdapOptions:             ldap.NewLdapOptions(),
		S3Options:               s2is3.NewS3Options(),
		RedisOptions:            redis.NewRedisOptions(),
		OpenPitrixOptions:       openpitrix.NewOpenPitrixOptions(),
	}

	return &s
}

func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {

	s.GenericServerRunOptions.AddFlags(fss.FlagSet("generic"))
	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"))
	s.LdapOptions.AddFlags(fss.FlagSet("ldap"))
	s.MySQLOptions.AddFlags(fss.FlagSet("mysql"))
	s.DevopsOptions.AddFlags(fss.FlagSet("devops"))
	s.SonarQubeOptions.AddFlags(fss.FlagSet("sonarqube"))
	s.S3Options.AddFlags(fss.FlagSet("s3"))
	s.RedisOptions.AddFlags(fss.FlagSet("redis"))
	s.OpenPitrixOptions.AddFlags(fss.FlagSet("openpitrix"))
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
