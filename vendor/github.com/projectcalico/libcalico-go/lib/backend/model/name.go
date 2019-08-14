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

package model

import "strings"

// escapeName removes any "/" from the name and URL encodes it to %2f,
// and necessarily removes % and encodes to %25.
func escapeName(name string) string {
	name = strings.Replace(name, "%", "%25", -1)
	return strings.Replace(name, "/", "%2f", -1)
}

// unescapeName replaces %2f and %25 in the name back to be a / and %.
func unescapeName(name string) string {
	name = strings.Replace(name, "%2f", "/", -1)
	return strings.Replace(name, "%25", "%", -1)
}
