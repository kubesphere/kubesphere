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
