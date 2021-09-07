/*
Copyright 2020 The KubeSphere Authors.

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

package install

import (
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"

	devopsv1alpha1 "kubesphere.io/api/devops/v1alpha1"
	devopsv1alpha3 "kubesphere.io/api/devops/v1alpha3"
)

func Install(scheme *k8sruntime.Scheme) {
	urlruntime.Must(devopsv1alpha1.AddToScheme(scheme))
	urlruntime.Must(devopsv1alpha3.AddToScheme(scheme))
	urlruntime.Must(scheme.SetVersionPriority(devopsv1alpha3.SchemeGroupVersion, devopsv1alpha1.SchemeGroupVersion))
}
