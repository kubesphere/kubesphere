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

package esclient

import "time"

type Client interface {
	// Perform Search API
	Search(body []byte, scrollTimeout time.Duration) ([]byte, error)
	Scroll(scrollId string, scrollTimeout time.Duration) ([]byte, error)
	ClearScroll(scrollId string)
	GetTotalHitCount(v interface{}) int64
}
