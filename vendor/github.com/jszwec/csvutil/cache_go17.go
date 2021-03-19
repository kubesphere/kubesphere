// +build !go1.9

package csvutil

import (
	"sync"
)

var fieldCache = struct {
	mtx sync.RWMutex
	m   map[typeKey][]field
}{m: make(map[typeKey][]field)}

func cachedFields(k typeKey) fields {
	fieldCache.mtx.RLock()
	fields, ok := fieldCache.m[k]
	fieldCache.mtx.RUnlock()

	if ok {
		return fields
	}

	fields = buildFields(k)

	fieldCache.mtx.Lock()
	fieldCache.m[k] = fields
	fieldCache.mtx.Unlock()

	return fields
}
