/*
Copyright 2018 The KubeSphere Authors.

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

package version

import (
	"flag"
	"fmt"
	"os"
)

var (
	Version     = ""
	BuildTime   = ""
	versionFlag = false
)

// PrintAndExitIfRequested will check if the -version flag was passed
// and, if so, print the version and exit.
func init() {
	flag.BoolVar(&versionFlag, "version", false, "print the version of kubesphere")
}

func PrintAndExitIfRequested() {
	if versionFlag {
		fmt.Printf("Version: %s\nBuildTime: %s\n", Version, BuildTime)
		os.Exit(0)
	}
}
