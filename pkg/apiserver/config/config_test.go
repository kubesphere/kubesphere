package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	iamapi "kubesphere.io/kubesphere/pkg/api/iam"
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
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
	"os"
	"testing"
	"time"
)

func newTestConfig() *Config {
	conf := &Config{
		MySQLOptions: &mysql.Options{
			Host:                  "10.68.96.5:3306",
			Username:              "root",
			Password:              "admin",
			MaxIdleConnections:    10,
			MaxOpenConnections:    20,
			MaxConnectionLifeTime: time.Duration(10) * time.Second,
		},
		DevopsOptions: &jenkins.Options{
			Host:           "http://ks-devops.kubesphere-devops-system.svc",
			Username:       "jenkins",
			Password:       "kubesphere",
			MaxConnections: 10,
		},
		SonarQubeOptions: &sonarqube.Options{
			Host:  "http://sonarqube.kubesphere-devops-system.svc",
			Token: "ABCDEFG",
		},
		KubernetesOptions: &k8s.KubernetesOptions{
			KubeConfig: "/Users/zry/.kube/config",
			Master:     "https://127.0.0.1:6443",
			QPS:        1e6,
			Burst:      1e6,
		},
		ServiceMeshOptions: &servicemesh.Options{
			IstioPilotHost:            "http://istio-pilot.istio-system.svc:9090",
			JaegerQueryHost:           "http://jaeger-query.istio-system.svc:80",
			ServicemeshPrometheusHost: "http://prometheus-k8s.kubesphere-monitoring-system.svc",
		},
		LdapOptions: &ldap.Options{
			Host:            "http://openldap.kubesphere-system.svc",
			ManagerDN:       "cn=admin,dc=example,dc=org",
			ManagerPassword: "P@88w0rd",
			UserSearchBase:  "ou=Users,dc=example,dc=org",
			GroupSearchBase: "ou=Groups,dc=example,dc=org",
		},
		RedisOptions: &cache.Options{
			RedisURL: "redis://:qwerty@localhost:6379/1",
		},
		S3Options: &s3.Options{
			Endpoint:        "http://minio.openpitrix-system.svc",
			Region:          "",
			DisableSSL:      false,
			ForcePathStyle:  false,
			AccessKeyID:     "ABCDEFGHIJKLMN",
			SecretAccessKey: "OPQRSTUVWXYZ",
			SessionToken:    "abcdefghijklmn",
			Bucket:          "ssss",
		},
		OpenPitrixOptions: &openpitrix.Options{
			RuntimeManagerEndpoint:    "openpitrix-hyperpitrix.openpitrix-system.svc:9103",
			ClusterManagerEndpoint:    "openpitrix-hyperpitrix.openpitrix-system.svc:9104",
			RepoManagerEndpoint:       "openpitrix-hyperpitrix.openpitrix-system.svc:9101",
			AppManagerEndpoint:        "openpitrix-hyperpitrix.openpitrix-system.svc:9102",
			CategoryManagerEndpoint:   "openpitrix-hyperpitrix.openpitrix-system.svc:9113",
			AttachmentManagerEndpoint: "openpitrix-hyperpitrix.openpitrix-system.svc:9122",
		},
		MonitoringOptions: &prometheus.Options{
			Endpoint:          "http://prometheus.kubesphere-monitoring-system.svc",
			SecondaryEndpoint: "http://prometheus.kubesphere-monitoring-system.svc",
		},
		LoggingOptions: &elasticsearch.Options{
			Host:        "http://elasticsearch-logging.kubesphere-logging-system.svc:9200",
			IndexPrefix: "elk",
			Version:     "6",
		},
		KubeSphereOptions: &kubesphere.Options{
			APIServer:     "http://ks-apiserver.kubesphere-system.svc",
			AccountServer: "http://ks-account.kubesphere-system.svc",
		},
		AlertingOptions: &alerting.Options{
			Endpoint: "http://alerting.kubesphere-alerting-system.svc:9200",
		},
		NotificationOptions: &notification.Options{
			Endpoint: "http://notification.kubesphere-alerting-system.svc:9200",
		},
		AuthenticateOptions: &iamapi.AuthenticationOptions{
			AuthenticateRateLimiterMaxTries: 5,
			AuthenticateRateLimiterDuration: 30 * time.Minute,
			MaxAuthenticateRetries:          6,
			TokenExpiration:                 30 * time.Minute,
			MultipleLogin:                   false,
		},
	}
	return conf
}

func saveTestConfig(t *testing.T, conf *Config) {
	content, err := yaml.Marshal(conf)
	if err != nil {
		t.Fatalf("error marshal config. %v", err)
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s.yaml", defaultConfigurationName), content, 0640)
	if err != nil {
		t.Fatalf("error write configuration file, %v", err)
	}
}

func cleanTestConfig(t *testing.T) {
	file := fmt.Sprintf("%s.yaml", defaultConfigurationName)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Log("file not exists, skipping")
		return
	}

	err := os.Remove(file)
	if err != nil {
		t.Fatalf("remove %s file failed", file)
	}

}

func TestGet(t *testing.T) {
	conf := newTestConfig()
	saveTestConfig(t, conf)
	defer cleanTestConfig(t)

	conf2, err := TryLoadFromDisk()
	if err != nil {
		t.Fatal(err)
	}

	if diff := reflectutils.Equal(conf, conf2); diff != nil {
		t.Fatal(diff)
	}
}

func TestKubeSphereOptions(t *testing.T) {
	conf := newTestConfig()

	t.Run("save nil kubesphere options", func(t *testing.T) {
		savedConf := *conf
		savedConf.KubeSphereOptions = nil
		saveTestConfig(t, &savedConf)
		defer cleanTestConfig(t)

		loadedConf, err := TryLoadFromDisk()
		if err != nil {
			t.Fatal(err)
		}

		if diff := reflectutils.Equal(conf, loadedConf); diff != nil {
			t.Fatal(diff)
		}
	})

	t.Run("save partially kubesphere options", func(t *testing.T) {
		savedConf := *conf
		savedConf.KubeSphereOptions.APIServer = "http://example.com"
		savedConf.KubeSphereOptions.AccountServer = ""

		saveTestConfig(t, &savedConf)
		defer cleanTestConfig(t)

		loadedConf, err := TryLoadFromDisk()
		if err != nil {
			t.Fatal(err)
		}

		savedConf.KubeSphereOptions.AccountServer = "http://ks-account.kubesphere-system.svc"

		if diff := reflectutils.Equal(&savedConf, loadedConf); diff != nil {
			t.Fatal(diff)
		}
	})
}
