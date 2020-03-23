// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package merge contains helpers to merge data structures
// frequently encountered in OPA.
package merge

// InterfaceMaps returns the result of merging a and b. If a and b cannot be
// merged because of conflicting key-value pairs, ok is false.
func InterfaceMaps(a map[string]interface{}, b map[string]interface{}) (c map[string]interface{}, ok bool) {

	c = map[string]interface{}{}
	for k := range a {
		c[k] = a[k]
	}

	for k := range b {

		add := b[k]
		exist, ok := c[k]
		if !ok {
			c[k] = add
			continue
		}

		existObj, existOk := exist.(map[string]interface{})
		addObj, addOk := add.(map[string]interface{})
		if !existOk || !addOk {
			return nil, false
		}

		c[k], ok = InterfaceMaps(existObj, addObj)
		if !ok {
			return nil, false
		}
	}

	return c, true
}
