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

package oauth

import "golang.org/x/oauth2"

type SimpleConfigManager struct {
}

func (s *SimpleConfigManager) Load(clientId string) (*oauth2.Config, error) {
	if clientId == "kubesphere-console-client" {
		return &oauth2.Config{
			ClientID:     "8b21fef43889a28f2bd6",
			ClientSecret: "xb21fef43889a28f2bd6",
			Endpoint:     oauth2.Endpoint{AuthURL: "http://ks-apiserver.kubesphere-system.svc/oauth/authorize", TokenURL: "http://ks-apiserver.kubesphere.io/oauth/token"},
			RedirectURL:  "http://ks-console.kubesphere-system.svc/oauth/token/implicit",
			Scopes:       nil,
		}, nil
	}
	return nil, ConfigNotFound
}
