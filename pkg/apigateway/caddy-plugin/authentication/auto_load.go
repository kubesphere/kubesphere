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
package authentication

import (
	"fmt"
	"strings"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"

	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/signals"
)

func init() {
	caddy.RegisterPlugin("authentication", caddy.Plugin{
		ServerType: "http",
		Action:     Setup,
	})
}

// Setup is called by Caddy to parse the config block
func Setup(c *caddy.Controller) error {

	rule, err := parse(c)

	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	c.OnStartup(func() error {
		stopChan := signals.SetupSignalHandler()
		informerFactory := informers.SharedInformerFactory()
		informerFactory.Rbac().V1().Roles().Lister()
		informerFactory.Rbac().V1().RoleBindings().Lister()
		informerFactory.Rbac().V1().ClusterRoles().Lister()
		informerFactory.Rbac().V1().ClusterRoleBindings().Lister()
		informerFactory.Start(stopChan)
		informerFactory.WaitForCacheSync(stopChan)
		fmt.Println("Authentication middleware is initiated")
		return nil
	})

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return &Authentication{Next: next, Rule: rule}
	})
	return nil
}

func parse(c *caddy.Controller) (Rule, error) {

	rule := Rule{ExceptedPath: make([]string, 0)}

	if c.Next() {
		args := c.RemainingArgs()
		switch len(args) {
		case 0:
			for c.NextBlock() {
				switch c.Val() {
				case "path":
					if !c.NextArg() {
						return rule, c.ArgErr()
					}

					rule.Path = c.Val()

					if c.NextArg() {
						return rule, c.ArgErr()
					}

					break
				case "except":
					if !c.NextArg() {
						return rule, c.ArgErr()
					}

					rule.ExceptedPath = strings.Split(c.Val(), ",")

					for i := 0; i < len(rule.ExceptedPath); i++ {
						rule.ExceptedPath[i] = strings.TrimSpace(rule.ExceptedPath[i])
					}

					if c.NextArg() {
						return rule, c.ArgErr()
					}
					break
				}
			}
		case 1:
			rule.Path = args[0]
			if c.NextBlock() {
				return rule, c.ArgErr()
			}
		default:
			return rule, c.ArgErr()
		}
	}

	if c.Next() {
		return rule, c.ArgErr()
	}

	return rule, nil
}
