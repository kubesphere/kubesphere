/*
Copyright 2020 The KubeSphere Authors.

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

package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"

	networkv1alpha1 "kubesphere.io/api/network/v1alpha1"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/terminal"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
	"kubesphere.io/kubesphere/pkg/simple/client/auditing"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/edgeruntime"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
	"kubesphere.io/kubesphere/pkg/simple/client/gateway"
	"kubesphere.io/kubesphere/pkg/simple/client/gpu"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/kubeedge"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"kubesphere.io/kubesphere/pkg/simple/client/metering"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring/prometheus"
	"kubesphere.io/kubesphere/pkg/simple/client/multicluster"
	"kubesphere.io/kubesphere/pkg/simple/client/network"
	"kubesphere.io/kubesphere/pkg/simple/client/notification"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/simple/client/servicemesh"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
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

var (
	// singleton instance of config package
	_config = defaultConfig()
)

const (
	// DefaultConfigurationName is the default name of configuration
	defaultConfigurationName = "kubesphere"

	// DefaultConfigurationPath the default location of the configuration file
	defaultConfigurationPath = "/etc/kubesphere"
)

type config struct {
	cfg         *Config
	cfgChangeCh chan Config
	watchOnce   sync.Once
	loadOnce    sync.Once
}

func (c *config) watchConfig() <-chan Config {
	c.watchOnce.Do(func() {
		viper.WatchConfig()
		viper.OnConfigChange(func(in fsnotify.Event) {
			cfg := New()
			if err := viper.Unmarshal(cfg); err != nil {
				klog.Warningf("config reload error: %v", err)
			} else {
				c.cfgChangeCh <- *cfg
			}
		})
	})
	return c.cfgChangeCh
}

func (c *config) loadFromDisk() (*Config, error) {
	var err error
	c.loadOnce.Do(func() {
		if err = viper.ReadInConfig(); err != nil {
			return
		}
		err = viper.Unmarshal(c.cfg)
	})
	return c.cfg, err
}

func defaultConfig() *config {
	viper.SetConfigName(defaultConfigurationName)
	viper.AddConfigPath(defaultConfigurationPath)

	// Load from current working directory, only used for debugging
	viper.AddConfigPath(".")

	// Load from Environment variables
	viper.SetEnvPrefix("kubesphere")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return &config{
		cfg:         New(),
		cfgChangeCh: make(chan Config),
		watchOnce:   sync.Once{},
		loadOnce:    sync.Once{},
	}
}

// Config defines everything needed for apiserver to deal with external services
type Config struct {
	DevopsOptions         *jenkins.Options        `json:"devops,omitempty" yaml:"devops,omitempty" mapstructure:"devops"`
	SonarQubeOptions      *sonarqube.Options      `json:"sonarqube,omitempty" yaml:"sonarQube,omitempty" mapstructure:"sonarqube"`
	KubernetesOptions     *k8s.KubernetesOptions  `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes"`
	ServiceMeshOptions    *servicemesh.Options    `json:"servicemesh,omitempty" yaml:"servicemesh,omitempty" mapstructure:"servicemesh"`
	NetworkOptions        *network.Options        `json:"network,omitempty" yaml:"network,omitempty" mapstructure:"network"`
	LdapOptions           *ldap.Options           `json:"-,omitempty" yaml:"ldap,omitempty" mapstructure:"ldap"`
	CacheOptions          *cache.Options          `json:"cache,omitempty" yaml:"cache,omitempty" mapstructure:"cache"`
	S3Options             *s3.Options             `json:"s3,omitempty" yaml:"s3,omitempty" mapstructure:"s3"`
	OpenPitrixOptions     *openpitrix.Options     `json:"openpitrix,omitempty" yaml:"openpitrix,omitempty" mapstructure:"openpitrix"`
	MonitoringOptions     *prometheus.Options     `json:"monitoring,omitempty" yaml:"monitoring,omitempty" mapstructure:"monitoring"`
	LoggingOptions        *logging.Options        `json:"logging,omitempty" yaml:"logging,omitempty" mapstructure:"logging"`
	AuthenticationOptions *authentication.Options `json:"authentication,omitempty" yaml:"authentication,omitempty" mapstructure:"authentication"`
	AuthorizationOptions  *authorization.Options  `json:"authorization,omitempty" yaml:"authorization,omitempty" mapstructure:"authorization"`
	MultiClusterOptions   *multicluster.Options   `json:"multicluster,omitempty" yaml:"multicluster,omitempty" mapstructure:"multicluster"`
	EventsOptions         *events.Options         `json:"events,omitempty" yaml:"events,omitempty" mapstructure:"events"`
	AuditingOptions       *auditing.Options       `json:"auditing,omitempty" yaml:"auditing,omitempty" mapstructure:"auditing"`
	AlertingOptions       *alerting.Options       `json:"alerting,omitempty" yaml:"alerting,omitempty" mapstructure:"alerting"`
	NotificationOptions   *notification.Options   `json:"notification,omitempty" yaml:"notification,omitempty" mapstructure:"notification"`
	KubeEdgeOptions       *kubeedge.Options       `json:"kubeedge,omitempty" yaml:"kubeedge,omitempty" mapstructure:"kubeedge"`
	EdgeRuntimeOptions    *edgeruntime.Options    `json:"edgeruntime,omitempty" yaml:"edgeruntime,omitempty" mapstructure:"edgeruntime"`
	MeteringOptions       *metering.Options       `json:"metering,omitempty" yaml:"metering,omitempty" mapstructure:"metering"`
	GatewayOptions        *gateway.Options        `json:"gateway,omitempty" yaml:"gateway,omitempty" mapstructure:"gateway"`
	GPUOptions            *gpu.Options            `json:"gpu,omitempty" yaml:"gpu,omitempty" mapstructure:"gpu"`
	TerminalOptions       *terminal.Options       `json:"terminal,omitempty" yaml:"terminal,omitempty" mapstructure:"terminal"`
}

// newConfig creates a default non-empty Config
func New() *Config {
	return &Config{
		DevopsOptions:         jenkins.NewDevopsOptions(),
		SonarQubeOptions:      sonarqube.NewSonarQubeOptions(),
		KubernetesOptions:     k8s.NewKubernetesOptions(),
		ServiceMeshOptions:    servicemesh.NewServiceMeshOptions(),
		NetworkOptions:        network.NewNetworkOptions(),
		LdapOptions:           ldap.NewOptions(),
		CacheOptions:          cache.NewCacheOptions(),
		S3Options:             s3.NewS3Options(),
		OpenPitrixOptions:     openpitrix.NewOptions(),
		MonitoringOptions:     prometheus.NewPrometheusOptions(),
		AlertingOptions:       alerting.NewAlertingOptions(),
		NotificationOptions:   notification.NewNotificationOptions(),
		LoggingOptions:        logging.NewLoggingOptions(),
		AuthenticationOptions: authentication.NewOptions(),
		AuthorizationOptions:  authorization.NewOptions(),
		MultiClusterOptions:   multicluster.NewOptions(),
		EventsOptions:         events.NewEventsOptions(),
		AuditingOptions:       auditing.NewAuditingOptions(),
		KubeEdgeOptions:       kubeedge.NewKubeEdgeOptions(),
		EdgeRuntimeOptions:    edgeruntime.NewEdgeRuntimeOptions(),
		MeteringOptions:       metering.NewMeteringOptions(),
		GatewayOptions:        gateway.NewGatewayOptions(),
		GPUOptions:            gpu.NewGPUOptions(),
		TerminalOptions:       terminal.NewTerminalOptions(),
	}
}

// TryLoadFromDisk loads configuration from default location after server startup
// return nil error if configuration file not exists
func TryLoadFromDisk() (*Config, error) {
	return _config.loadFromDisk()
}

// WatchConfigChange return config change channel
func WatchConfigChange() <-chan Config {
	return _config.watchConfig()
}

// convertToMap simply converts config to map[string]bool
// to hide sensitive information
func (conf *Config) ToMap() map[string]bool {
	conf.stripEmptyOptions()
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

		if name == "network" {
			ippoolName := "network.ippool"
			nsnpName := "network"
			networkTopologyName := "network.topology"
			if conf.NetworkOptions == nil {
				result[nsnpName] = false
				result[ippoolName] = false
			} else {
				if conf.NetworkOptions.EnableNetworkPolicy {
					result[nsnpName] = true
				} else {
					result[nsnpName] = false
				}

				if conf.NetworkOptions.IPPoolType == networkv1alpha1.IPPoolTypeNone {
					result[ippoolName] = false
				} else {
					result[ippoolName] = true
				}

				if conf.NetworkOptions.WeaveScopeHost == "" {
					result[networkTopologyName] = false
				} else {
					result[networkTopologyName] = true
				}
			}
			continue
		}

		if name == "openpitrix" {
			// openpitrix is always true
			result[name] = true
			if conf.OpenPitrixOptions == nil {
				result["openpitrix.appstore"] = false
			} else {
				result["openpitrix.appstore"] = !conf.OpenPitrixOptions.AppStoreConfIsEmpty()
			}
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

	if conf.CacheOptions != nil && conf.CacheOptions.Type == "" {
		conf.CacheOptions = nil
	}

	if conf.DevopsOptions != nil && conf.DevopsOptions.Host == "" {
		conf.DevopsOptions = nil
	}

	if conf.MonitoringOptions != nil && conf.MonitoringOptions.Endpoint == "" {
		conf.MonitoringOptions = nil
	}

	if conf.SonarQubeOptions != nil && conf.SonarQubeOptions.Host == "" {
		conf.SonarQubeOptions = nil
	}

	if conf.LdapOptions != nil && conf.LdapOptions.Host == "" {
		conf.LdapOptions = nil
	}

	if conf.NetworkOptions != nil && conf.NetworkOptions.IsEmpty() {
		conf.NetworkOptions = nil
	}

	if conf.ServiceMeshOptions != nil && conf.ServiceMeshOptions.IstioPilotHost == "" &&
		conf.ServiceMeshOptions.ServicemeshPrometheusHost == "" &&
		conf.ServiceMeshOptions.JaegerQueryHost == "" {
		conf.ServiceMeshOptions = nil
	}

	if conf.S3Options != nil && conf.S3Options.Endpoint == "" {
		conf.S3Options = nil
	}

	if conf.AlertingOptions != nil && conf.AlertingOptions.Endpoint == "" &&
		conf.AlertingOptions.PrometheusEndpoint == "" && conf.AlertingOptions.ThanosRulerEndpoint == "" {
		conf.AlertingOptions = nil
	}

	if conf.LoggingOptions != nil && conf.LoggingOptions.Host == "" {
		conf.LoggingOptions = nil
	}

	if conf.NotificationOptions != nil && conf.NotificationOptions.Endpoint == "" {
		conf.NotificationOptions = nil
	}

	if conf.MultiClusterOptions != nil && !conf.MultiClusterOptions.Enable {
		conf.MultiClusterOptions = nil
	}

	if conf.EventsOptions != nil && conf.EventsOptions.Host == "" {
		conf.EventsOptions = nil
	}

	if conf.AuditingOptions != nil && conf.AuditingOptions.Host == "" {
		conf.AuditingOptions = nil
	}

	if conf.KubeEdgeOptions != nil && conf.KubeEdgeOptions.Endpoint == "" {
		conf.KubeEdgeOptions = nil
	}

	if conf.EdgeRuntimeOptions != nil && conf.EdgeRuntimeOptions.Endpoint == "" {
		conf.EdgeRuntimeOptions = nil
	}

	if conf.GPUOptions != nil && len(conf.GPUOptions.Kinds) == 0 {
		conf.GPUOptions = nil
	}
}

// GetFromConfigMap returns KubeSphere ruuning config by the given ConfigMap.
func GetFromConfigMap(cm *corev1.ConfigMap) (*Config, error) {
	c := &Config{}
	value, ok := cm.Data[constants.KubeSphereConfigMapDataKey]
	if !ok {
		return nil, fmt.Errorf("failed to get configmap kubesphere.yaml value")
	}

	if err := yaml.Unmarshal([]byte(value), c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value from configmap. err: %s", err)
	}
	return c, nil
}
