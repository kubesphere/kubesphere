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
	"bytes"
	"encoding/json"
	"fmt"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/klog"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/chart/loader"
)

func LoadPackage(pkg []byte) (VersionInterface, error) {
	p, err := loader.LoadArchive(bytes.NewReader(pkg))
	if err != nil {
		klog.Errorf("Failed to load package, error: %+v", err)
		return nil, err
	}
	return HelmVersionWrapper{ChartVersion: &repo.ChartVersion{Metadata: p.Metadata}}, nil
}

type HelmVersionWrapper struct {
	*repo.ChartVersion
}

func (h HelmVersionWrapper) GetIcon() string          { return h.ChartVersion.Icon }
func (h HelmVersionWrapper) GetName() string          { return h.ChartVersion.Name }
func (h HelmVersionWrapper) GetHome() string          { return h.ChartVersion.Home }
func (h HelmVersionWrapper) GetVersion() string       { return h.ChartVersion.Version }
func (h HelmVersionWrapper) GetAppVersion() string    { return h.ChartVersion.AppVersion }
func (h HelmVersionWrapper) GetDescription() string   { return h.ChartVersion.Description }
func (h HelmVersionWrapper) GetCreateTime() time.Time { return h.ChartVersion.Created }
func (h HelmVersionWrapper) GetUrls() string {
	if len(h.ChartVersion.URLs) == 0 {
		return ""
	}
	return h.ChartVersion.URLs[0]
}

func (h HelmVersionWrapper) GetSources() string {
	if len(h.ChartVersion.Sources) == 0 {
		return ""
	}
	s, _ := json.Marshal(h.ChartVersion.Sources)
	return string(s)
}

func (h HelmVersionWrapper) GetKeywords() string {
	return strings.Join(h.ChartVersion.Keywords, ",")
}

func (h HelmVersionWrapper) GetMaintainers() string {
	if len(h.ChartVersion.Maintainers) == 0 {
		return ""
	}
	s, _ := json.Marshal(h.ChartVersion.Maintainers)
	return string(s)
}

func (h HelmVersionWrapper) GetScreenshots() string {
	return ""
}

func (h HelmVersionWrapper) GetVersionName() string {
	versionName := h.GetVersion()
	if h.GetAppVersion() != "" {
		versionName += fmt.Sprintf(" [%s]", h.GetAppVersion())
	}
	return versionName
}

func (h HelmVersionWrapper) GetPackageName() string {
	file := h.GetUrls()
	if len(file) == 0 {
		return fmt.Sprintf("%s-%s.tgz", h.Name, h.Version)
	}
	return file
}
