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

package proxy

import (
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
)

// NewAuthorizer returns an authorizer which accepts cluster proxy request.
// If multi-cluster mode is enabled, request should authorize by target apiserver.
func NewAuthorizer(multiClusterEnabled bool) authorizer.Authorizer {
	return authorizer.AuthorizerFunc(func(a authorizer.Attributes) (authorizer.Decision, string, error) {
		// in multi cluster mode, the request will be dispatch.
		if multiClusterEnabled && a.GetCluster() != "" {
			return authorizer.DecisionAllow, "", nil
		}
		return authorizer.DecisionNoOpinion, "", nil
	})
}
