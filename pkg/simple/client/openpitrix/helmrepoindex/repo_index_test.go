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
	"strings"
	"testing"

	"kubesphere.io/api/application/v1alpha1"
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

var indexData1 = `
apiVersion: v1
entries:
  apisix: []
  apisix-dashboard:
  - apiVersion: v2
    appVersion: 2.9.0
    created: "2021-11-15T08:23:00.343784368Z"
    description: A Helm chart for Apache APISIX Dashboard
    digest: 76f794b1300f7bfb756ede352fe71eb863b89f1995b495e8b683990709e310ad
    icon: https://apache.org/logos/res/apisix/apisix.png
    maintainers:
    - email: zhangjintao@apache.org
      name: tao12345666333
    name: apisix-dashboard
    type: application
    urls:
    - https://charts.kubesphere.io/main/apisix-dashboard-0.3.0.tgz
    version: 0.3.0
`
var indexData2 = `
apiVersion: v1
entries:
  apisix:
  - apiVersion: v2
    appVersion: 2.10.0
    created: "2021-11-15T08:23:00.343234584Z"
    dependencies:
    - condition: etcd.enabled
      name: etcd
      repository: https://charts.bitnami.com/bitnami
      version: 6.2.6
    - alias: dashboard
      condition: dashboard.enabled
      name: apisix-dashboard
      repository: https://charts.apiseven.com
      version: 0.3.0
    - alias: ingress-controller
      condition: ingress-controller.enabled
      name: apisix-ingress-controller
      repository: https://charts.apiseven.com
      version: 0.8.0
    description: A Helm chart for Apache APISIX
    digest: fed38a11c0fb54d385144767227e43cb2961d1b50d36ea207fdd122bddd3de28
    icon: https://apache.org/logos/res/apisix/apisix.png
    maintainers:
    - email: zhangjintao@apache.org
      name: tao12345666333
    name: apisix
    type: application
    urls:
    - https://charts.kubesphere.io/main/apisix-0.7.2.tgz
    version: 0.7.2
  apisix-dashboard:
  - apiVersion: v2
    appVersion: 2.9.0
    created: "2021-11-15T08:23:00.343784368Z"
    description: A Helm chart for Apache APISIX Dashboard
    digest: 76f794b1300f7bfb756ede352fe71eb863b89f1995b495e8b683990709e310ad
    icon: https://apache.org/logos/res/apisix/apisix.png
    maintainers:
    - email: zhangjintao@apache.org
      name: tao12345666333
    name: apisix-dashboard
    type: application
    urls:
    - https://charts.kubesphere.io/main/apisix-dashboard-0.3.0.tgz
    version: 0.3.0
`

func TestMergeRepo(t *testing.T) {
	repoIndex1, err := loadIndex([]byte(indexData1))
	if err != nil {
		t.Errorf("failed to load repo index")
		t.Failed()
	}
	existsSavedIndex := &SavedIndex{}
	repoCR := &v1alpha1.HelmRepo{}

	savedIndex1 := MergeRepoIndex(repoCR, repoIndex1, existsSavedIndex)
	if len(savedIndex1.Applications) != 1 {
		t.Errorf("faied to merge repo index with empty repo")
		t.Failed()
	}

	repoIndex2, err := loadIndex([]byte(indexData2))
	if err != nil {
		t.Errorf("failed to load repo index")
		t.Failed()
	}

	savedIndex2 := MergeRepoIndex(repoCR, repoIndex2, savedIndex1)
	if len(savedIndex2.Applications) != 2 {
		t.Errorf("faied to merge two repo index")
		t.Failed()
	}
}
