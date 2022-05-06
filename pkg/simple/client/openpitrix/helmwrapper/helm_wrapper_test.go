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
	"io/ioutil"
	"os"
	"testing"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"

	"kubesphere.io/kubesphere/pkg/constants"
)

func TestHelmInstall(t *testing.T) {
	wr := NewHelmWrapper("", "dummy", "dummy",
		SetAnnotations(map[string]string{constants.CreatorAnnotationKey: "1234"}),
		SetMock(true))
	charData := GenerateChartData(t, "dummy-chart")
	chartValues := `helm-wrapper: "test-val"`

	err := wr.writeAction("dummy-chart", charData, chartValues, false)
	if err != nil {
		t.Fail()
	}
}

func TempDir(t *testing.T) string {
	t.Helper()
	d, err := ioutil.TempDir("", "kubesphere")
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func GenerateChartData(t *testing.T, name string) string {
	tmpChart := TempDir(t)
	defer os.RemoveAll(tmpChart)

	cfile := &chart.Chart{
		Metadata: &chart.Metadata{
			APIVersion:  chart.APIVersionV1,
			Name:        name,
			Description: "A Helm chart for Kubernetes",
			Version:     "0.1.0",
		},
	}

	filename, err := chartutil.Save(cfile, tmpChart)
	if err != nil {
		t.Fatalf("Error creating chart for test: %v", err)
	}
	charData, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("Error loading chart data %v", err)
	}
	return string(charData)
}
