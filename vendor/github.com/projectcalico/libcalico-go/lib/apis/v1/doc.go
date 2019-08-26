// Copyright (c) 2016 Tigera, Inc. All rights reserved.

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
Package v1 implements the struct definitions used on the northbound client API.

In particular this includes:
	-  The various resource types as exposed on the Northbound API.  The valid resource
	   types are:
		-  BGPPeer
		-  HostEndpoint
		-  Policy
		-  Pool
		-  Profile
	-  The client configuration

The resource structures include the JSON tags for each exposed field.  These are standard
go-lang tags that define the JSON format of the structures as used by calicoctl.  The YAML
format also used by calicoctl is directly mapped from the JSON.
*/
package v1
