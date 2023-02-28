// Copyright (c) 2017 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package names

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	cerrors "github.com/projectcalico/calico/libcalico-go/lib/errors"
)

// WorkloadEndpointIdentifiers is a collection of identifiers that are used to uniquely
// identify a WorkloadEndpoint resource.  Since a resource is identified by a single
// name field, Calico requires the name to be constructed in a very specific format.
// The format is dependent on the Orchestrator type:
// -  k8s:  <node>-k8s-<pod>-<endpoint>
// -  cni:  <node>-cni-<containerID>-<endpoint>
// -  libnetwork:  <node>-libnetwork-libnetwork-<endpoint>
// -  (other):  <node>-<orchestrator>-<workload>-<endpoint>
//
// Each parameter cannot start or end with a dash (-), and dashes within the parameter
// will be escaped to a double-dash (--) in the constructed name.
//
// List queries allow for prefix lists (for non-KDD), the client should verify that
// the items returned in the list match the supplied identifiers using the
// NameMatches() method.  This is necessary because a prefix match may return endpoints
// that do not exactly match the required identifiers.  For example, suppose you are
// querying endpoints with node=node1, orch=k8s, pod=pod and endpoints is wild carded:
//   - The name prefix would be `node1-k8s-pod-`
//   - A list query using that prefix would also return endpoints with, for example,
//     a pod call "pod-1", because the name of the endpoint might be `node1-k8s-pod--1-eth0`
//     which matches the required name prefix.
//
// The Node and Orchestrator are always required for both prefix and non-prefix name
// construction.
type WorkloadEndpointIdentifiers struct {
	Node         string
	Orchestrator string
	Endpoint     string
	Workload     string
	Pod          string
	ContainerID  string
}

// NameMatches returns true if the supplied WorkloadEndpoint name matches the
// supplied identifiers.
// This will return an error if the identifiers are not valid.
func (ids WorkloadEndpointIdentifiers) NameMatches(name string) (bool, error) {
	// Extract the required segments for this orchestrator type.
	req, err := ids.getSegments()
	if err != nil {
		return false, err
	}

	// Extract the parameters from the name.
	parts := ExtractDashSeparatedParms(name, len(req))
	if len(parts) == 0 {
		return false, nil
	}

	// Check each name segment for a non-match.
	for i, r := range req {
		if r.value != "" && r.value != parts[i] {
			return false, nil
		}
	}
	return true, nil
}

// CalculateWorkloadEndpointName calculates the expected name for a workload
// endpoint given the supplied Spec.  Calico requires a precise naming convention
// for workload endpoints that is based on orchestrator and various other orchestrator
// specific parameters.
//
// If allowPrefix is true, we construct the name prefix up to the last specified index
// and terminate with a dash.
func (ids WorkloadEndpointIdentifiers) CalculateWorkloadEndpointName(allowPrefix bool) (string, error) {
	req, err := ids.getSegments()
	if err != nil {
		return "", err
	}

	parts := []string{}
	for _, s := range req {
		part := ""

		if len(s.value) == 0 {
			// This segment has no value associated with it.
			if !allowPrefix {
				// We are not allowing prefixes.  This is an error scenario
				return "", cerrors.ErrorValidation{
					ErroredFields: []cerrors.ErroredField{
						{Name: s.field, Value: s.value, Reason: "field should be assigned"},
					},
				}
			}

			// We are allowing prefixes, so return the prefix that we have constructed thus far,
			// terminating with a "-".
			return strings.Join(parts, "-") + "-", nil
		}

		part, ef := escapeDashes(s)
		if ef != nil {
			return "", cerrors.ErrorValidation{ErroredFields: []cerrors.ErroredField{*ef}}
		}
		parts = append(parts, part)
	}

	// We have extracted all of the required segments, join the segments with a "-" and
	// return that as the name.
	return strings.Join(parts, "-"), nil
}

// getSegments returns the ID segments specific to the orchestrator.
func (ids WorkloadEndpointIdentifiers) getSegments() ([]segment, error) {
	node := segment{value: ids.Node, field: "node", structField: "Node"}
	orch := segment{value: ids.Orchestrator, field: "orchestrator", structField: "Orchestrator"}
	cont := segment{value: ids.ContainerID, field: "containerID", structField: "ContainerID"}
	pod := segment{value: ids.Pod, field: "pod", structField: "Pod"}
	endp := segment{value: ids.Endpoint, field: "endpoint", structField: "Endpoint"}
	workl := segment{value: ids.Workload, field: "workload", structField: "Workload"}

	// Node is *always* required.
	if len(node.value) == 0 {
		return nil, cerrors.ErrorValidation{
			ErroredFields: []cerrors.ErroredField{
				{Name: node.field, Reason: "field should be assigned"},
			},
		}
	}

	// Extract the segment values based on the orchestrator.
	var segments []segment
	switch orch.value {
	case "k8s":
		segments = []segment{node, orch, pod, endp}
	case "cni":
		segments = []segment{node, orch, cont, endp}
	case "libnetwork":
		segments = []segment{node, orch, orch, endp}
	default:
		segments = []segment{node, orch, workl, endp}
	}

	return segments, nil
}

// Segment contains the information of a single name segment.  The field names
// are geared towards the struct definition of the corresponding resource.
type segment struct {
	// The value of the field.
	value string
	// The JSON/YAML name of the corresponding field in the WorkloadEndpointSpec
	field string
	// The structure name of the corresponding field in the WorkloadEndpointSpec
	structField string
}

// escapeDashes replaces a single dash with a double dash.  This type of escaping is
// used for names constructed by joining a set of names with dashes - it assumes that
// each name segment cannot begin or end in a dash.
func escapeDashes(seg segment) (string, *cerrors.ErroredField) {
	if seg.value[0] == '-' {
		return "", &cerrors.ErroredField{Name: seg.field, Value: seg.value, Reason: "field must not begin with a '-'"}
	}
	if seg.value[len(seg.value)-1] == '-' {
		return "", &cerrors.ErroredField{Name: seg.field, Value: seg.value, Reason: "field must not end with a '-'"}
	}
	return strings.Replace(seg.value, "-", "--", -1), nil
}

func extractParts(name string) []string {
	parts := []string{}
	lastDash := -1
	for i := 1; i < len(name); i++ {
		// Skip non-dashes.
		if name[i] != '-' {
			continue
		}
		// Skip over double dashes
		if i < len(name)-1 && name[i+1] == '-' {
			i++
			continue
		}
		// This is a dash separator.
		parts = append(parts, strings.Replace(name[lastDash+1:i], "--", "-", -1))
		lastDash = i
	}
	// Add the last segment.
	parts = append(parts, strings.Replace(name[lastDash+1:], "--", "-", -1))

	return parts
}

// Extract the dash separated parms from the name.  Each parm will have had their dashes escaped,
// this also removes that escaping.  Returns nil if the parameters could not be extracted.
func ExtractDashSeparatedParms(name string, numParms int) []string {
	// The name must be at least as long as the number of parameters plus the separators.
	if len(name) < (2*numParms - 1) {
		return nil
	}

	parts := extractParts(name)

	// We should have extracted the correct number of name segments.
	if len(parts) != numParms {
		return nil
	}

	return parts
}

var (
	k8sFields        = []string{"Pod", "Endpoint"}
	cniFields        = []string{"ContainerID", "Endpoint"}
	libnetworkFields = []string{"Orchestrator", "Endpoint"}
	otherFields      = []string{"Workload", "Endpoint"}
)

// ParseWorkloadEndpointName parses a given name and returns a WorkloadEndpointIdentifiers
// instance with fields populated according to the WorkloadEndpoint name format.
func ParseWorkloadEndpointName(wepName string) (WorkloadEndpointIdentifiers, error) {
	if len(wepName) == 0 {
		return WorkloadEndpointIdentifiers{}, errors.New("Cannot parse empty string")
	}
	parts := extractParts(wepName)
	if parts == nil || len(parts) == 0 {
		return WorkloadEndpointIdentifiers{}, fmt.Errorf("Cannot parse %s", wepName)
	}
	pl := len(parts)
	weid := WorkloadEndpointIdentifiers{Node: parts[0]}
	if pl > 1 {
		weid.Orchestrator = parts[1]
		var orchFlds []string
		switch parts[1] {
		case "k8s":
			orchFlds = k8sFields
		case "cni":
			orchFlds = cniFields
		case "libnetwork":
			orchFlds = libnetworkFields
		default:
			orchFlds = otherFields
		}
		if pl > 2 {
			weidR := reflect.ValueOf(&weid)
			weidStruct := weidR.Elem()
			for i, part := range parts[2:] {
				fld := weidStruct.FieldByName(orchFlds[i])
				fld.SetString(part)
			}
		}
	}
	return weid, nil
}
