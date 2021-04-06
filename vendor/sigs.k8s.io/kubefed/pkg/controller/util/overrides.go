/*
Copyright 2018 The Kubernetes Authors.

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

package util

import (
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
)

type ClusterOverride struct {
	Op    string      `json:"op,omitempty"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

type GenericOverrideItem struct {
	ClusterName      string            `json:"clusterName"`
	ClusterOverrides []ClusterOverride `json:"clusterOverrides,omitempty"`
}

type GenericOverrideSpec struct {
	Overrides []GenericOverrideItem `json:"overrides,omitempty"`
}

type GenericOverride struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec *GenericOverrideSpec `json:"spec,omitempty"`
}

// Namespace and name may not be overridden since these fields are the
// primary mechanism of association between a federated resource in
// the host cluster and the target resources in the member clusters.
//
// Kind should always be sourced from the FTC and not vary across
// member clusters.
//
// apiVersion can be overridden to support managing resources like
// Ingress which can exist in different groups at different
// versions. Users will need to take care not to abuse this
// capability.
var invalidPaths = sets.NewString(
	"/metadata/namespace",
	"/metadata/name",
	"/metadata/generateName",
	"/kind",
)

// Slice of ClusterOverride
type ClusterOverrides []ClusterOverride

// Mapping of clusterName to overrides for the cluster
type OverridesMap map[string]ClusterOverrides

// ToUnstructuredSlice converts the map of overrides to a slice of
// interfaces that can be set in an unstructured object.
func (m OverridesMap) ToUnstructuredSlice() []interface{} {
	overrides := []interface{}{}
	for clusterName, clusterOverrides := range m {
		overridesItem := map[string]interface{}{
			ClusterNameField:      clusterName,
			ClusterOverridesField: clusterOverrides,
		}
		overrides = append(overrides, overridesItem)
	}
	return overrides
}

// GetOverrides returns a map of overrides populated from the given
// unstructured object.
func GetOverrides(rawObj *unstructured.Unstructured) (OverridesMap, error) {
	overridesMap := make(OverridesMap)

	if rawObj == nil {
		return overridesMap, nil
	}

	genericFedObject := GenericOverride{}
	err := UnstructuredToInterface(rawObj, &genericFedObject)
	if err != nil {
		return nil, err
	}

	if genericFedObject.Spec == nil || genericFedObject.Spec.Overrides == nil {
		// No overrides defined for the federated type
		return overridesMap, nil
	}

	for _, overrideItem := range genericFedObject.Spec.Overrides {
		clusterName := overrideItem.ClusterName
		if _, ok := overridesMap[clusterName]; ok {
			return nil, errors.Errorf("cluster %q appears more than once", clusterName)
		}

		clusterOverrides := overrideItem.ClusterOverrides

		paths := sets.NewString()
		for i, clusterOverride := range clusterOverrides {
			path := clusterOverride.Path
			if invalidPaths.Has(path) {
				return nil, errors.Errorf("override[%d] for cluster %q has an invalid path: %s", i, clusterName, path)
			}
			if paths.Has(path) {
				return nil, errors.Errorf("path %q appears more than once for cluster %q", path, clusterName)
			}
			paths.Insert(path)
		}
		overridesMap[clusterName] = clusterOverrides
	}

	return overridesMap, nil
}

// SetOverrides sets the spec.overrides field of the unstructured
// object from the provided overrides map.
func SetOverrides(fedObject *unstructured.Unstructured, overridesMap OverridesMap) error {
	rawSpec := fedObject.Object[SpecField]
	if rawSpec == nil {
		rawSpec = map[string]interface{}{}
		fedObject.Object[SpecField] = rawSpec
	}

	spec, ok := rawSpec.(map[string]interface{})
	if !ok {
		return errors.Errorf("Unable to set overrides since %q is not an object: %T", SpecField, rawSpec)
	}
	spec[OverridesField] = overridesMap.ToUnstructuredSlice()
	return nil
}

// UnstructuredToInterface converts an unstructured object to the
// provided interface by json marshalling/unmarshalling.
func UnstructuredToInterface(rawObj *unstructured.Unstructured, obj interface{}) error {
	content, err := rawObj.MarshalJSON()
	if err != nil {
		return err
	}
	return json.Unmarshal(content, obj)
}

// ApplyJSONPatch applies the override on to the given unstructured object.
func ApplyJSONPatch(obj *unstructured.Unstructured, overrides ClusterOverrides) error {
	// TODO: Do the defaulting of "op" field to "replace" in API defaulting
	for i, overrideItem := range overrides {
		if overrideItem.Op == "" {
			overrides[i].Op = "replace"
		}
	}
	jsonPatchBytes, err := json.Marshal(overrides)
	if err != nil {
		return err
	}

	patch, err := jsonpatch.DecodePatch(jsonPatchBytes)
	if err != nil {
		return err
	}

	objectJSONBytes, err := obj.MarshalJSON()
	if err != nil {
		return err
	}

	patchedObjectJSONBytes, err := patch.Apply(objectJSONBytes)
	if err != nil {
		return err
	}

	err = obj.UnmarshalJSON(patchedObjectJSONBytes)
	return err
}
