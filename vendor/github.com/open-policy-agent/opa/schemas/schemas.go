// Copyright 2023 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package schemas

import (
	"embed"
)

// FS contains the known schemas for OPA's Authorization Policy etc.
// "authorizationPolicy.json" contains the input schema for OPA's Authorization Policy
//
//go:embed *.json
var FS embed.FS
