/*
Copyright 2016 The Kubernetes Authors.

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

package set

import (
	"fmt"
	"io"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

// selectContainers allows one or more containers to be matched against a string or wildcard
func selectContainers(containers []v1.Container, spec string) ([]*v1.Container, []*v1.Container) {
	out := []*v1.Container{}
	skipped := []*v1.Container{}
	for i, c := range containers {
		if selectString(c.Name, spec) {
			out = append(out, &containers[i])
		} else {
			skipped = append(skipped, &containers[i])
		}
	}
	return out, skipped
}

// handlePodUpdateError prints a more useful error to the end user when mutating a pod.
func handlePodUpdateError(out io.Writer, err error, resource string) {
	if statusError, ok := err.(*errors.StatusError); ok && errors.IsInvalid(err) {
		errorDetails := statusError.Status().Details
		if errorDetails.Kind == "Pod" {
			all, match := true, false
			for _, cause := range errorDetails.Causes {
				if cause.Field == "spec" && strings.Contains(cause.Message, "may not update fields other than") {
					fmt.Fprintf(out, "error: may not update %s in pod %q directly\n", resource, errorDetails.Name)
					match = true
				} else {
					all = false
				}
			}
			if all && match {
				return
			}
		} else {
			if ok := cmdutil.PrintErrorWithCauses(err, out); ok {
				return
			}
		}
	}

	fmt.Fprintf(out, "error: %v\n", err)
}

// selectString returns true if the provided string matches spec, where spec is a string with
// a non-greedy '*' wildcard operator.
// TODO: turn into a regex and handle greedy matches and backtracking.
func selectString(s, spec string) bool {
	if spec == "*" {
		return true
	}
	if !strings.Contains(spec, "*") {
		return s == spec
	}

	pos := 0
	match := true
	parts := strings.Split(spec, "*")
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		next := strings.Index(s[pos:], part)
		switch {
		// next part not in string
		case next < pos:
			fallthrough
		// first part does not match start of string
		case i == 0 && pos != 0:
			fallthrough
		// last part does not exactly match remaining part of string
		case i == (len(parts)-1) && len(s) != (len(part)+next):
			match = false
			break
		default:
			pos = next
		}
	}
	return match
}

// Patch represents the result of a mutation to an object.
type Patch struct {
	Info *resource.Info
	Err  error

	Before []byte
	After  []byte
	Patch  []byte
}

// patchFn is a function type that accepts an info object and returns a byte slice.
// Implementations of patchFn should update the object and return it encoded.
type patchFn func(*resource.Info) ([]byte, error)

// CalculatePatch calls the mutation function on the provided info object, and generates a strategic merge patch for
// the changes in the object. Encoder must be able to encode the info into the appropriate destination type.
// This function returns whether the mutation function made any change in the original object.
func CalculatePatch(patch *Patch, encoder runtime.Encoder, mutateFn patchFn) bool {
	patch.Before, patch.Err = runtime.Encode(encoder, patch.Info.Object)
	patch.After, patch.Err = mutateFn(patch.Info)
	if patch.Err != nil {
		return true
	}
	if patch.After == nil {
		return false
	}

	patch.Patch, patch.Err = strategicpatch.CreateTwoWayMergePatch(patch.Before, patch.After, patch.Info.Object)
	return true
}

// CalculatePatches calculates patches on each provided info object. If the provided mutateFn
// makes no change in an object, the object is not included in the final list of patches.
func CalculatePatches(infos []*resource.Info, encoder runtime.Encoder, mutateFn patchFn) []*Patch {
	var patches []*Patch
	for _, info := range infos {
		patch := &Patch{Info: info}
		if CalculatePatch(patch, encoder, mutateFn) {
			patches = append(patches, patch)
		}
	}
	return patches
}

func findEnv(env []v1.EnvVar, name string) (v1.EnvVar, bool) {
	for _, e := range env {
		if e.Name == name {
			return e, true
		}
	}
	return v1.EnvVar{}, false
}

func updateEnv(existing []v1.EnvVar, env []v1.EnvVar, remove []string) []v1.EnvVar {
	out := []v1.EnvVar{}
	covered := sets.NewString(remove...)
	for _, e := range existing {
		if covered.Has(e.Name) {
			continue
		}
		newer, ok := findEnv(env, e.Name)
		if ok {
			covered.Insert(e.Name)
			out = append(out, newer)
			continue
		}
		out = append(out, e)
	}
	for _, e := range env {
		if covered.Has(e.Name) {
			continue
		}
		covered.Insert(e.Name)
		out = append(out, e)
	}
	return out
}
