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
package client

import (
	"os"
	"strconv"
	"sync"

	"github.com/go-redis/redis"
)

const (
	redisHostEnv     = "REDIS_HOST"
	redisPasswordEnv = "REDIS_PASSWORD"
	redisDbEnv       = "REDIS_DB"
)

var (
	redisHost     = "localhost:6379"
	redisPassword = ""
	redisDB       = 0
	once          sync.Once
	redisClient   *redis.Client
)

func init() {
	if env := os.Getenv(redisHostEnv); env != "" {
		redisHost = env
	}
	if env := os.Getenv(redisPasswordEnv); env != "" {
		redisPassword = env
	}
	if env := os.Getenv(redisDbEnv); env != "" {
		if i, err := strconv.Atoi(env); err == nil {
			redisDB = i
		}
	}
}

func RedisClient() *redis.Client {

	once.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisHost,
			Password: redisPassword,
			DB:       redisDB,
		})
	})

	return redisClient
}
