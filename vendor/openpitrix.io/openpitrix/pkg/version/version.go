// Copyright 2017 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

//go:generate go run gen_helper.go
//go:generate go fmt

package version

import "fmt"

var (
	ShortVersion   = "dev"
	GitSha1Version = "git-sha1"
	BuildDate      = "2017-01-01"
)

func PrintVersionInfo(printer func(string, ...interface{})) {
	printer("Release OpVersion: %s", ShortVersion)
	printer("Git Commit Hash: %s", GitSha1Version)
	printer("Build Time: %s", BuildDate)
}

func GetVersionString() string {
	return fmt.Sprintf("%s; git: %s; build time: %s", ShortVersion, GitSha1Version, BuildDate)
}
