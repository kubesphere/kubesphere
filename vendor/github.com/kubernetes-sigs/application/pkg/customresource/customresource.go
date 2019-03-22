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

package customresource

import (
	component "github.com/kubernetes-sigs/application/pkg/component"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"strings"
)

// Labels return custom resource label
func (hv *HandleValue) Labels() map[string]string {
	c := hv.Handle.(metav1.Object)
	return map[string]string{
		component.LabelCR:          strings.Trim(reflect.TypeOf(c).String(), "*"),
		component.LabelCRName:      c.GetName(),
		component.LabelCRNamespace: c.GetNamespace(),
	}
}
