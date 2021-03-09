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

package helmrepoindex

import (
	"context"
	"fmt"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"strings"
	"testing"
)

func TestLoadRepo(t *testing.T) {

	u := "https://charts.kubesphere.io/main"

	index, err := LoadRepoIndex(context.TODO(), u, &v1alpha1.HelmRepoCredential{})
	if err != nil {
		t.Errorf("load repo failed, err: %s", err)
		t.Failed()
		return
	}

	for _, entry := range index.Entries {
		chartUrl := entry[0].URLs[0]

		if !(strings.HasPrefix(chartUrl, "https://") || strings.HasPrefix(chartUrl, "http://")) {
			chartUrl = fmt.Sprintf("%s/%s", u, chartUrl)
		}
		chartData, err := LoadChart(context.TODO(), chartUrl, &v1alpha1.HelmRepoCredential{})
		if err != nil {
			t.Errorf("load chart data failed, err: %s", err)
			t.Failed()
		}
		_ = chartData
		break
	}

}
