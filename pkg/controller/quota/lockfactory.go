/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package quota

import (
	"sync"
)

// Following code copied from github.com/openshift/apiserver-library-go/pkg/admission/quota/clusterresourcequota
type LockFactory interface {
	GetLock(string) sync.Locker
}

type DefaultLockFactory struct {
	lock sync.RWMutex

	locks map[string]sync.Locker
}

func NewDefaultLockFactory() *DefaultLockFactory {
	return &DefaultLockFactory{locks: map[string]sync.Locker{}}
}

func (f *DefaultLockFactory) GetLock(key string) sync.Locker {
	lock, exists := f.getExistingLock(key)
	if exists {
		return lock
	}

	f.lock.Lock()
	defer f.lock.Unlock()
	lock = &sync.Mutex{}
	f.locks[key] = lock
	return lock
}

func (f *DefaultLockFactory) getExistingLock(key string) (sync.Locker, bool) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	lock, exists := f.locks[key]
	return lock, exists
}
