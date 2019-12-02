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
package authenticate

import (
	"net/http"

	"github.com/caddyserver/caddy/caddyhttp/httpserver"
)

type Swagger struct {
	Handler Handler
	Next    httpserver.Handler
}

type Handler struct {
	URL      string
	FilePath string
	Handler  http.Handler
}

func (h Swagger) ServeHTTP(resp http.ResponseWriter, req *http.Request) (int, error) {

	if httpserver.Path(req.URL.Path).Matches(h.Handler.URL) {
		h.Handler.Handler.ServeHTTP(resp, req)
		return http.StatusOK, nil
	}

	return h.Next.ServeHTTP(resp, req)
}
