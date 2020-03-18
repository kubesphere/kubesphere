/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package authorizerfactory

import "kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"

type opaAuthorizer struct{}

func (opaAuthorizer) Authorize(a authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {
	// TODO implement.
	return authorizer.DecisionAllow, "", nil
}

func NewOPAAuthorizer() *opaAuthorizer {
	return new(opaAuthorizer)
}
