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
	"github.com/go-redis/redis"
	"k8s.io/klog"
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

func NewRedisClient(option *RedisOptions, stopCh <-chan struct{}) (*RedisClient, error) {
	var r RedisClient

	options, err := redis.ParseURL(option.RedisURL)

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
