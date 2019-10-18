package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
	"os"
	"testing"
	"time"
)

func newTestConfig() *Config {
	conf := &Config{
		MySQLOptions: &mysql.MySQLOptions{
			Host:                  "10.68.96.5:3306",
			Username:              "root",
			Password:              "admin",
			MaxIdleConnections:    10,
			MaxOpenConnections:    20,
			MaxConnectionLifeTime: time.Duration(10) * time.Second,
		},
		DevopsOptions: &devops.DevopsOptions{
			Host:           "http://ks-devops.kubesphere-devops-system.svc",
			Username:       "jenkins",
			Password:       "kubesphere",
			MaxConnections: 10,
		},
		SonarQubeOptions: &sonarqube.SonarQubeOptions{
			Host:  "http://sonarqube.kubesphere-devops-system.svc",
			Token: "ABCDEFG",
		},
		KubernetesOptions: &k8s.KubernetesOptions{
			KubeConfig: "/Users/zry/.kube/config",
			Master:     "https://127.0.0.1:6443",
			QPS:        1e6,
			Burst:      1e6,
		},
		ServiceMeshOptions: &servicemesh.ServiceMeshOptions{
			IstioPilotHost:            "http://istio-pilot.istio-system.svc:9090",
			JaegerQueryHost:           "http://jaeger-query.istio-system.svc:80",
			ServicemeshPrometheusHost: "http://prometheus-k8s.kubesphere-monitoring-system.svc",
		},
		LdapOptions: &ldap.LdapOptions{
			Host:            "http://openldap.kubesphere-system.svc",
			ManagerDN:       "cn=admin,dc=example,dc=org",
			ManagerPassword: "P@88w0rd",
			UserSearchBase:  "ou=Users,dc=example,dc=org",
			GroupSearchBase: "ou=Groups,dc=example,dc=org",
		},
		RedisOptions: &redis.RedisOptions{
			RedisURL: "redis://:qwerty@localhost:6379/1",
		},
		S3Options: &s2is3.S3Options{
			Endpoint:        "http://minio.openpitrix-system.svc",
			Region:          "",
			DisableSSL:      false,
			ForcePathStyle:  false,
			AccessKeyID:     "ABCDEFGHIJKLMN",
			SecretAccessKey: "OPQRSTUVWXYZ",
			SessionToken:    "abcdefghijklmn",
			Bucket:          "ssss",
		},
		OpenPitrixOptions: &openpitrix.OpenPitrixOptions{
			RuntimeManagerEndpoint:    "openpitrix-hyperpitrix.openpitrix-system.svc:9103",
			ClusterManagerEndpoint:    "openpitrix-hyperpitrix.openpitrix-system.svc:9104",
			RepoManagerEndpoint:       "openpitrix-hyperpitrix.openpitrix-system.svc:9101",
			AppManagerEndpoint:        "openpitrix-hyperpitrix.openpitrix-system.svc:9102",
			CategoryManagerEndpoint:   "openpitrix-hyperpitrix.openpitrix-system.svc:9113",
			AttachmentManagerEndpoint: "openpitrix-hyperpitrix.openpitrix-system.svc:9122",
		},
		MonitoringOptions: &prometheus.PrometheusOptions{
			Endpoint:          "http://prometheus.kubesphere-monitoring-system.svc",
			SecondaryEndpoint: "http://prometheus.kubesphere-monitoring-system.svc",
		},
		LoggingOptions: &esclient.ElasticSearchOptions{
			Host:        "http://elasticsearch-logging.kubesphere-logging-system.svc:9200",
			IndexPrefix: "elk",
			Version:     "6",
		},
		KubeSphereOptions: &kubesphere.KubeSphereOptions{
			APIServer:     "http://ks-apiserver.kubesphere-system.svc",
			AccountServer: "http://ks-account.kubesphere-system.svc",
		},
		AlertingOptions: &alerting.AlertingOptions{
			Endpoint: "http://alerting.kubesphere-alerting-system.svc:9200",
		},
		NotificationOptions: &notification.NotificationOptions{
			Endpoint: "http://notification.kubesphere-alerting-system.svc:9200",
		},
	}
	return conf
}

func saveTestConfig(t *testing.T, conf *Config) {
	content, err := yaml.Marshal(conf)
	if err != nil {
		t.Fatalf("error marshal config. %v", err)
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s.yaml", DefaultConfigurationName), content, 0640)
	if err != nil {
		t.Fatalf("error write configuration file, %v", err)
	}
}

func cleanTestConfig(t *testing.T) {
	file := fmt.Sprintf("%s.yaml", DefaultConfigurationName)
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

	err := Load()
	if err != nil {
		t.Fatal(err)
	}
	conf2 := Get()

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

		err := Load()
		if err != nil {
			t.Fatal(err)
		}
		loadedConf := Get()

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

		err := Load()
		if err != nil {
			t.Fatal(err)
		}
		loadedConf := Get()

		savedConf.KubeSphereOptions.AccountServer = "http://ks-account.kubesphere-system.svc"

		if diff := reflectutils.Equal(&savedConf, loadedConf); diff != nil {
			t.Fatal(diff)
		}
	})
}
