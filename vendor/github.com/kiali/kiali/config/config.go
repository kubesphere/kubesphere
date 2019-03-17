package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/kiali/kiali/config/security"
	"github.com/kiali/kiali/log"
)

// Environment vars can define some default values.
// NOTE: If you add a new variable, don't forget to update README.adoc
const (
	EnvIdentityCertFile       = "IDENTITY_CERT_FILE"
	EnvIdentityPrivateKeyFile = "IDENTITY_PRIVATE_KEY_FILE"

	EnvPrometheusServiceURL       = "PROMETHEUS_SERVICE_URL"
	EnvPrometheusCustomMetricsURL = "PROMETHEUS_CUSTOM_METRICS_URL"
	EnvInCluster                  = "IN_CLUSTER"
	EnvIstioIdentityDomain        = "ISTIO_IDENTITY_DOMAIN"
	EnvIstioSidecarAnnotation     = "ISTIO_SIDECAR_ANNOTATION"
	EnvIstioUrlServiceVersion     = "ISTIO_URL_SERVICE_VERSION"
	EnvApiNamespacesExclude       = "API_NAMESPACES_EXCLUDE"

	EnvServerAddress                    = "SERVER_ADDRESS"
	EnvServerPort                       = "SERVER_PORT"
	EnvServerCredentialsUsername        = "SERVER_CREDENTIALS_USERNAME"
	EnvServerCredentialsPassword        = "SERVER_CREDENTIALS_PASSWORD"
	EnvServerAllowAnonymousAccess       = "SERVER_ALLOW_ANONYMOUS_ACCESS"
	EnvWebRoot                          = "SERVER_WEB_ROOT"
	EnvServerStaticContentRootDirectory = "SERVER_STATIC_CONTENT_ROOT_DIRECTORY"
	EnvServerCORSAllowAll               = "SERVER_CORS_ALLOW_ALL"
	EnvServerAuditLog                   = "SERVER_AUDIT_LOG"

	EnvGrafanaDisplayLink              = "GRAFANA_DISPLAY_LINK"
	EnvGrafanaURL                      = "GRAFANA_URL"
	EnvGrafanaServiceNamespace         = "GRAFANA_SERVICE_NAMESPACE"
	EnvGrafanaService                  = "GRAFANA_SERVICE"
	EnvGrafanaWorkloadDashboardPattern = "GRAFANA_WORKLOAD_DASHBOARD_PATTERN"
	EnvGrafanaServiceDashboardPattern  = "GRAFANA_SERVICE_DASHBOARD_PATTERN"
	EnvGrafanaVarNamespace             = "GRAFANA_VAR_NAMESPACE"
	EnvGrafanaVarService               = "GRAFANA_VAR_SERVICE"
	EnvGrafanaVarWorkload              = "GRAFANA_VAR_WORKLOAD"
	EnvGrafanaAPIKey                   = "GRAFANA_API_KEY"
	EnvGrafanaUsername                 = "GRAFANA_USERNAME"
	EnvGrafanaPassword                 = "GRAFANA_PASSWORD"

	EnvJaegerURL     = "JAEGER_URL"
	EnvJaegerService = "JAEGER_SERVICE"

	EnvLoginTokenSigningKey        = "LOGIN_TOKEN_SIGNING_KEY"
	EnvLoginTokenExpirationSeconds = "LOGIN_TOKEN_EXPIRATION_SECONDS"
	EnvIstioNamespace              = "ISTIO_NAMESPACE"

	EnvIstioLabelNameApp     = "ISTIO_LABEL_NAME_APP"
	EnvIstioLabelNameVersion = "ISTIO_LABEL_NAME_VERSION"

	EnvKubernetesBurst         = "KUBERNETES_BURST"
	EnvKubernetesQPS           = "KUBERNETES_QPS"
	EnvKubernetesCacheEnabled  = "KUBERNETES_CACHE_ENABLED"
	EnvKubernetesCacheDuration = "KUBERNETES_CACHE_DURATION"
)

// The versions that Kiali requires
const (
	IstioVersionSupported   = ">= 1.0"
	MaistraVersionSupported = ">= 0.1.0"
)

// Global configuration for the application.
var configuration *Config

// Server configuration
type Server struct {
	Address                    string               `yaml:",omitempty"`
	Port                       int                  `yaml:",omitempty"`
	Credentials                security.Credentials `yaml:",omitempty"`
	WebRoot                    string               `yaml:"web_root,omitempty"`
	StaticContentRootDirectory string               `yaml:"static_content_root_directory,omitempty"`
	CORSAllowAll               bool                 `yaml:"cors_allow_all,omitempty"`
	AuditLog                   bool                 `yaml:"audit_log,omitempty"`
}

// GrafanaConfig describes configuration used for Grafana links
type GrafanaConfig struct {
	DisplayLink              bool   `yaml:"display_link"`
	URL                      string `yaml:"url"`
	ServiceNamespace         string `yaml:"service_namespace"`
	Service                  string `yaml:"service"`
	WorkloadDashboardPattern string `yaml:"workload_dashboard_pattern"`
	ServiceDashboardPattern  string `yaml:"service_dashboard_pattern"`
	VarNamespace             string `yaml:"var_namespace"`
	VarService               string `yaml:"var_service"`
	VarWorkload              string `yaml:"var_workload"`
	APIKey                   string `yaml:"api_key"`
	Username                 string `yaml:"username"`
	Password                 string `yaml:"password"`
}

// JaegerConfig describes configuration used for jaeger links
type JaegerConfig struct {
	URL     string `yaml:"url"`
	Service string `yaml:"service"`
}

// IstioConfig describes configuration used for istio links
type IstioConfig struct {
	UrlServiceVersion      string `yaml:"url_service_version"`
	IstioIdentityDomain    string `yaml:"istio_identity_domain,omitempty"`
	IstioSidecarAnnotation string `yaml:"istio_sidecar_annotation,omitempty"`
}

// ExternalServices holds configurations for other systems that Kiali depends on
type ExternalServices struct {
	Istio                      IstioConfig   `yaml:"istio,omitempty"`
	PrometheusServiceURL       string        `yaml:"prometheus_service_url,omitempty"`
	PrometheusCustomMetricsURL string        `yaml:"prometheus_custom_metrics_url,omitempty"`
	Grafana                    GrafanaConfig `yaml:"grafana,omitempty"`
	Jaeger                     JaegerConfig  `yaml:"jaeger,omitempty"`
}

// LoginToken holds config used in token-based authentication
type LoginToken struct {
	SigningKey        []byte `yaml:"signing_key,omitempty"`
	ExpirationSeconds int64  `yaml:"expiration_seconds,omitempty"`
}

// IstioLabels holds configuration about the labels required by Istio
type IstioLabels struct {
	AppLabelName     string `yaml:"app_label_name,omitempty" json:"appLabelName"`
	VersionLabelName string `yaml:"version_label_name,omitempty" json:"versionLabelName"`
}

// Kubernetes client configuration
type KubernetesConfig struct {
	Burst         int     `yaml:"burst,omitempty"`
	QPS           float32 `yaml:"qps,omitempty"`
	CacheEnabled  bool    `yaml:"cache_enabled,omitempty"`
	CacheDuration int64   `yaml:"cache_duration,omitempty"`
}

// Exclude Blacklist holds regex strings defining a blacklist
type ApiConfig struct {
	Namespaces ApiNamespacesConfig
}

// Exclude Blacklist holds regex strings defining a blacklist
type ApiNamespacesConfig struct {
	Exclude []string
}

// Config defines full YAML configuration.
type Config struct {
	Identity         security.Identity `yaml:",omitempty"`
	Server           Server            `yaml:",omitempty"`
	InCluster        bool              `yaml:"in_cluster,omitempty"`
	ExternalServices ExternalServices  `yaml:"external_services,omitempty"`
	LoginToken       LoginToken        `yaml:"login_token,omitempty"`
	IstioNamespace   string            `yaml:"istio_namespace,omitempty"`
	IstioLabels      IstioLabels       `yaml:"istio_labels,omitempty"`
	KubernetesConfig KubernetesConfig  `yaml:"kubernetes_config,omitempty"`
	API              ApiConfig         `yaml:"api,omitempty"`
}

// NewConfig creates a default Config struct
func NewConfig() (c *Config) {
	c = new(Config)

	c.Identity.CertFile = getDefaultString(EnvIdentityCertFile, "")
	c.Identity.PrivateKeyFile = getDefaultString(EnvIdentityPrivateKeyFile, "")
	c.InCluster = getDefaultBool(EnvInCluster, true)
	c.IstioNamespace = strings.TrimSpace(getDefaultString(EnvIstioNamespace, "istio-system"))
	c.IstioLabels.AppLabelName = strings.TrimSpace(getDefaultString(EnvIstioLabelNameApp, "app"))
	c.IstioLabels.VersionLabelName = strings.TrimSpace(getDefaultString(EnvIstioLabelNameVersion, "version"))
	c.API.Namespaces.Exclude = getDefaultStringArray(EnvApiNamespacesExclude, "istio-operator,kube.*,openshift.*,ibm.*")

	// Server configuration
	c.Server.Address = strings.TrimSpace(getDefaultString(EnvServerAddress, ""))
	c.Server.Port = getDefaultInt(EnvServerPort, 20000)
	c.Server.Credentials = security.Credentials{
		Username:       getDefaultString(EnvServerCredentialsUsername, ""),
		Password:       getDefaultString(EnvServerCredentialsPassword, ""),
		AllowAnonymous: getDefaultBool(EnvServerAllowAnonymousAccess, false),
	}

	c.Server.WebRoot = strings.TrimSpace(getDefaultString(EnvWebRoot, "/"))
	c.Server.StaticContentRootDirectory = strings.TrimSpace(getDefaultString(EnvServerStaticContentRootDirectory, "/opt/kiali/console"))
	c.Server.CORSAllowAll = getDefaultBool(EnvServerCORSAllowAll, false)
	c.Server.AuditLog = getDefaultBool(EnvServerAuditLog, true)

	// Prometheus configuration
	c.ExternalServices.PrometheusServiceURL = strings.TrimSpace(getDefaultString(EnvPrometheusServiceURL, "http://prometheus.istio-system:9090"))
	c.ExternalServices.PrometheusCustomMetricsURL = strings.TrimSpace(getDefaultString(EnvPrometheusCustomMetricsURL, c.ExternalServices.PrometheusServiceURL))

	// Grafana Configuration
	c.ExternalServices.Grafana.DisplayLink = getDefaultBool(EnvGrafanaDisplayLink, true)
	c.ExternalServices.Grafana.URL = strings.TrimSpace(getDefaultString(EnvGrafanaURL, ""))
	c.ExternalServices.Grafana.ServiceNamespace = strings.TrimSpace(getDefaultString(EnvGrafanaServiceNamespace, "istio-system"))
	c.ExternalServices.Grafana.Service = strings.TrimSpace(getDefaultString(EnvGrafanaService, "grafana"))
	c.ExternalServices.Grafana.WorkloadDashboardPattern = strings.TrimSpace(getDefaultString(EnvGrafanaWorkloadDashboardPattern, "Istio%20Workload%20Dashboard"))
	c.ExternalServices.Grafana.ServiceDashboardPattern = strings.TrimSpace(getDefaultString(EnvGrafanaServiceDashboardPattern, "Istio%20Service%20Dashboard"))
	c.ExternalServices.Grafana.VarNamespace = strings.TrimSpace(getDefaultString(EnvGrafanaVarNamespace, "var-namespace"))
	c.ExternalServices.Grafana.VarService = strings.TrimSpace(getDefaultString(EnvGrafanaVarService, "var-service"))
	c.ExternalServices.Grafana.VarWorkload = strings.TrimSpace(getDefaultString(EnvGrafanaVarWorkload, "var-workload"))
	c.ExternalServices.Grafana.APIKey = strings.TrimSpace(getDefaultString(EnvGrafanaAPIKey, ""))
	c.ExternalServices.Grafana.Username = strings.TrimSpace(getDefaultString(EnvGrafanaUsername, ""))
	c.ExternalServices.Grafana.Password = strings.TrimSpace(getDefaultString(EnvGrafanaPassword, ""))
	if c.ExternalServices.Grafana.Username != "" && c.ExternalServices.Grafana.Password == "" {
		log.Error("Grafana username (\"GRAFANA_USERNAME\") requires that Grafana password (\"GRAFANA_PASSWORD\") is set.")
	}

	// Jaeger Configuration
	c.ExternalServices.Jaeger.URL = strings.TrimSpace(getDefaultString(EnvJaegerURL, ""))
	c.ExternalServices.Jaeger.Service = strings.TrimSpace(getDefaultString(EnvJaegerService, "jaeger-query"))

	// Istio Configuration
	c.ExternalServices.Istio.IstioIdentityDomain = strings.TrimSpace(getDefaultString(EnvIstioIdentityDomain, "svc.cluster.local"))
	c.ExternalServices.Istio.IstioSidecarAnnotation = strings.TrimSpace(getDefaultString(EnvIstioSidecarAnnotation, "sidecar.istio.io/status"))
	c.ExternalServices.Istio.UrlServiceVersion = strings.TrimSpace(getDefaultString(EnvIstioUrlServiceVersion, "http://istio-pilot:8080/version"))

	// Token-based authentication Configuration
	c.LoginToken.SigningKey = []byte(strings.TrimSpace(getDefaultString(EnvLoginTokenSigningKey, "kiali")))
	c.LoginToken.ExpirationSeconds = getDefaultInt64(EnvLoginTokenExpirationSeconds, 24*3600)

	// Kubernetes client Configuration
	c.KubernetesConfig.Burst = getDefaultInt(EnvKubernetesBurst, 200)
	c.KubernetesConfig.QPS = getDefaultFloat32(EnvKubernetesQPS, 175)
	c.KubernetesConfig.CacheEnabled = getDefaultBool(EnvKubernetesCacheEnabled, false)
	c.KubernetesConfig.CacheDuration = getDefaultInt64(EnvKubernetesCacheDuration, time.Duration(5*time.Minute).Nanoseconds())

	trimmedExclusionPatterns := []string{}
	for _, entry := range c.API.Namespaces.Exclude {
		exclusionPattern := strings.TrimSpace(entry)
		if _, err := regexp.Compile(exclusionPattern); err != nil {
			log.Errorf("Invalid namespace exclude entry, [%s] is not a valid regex pattern: %v", exclusionPattern, err)
		} else {
			trimmedExclusionPatterns = append(trimmedExclusionPatterns, strings.TrimSpace(exclusionPattern))
		}
	}
	c.API.Namespaces.Exclude = trimmedExclusionPatterns

	return
}

// Get the global Config
func Get() (conf *Config) {
	return configuration
}

// Set the global Config
func Set(conf *Config) {
	configuration = conf
}

func getDefaultString(envvar string, defaultValue string) (retVal string) {
	retVal = os.Getenv(envvar)
	if retVal == "" {
		retVal = defaultValue
	}
	return
}

func getDefaultStringArray(envvar string, defaultValue string) (retVal []string) {
	csv := os.Getenv(envvar)
	if csv == "" {
		csv = defaultValue
	}
	retVal = strings.Split(csv, ",")
	return
}

func getDefaultInt(envvar string, defaultValue int) (retVal int) {
	retValString := os.Getenv(envvar)
	if retValString == "" {
		retVal = defaultValue
	} else {
		if num, err := strconv.Atoi(retValString); err != nil {
			log.Warningf("Invalid number for envvar [%v]. err=%v", envvar, err)
			retVal = defaultValue
		} else {
			retVal = num
		}
	}
	return
}

func getDefaultInt64(envvar string, defaultValue int64) (retVal int64) {
	retValString := os.Getenv(envvar)
	if retValString == "" {
		retVal = defaultValue
	} else {
		if num, err := strconv.ParseInt(retValString, 10, 64); err != nil {
			log.Warningf("Invalid number for envvar [%v]. err=%v", envvar, err)
			retVal = defaultValue
		} else {
			retVal = num
		}
	}
	return
}

func getDefaultBool(envvar string, defaultValue bool) (retVal bool) {
	retValString := os.Getenv(envvar)
	if retValString == "" {
		retVal = defaultValue
	} else {
		if b, err := strconv.ParseBool(retValString); err != nil {
			log.Warningf("Invalid boolean for envvar [%v]. err=%v", envvar, err)
			retVal = defaultValue
		} else {
			retVal = b
		}
	}
	return
}

func getDefaultFloat32(envvar string, defaultValue float32) (retVal float32) {
	retValString := os.Getenv(envvar)
	if retValString == "" {
		retVal = defaultValue
	} else {
		if f, err := strconv.ParseFloat(retValString, 32); err != nil {
			log.Warningf("Invalid float number for envvar [%v]. err=%v", envvar, err)
			retVal = defaultValue
		} else {
			retVal = float32(f)
		}
	}
	return
}

// String marshals the given Config into a YAML string
func (conf Config) String() (str string) {
	str, err := Marshal(&conf)
	if err != nil {
		str = fmt.Sprintf("Failed to marshal config to string. err=%v", err)
		log.Debugf(str)
	}

	return
}

// Unmarshal parses the given YAML string and returns its Config object representation.
func Unmarshal(yamlString string) (conf *Config, err error) {
	conf = NewConfig()
	err = yaml.Unmarshal([]byte(yamlString), &conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse yaml data. error=%v", err)
	}
	return
}

// Marshal converts the Config object and returns its YAML string.
func Marshal(conf *Config) (yamlString string, err error) {
	yamlBytes, err := yaml.Marshal(&conf)
	if err != nil {
		return "", fmt.Errorf("Failed to produce yaml. error=%v", err)
	}

	yamlString = string(yamlBytes)
	return
}

// LoadFromFile reads the YAML from the given file, parses the content, and returns its Config object representation.
func LoadFromFile(filename string) (conf *Config, err error) {
	log.Debugf("Reading YAML config from [%s]", filename)
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to load config file [%v]. error=%v", filename, err)
	}

	return Unmarshal(string(fileContent))
}

// SaveToFile converts the Config object and stores its YAML string into the given file, overwriting any data that is in the file.
func SaveToFile(filename string, conf *Config) (err error) {
	fileContent, err := Marshal(conf)
	if err != nil {
		return fmt.Errorf("Failed to save config file [%v]. error=%v", filename, err)
	}

	log.Debugf("Writing YAML config to [%s]", filename)
	err = ioutil.WriteFile(filename, []byte(fileContent), 0640)
	return
}
