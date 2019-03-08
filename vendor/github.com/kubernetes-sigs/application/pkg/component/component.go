/*
Copyright 2018 The Kubernetes Authors
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

package component

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"strings"
)

// Constants defining labels
const (
	LabelCR          = "custom-resource"
	LabelCRName      = "custom-resource-name"
	LabelCRNamespace = "custom-resource-namespace"
	LabelComponent   = "component"
)

// Labels return
func Labels(cr metav1.Object, component string) map[string]string {
	return map[string]string{
		LabelCR:          strings.Trim(reflect.TypeOf(cr).String(), "*"),
		LabelCRName:      cr.GetName(),
		LabelCRNamespace: cr.GetNamespace(),
		LabelComponent:   component,
	}
}

// Labels return the common labels for a resource
func (c *Component) Labels() map[string]string {
	return Labels(c.CR, c.Name)
}

// Merge is used to merge multiple maps into the target map
func (out KVMap) Merge(kvmaps ...KVMap) {
	for _, kvmap := range kvmaps {
		for k, v := range kvmap {
			out[k] = v
		}
	}
}
