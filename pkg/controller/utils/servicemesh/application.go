/*
Copyright 2020 KubeSphere Authors

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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AppLabel                     = "app"
	VersionLabel                 = "version"
	ApplicationNameLabel         = "app.kubernetes.io/name"
	ApplicationVersionLabel      = "app.kubernetes.io/version"
	ServiceMeshEnabledAnnotation = "servicemesh.kubesphere.io/enabled"
	SidecarInjectAnnotation      = "sidecar.istio.io/inject"
	DefaultDeploymentVersion     = "v1"
)

// Deployment should have these labels, to satisfy ServiceMesh Application
var DeploymentLabels = []string{
	ApplicationNameLabel,
	ApplicationVersionLabel,
	AppLabel,
	VersionLabel,
}

// resource with these following labels considered as part of servicemesh
var ApplicationLabels = []string{
	ApplicationNameLabel,
	ApplicationVersionLabel,
	AppLabel,
}

// resource with these following labels considered as part of kubernetes-sigs/application
var AppLabels = []string{
	ApplicationNameLabel,
	ApplicationVersionLabel,
}

var TrimChars = [...]string{".", "_", "-"}

func GetApplictionName(lbs map[string]string) string {
	if name, ok := lbs[ApplicationNameLabel]; ok {
		return name
	}
	return ""
}

func GetComponentVersion(meta *v1.ObjectMeta) string {
	if len(meta.Labels[VersionLabel]) > 0 {
		return meta.Labels[VersionLabel]
	}
	return ""
}

func ExtractApplicationLabels(lbs map[string]string) map[string]string {
	objLabels := make(map[string]string, len(ApplicationLabels))
	for _, label := range ApplicationLabels {
		if _, ok := lbs[label]; !ok {
			return nil
		} else {
			objLabels[label] = lbs[label]
		}
	}

	return objLabels
}

func IsApplicationComponent(lbs map[string]string, appLabels []string) bool {
	for _, label := range appLabels {
		if _, ok := lbs[label]; !ok {
			return false
		}
	}

	return true
}
