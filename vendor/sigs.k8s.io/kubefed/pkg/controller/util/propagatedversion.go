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
	"fmt"
	"reflect"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	fedv1a1 "sigs.k8s.io/kubefed/pkg/apis/core/v1alpha1"
)

const (
	generationPrefix      = "gen:"
	resourceVersionPrefix = "rv:"
)

// ObjectVersion retrieves the field type-prefixed value used for
// determining currency of the given cluster object.
func ObjectVersion(clusterObj *unstructured.Unstructured) string {
	generation := clusterObj.GetGeneration()
	if generation != 0 {
		return fmt.Sprintf("%s%d", generationPrefix, generation)
	}
	return fmt.Sprintf("%s%s", resourceVersionPrefix, clusterObj.GetResourceVersion())
}

// ObjectNeedsUpdate determines whether the 2 objects provided cluster
// object needs to be updated according to the desired object and the
// recorded version.
func ObjectNeedsUpdate(desiredObj, clusterObj *unstructured.Unstructured, recordedVersion string) bool {
	targetVersion := ObjectVersion(clusterObj)

	if recordedVersion != targetVersion {
		return true
	}

	// If versions match and the version is sourced from the
	// generation field, a further check of metadata equivalency is
	// required.
	return strings.HasPrefix(targetVersion, generationPrefix) && !ObjectMetaObjEquivalent(desiredObj, clusterObj)
}

// SortClusterVersions ASCII sorts the given cluster versions slice
// based on cluster name.
func SortClusterVersions(versions []fedv1a1.ClusterObjectVersion) {
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].ClusterName < versions[j].ClusterName
	})
}

// PropagatedVersionStatusEquivalent returns true if both statuses are equal by
// comparing Template and Override version, and their ClusterVersion slices;
// false otherwise.
func PropagatedVersionStatusEquivalent(pvs1, pvs2 *fedv1a1.PropagatedVersionStatus) bool {
	return pvs1.TemplateVersion == pvs2.TemplateVersion &&
		pvs1.OverrideVersion == pvs2.OverrideVersion &&
		reflect.DeepEqual(pvs1.ClusterVersions, pvs2.ClusterVersions)
}
