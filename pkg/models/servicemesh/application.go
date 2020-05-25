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

package servicemesh

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	AppLabel                = "app"
	VersionLabel            = "version"
	ApplicationNameLabel    = "app.kubernetes.io/name"
	ApplicationVersionLabel = "app.kubernetes.io/version"
)

var ApplicationLabels = [...]string{
	ApplicationNameLabel,
	ApplicationVersionLabel,
	AppLabel,
}

var TrimChars = [...]string{".", "_", "-"}

// normalize version names
// strip [_.-]
func NormalizeVersionName(version string) string {
	for _, char := range TrimChars {
		version = strings.ReplaceAll(version, char, "")
	}
	return version
}

func GetComponentName(meta *metav1.ObjectMeta) string {
	if len(meta.Labels[AppLabel]) > 0 {
		return meta.Labels[AppLabel]
	}
	return ""
}

func GetComponentVersion(meta *metav1.ObjectMeta) string {
	if len(meta.Labels[VersionLabel]) > 0 {
		return meta.Labels[VersionLabel]
	}
	return ""
}

func ExtractApplicationLabels(meta *metav1.ObjectMeta) map[string]string {

	labels := make(map[string]string, 0)
	for _, label := range ApplicationLabels {
		if len(meta.Labels[label]) == 0 {
			return nil
		} else {
			labels[label] = meta.Labels[label]
		}
	}

	return labels
}

func IsApplicationComponent(meta *metav1.ObjectMeta) bool {

	for _, label := range ApplicationLabels {
		if len(meta.Labels[label]) == 0 {
			return false
		}
	}

	return true
}
