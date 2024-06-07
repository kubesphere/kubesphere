// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package config implements OPA configuration file parsing and validation.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/internal/ref"
	"github.com/open-policy-agent/opa/util"
	"github.com/open-policy-agent/opa/version"
)

// Config represents the configuration file that OPA can be started with.
type Config struct {
	Services                     json.RawMessage            `json:"services,omitempty"`
	Labels                       map[string]string          `json:"labels,omitempty"`
	Discovery                    json.RawMessage            `json:"discovery,omitempty"`
	Bundle                       json.RawMessage            `json:"bundle,omitempty"` // Deprecated: Use `bundles` instead
	Bundles                      json.RawMessage            `json:"bundles,omitempty"`
	DecisionLogs                 json.RawMessage            `json:"decision_logs,omitempty"`
	Status                       json.RawMessage            `json:"status,omitempty"`
	Plugins                      map[string]json.RawMessage `json:"plugins,omitempty"`
	Keys                         json.RawMessage            `json:"keys,omitempty"`
	DefaultDecision              *string                    `json:"default_decision,omitempty"`
	DefaultAuthorizationDecision *string                    `json:"default_authorization_decision,omitempty"`
	Caching                      json.RawMessage            `json:"caching,omitempty"`
	NDBuiltinCache               bool                       `json:"nd_builtin_cache,omitempty"`
	PersistenceDirectory         *string                    `json:"persistence_directory,omitempty"`
	DistributedTracing           json.RawMessage            `json:"distributed_tracing,omitempty"`
	Server                       *struct {
		Encoding json.RawMessage `json:"encoding,omitempty"`
		Metrics  json.RawMessage `json:"metrics,omitempty"`
	} `json:"server,omitempty"`
	Storage *struct {
		Disk json.RawMessage `json:"disk,omitempty"`
	} `json:"storage,omitempty"`
	Extra map[string]json.RawMessage `json:"-"`
}

// ParseConfig returns a valid Config object with defaults injected. The id
// and version parameters will be set in the labels map.
func ParseConfig(raw []byte, id string) (*Config, error) {
	// NOTE(sr): based on https://stackoverflow.com/a/33499066/993018
	var result Config
	objValue := reflect.ValueOf(&result).Elem()
	knownFields := map[string]reflect.Value{}
	for i := 0; i != objValue.NumField(); i++ {
		jsonName := strings.Split(objValue.Type().Field(i).Tag.Get("json"), ",")[0]
		knownFields[jsonName] = objValue.Field(i)
	}

	if err := util.Unmarshal(raw, &result.Extra); err != nil {
		return nil, err
	}

	for key, chunk := range result.Extra {
		if field, found := knownFields[key]; found {
			if err := util.Unmarshal(chunk, field.Addr().Interface()); err != nil {
				return nil, err
			}
			delete(result.Extra, key)
		}
	}
	if len(result.Extra) == 0 {
		result.Extra = nil
	}
	return &result, result.validateAndInjectDefaults(id)
}

// PluginNames returns a sorted list of names of enabled plugins.
func (c Config) PluginNames() (result []string) {
	if c.Bundle != nil || c.Bundles != nil {
		result = append(result, "bundles")
	}
	if c.Status != nil {
		result = append(result, "status")
	}
	if c.DecisionLogs != nil {
		result = append(result, "decision_logs")
	}
	for name := range c.Plugins {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

// PluginsEnabled returns true if one or more plugin features are enabled.
//
// Deprecated. Use PluginNames instead.
func (c Config) PluginsEnabled() bool {
	return c.Bundle != nil || c.Bundles != nil || c.DecisionLogs != nil || c.Status != nil || len(c.Plugins) > 0
}

// DefaultDecisionRef returns the default decision as a reference.
func (c Config) DefaultDecisionRef() ast.Ref {
	r, _ := ref.ParseDataPath(*c.DefaultDecision)
	return r
}

// DefaultAuthorizationDecisionRef returns the default authorization decision
// as a reference.
func (c Config) DefaultAuthorizationDecisionRef() ast.Ref {
	r, _ := ref.ParseDataPath(*c.DefaultAuthorizationDecision)
	return r
}

// NDBuiltinCacheEnabled returns if the ND builtins cache should be used.
func (c Config) NDBuiltinCacheEnabled() bool {
	return c.NDBuiltinCache
}

func (c *Config) validateAndInjectDefaults(id string) error {

	if c.DefaultDecision == nil {
		s := defaultDecisionPath
		c.DefaultDecision = &s
	}

	_, err := ref.ParseDataPath(*c.DefaultDecision)
	if err != nil {
		return err
	}

	if c.DefaultAuthorizationDecision == nil {
		s := defaultAuthorizationDecisionPath
		c.DefaultAuthorizationDecision = &s
	}

	_, err = ref.ParseDataPath(*c.DefaultAuthorizationDecision)
	if err != nil {
		return err
	}

	if c.Labels == nil {
		c.Labels = map[string]string{}
	}

	c.Labels["id"] = id
	c.Labels["version"] = version.Version

	return nil
}

// GetPersistenceDirectory returns the configured persistence directory, or $PWD/.opa if none is configured
func (c Config) GetPersistenceDirectory() (string, error) {
	if c.PersistenceDirectory == nil {
		pwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(pwd, ".opa"), nil
	}
	return *c.PersistenceDirectory, nil
}

// ActiveConfig returns OPA's active configuration
// with the credentials and crypto keys removed
func (c *Config) ActiveConfig() (interface{}, error) {
	bs, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := util.UnmarshalJSON(bs, &result); err != nil {
		return nil, err
	}
	for k, e := range c.Extra {
		var v any
		if err := util.UnmarshalJSON(e, &v); err != nil {
			return nil, err
		}
		result[k] = v
	}

	if err := removeServiceCredentials(result["services"]); err != nil {
		return nil, err
	}

	if err := removeCryptoKeys(result["keys"]); err != nil {
		return nil, err
	}

	return result, nil
}

func removeServiceCredentials(x interface{}) error {
	switch x := x.(type) {
	case nil:
		return nil
	case []interface{}:
		for _, v := range x {
			err := removeKey(v, "credentials")
			if err != nil {
				return err
			}
		}

	case map[string]interface{}:
		for _, v := range x {
			err := removeKey(v, "credentials")
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("illegal service config type: %T", x)
	}

	return nil
}

func removeCryptoKeys(x interface{}) error {
	switch x := x.(type) {
	case nil:
		return nil
	case map[string]interface{}:
		for _, v := range x {
			err := removeKey(v, "key", "private_key")
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("illegal keys config type: %T", x)
	}

	return nil
}

func removeKey(x interface{}, keys ...string) error {
	val, ok := x.(map[string]interface{})
	if !ok {
		return fmt.Errorf("type assertion error")
	}

	for _, key := range keys {
		delete(val, key)
	}

	return nil
}

const (
	defaultDecisionPath              = "/system/main"
	defaultAuthorizationDecisionPath = "/system/authz/allow"
)
