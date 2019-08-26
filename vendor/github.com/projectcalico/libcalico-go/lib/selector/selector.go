// Copyright (c) 2016-2017 Tigera, Inc. All rights reserved.

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

package selector

import "github.com/projectcalico/libcalico-go/lib/selector/parser"

// Selector represents a label selector.
type Selector interface {
	// Evaluate evaluates the selector against the given labels expressed as a concrete map.
	Evaluate(labels map[string]string) bool

	// EvaluateLabels evaluates the selector against the given labels expressed as an interface.
	// This allows for labels that are calculated on the fly.
	EvaluateLabels(labels parser.Labels) bool

	// String returns a string that represents this selector.
	String() string

	// UniqueID returns the unique ID that represents this selector.
	UniqueID() string
}

// Parse a string representation of a selector expression into a Selector.
func Parse(selector string) (sel Selector, err error) {
	return parser.Parse(selector)
}
