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
	"flag"
	"log"
	"sync"

	"github.com/go-redis/redis"
)

var (
	redisHost       string
	redisPassword   string
	redisDB         int
	redisClientOnce sync.Once
	redisClient     *redis.Client
)

func init() {
	flag.StringVar(&redisHost, "redis-server", "localhost:6379", "redis server host")
	flag.StringVar(&redisPassword, "redis-password", "", "redis password")
	flag.IntVar(&redisDB, "redis-db", 0, "redis db")
}

func RedisClient() *redis.Client {

	redisClientOnce.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisHost,
			Password: redisPassword,
			DB:       redisDB,
		})
		if err := redisClient.Ping().Err(); err != nil {
			log.Fatalln(err)
		}
	})

	return redisClient
}
