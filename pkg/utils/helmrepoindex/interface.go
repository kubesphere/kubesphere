package helmrepoindex

import "time"

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
	GetKeywords() string
	GetMaintainers() string
	GetScreenshots() string
	GetPackageName() string
	GetCreateTime() time.Time
}
