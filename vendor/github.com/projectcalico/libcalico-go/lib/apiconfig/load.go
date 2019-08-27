package apiconfig

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/kelseyhightower/envconfig"
	yaml "github.com/projectcalico/go-yaml-wrapper"
	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	log "github.com/sirupsen/logrus"
)

// LoadClientConfig loads the ClientConfig from the specified file (if specified)
// or from environment variables (if the file is not specified).
func LoadClientConfig(filename string) (*CalicoAPIConfig, error) {

	// Override / merge with values loaded from the specified file.
	if filename != "" {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		c, err := LoadClientConfigFromBytes(b)
		if err != nil {
			return nil, fmt.Errorf("syntax error in %s: %v", filename, err)
		}
		return c, nil
	}
	return LoadClientConfigFromEnvironment()
}

// LoadClientConfig loads the ClientConfig from the supplied bytes containing
// YAML or JSON format data.
func LoadClientConfigFromBytes(b []byte) (*CalicoAPIConfig, error) {
	var c CalicoAPIConfig

	// Default the backend type to be etcd v3.  This will be overridden if
	// explicitly specified in the file.
	log.Debug("Loading config from JSON or YAML data")
	c = CalicoAPIConfig{
		Spec: CalicoAPIConfigSpec{
			DatastoreType: EtcdV3,
		},
	}

	if err := yaml.UnmarshalStrict(b, &c); err != nil {
		return nil, err
	}

	// Validate the version and kind.
	if c.APIVersion != apiv3.GroupVersionCurrent {
		return nil, errors.New("invalid config file: unknown APIVersion '" + c.APIVersion + "'")
	}
	if c.Kind != KindCalicoAPIConfig {
		return nil, errors.New("invalid config file: expected kind '" + KindCalicoAPIConfig + "', got '" + c.Kind + "'")
	}

	log.Debug("Datastore type: ", c.Spec.DatastoreType)
	return &c, nil
}

// LoadClientConfig loads the ClientConfig from the specified file (if specified)
// or from environment variables (if the file is not specified).
func LoadClientConfigFromEnvironment() (*CalicoAPIConfig, error) {
	c := NewCalicoAPIConfig()

	// Load client config from environment variables.
	log.Debug("Loading config from environment")
	if err := envconfig.Process("calico", &c.Spec); err != nil {
		return nil, err
	}

	return c, nil
}
