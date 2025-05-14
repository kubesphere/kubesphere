// Copyright 2023 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package rego

import (
	v1 "github.com/open-policy-agent/opa/v1/rego"
)

type TargetPlugin = v1.TargetPlugin

type TargetPluginEval = v1.TargetPluginEval

func RegisterPlugin(name string, p TargetPlugin) {
	v1.RegisterPlugin(name, p)
}
