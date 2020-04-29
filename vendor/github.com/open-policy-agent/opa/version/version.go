// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package version contains version information that is set at build time.
package version

// Version information that is displayed by the "version" command and used to
// identify the version of running instances of OPA.
var (
	Version   = ""
	Vcs       = ""
	Timestamp = ""
	Hostname  = ""
)
