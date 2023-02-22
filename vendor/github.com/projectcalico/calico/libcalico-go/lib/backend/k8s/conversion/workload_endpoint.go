// Copyright (c) 2016-2020 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// TODO move the WorkloadEndpoint converters to is own package. Some refactoring of the annotation and label constants
// is necessary to avoid circular imports, which is why this has been deferred.
package conversion

import (
	kapiv1 "k8s.io/api/core/v1"

	"github.com/projectcalico/calico/libcalico-go/lib/backend/model"
)

type WorkloadEndpointConverter interface {
	VethNameForWorkload(namespace, podName string) string
	PodToWorkloadEndpoints(pod *kapiv1.Pod) ([]*model.KVPair, error)
}

func NewWorkloadEndpointConverter() WorkloadEndpointConverter {
	return &defaultWorkloadEndpointConverter{}
}
