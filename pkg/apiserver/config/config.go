package config

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/api/iam"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/kubesphere"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring/prometheus"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/simple/client/notification"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
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
// we will `root:password@mysql.kubesphere-system.svc` as input, cause
// mysql-host is missing in command line flags, all other mysql command line flags
// will be ignored.

const (
	// DefaultConfigurationName is the default name of configuration
	defaultConfigurationName = "kubesphere"

	// DefaultConfigurationPath the default location of the configuration file
	defaultConfigurationPath = "/etc/kubesphere"
)

// Config defines everything needed for apiserver to deal with external services
type Config struct {
	MySQLOptions       *mysql.Options         `json:"mysql,omitempty" yaml:"mysql,omitempty" mapstructure:"mysql"`
	DevopsOptions      *jenkins.Options       `json:"devops,omitempty" yaml:"devops,omitempty" mapstructure:"devops"`
	SonarQubeOptions   *sonarqube.Options     `json:"sonarqube,omitempty" yaml:"sonarQube,omitempty" mapstructure:"sonarqube"`
	KubernetesOptions  *k8s.KubernetesOptions `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes"`
	ServiceMeshOptions *servicemesh.Options   `json:"servicemesh,omitempty" yaml:"servicemesh,omitempty" mapstructure:"servicemesh"`
	LdapOptions        *ldap.Options          `json:"ldap,omitempty" yaml:"ldap,omitempty" mapstructure:"ldap"`
	RedisOptions       *cache.Options         `json:"redis,omitempty" yaml:"redis,omitempty" mapstructure:"redis"`
	S3Options          *s3.Options            `json:"s3,omitempty" yaml:"s3,omitempty" mapstructure:"s3"`
	OpenPitrixOptions  *openpitrix.Options    `json:"openpitrix,omitempty" yaml:"openpitrix,omitempty" mapstructure:"openpitrix"`
	MonitoringOptions  *prometheus.Options    `json:"monitoring,omitempty" yaml:"monitoring,omitempty" mapstructure:"monitoring"`
	LoggingOptions     *elasticsearch.Options `json:"logging,omitempty" yaml:"logging,omitempty" mapstructure:"logging"`

	// Options below are only loaded from configuration file, no command line flags for these options now.
	KubeSphereOptions *kubesphere.Options `json:"-" yaml:"kubesphere,omitempty" mapstructure:"kubesphere"`

	AuthenticateOptions *iam.AuthenticationOptions `json:"authenticate,omitempty" yaml:"authenticate,omitempty" mapstructure:"authenticate"`

	// Options used for enabling components, not actually used now. Once we switch Alerting/Notification API to kubesphere,
	// we can add these options to kubesphere command lines
	AlertingOptions     *alerting.Options     `json:"alerting,omitempty" yaml:"alerting,omitempty" mapstructure:"alerting"`
	NotificationOptions *notification.Options `json:"notification,omitempty" yaml:"notification,omitempty" mapstructure:"notification"`
}

// newConfig creates a default non-empty Config
func New() *Config {
	return &Config{
		MySQLOptions:        mysql.NewMySQLOptions(),
		DevopsOptions:       jenkins.NewDevopsOptions(),
		SonarQubeOptions:    sonarqube.NewSonarQubeOptions(),
		KubernetesOptions:   k8s.NewKubernetesOptions(),
		ServiceMeshOptions:  servicemesh.NewServiceMeshOptions(),
		LdapOptions:         ldap.NewOptions(),
		RedisOptions:        cache.NewRedisOptions(),
		S3Options:           s3.NewS3Options(),
		OpenPitrixOptions:   openpitrix.NewOptions(),
		MonitoringOptions:   prometheus.NewPrometheusOptions(),
		KubeSphereOptions:   kubesphere.NewKubeSphereOptions(),
		AlertingOptions:     alerting.NewAlertingOptions(),
		NotificationOptions: notification.NewNotificationOptions(),
		LoggingOptions:      elasticsearch.NewElasticSearchOptions(),
		AuthenticateOptions: iam.NewAuthenticateOptions(),
	}
}

// TryLoadFromDisk loads configuration from default location after server startup
// return nil error if configuration file not exists
func TryLoadFromDisk() (*Config, error) {
	viper.SetConfigName(defaultConfigurationName)
	viper.AddConfigPath(defaultConfigurationPath)

	// Load from current working directory, only used for debugging
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, err
		} else {
			return nil, fmt.Errorf("error parsing configuration file %s", err)
		}
	}

	conf := New()
	if err := viper.Unmarshal(conf); err != nil {
		return nil, err
	} else {
		// make sure kubesphere options always exists
		if conf.KubeSphereOptions == nil {
			conf.KubeSphereOptions = kubesphere.NewKubeSphereOptions()
		} else {
			ksOptions := kubesphere.NewKubeSphereOptions()
			conf.KubeSphereOptions.ApplyTo(ksOptions)
			conf.KubeSphereOptions = ksOptions
		}
	}

	return conf, nil
}

// InstallAPI installs api for config
func (conf *Config) InstallAPI(c *restful.Container) {
	ws := runtime.NewWebService(schema.GroupVersion{
		Group:   "",
		Version: "v1alpha1",
	})

	ws.Route(ws.GET("/configz").
		To(func(request *restful.Request, response *restful.Response) {
			conf.stripEmptyOptions()
			response.WriteAsJson(conf.toMap())
		}).
		Doc("Get system components configuration").
		Produces(restful.MIME_JSON).
		Writes(Config{}).
		Returns(http.StatusOK, "ok", Config{}))

	c.Add(ws)
}

// convertToMap simply converts config to map[string]bool
// to hide sensitive information
func (conf *Config) toMap() map[string]bool {
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

// Remove invalid options before serializing to json or yaml
func (conf *Config) stripEmptyOptions() {
	if conf.MySQLOptions != nil && conf.MySQLOptions.Host == "" {
		conf.MySQLOptions = nil
	}

	if conf.RedisOptions != nil && conf.RedisOptions.RedisURL == "" {
		conf.RedisOptions = nil
	}

	if conf.DevopsOptions != nil && conf.DevopsOptions.Host == "" {
		conf.DevopsOptions = nil
	}

	if conf.MonitoringOptions != nil && conf.MonitoringOptions.Endpoint == "" &&
		conf.MonitoringOptions.SecondaryEndpoint == "" {
		conf.MonitoringOptions = nil
	}

	if conf.SonarQubeOptions != nil && conf.SonarQubeOptions.Host == "" {
		conf.SonarQubeOptions = nil
	}

	if conf.LdapOptions != nil && conf.LdapOptions.Host == "" {
		conf.LdapOptions = nil
	}

	if conf.OpenPitrixOptions != nil && conf.OpenPitrixOptions.IsEmpty() {
		conf.OpenPitrixOptions = nil
	}

	if conf.ServiceMeshOptions != nil && conf.ServiceMeshOptions.IstioPilotHost == "" &&
		conf.ServiceMeshOptions.ServicemeshPrometheusHost == "" &&
		conf.ServiceMeshOptions.JaegerQueryHost == "" {
		conf.ServiceMeshOptions = nil
	}

	if conf.S3Options != nil && conf.S3Options.Endpoint == "" {
		conf.S3Options = nil
	}

	if conf.AlertingOptions != nil && conf.AlertingOptions.Endpoint == "" {
		conf.AlertingOptions = nil
	}

	if conf.LoggingOptions != nil && conf.LoggingOptions.Host == "" {
		conf.LoggingOptions = nil
	}

	if conf.NotificationOptions != nil && conf.NotificationOptions.Endpoint == "" {
		conf.NotificationOptions = nil
	}

}
