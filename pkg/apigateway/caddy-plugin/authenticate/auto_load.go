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
	"github.com/go-redis/redis"
	"strings"
	"time"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func Setup(c *caddy.Controller) error {

	rule, err := parse(c)

	if err != nil {
		return err
	}

	rule.RedisClient = redis.NewClient(rule.RedisOptions)

	c.OnStartup(func() error {
		if err := rule.RedisClient.Ping().Err(); err != nil {
			return err
		}
		fmt.Println("Authenticate middleware is initiated")
		return nil
	})

	c.OnShutdown(func() error {
		return rule.RedisClient.Close()
	})

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return &Auth{Next: next, Rule: rule}
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
				case "token-idle-timeout":
					if !c.NextArg() {
						return rule, c.ArgErr()
					}

					if timeout, err := time.ParseDuration(c.Val()); err != nil {
						return rule, c.ArgErr()
					} else {
						rule.TokenIdleTimeout = timeout
					}

					if c.NextArg() {
						return rule, c.ArgErr()
					}
				case "redis-url":
					if !c.NextArg() {
						return rule, c.ArgErr()
					}

					if redisOptions, err := redis.ParseURL(c.Val()); err != nil {
						return rule, c.ArgErr()
					} else {
						rule.RedisOptions = redisOptions
					}

					if c.NextArg() {
						return rule, c.ArgErr()
					}
				case "secret":
					if !c.NextArg() {
						return rule, c.ArgErr()
					}

					rule.Secret = []byte(c.Val())

					if c.NextArg() {
						return rule, c.ArgErr()
					}
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
				}
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
