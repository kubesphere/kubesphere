package config

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
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
	"net/http"
)

// install api for config
func InstallAPI(c *restful.Container) {
	ws := runtime.NewWebService(schema.GroupVersion{
		Group:   "",
		Version: "v1alpha1",
	})

	ws.Route(ws.GET("/configz").
		To(func(request *restful.Request, response *restful.Response) {
			response.WriteAsJson(sharedConfig)
		}).
		Doc("Get system components configuration").
		Produces(restful.MIME_JSON).
		Writes(Config{}).
		Returns(http.StatusOK, "ok", Config{}))

	c.Add(ws)
}

// load configuration after setup
func Load() error {
	sharedConfig = newConfig()

	viper.SetConfigName(DefaultConfigurationName)
	viper.AddConfigPath(DefaultConfigurationPath)
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			klog.Warning("configuration file not found")
			return nil
		} else {
			panic(fmt.Errorf("error parsing configuration file %s", err))
		}
	}

	conf := &Config{}
	if err := viper.Unmarshal(&conf); err != nil {
		klog.Error(fmt.Errorf("error unmarshal configuration %v", err))
		return err
	} else {
		conf.Apply(shadowConfig)
		sharedConfig = conf
	}

	return nil
}

const (
	DefaultConfigurationName = "kubesphere"
	DefaultConfigurationPath = "/etc/kubesphere"
)

var (
	// sharedConfig holds configuration across kubesphere
	sharedConfig *Config

	// shadowConfig contains options from commandline options
	shadowConfig = &Config{}
)

type Config struct {
	MySQLOptions       *mysql.MySQLOptions             `json:"mysql,omitempty" yaml:"mysql,omitempty" mapstructure:"mysql"`
	DevopsOptions      *devops.DevopsOptions           `json:"devops,omitempty" yaml:"devops,omitempty" mapstructure:"devops"`
	SonarQubeOptions   *sonarqube.SonarQubeOptions     `json:"sonarqube,omitempty" yaml:"sonarQube,omitempty" mapstructure:"sonarqube"`
	KubernetesOptions  *k8s.KubernetesOptions          `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes"`
	ServiceMeshOptions *servicemesh.ServiceMeshOptions `json:"servicemesh,omitempty" yaml:"servicemesh,omitempty" mapstructure:"servicemesh"`
	LdapOptions        *ldap.LdapOptions               `json:"ldap,omitempty" yaml:"ldap,omitempty" mapstructure:"ldap"`
	RedisOptions       *redis.RedisOptions             `json:"redis,omitempty" yaml:"redis,omitempty" mapstructure:"redis"`
	S3Options          *s2is3.S3Options                `json:"s3,omitempty" yaml:"s3,omitempty" mapstructure:"s3"`
	OpenPitrixOptions  *openpitrix.OpenPitrixOptions   `json:"openpitrix,omitempty" yaml:"openpitrix,omitempty" mapstructure:"openpitrix"`
	MonitoringOptions  *prometheus.PrometheusOptions   `json:"monitoring,omitempty" yaml:"monitoring,omitempty" mapstructure:"monitoring"`
}

func newConfig() *Config {
	return &Config{
		MySQLOptions:       mysql.NewMySQLOptions(),
		DevopsOptions:      devops.NewDevopsOptions(),
		SonarQubeOptions:   sonarqube.NewSonarQubeOptions(),
		KubernetesOptions:  k8s.NewKubernetesOptions(),
		ServiceMeshOptions: servicemesh.NewServiceMeshOptions(),
		LdapOptions:        ldap.NewLdapOptions(),
		RedisOptions:       redis.NewRedisOptions(),
		S3Options:          s2is3.NewS3Options(),
		OpenPitrixOptions:  openpitrix.NewOpenPitrixOptions(),
		MonitoringOptions:  prometheus.NewPrometheusOptions(),
	}
}

func Get() *Config {
	return sharedConfig
}

func (c *Config) Apply(conf *Config) {
	shadowConfig = conf

	if conf.MonitoringOptions != nil {
		conf.MonitoringOptions.ApplyTo(c.MonitoringOptions)
	}
	if conf.OpenPitrixOptions != nil {
		conf.OpenPitrixOptions.ApplyTo(c.OpenPitrixOptions)
	}

	if conf.S3Options != nil {
		conf.S3Options.ApplyTo(c.S3Options)
	}

	if conf.RedisOptions != nil {
		conf.RedisOptions.ApplyTo(c.RedisOptions)
	}

	if conf.LdapOptions != nil {
		conf.LdapOptions.ApplyTo(c.LdapOptions)
	}

	if conf.ServiceMeshOptions != nil {
		conf.ServiceMeshOptions.ApplyTo(c.ServiceMeshOptions)
	}

	if conf.KubernetesOptions != nil {
		conf.KubernetesOptions.ApplyTo(c.KubernetesOptions)
	}

	if conf.SonarQubeOptions != nil {
		conf.SonarQubeOptions.ApplyTo(c.SonarQubeOptions)
	}

	if conf.DevopsOptions != nil {
		conf.DevopsOptions.ApplyTo(c.DevopsOptions)
	}

	if conf.MySQLOptions != nil {
		conf.MySQLOptions.ApplyTo(c.MySQLOptions)
	}
}
