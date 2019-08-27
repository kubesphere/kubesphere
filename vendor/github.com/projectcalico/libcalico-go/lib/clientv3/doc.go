// Copyright (c) 2018 Tigera, Inc. All rights reserved.

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

/*
Package client implements the northbound client used to manage Calico configuration.

This client is the main entry point for applications that are managing or querying
Calico configuration.

This client provides a typed interface for managing different resource types.  The
definitions for each resource type are defined in the following package:
	github.com/projectcalico/libcalico-go/lib/api

The client has a number of methods that return interfaces for managing:
    -  BGP Peer resources
	-  Policy resources
	-  IP Pool resources
	-  Global network sets resources
	-  Host endpoint resources
	-  Workload endpoint resources
	-  Profile resources
	-  IP Address Management (IPAM)

See [resource definitions](http://docs.projectcalico.org/latest/reference/calicoctl/resources/) for details about the set of management commands for each
resource type.
*/
package clientv3
