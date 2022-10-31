// Copyright 2022 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

//go:build go1.18
// +build go1.18

package version

import (
	"runtime/debug"
)

func init() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	dirty := false
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.time":
			Timestamp = s.Value
		case "vcs.revision":
			Vcs = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if dirty {
		Vcs = Vcs + "-dirty"
	}
}
