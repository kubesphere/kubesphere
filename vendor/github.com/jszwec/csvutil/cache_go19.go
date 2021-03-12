// +build go1.9

package csvutil

import (
	"sync"
)

var fieldCache sync.Map // map[typeKey][]field

func cachedFields(k typeKey) fields {
	if v, ok := fieldCache.Load(k); ok {
		return v.(fields)
	}

	v, _ := fieldCache.LoadOrStore(k, buildFields(k))
	return v.(fields)
}
