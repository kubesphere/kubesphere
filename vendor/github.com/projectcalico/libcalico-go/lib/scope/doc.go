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
Package scope implements field types that represent different scopes for resource
types.  For example, a resource may be valid at the global scope in that applies to
all Calico nodes, or may be at a node scope in that applies to a specific node.

The internal representation is an integer, but the JSON serialization of these
values is a string.
*/
package scope
