// Copyright (c) 2017-2021 Tigera, Inc. All rights reserved.

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

package v3

const (
	// API group details for the Calico v3 API.
	Group               = "projectcalico.org"
	VersionCurrent      = "v3"
	GroupVersionCurrent = Group + "/" + VersionCurrent

	// AllNamepaces is used for client instantiation, either for when the namespace
	// will be specified in the resource request, or for List or Watch queries across
	// all namespaces.
	AllNamespaces = ""

	// AllNames is used for List or Watch queries to wildcard the name.
	AllNames = ""

	// Label used to denote the Namespace.  This is added to workload endpoints and network sets by Calico
	// and may be used for label matches by Policy selectors.
	LabelNamespace = "projectcalico.org/namespace"

	// Label used to denote the ServiceAccount.  This is added to the workload endpoints by Calico
	// and may be used for label matches by Policy selectors.
	LabelServiceAccount = "projectcalico.org/serviceaccount"

	// Label used to denote the Orchestrator.  This is added to the workload endpoints by an
	// orchestrator.
	LabelOrchestrator = "projectcalico.org/orchestrator"

	// Known orchestrators.  Orchestrators are not limited to this list.
	OrchestratorKubernetes = "k8s"
	OrchestratorCNI        = "cni"
	OrchestratorDocker     = "libnetwork"
	OrchestratorOpenStack  = "openstack"

	// Enum options for enable/disable fields
	Enabled  = "Enabled"
	Disabled = "Disabled"
)
