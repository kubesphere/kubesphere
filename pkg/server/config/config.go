package config

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/kubesphere"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/simple/client/notification"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/prometheus"
	"kubesphere.io/kubesphere/pkg/simple/client/redis"
	"kubesphere.io/kubesphere/pkg/simple/client/s2is3"
	"kubesphere.io/kubesphere/pkg/simple/client/servicemesh"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
	"net/http"
	"reflect"
	"strings"
)

// Package config saves configuration for running KubeSphere components
//
// Config can be configured from command line flags and configuration file.
// Command line flags hold higher priority than configuration file. But if
// component Endpoint/Host/APIServer was left empty, all of that component
// command line flags will be ignored, use configuration file instead.
// For example, we have configuration file
//
// mysql:
//   host: mysql.kubesphere-system.svc
//   username: root
//   password: password
//
// At the same time, have command line flags like following:
//
// --mysql-host mysql.openpitrix-system.svc --mysql-username king --mysql-password 1234
//
// We will use `king:1234@mysql.openpitrix-system.svc` from command line flags rather
// than `root:password@mysql.kubesphere-system.svc` from configuration file,
// cause command line has higher priority. But if command line flags like following:
//
// --mysql-username root --mysql-password password
//
// we will `root:password@mysql.kubesphere-system.svc` as input, case
// mysql-host is missing in command line flags, all other mysql command line flags
// will be ignored.

// InstallAPI installs api for config
func InstallAPI(c *restful.Container) {
	ws := runtime.NewWebService(schema.GroupVersion{
		Group:   "",
		Version: "v1alpha1",
	})

	ws.Route(ws.GET("/configz").
		To(func(request *restful.Request, response *restful.Response) {
			var conf = *sharedConfig

			conf.stripEmptyOptions()

			response.WriteAsJson(convertToMap(&conf))
		}).
		Doc("Get system components configuration").
		Produces(restful.MIME_JSON).
		Writes(Config{}).
		Returns(http.StatusOK, "ok", Config{}))

	c.Add(ws)
}

// convertToMap simply converts config to map[string]bool
// to hide sensitive information
func convertToMap(conf *Config) map[string]bool {
	result := make(map[string]bool, 0)

	if conf == nil {
		return result
	}

	c := reflect.Indirect(reflect.ValueOf(conf))

	for i := 0; i < c.NumField(); i++ {
		name := strings.Split(c.Type().Field(i).Tag.Get("json"), ",")[0]
		if strings.HasPrefix(name, "-") {
			continue
		}

		if c.Field(i).IsNil() {
			result[name] = false
		} else {
			result[name] = true
		}
	}

	return result
}

// Load loads configuration after setup
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

	conf := newConfig()
	if err := viper.Unmarshal(conf); err != nil {
		klog.Error(fmt.Errorf("error unmarshal configuration %v", err))
		return err
	} else {
		// make sure kubesphere options always exists
		if conf.KubeSphereOptions == nil {
			conf.KubeSphereOptions = kubesphere.NewKubeSphereOptions()
		} else {
			ksOptions := kubesphere.NewKubeSphereOptions()
			conf.KubeSphereOptions.ApplyTo(ksOptions)
			conf.KubeSphereOptions = ksOptions
		}

		conf.Apply(shadowConfig)
		sharedConfig = conf
	}

	return nil
}

const (
	// DefaultConfigurationName is the default name of configuration
	DefaultConfigurationName = "kubesphere"

	// DefaultConfigurationPath the default location of the configuration file
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
	LoggingOptions     *esclient.ElasticSearchOptions  `json:"logging,omitempty" yaml:"logging,omitempty" mapstructure:"logging"`

	// Options below are only loaded from configuration file, no command line flags for these options now.
	KubeSphereOptions *kubesphere.KubeSphereOptions `json:"-" yaml:"kubesphere,omitempty" mapstructure:"kubesphere"`

	// Options used for enabling components, not actually used now. Once we switch Alerting/Notification API to kubesphere,
	// we can add these options to kubesphere command lines
	AlertingOptions     *alerting.AlertingOptions         `json:"alerting,omitempty" yaml:"alerting,omitempty" mapstructure:"alerting"`
	NotificationOptions *notification.NotificationOptions `json:"notification,omitempty" yaml:"notification,omitempty" mapstructure:"notification"`
}

func newConfig() *Config {
	return &Config{
		MySQLOptions:        mysql.NewMySQLOptions(),
		DevopsOptions:       devops.NewDevopsOptions(),
		SonarQubeOptions:    sonarqube.NewSonarQubeOptions(),
		KubernetesOptions:   k8s.NewKubernetesOptions(),
		ServiceMeshOptions:  servicemesh.NewServiceMeshOptions(),
		LdapOptions:         ldap.NewLdapOptions(),
		RedisOptions:        redis.NewRedisOptions(),
		S3Options:           s2is3.NewS3Options(),
		OpenPitrixOptions:   openpitrix.NewOpenPitrixOptions(),
		MonitoringOptions:   prometheus.NewPrometheusOptions(),
		KubeSphereOptions:   kubesphere.NewKubeSphereOptions(),
		AlertingOptions:     alerting.NewAlertingOptions(),
		NotificationOptions: notification.NewNotificationOptions(),
		LoggingOptions:      esclient.NewElasticSearchOptions(),
	}
}

func Get() *Config {
	return sharedConfig
}

func (c *Config) Apply(conf *Config) {
	shadowConfig = conf

	if conf.LoggingOptions != nil {
		conf.LoggingOptions.ApplyTo(c.LoggingOptions)
	}

	if conf.KubeSphereOptions != nil {
		conf.KubeSphereOptions.ApplyTo(c.KubeSphereOptions)
	}

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

func (c *Config) stripEmptyOptions() {
	if c.MySQLOptions != nil && c.MySQLOptions.Host == "" {
		c.MySQLOptions = nil
	}

	if c.RedisOptions != nil && c.RedisOptions.RedisURL == "" {
		c.RedisOptions = nil
	}

	if c.DevopsOptions != nil && c.DevopsOptions.Host == "" {
		c.DevopsOptions = nil
	}

	if c.MonitoringOptions != nil && c.MonitoringOptions.Endpoint == "" &&
		c.MonitoringOptions.SecondaryEndpoint == "" {
		c.MonitoringOptions = nil
	}

	if c.SonarQubeOptions != nil && c.SonarQubeOptions.Host == "" {
		c.SonarQubeOptions = nil
	}

	if c.LdapOptions != nil && c.LdapOptions.Host == "" {
		c.LdapOptions = nil
	}

	if c.OpenPitrixOptions != nil && c.OpenPitrixOptions.IsEmpty() {
		c.OpenPitrixOptions = nil
	}

	if c.ServiceMeshOptions != nil && c.ServiceMeshOptions.IstioPilotHost == "" &&
		c.ServiceMeshOptions.ServicemeshPrometheusHost == "" &&
		c.ServiceMeshOptions.JaegerQueryHost == "" {
		c.ServiceMeshOptions = nil
	}

	if c.S3Options != nil && c.S3Options.Endpoint == "" {
		c.S3Options = nil
	}

	if c.AlertingOptions != nil && c.AlertingOptions.Endpoint == "" {
		c.AlertingOptions = nil
	}

	if c.LoggingOptions != nil && c.LoggingOptions.Host == "" {
		c.LoggingOptions = nil
	}

	if c.NotificationOptions != nil && c.NotificationOptions.Endpoint == "" {
		c.NotificationOptions = nil
	}

}
