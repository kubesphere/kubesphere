/*
Copyright 2019 The KubeSphere Authors.

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

package k8sutil

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func IsControlledBy(reference []metav1.OwnerReference, kind string, name string) bool {
	for _, ref := range reference {
		if ref.Kind == kind && (name == "" || ref.Name == name) {
			return true
		}
	}
	return false
}

func GetControlledWorkspace(reference []metav1.OwnerReference) string {
	for _, ref := range reference {
		if ref.Kind == "Workspace" {
			return ref.Name
		}
	}
	return ""
}
