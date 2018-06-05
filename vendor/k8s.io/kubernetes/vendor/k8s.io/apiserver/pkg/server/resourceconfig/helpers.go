/*
Copyright 2017 The Kubernetes Authors.

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

package resourceconfig

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serverstore "k8s.io/apiserver/pkg/server/storage"
	utilflag "k8s.io/apiserver/pkg/util/flag"
)

// GroupVersionRegistry provides access to registered group versions.
type GroupVersionRegistry interface {
	// IsRegistered returns true if given group is registered.
	IsRegistered(group string) bool
	// IsRegisteredVersion returns true if given version is registered.
	IsRegisteredVersion(v schema.GroupVersion) bool
	// RegisteredGroupVersions returns all registered group versions.
	RegisteredGroupVersions() []schema.GroupVersion
}

// MergeResourceEncodingConfigs merges the given defaultResourceConfig with specific GroupVersionResource overrides.
func MergeResourceEncodingConfigs(
	defaultResourceEncoding *serverstore.DefaultResourceEncodingConfig,
	resourceEncodingOverrides []schema.GroupVersionResource,
) *serverstore.DefaultResourceEncodingConfig {
	resourceEncodingConfig := defaultResourceEncoding
	for _, gvr := range resourceEncodingOverrides {
		resourceEncodingConfig.SetResourceEncoding(gvr.GroupResource(), gvr.GroupVersion(),
			schema.GroupVersion{Group: gvr.Group, Version: runtime.APIVersionInternal})
	}
	return resourceEncodingConfig
}

// MergeGroupEncodingConfigs merges the given defaultResourceConfig with specific GroupVersion overrides.
func MergeGroupEncodingConfigs(
	defaultResourceEncoding *serverstore.DefaultResourceEncodingConfig,
	storageEncodingOverrides map[string]schema.GroupVersion,
) *serverstore.DefaultResourceEncodingConfig {
	resourceEncodingConfig := defaultResourceEncoding
	for group, storageEncodingVersion := range storageEncodingOverrides {
		resourceEncodingConfig.SetVersionEncoding(group, storageEncodingVersion, schema.GroupVersion{Group: group, Version: runtime.APIVersionInternal})
	}
	return resourceEncodingConfig
}

// MergeAPIResourceConfigs merges the given defaultAPIResourceConfig with the given resourceConfigOverrides.
// Exclude the groups not registered in registry, and check if version is
// not registered in group, then it will fail.
func MergeAPIResourceConfigs(
	defaultAPIResourceConfig *serverstore.ResourceConfig,
	resourceConfigOverrides utilflag.ConfigurationMap,
	registry GroupVersionRegistry,
) (*serverstore.ResourceConfig, error) {
	resourceConfig := defaultAPIResourceConfig
	overrides := resourceConfigOverrides

	// "api/all=false" allows users to selectively enable specific api versions.
	allAPIFlagValue, ok := overrides["api/all"]
	if ok {
		if allAPIFlagValue == "false" {
			// Disable all group versions.
			resourceConfig.DisableAll()
		} else if allAPIFlagValue == "true" {
			resourceConfig.EnableAll()
		}
	}

	// "<resourceSpecifier>={true|false} allows users to enable/disable API.
	// This takes preference over api/all, if specified.
	// Iterate through all group/version overrides specified in runtimeConfig.
	for key := range overrides {
		// Have already handled them above. Can skip them here.
		if key == "api/all" {
			continue
		}

		tokens := strings.Split(key, "/")
		if len(tokens) != 2 {
			continue
		}
		groupVersionString := tokens[0] + "/" + tokens[1]
		groupVersion, err := schema.ParseGroupVersion(groupVersionString)
		if err != nil {
			return nil, fmt.Errorf("invalid key %s", key)
		}

		// Exclude group not registered into the registry.
		if !registry.IsRegistered(groupVersion.Group) {
			continue
		}

		// Verify that the groupVersion is registered into registry.
		if !registry.IsRegisteredVersion(groupVersion) {
			return nil, fmt.Errorf("group version %s that has not been registered", groupVersion.String())
		}
		enabled, err := getRuntimeConfigValue(overrides, key, false)
		if err != nil {
			return nil, err
		}
		if enabled {
			resourceConfig.EnableVersions(groupVersion)
		} else {
			resourceConfig.DisableVersions(groupVersion)
		}
	}

	return resourceConfig, nil
}

func getRuntimeConfigValue(overrides utilflag.ConfigurationMap, apiKey string, defaultValue bool) (bool, error) {
	flagValue, ok := overrides[apiKey]
	if ok {
		if flagValue == "" {
			return true, nil
		}
		boolValue, err := strconv.ParseBool(flagValue)
		if err != nil {
			return false, fmt.Errorf("invalid value of %s: %s, err: %v", apiKey, flagValue, err)
		}
		return boolValue, nil
	}
	return defaultValue, nil
}

// ParseGroups takes in resourceConfig and returns parsed groups.
func ParseGroups(resourceConfig utilflag.ConfigurationMap) ([]string, error) {
	groups := []string{}
	for key := range resourceConfig {
		if key == "api/all" {
			continue
		}
		tokens := strings.Split(key, "/")
		if len(tokens) != 2 && len(tokens) != 3 {
			return groups, fmt.Errorf("runtime-config invalid key %s", key)
		}
		groupVersionString := tokens[0] + "/" + tokens[1]
		groupVersion, err := schema.ParseGroupVersion(groupVersionString)
		if err != nil {
			return nil, fmt.Errorf("runtime-config invalid key %s", key)
		}
		groups = append(groups, groupVersion.Group)
	}

	return groups, nil
}
