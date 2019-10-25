/*
 *
 * Copyright 2019 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package internal

import "net/http"

const AllMethod = "*"

var HttpMethods = []string{AllMethod, http.MethodPost, http.MethodDelete,
	http.MethodPatch, http.MethodPut, http.MethodGet, http.MethodOptions, http.MethodConnect}

// Path exclusion rule
type ExclusionRule struct {
	Method string
	Path   string
}
