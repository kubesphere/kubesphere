// status is a simple package for offering up various status information from Kiali.
package status

const (
	name           = "Kiali"
	ConsoleVersion = name + " console version"
	CoreVersion    = name + " core version"
	CoreCommitHash = name + " core commit hash"
	State          = name + " state"
	ClusterMTLS    = "Istio mTLS"
	StateRunning   = "running"
)

// StatusInfo statusInfo
//
// This is used for returning a response of Kiali Status
//
// swagger:model StatusInfo
type StatusInfo struct {
	// The state of Kiali
	// A hash of key,values with versions of Kiali and state
	//
	// required: true
	Status map[string]string `json:"status"`
	// An array of external services installed
	//
	// required: true
	// swagger:allOf
	ExternalServices []ExternalServiceInfo `json:"externalServices"`
	// An array of warningMessages
	// items.example: Istio version 0.7.1 is not supported, the version should be 0.8.0
	// swagger:allOf
	WarningMessages []string `json:"warningMessages"`
}

var info StatusInfo

// Status response model
//
// This is used for returning a response of Kiali Status
//
// swagger:model externalServiceInfo
type ExternalServiceInfo struct {
	// The name of the service
	//
	// required: true
	// example: Istio
	Name string `json:"name"`

	// The installed version of the service
	//
	// required: false
	// example: 0.8.0
	Version string `json:"version,omitempty"`

	// The service url
	//
	// required: false
	// example: jaeger-query-istio-system.127.0.0.1.nip.io
	Url string `json:"url,omitempty"`
}

func init() {
	info = StatusInfo{Status: make(map[string]string)}
	info.Status[State] = StateRunning
}

// Put adds or replaces status info for the provided name. Any previous setting is returned.
func Put(name, value string) (previous string, hasPrevious bool) {
	previous, hasPrevious = info.Status[name]
	info.Status[name] = value
	return previous, hasPrevious
}

// Get returns a copy of the current status info.
func Get() (status StatusInfo) {
	info.ExternalServices = []ExternalServiceInfo{}
	info.WarningMessages = []string{}
	info.getmTLSStatus()
	getVersions()
	return info
}
