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
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"

	networkv1alpha1 "kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	authorizationoptions "kubesphere.io/kubesphere/pkg/apiserver/authorization/options"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
	"kubesphere.io/kubesphere/pkg/simple/client/auditing"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
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

func newTestConfig() (*Config, error) {

	var conf = &Config{
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
			Host:     "localhost",
			Port:     6379,
			Password: "P@88w0rd",
			DB:       0,
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
			S3Options: &s3.Options{
				Endpoint:        "http://minio.openpitrix-system.svc",
				Region:          "",
				DisableSSL:      false,
				ForcePathStyle:  false,
				AccessKeyID:     "ABCDEFGHIJKLMN",
				SecretAccessKey: "OPQRSTUVWXYZ",
				SessionToken:    "abcdefghijklmn",
				Bucket:          "app",
			},
			ReleaseControllerOptions: &openpitrix.ReleaseControllerOptions{
				MaxConcurrent: 10,
				WaitTime:      30 * time.Second,
			},
		},
		NetworkOptions: &network.Options{
			EnableNetworkPolicy: true,
			NSNPOptions: network.NSNPOptions{
				AllowedIngressNamespaces: []string{},
			},
			WeaveScopeHost: "weave-scope-app.weave",
			IPPoolType:     networkv1alpha1.IPPoolTypeNone,
		},
		MonitoringOptions: &prometheus.Options{
			Endpoint: "http://prometheus.kubesphere-monitoring-system.svc",
		},
		LoggingOptions: &logging.Options{
			Host:        "http://elasticsearch-logging.kubesphere-logging-system.svc:9200",
			IndexPrefix: "elk",
			Version:     "6",
		},
		AlertingOptions: &alerting.Options{
			Endpoint: "http://alerting-client-server.kubesphere-alerting-system.svc:9200/api",

			PrometheusEndpoint:       "http://prometheus-operated.kubesphere-monitoring-system.svc",
			ThanosRulerEndpoint:      "http://thanos-ruler-operated.kubesphere-monitoring-system.svc",
			ThanosRuleResourceLabels: "thanosruler=thanos-ruler,role=thanos-alerting-rules",
		},
		NotificationOptions: &notification.Options{
			Endpoint: "http://notification.kubesphere-alerting-system.svc:9200",
		},
		AuthorizationOptions: authorizationoptions.NewAuthorizationOptions(),
		AuthenticationOptions: &authoptions.AuthenticationOptions{
			AuthenticateRateLimiterMaxTries: 5,
			AuthenticateRateLimiterDuration: 30 * time.Minute,
			JwtSecret:                       "xxxxxx",
			MultipleLogin:                   false,
			OAuthOptions: &oauth.Options{
				IdentityProviders: []oauth.IdentityProviderOptions{},
				Clients: []oauth.Client{{
					Name:                         "kubesphere-console-client",
					Secret:                       "xxxxxx-xxxxxx-xxxxxx",
					RespondWithChallenges:        true,
					RedirectURIs:                 []string{"http://ks-console.kubesphere-system.svc/oauth/token/implicit"},
					GrantMethod:                  oauth.GrantHandlerAuto,
					AccessTokenInactivityTimeout: nil,
				}},
				AccessTokenMaxAge:            time.Hour * 24,
				AccessTokenInactivityTimeout: 0,
			},
		},
		MultiClusterOptions: &multicluster.Options{
			Enable: false,
		},
		EventsOptions: &events.Options{
			Host:        "http://elasticsearch-logging-data.kubesphere-logging-system.svc:9200",
			IndexPrefix: "ks-logstash-events",
			Version:     "6",
		},
		AuditingOptions: &auditing.Options{
			Host:        "http://elasticsearch-logging-data.kubesphere-logging-system.svc:9200",
			IndexPrefix: "ks-logstash-auditing",
			Version:     "6",
		},
		KubeEdgeOptions: &kubeedge.Options{
			Endpoint: "http://edge-watcher.kubeedge.svc/api/",
		},
		MeteringOptions: &metering.Options{
			RetentionDay: "7d",
		},
	}
	return conf, nil
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
	conf, err := newTestConfig()
	if err != nil {
		t.Fatal(err)
	}
	saveTestConfig(t, conf)
	defer cleanTestConfig(t)

	conf2, err := TryLoadFromDisk()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(conf, conf2); diff != "" {
		t.Fatal(diff)
	}
}

func TestStripEmptyOptions(t *testing.T) {
	var config Config

	config.RedisOptions = &cache.Options{Host: ""}
	config.DevopsOptions = &jenkins.Options{Host: ""}
	config.MonitoringOptions = &prometheus.Options{Endpoint: ""}
	config.SonarQubeOptions = &sonarqube.Options{Host: ""}
	config.LdapOptions = &ldap.Options{Host: ""}
	config.NetworkOptions = &network.Options{
		EnableNetworkPolicy: false,
		WeaveScopeHost:      "",
		IPPoolType:          networkv1alpha1.IPPoolTypeNone,
	}
	config.ServiceMeshOptions = &servicemesh.Options{
		IstioPilotHost:            "",
		ServicemeshPrometheusHost: "",
		JaegerQueryHost:           "",
	}
	config.S3Options = &s3.Options{
		Endpoint: "",
	}
	config.AlertingOptions = &alerting.Options{
		Endpoint:            "",
		PrometheusEndpoint:  "",
		ThanosRulerEndpoint: "",
	}
	config.LoggingOptions = &logging.Options{Host: ""}
	config.NotificationOptions = &notification.Options{Endpoint: ""}
	config.MultiClusterOptions = &multicluster.Options{Enable: false}
	config.EventsOptions = &events.Options{Host: ""}
	config.AuditingOptions = &auditing.Options{Host: ""}
	config.KubeEdgeOptions = &kubeedge.Options{Endpoint: ""}

	config.stripEmptyOptions()

	if config.RedisOptions != nil ||
		config.DevopsOptions != nil ||
		config.MonitoringOptions != nil ||
		config.SonarQubeOptions != nil ||
		config.LdapOptions != nil ||
		config.NetworkOptions != nil ||
		config.ServiceMeshOptions != nil ||
		config.S3Options != nil ||
		config.AlertingOptions != nil ||
		config.LoggingOptions != nil ||
		config.NotificationOptions != nil ||
		config.MultiClusterOptions != nil ||
		config.EventsOptions != nil ||
		config.AuditingOptions != nil ||
		config.KubeEdgeOptions != nil {
		t.Fatal("config stripEmptyOptions failed")
	}
}
