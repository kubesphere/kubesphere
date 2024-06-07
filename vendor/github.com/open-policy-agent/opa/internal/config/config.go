// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package config implements helper functions to parse OPA's configuration.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/open-policy-agent/opa/internal/strvals"
	"github.com/open-policy-agent/opa/keys"
	"github.com/open-policy-agent/opa/logging"
	"github.com/open-policy-agent/opa/plugins/rest"
	"github.com/open-policy-agent/opa/tracing"
	"github.com/open-policy-agent/opa/util"
)

// ServiceOptions stores the options passed to ParseServicesConfig
type ServiceOptions struct {
	Raw                   json.RawMessage
	AuthPlugin            rest.AuthPluginLookupFunc
	Keys                  map[string]*keys.Config
	Logger                logging.Logger
	DistributedTacingOpts tracing.Options
}

// ParseServicesConfig returns a set of named service clients. The service
// clients can be specified either as an array or as a map. Some systems (e.g.,
// Helm) do not have proper support for configuration values nested under
// arrays, so just support both here.
func ParseServicesConfig(opts ServiceOptions) (map[string]rest.Client, error) {

	services := map[string]rest.Client{}

	var arr []json.RawMessage
	var obj map[string]json.RawMessage

	if err := util.Unmarshal(opts.Raw, &arr); err == nil {
		for _, s := range arr {
			client, err := rest.New(s, opts.Keys, rest.AuthPluginLookup(opts.AuthPlugin), rest.Logger(opts.Logger), rest.DistributedTracingOpts(opts.DistributedTacingOpts))
			if err != nil {
				return nil, err
			}
			services[client.Service()] = client
		}
	} else if util.Unmarshal(opts.Raw, &obj) == nil {
		for k := range obj {
			client, err := rest.New(obj[k], opts.Keys, rest.Name(k), rest.AuthPluginLookup(opts.AuthPlugin), rest.Logger(opts.Logger), rest.DistributedTracingOpts(opts.DistributedTacingOpts))
			if err != nil {
				return nil, err
			}
			services[client.Service()] = client
		}
	} else {
		// Return error from array decode as that is the default format.
		return nil, err
	}

	return services, nil
}

// Load implements configuration file loading. The supplied config file will be
// read from disk (if specified) and overrides will be applied. If no config file is
// specified, the overrides can still be applied to an empty config.
func Load(configFile string, overrides []string, overrideFiles []string) ([]byte, error) {
	baseConf := map[string]interface{}{}

	// User specified config file
	if configFile != "" {
		var bytes []byte
		var err error
		bytes, err = os.ReadFile(configFile)
		if err != nil {
			return nil, err
		}

		processedConf := subEnvVars(string(bytes))

		if err := yaml.Unmarshal([]byte(processedConf), &baseConf); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %s", configFile, err)
		}
	}

	overrideConf := map[string]interface{}{}

	// User specified a config override via --set
	for _, override := range overrides {
		processedOverride := subEnvVars(override)
		if err := strvals.ParseInto(processedOverride, overrideConf); err != nil {
			return nil, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	// User specified a config override value via --set-file
	for _, override := range overrideFiles {
		reader := func(rs []rune) (interface{}, error) {
			bytes, err := os.ReadFile(string(rs))
			value := strings.TrimSpace(string(bytes))
			return value, err
		}
		if err := strvals.ParseIntoFile(override, overrideConf, reader); err != nil {
			return nil, fmt.Errorf("failed parsing --set-file data: %s", err)
		}
	}

	// Merge together base config file and overrides, prefer the overrides
	conf := mergeValues(baseConf, overrideConf)

	// Take the patched config and marshal back to YAML
	return yaml.Marshal(conf)
}

// regex looking for ${...} notation strings
var envRegex = regexp.MustCompile(`(?U:\${.*})`)

// subEnvVars will look for any environment variables in the passed in string
// with the syntax of ${VAR_NAME} and replace that string with ENV[VAR_NAME]
func subEnvVars(s string) string {
	updatedConfig := envRegex.ReplaceAllStringFunc(s, func(s string) string {
		// Trim off the '${' and '}'
		if len(s) <= 3 {
			// This should never happen..
			return ""
		}
		varName := s[2 : len(s)-1]

		// Lookup the variable in the environment. We play by
		// bash rules.. if its undefined we'll treat it as an
		// empty string instead of raising an error.
		return os.Getenv(varName)
	})

	return updatedConfig
}

// mergeValues will merge source and destination map, preferring values from the source map
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}
