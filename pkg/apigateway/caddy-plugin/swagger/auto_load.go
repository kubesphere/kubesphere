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
	"fmt"
	"net/http"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func Setup(c *caddy.Controller) error {

	handler, err := parse(c)

	if err != nil {
		return err
	}

	c.OnStartup(func() error {
		fmt.Println("Swagger middleware is initiated")
		return nil
	})

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return &Swagger{Next: next, Handler: handler}
	})

	return nil
}
func parse(c *caddy.Controller) (Handler, error) {

	handler := Handler{URL: "/swagger-ui", FilePath: "/var/static/swagger-ui"}

	if c.Next() {
		args := c.RemainingArgs()
		switch len(args) {
		case 0:
			for c.NextBlock() {
				switch c.Val() {
				case "url":
					if !c.NextArg() {
						return handler, c.ArgErr()
					}

					handler.URL = c.Val()

					if c.NextArg() {
						return handler, c.ArgErr()
					}
				case "filePath":
					if !c.NextArg() {
						return handler, c.ArgErr()
					}

					handler.FilePath = c.Val()

					if c.NextArg() {
						return handler, c.ArgErr()
					}
				default:
					return handler, c.ArgErr()
				}
			}
		default:
			return handler, c.ArgErr()
		}
	}

	if c.Next() {
		return handler, c.ArgErr()
	}

	handler.Handler = http.StripPrefix(handler.URL, http.FileServer(http.Dir(handler.FilePath)))

	return handler, nil
}
