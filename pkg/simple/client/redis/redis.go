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
package redis

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"k8s.io/klog"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClientOrDie(options *RedisOptions, stopCh <-chan struct{}) *RedisClient {
	client, err := NewRedisClient(options, stopCh)
	if err != nil {
		panic(err)
	}

	return client
}

// ref https://github.com/go-redis/redis/blob/v6.15.2/options.go#L163
func parseURL(redisURL string) (*redis.Options, error) {
	o := &redis.Options{Network: "tcp"}
	u, err := url.Parse(redisURL)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "redis" && u.Scheme != "rediss" {
		return nil, errors.New("invalid redis URL scheme: " + u.Scheme)
	}

	if u.User != nil {
		if p, ok := u.User.Password(); ok {
			o.Password = p
		}
	}

	if val := u.Query().Get("maxRetries"); val != "" {
		if maxRetries, err := strconv.Atoi(val); err != nil {
			return nil, errors.New("invalid options")
		} else {
			o.MaxRetries = maxRetries
		}
	}

	if val := u.Query().Get("minIdleConns"); val != "" {
		if minIdleConns, err := strconv.Atoi(val); err != nil {
			return nil, errors.New("invalid options")
		} else {
			o.MinIdleConns = minIdleConns
		}
	}

	if val := u.Query().Get("idleTimeout"); val != "" {
		if idleTimeout, err := time.ParseDuration(val); err != nil {
			return nil, errors.New("invalid options")
		} else {
			o.IdleTimeout = idleTimeout
		}
	}

	if val := u.Query().Get("idleCheckFrequency"); val != "" {
		if idleCheckFrequency, err := time.ParseDuration(val); err != nil {
			return nil, errors.New("invalid options")
		} else {
			o.IdleCheckFrequency = idleCheckFrequency
		}
	}

	if val := u.Query().Get("dialTimeout"); val != "" {
		if dialTimeout, err := time.ParseDuration(val); err != nil {
			return nil, errors.New("invalid options")
		} else {
			o.DialTimeout = dialTimeout
		}
	}

	if val := u.Query().Get("poolTimeout"); val != "" {
		if poolTimeout, err := time.ParseDuration(val); err != nil {
			return nil, errors.New("invalid options")
		} else {
			o.PoolTimeout = poolTimeout
		}
	}

	if val := u.Query().Get("poolSize"); val != "" {
		if poolSize, err := strconv.Atoi(val); err != nil {
			return nil, errors.New("invalid options")
		} else {
			o.PoolSize = poolSize
		}
	}

	if val := u.Query().Get("readTimeout"); val != "" {
		if readTimeout, err := time.ParseDuration(val); err != nil {
			return nil, errors.New("invalid options")
		} else {
			o.ReadTimeout = readTimeout
		}
	}

	if val := u.Query().Get("maxRetryBackoff"); val != "" {
		if maxRetryBackoff, err := time.ParseDuration(val); err != nil {
			return nil, errors.New("invalid options")
		} else {
			o.MaxRetryBackoff = maxRetryBackoff
		}
	}

	if val := u.Query().Get("maxConnAge"); val != "" {
		if maxConnAge, err := time.ParseDuration(val); err != nil {
			return nil, errors.New("invalid options")
		} else {
			o.MaxConnAge = maxConnAge
		}
	}

	h, p, err := net.SplitHostPort(u.Host)
	if err != nil {
		h = u.Host
	}
	if h == "" {
		h = "localhost"
	}
	if p == "" {
		p = "6379"
	}
	o.Addr = net.JoinHostPort(h, p)

	f := strings.FieldsFunc(u.Path, func(r rune) bool {
		return r == '/'
	})
	switch len(f) {
	case 0:
		o.DB = 0
	case 1:
		if o.DB, err = strconv.Atoi(f[0]); err != nil {
			return nil, fmt.Errorf("invalid redis database number: %q", f[0])
		}
	default:
		return nil, errors.New("invalid redis URL path: " + u.Path)
	}

	if u.Scheme == "rediss" {
		o.TLSConfig = &tls.Config{ServerName: h}
	}
	return o, nil
}

func NewRedisClient(option *RedisOptions, stopCh <-chan struct{}) (*RedisClient, error) {
	var r RedisClient

	options, err := parseURL(option.RedisURL)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	r.client = redis.NewClient(options)

	if err := r.client.Ping().Err(); err != nil {
		klog.Error("unable to reach redis host", err)
		r.client.Close()
		return nil, err
	}

	if stopCh != nil {
		go func() {
			<-stopCh
			if err := r.client.Close(); err != nil {
				klog.Error(err)
			}
		}()
	}

	return &r, nil
}

func (r *RedisClient) Redis() *redis.Client {
	return r.client
}
