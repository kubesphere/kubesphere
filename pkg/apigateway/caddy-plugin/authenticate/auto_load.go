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
	"kubesphere.io/kubesphere/pkg/apigateway/caddy-plugin/internal"
	"kubesphere.io/kubesphere/pkg/simple/client/redis"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"time"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func Setup(c *caddy.Controller) error {

	rule, err := parse(c)

	if err != nil {
		return err
	}

	c.OnStartup(func() error {
		rule.RedisClient, err = redis.NewRedisClient(rule.RedisOptions, nil)
		// ensure redis is connected  when startup
		if err != nil {
			return err
		}
		fmt.Println("Authenticate middleware is initiated")
		return nil
	})

	c.OnShutdown(func() error {
		return rule.RedisClient.Redis().Close()
	})

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return &Auth{Next: next, Rule: rule}
	})

	return nil
}

func parse(c *caddy.Controller) (*Rule, error) {

	rule := &Rule{}
	rule.ExclusionRules = make([]internal.ExclusionRule, 0)
	if c.Next() {
		args := c.RemainingArgs()
		switch len(args) {
		case 0:
			for c.NextBlock() {
				switch c.Val() {
				case "path":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}

					rule.Path = c.Val()

					if c.NextArg() {
						return nil, c.ArgErr()
					}
				case "token-idle-timeout":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}

					if timeout, err := time.ParseDuration(c.Val()); err != nil {
						return nil, c.ArgErr()
					} else {
						rule.TokenIdleTimeout = timeout
					}

					if c.NextArg() {
						return nil, c.ArgErr()
					}
				case "redis-url":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}

					options := &redis.RedisOptions{RedisURL: c.Val()}

					if err := options.Validate(); len(err) > 0 {
						return nil, c.ArgErr()
					} else {
						rule.RedisOptions = options
					}

					if c.NextArg() {
						return nil, c.ArgErr()
					}
				case "secret":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}

					rule.Secret = []byte(c.Val())

					if c.NextArg() {
						return nil, c.ArgErr()
					}
				case "except":

					if !c.NextArg() {
						return nil, c.ArgErr()
					}

					method := c.Val()

					if !sliceutil.HasString(internal.HttpMethods, method) {
						return nil, c.ArgErr()
					}

					for c.NextArg() {
						path := c.Val()
						rule.ExclusionRules = append(rule.ExclusionRules, internal.ExclusionRule{Method: method, Path: path})
					}
				}
			}
		default:
			return nil, c.ArgErr()
		}
	}

	if c.Next() {
		return nil, c.ArgErr()
	}

	if rule.RedisOptions == nil {
		return nil, c.Err("redis-url must be specified")
	}

	return rule, nil
}
