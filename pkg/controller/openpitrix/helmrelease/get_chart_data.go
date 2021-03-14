/*
Copyright 2019 The KubeSphere Authors.

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

package helmrelease

import (
	"context"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmrepoindex"
	"path"
	"strings"
)

func (r *ReconcileHelmRelease) GetChartData(rls *v1alpha1.HelmRelease) (chartName string, chartData []byte, err error) {
	if rls.Spec.RepoId != "" && rls.Spec.RepoId != v1alpha1.AppStoreRepoId {
		// load chart data from helm repo
		repo := v1alpha1.HelmRepo{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: rls.Spec.RepoId}, &repo)
		if err != nil {
			klog.Errorf("get helm repo %s failed, error: %v", rls.Spec.RepoId, err)
			return chartName, chartData, ErrGetRepoFailed
		}

		index, err := helmrepoindex.ByteArrayToSavedIndex([]byte(repo.Status.Data))

		if version := index.GetApplicationVersion(rls.Spec.ApplicationId, rls.Spec.ApplicationVersionId); version != nil {
			url := version.Spec.URLs[0]
			if !(strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "s3://")) {
				url = repo.Spec.Url + "/" + url
			}
			buf, err := helmrepoindex.LoadChart(context.TODO(), url, &repo.Spec.Credential)
			if err != nil {
				klog.Infof("load chart failed, error: %s", err)
				return chartName, chartData, ErrLoadChartFailed
			}
			chartData = buf.Bytes()
			chartName = version.Name
		} else {
			klog.Errorf("get app version: %s failed", rls.Spec.ApplicationVersionId)
			return chartName, chartData, ErrGetAppVersionFailed
		}
	} else {
		// load chart data from helm application version
		appVersion := &v1alpha1.HelmApplicationVersion{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: rls.Spec.ApplicationVersionId}, appVersion)
		if err != nil {
			klog.Errorf("get app version %s failed, error: %v", rls.Spec.ApplicationVersionId, err)
			return chartName, chartData, ErrGetAppVersionFailed
		}

		if r.StorageClient == nil {
			return "", nil, ErrS3Config
		}
		chartData, err = r.StorageClient.Read(path.Join(appVersion.GetWorkspace(), appVersion.Name))
		if err != nil {
			klog.Errorf("load chart from storage failed, error: %s", err)
			return chartName, chartData, ErrLoadChartFromStorageFailed
		}

		chartName = appVersion.GetTrueName()
	}
	return
}
