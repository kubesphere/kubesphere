/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
	"kubesphere.io/utils/s3"

	"kubesphere.io/kubesphere/pkg/apiserver/auditing"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization"
	"kubesphere.io/kubesphere/pkg/controller/options"
	"kubesphere.io/kubesphere/pkg/models/composedapp"
	"kubesphere.io/kubesphere/pkg/models/terminal"
	"kubesphere.io/kubesphere/pkg/multicluster"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
)

func newTestConfig() (*Config, error) {
	var conf = &Config{
		KubernetesOptions:     k8s.NewKubernetesOptions(),
		CacheOptions:          cache.NewCacheOptions(),
		AuthorizationOptions:  authorization.NewOptions(),
		AuthenticationOptions: authentication.NewOptions(),
		MultiClusterOptions:   multicluster.NewOptions(),
		AuditingOptions:       auditing.NewAuditingOptions(),
		TerminalOptions:       terminal.NewOptions(),
		TelemetryOptions:      options.NewTelemetryOptions(),
		HelmExecutorOptions:   options.NewHelmExecutorOptions(),
		ExtensionOptions:      options.NewExtensionOptions(),
		S3Options:             s3.NewS3Options(),
		KubeSphereOptions:     options.NewKubeSphereOptions(),
		ComposedAppOptions:    &composedapp.Options{},
		ExperimentalOptions:   NewExperimentalOptions(),
	}
	return conf, nil
}

func saveTestConfig(t *testing.T, conf *Config) {
	content, err := yaml.Marshal(conf)
	if err != nil {
		t.Fatalf("error marshal config. %v", err)
	}
	err = os.WriteFile(fmt.Sprintf("%s.yaml", defaultConfigurationName), content, 0640)
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
