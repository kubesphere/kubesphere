package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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
	"os"
	"reflect"
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
			Host:     "10.10.111.110",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		S3Options: &s2is3.S3Options{
			Endpoint:        "http://minio.openpitrix-system.svc",
			Region:          "us-east-1",
			DisableSSL:      true,
			ForcePathStyle:  false,
			AccessKeyID:     "ABCDEFGHIJKLMN",
			SecretAccessKey: "OPQRSTUVWXYZ",
			SessionToken:    "abcdefghijklmn",
			Bucket:          "ssss",
		},
		OpenPitrixOptions: &openpitrix.OpenPitrixOptions{
			APIServer: "http://api-gateway.openpitrix-system.svc",
			Token:     "ABCDEFGHIJKLMN",
		},
		MonitoringOptions: &prometheus.PrometheusOptions{
			Endpoint:          "http://prometheus.kubesphere-monitoring-system.svc",
			SecondaryEndpoint: "http://prometheus.kubesphere-monitoring-system.svc",
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

	err := Load()
	if err != nil {
		t.Fatal(err)
	}
	conf2 := Get()

	if !reflect.DeepEqual(conf2, conf) {
		t.Fatalf("Get %v\n expected %v\n", conf2, conf)
	}
	cleanTestConfig(t)
}
