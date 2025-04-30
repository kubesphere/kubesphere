/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

const (
	StateEnabled     = "Enabled"
	StateDisabled    = "Disabled"
	StateAvailable   = "Available"
	StateUnavailable = "Unavailable"
	DistPrefix       = "/dist"
	ProxyPrefix      = "/proxy"

	ReverseProxyTargetLabel     = "kubesphere.io/reverse-proxy-target"
	ReverseProxyTargetAPIServer = "ks-apiserver"
	ReverseProxyTargetConsole   = "ks-console"
)
