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

const typeFakeCache = "FAKE"

type fakeCacheFactory struct {
}

func (fc *fakeCacheFactory) Type() string {
	return typeFakeCache
}

// Create a fake cache. Just used for debug, compared to inMemoryCache,
// the fake cache has no timed memory cleanup. Do not use the fake cache
// in production environment or multi-replicas apiserver,
// which will cause risk of memory leaks and data inconsistency.
func (fc *fakeCacheFactory) Create(options DynamicOptions, stopCh <-chan struct{}) (Interface, error) {
	return NewInMemoryCache(nil, nil)
}

func init() {
	RegisterCacheFactory(&fakeCacheFactory{})
}
