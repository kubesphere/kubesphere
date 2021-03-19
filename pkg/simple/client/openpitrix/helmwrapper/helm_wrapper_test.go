/*
Copyright 2020 The KubeSphere Authors.

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

package helmwrapper

import (
	"fmt"
	"kubesphere.io/kubesphere/pkg/constants"
	"os"
	"testing"
)

func TestHelmInstall(t *testing.T) {
	wr := NewHelmWrapper("", "dummy", "dummy",
		SetAnnotations(map[string]string{constants.CreatorAnnotationKey: "1234"}),
		SetMock(true))

	res, err := wr.install("dummy-chart", "", "dummy-value", false)
	if err != nil {
		t.Fail()
	}

	_ = res
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// some code here to check arguments perhaps?
	fmt.Fprintf(os.Stdout, "helm mock success")
	os.Exit(0)
}
