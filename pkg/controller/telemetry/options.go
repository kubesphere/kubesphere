/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package telemetry

import (
	"fmt"

	"kubesphere.io/kubesphere/pkg/constants"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

const (
	ConfigName    = "io.kubesphere.config.platformconfig.telemetry"
	ConfigDataKey = "configuration.yaml"
)

// PlatformOptions store in constants.PlatformConfigurationName by hot loading.
type TelemetryOptions struct {
	// should enable the telemetry.
	Enabled bool `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`
	// Endpoint for kubesphere cloud
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty" mapstructure:"endpoint"`
	// collect period
	// The schedule in telemetry clusterInfo format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule,omitempty" yaml:"schedule,omitempty" mapstructure:"schedule"`
}

func NewTelemetryOptions() *TelemetryOptions {
	return &TelemetryOptions{
		Schedule: "0 1 * * *", // 1:00 each day
	}
}

// LoadPlatformConfig from given ConfigMap.
func LoadTelemetryConfig(secret *corev1.Secret) (*TelemetryOptions, error) {
	value, ok := secret.Data[constants.GenericPlatformConfigFileName]
	if !ok {
		return nil, fmt.Errorf("failed to get config %s from secret %s value", ConfigDataKey, ConfigName)
	}
	o := &TelemetryOptions{}
	if err := yaml.Unmarshal(value, o); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value from configmap. err: %s", err)
	}
	return o, nil
}
