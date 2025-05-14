// Copyright 2023 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package extension

import (
	"sync"
)

var pluginMtx sync.Mutex
var bundleExtensions map[string]Handler

// Handler is used to unmarshal a byte slice of a registered extension
// EXPERIMENTAL: Please don't rely on this functionality, it may go
// away or change in the future.
type Handler func([]byte, any) error

// RegisterExtension registers a Handler for a certain file extension, including
// the dot: ".json", not "json".
// EXPERIMENTAL: Please don't rely on this functionality, it may go
// away or change in the future.
func RegisterExtension(name string, handler Handler) {
	pluginMtx.Lock()
	defer pluginMtx.Unlock()

	if bundleExtensions == nil {
		bundleExtensions = map[string]Handler{}
	}
	bundleExtensions[name] = handler
}

// FindExtension ios used to look up a registered extension Handler
// EXPERIMENTAL: Please don't rely on this functionality, it may go
// away or change in the future.
func FindExtension(ext string) Handler {
	pluginMtx.Lock()
	defer pluginMtx.Unlock()
	return bundleExtensions[ext]
}
