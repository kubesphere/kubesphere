// Package Types defines type signatures used throughout SonyFlake. This allows for
// fine-tuned control over imports, and the ability to mock out imports as well
package types

import "net"

// InterfaceAddrs defines the interface used for retrieving network addresses
type InterfaceAddrs func() ([]net.Addr, error)
