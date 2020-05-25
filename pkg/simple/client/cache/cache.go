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

package cache

import "time"

var NeverExpire = time.Duration(0)

type Interface interface {
	// Keys retrieves all keys match the given pattern
	Keys(pattern string) ([]string, error)

	// Get retrieves the value of the given key, return error if key doesn't exist
	Get(key string) (string, error)

	// Set sets the value and living duration of the given key, zero duration means never expire
	Set(key string, value string, duration time.Duration) error

	// Del deletes the given key, no error returned if the key doesn't exists
	Del(keys ...string) error

	// Exists checks the existence of a give key
	Exists(keys ...string) (bool, error)

	// Expires updates object's expiration time, return err if key doesn't exist
	Expire(key string, duration time.Duration) error
}
