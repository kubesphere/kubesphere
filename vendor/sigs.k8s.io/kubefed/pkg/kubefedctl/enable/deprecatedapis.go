/*
Copyright 2019 The Kubernetes Authors.

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

package enable

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
)

// Deprecated APIs removed in 1.16 will be served by current equivalent APIs
// https://kubernetes.io/blog/2019/07/18/api-deprecations-in-1-16/
//
// Only allow one of the equivalent APIs for federation to avoid the possibility
// of multiple sync controllers fighting to update the same resource
var equivalentAPIs = map[string][]schema.GroupVersion{
	"deployments": {
		{
			Group:   "apps",
			Version: "v1",
		},
		{
			Group:   "apps",
			Version: "v1beta1",
		},
		{
			Group:   "apps",
			Version: "v1beta2",
		},
		{
			Group:   "extensions",
			Version: "v1beta1",
		},
	},
	"daemonsets": {
		{
			Group:   "apps",
			Version: "v1",
		},
		{
			Group:   "apps",
			Version: "v1beta1",
		},
		{
			Group:   "apps",
			Version: "v1beta2",
		},
		{
			Group:   "extensions",
			Version: "v1beta1",
		},
	},
	"statefulsets": {
		{
			Group:   "apps",
			Version: "v1",
		},
		{
			Group:   "apps",
			Version: "v1beta1",
		},
		{
			Group:   "apps",
			Version: "v1beta2",
		},
	},
	"replicasets": {
		{
			Group:   "apps",
			Version: "v1",
		},
		{
			Group:   "apps",
			Version: "v1beta1",
		},
		{
			Group:   "apps",
			Version: "v1beta2",
		},
		{
			Group:   "extensions",
			Version: "v1beta1",
		},
	},
	"networkpolicies": {
		{
			Group:   "networking.k8s.io",
			Version: "v1",
		},
		{
			Group:   "extensions",
			Version: "v1beta1",
		},
	},
	"podsecuritypolicies": {
		{
			Group:   "policy",
			Version: "v1beta1",
		},
		{
			Group:   "extensions",
			Version: "v1beta1",
		},
	},
	"ingresses": {
		{
			Group:   "networking.k8s.io",
			Version: "v1beta1",
		},
		{
			Group:   "extensions",
			Version: "v1beta1",
		},
	},
}

func IsEquivalentAPI(existingAPI, newAPI *fedv1b1.APIResource) bool {
	if existingAPI.PluralName != newAPI.PluralName {
		return false
	}

	apis, ok := equivalentAPIs[existingAPI.PluralName]
	if !ok {
		return false
	}

	for _, gv := range apis {
		if gv.Group == existingAPI.Group && gv.Version == existingAPI.Version {
			// skip exactly matched API from equivalent API list
			continue
		}

		if gv.Group == newAPI.Group && gv.Version == newAPI.Version {
			return true
		}
	}

	return false
}
