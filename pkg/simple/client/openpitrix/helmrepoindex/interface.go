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
	"time"

	"kubesphere.io/api/application/v1alpha1"
)

type VersionInterface interface {
	GetName() string
	GetVersion() string
	GetAppVersion() string
	GetDescription() string
	GetUrls() string
	GetVersionName() string
	GetIcon() string
	GetHome() string
	GetSources() string
	GetRawSources() []string
	GetKeywords() string
	GetMaintainers() string
	GetRawMaintainers() []*v1alpha1.Maintainer
	GetScreenshots() string
	GetPackageName() string
	GetCreateTime() time.Time
}
