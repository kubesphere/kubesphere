// Copyright 2019 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apigateway

import (
	"github.com/mholt/caddy"

	"kubesphere.io/kubesphere/pkg/apigateway/caddy-plugin/authenticate"
	"kubesphere.io/kubesphere/pkg/apigateway/caddy-plugin/authentication"
	swagger "kubesphere.io/kubesphere/pkg/apigateway/caddy-plugin/swagger"
)

func RegisterPlugins() {
	caddy.RegisterPlugin("swagger", caddy.Plugin{
		ServerType: "http",
		Action:     swagger.Setup,
	})

	caddy.RegisterPlugin("authenticate", caddy.Plugin{
		ServerType: "http",
		Action:     authenticate.Setup,
	})

	caddy.RegisterPlugin("authentication", caddy.Plugin{
		ServerType: "http",
		Action:     authentication.Setup,
	})
}
