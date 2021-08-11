/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package addr

import (
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/internal/flock"
)

// TODO(directxman12): interface / release functionality for external port managers

const (
	portReserveTime   = 10 * time.Minute
	portConflictRetry = 100
	portFilePrefix    = "port-"
)

var (
	cacheDir string
)

func init() {
	baseDir, err := os.UserCacheDir()
	if err != nil {
		baseDir = os.TempDir()
	}
	cacheDir = filepath.Join(baseDir, "kubebuilder-envtest")
	if err := os.MkdirAll(cacheDir, 0750); err != nil {
		panic(err)
	}
}

type portCache struct{}

func (c *portCache) add(port int) (bool, error) {
	// Remove outdated ports.
	if err := fs.WalkDir(os.DirFS(cacheDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !d.Type().IsRegular() || !strings.HasPrefix(path, portFilePrefix) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if time.Since(info.ModTime()) > portReserveTime {
			if err := os.Remove(filepath.Join(cacheDir, path)); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return false, err
	}
	// Try allocating new port, by acquiring a file.
	if err := flock.Acquire(fmt.Sprintf("%s/%s%d", cacheDir, portFilePrefix, port)); os.IsExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

var cache = &portCache{}

func suggest(listenHost string) (int, string, error) {
	if listenHost == "" {
		listenHost = "localhost"
	}
	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(listenHost, "0"))
	if err != nil {
		return -1, "", err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return -1, "", err
	}
	if err := l.Close(); err != nil {
		return -1, "", err
	}
	return l.Addr().(*net.TCPAddr).Port,
		addr.IP.String(),
		nil
}

// Suggest suggests an address a process can listen on. It returns
// a tuple consisting of a free port and the hostname resolved to its IP.
// It makes sure that new port allocated does not conflict with old ports
// allocated within 1 minute.
func Suggest(listenHost string) (int, string, error) {
	for i := 0; i < portConflictRetry; i++ {
		port, resolvedHost, err := suggest(listenHost)
		if err != nil {
			return -1, "", err
		}
		if ok, err := cache.add(port); ok {
			return port, resolvedHost, nil
		} else if err != nil {
			return -1, "", err
		}
	}
	return -1, "", fmt.Errorf("no free ports found after %d retries", portConflictRetry)
}
