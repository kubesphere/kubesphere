// Copyright (c) 2017 Tigera, Inc. All rights reserved.

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

package conversion

const (
	NamespaceLabelPrefix            = "pcns."
	NamespaceProfileNamePrefix      = "kns."
	K8sNetworkPolicyNamePrefix      = "knp.default."
	ServiceAccountLabelPrefix       = "pcsa."
	ServiceAccountProfileNamePrefix = "ksa."

	// AnnotationPodIP is an annotation we apply to pods when assigning them an IP.  It
	// duplicates the value of the Pod.Status.PodIP field, which is set by kubelet but,
	// since we write it ourselves, we can make sure that it is written synchronously
	// and quickly.
	AnnotationPodIP = "cni.projectcalico.org/podIP"
)
