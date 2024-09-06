/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package controllertest

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
)

func LoadCrdPath() ([]string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("could not determine path")
	}
	curDir, _ := filepath.Split(filename)
	crdDirPaths := make([]string, 0, 1)
	projectRoot := filepath.Join(curDir, "..", "..", "..")
	configRoot := filepath.Join(projectRoot, "config")
	if err := filepath.WalkDir(configRoot, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			_, file := filepath.Split(path)
			if file == "crds" {
				crdDirPaths = append(crdDirPaths, path)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return crdDirPaths, nil
}
