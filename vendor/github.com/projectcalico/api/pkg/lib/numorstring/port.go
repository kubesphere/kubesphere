// Copyright (c) 2016-2017 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package numorstring

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// Port represents either a range of numeric ports or a named port.
//
//   - For a named port, set the PortName, leaving MinPort and MaxPort as 0.
//   - For a port range, set MinPort and MaxPort to the (inclusive) port numbers.  Set
//     PortName to "".
//   - For a single port, set MinPort = MaxPort and PortName = "".
type Port struct {
	MinPort  uint16 `json:"minPort,omitempty"`
	MaxPort  uint16 `json:"maxPort,omitempty"`
	PortName string `json:"portName" validate:"omitempty,portName"`
}

// SinglePort creates a Port struct representing a single port.
func SinglePort(port uint16) Port {
	return Port{MinPort: port, MaxPort: port}
}

func NamedPort(name string) Port {
	return Port{PortName: name}
}

// PortFromRange creates a Port struct representing a range of ports.
func PortFromRange(minPort, maxPort uint16) (Port, error) {
	port := Port{MinPort: minPort, MaxPort: maxPort}
	if minPort > maxPort {
		msg := fmt.Sprintf("minimum port number (%d) is greater than maximum port number (%d) in port range", minPort, maxPort)
		return port, errors.New(msg)
	}
	return port, nil
}

var (
	allDigits = regexp.MustCompile(`^\d+$`)
	portRange = regexp.MustCompile(`^(\d+):(\d+)$`)
	nameRegex = regexp.MustCompile("^[a-zA-Z0-9_.-]{1,128}$")
)

// PortFromString creates a Port struct from its string representation.  A port
// may either be single value "1234", a range of values "100:200" or a named port: "name".
func PortFromString(s string) (Port, error) {
	if allDigits.MatchString(s) {
		// Port is all digits, it should parse as a single port.
		num, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			msg := fmt.Sprintf("invalid port format (%s)", s)
			return Port{}, errors.New(msg)
		}
		return SinglePort(uint16(num)), nil
	}

	if groups := portRange.FindStringSubmatch(s); len(groups) > 0 {
		// Port matches <digits>:<digits>, it should parse as a range of ports.
		if pmin, err := strconv.ParseUint(groups[1], 10, 16); err != nil {
			msg := fmt.Sprintf("invalid minimum port number in range (%s)", s)
			return Port{}, errors.New(msg)
		} else if pmax, err := strconv.ParseUint(groups[2], 10, 16); err != nil {
			msg := fmt.Sprintf("invalid maximum port number in range (%s)", s)
			return Port{}, errors.New(msg)
		} else {
			return PortFromRange(uint16(pmin), uint16(pmax))
		}
	}

	if !nameRegex.MatchString(s) {
		msg := fmt.Sprintf("invalid name for named port (%s)", s)
		return Port{}, errors.New(msg)
	}

	return NamedPort(s), nil
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (p *Port) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}

		if v, err := PortFromString(s); err != nil {
			return err
		} else {
			*p = v
			return nil
		}
	}

	// It's not a string, it must be a single int.
	var i uint16
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}
	v := SinglePort(i)
	*p = v
	return nil
}

// MarshalJSON implements the json.Marshaller interface.
func (p Port) MarshalJSON() ([]byte, error) {
	if p.PortName != "" {
		return json.Marshal(p.PortName)
	} else if p.MinPort == p.MaxPort {
		return json.Marshal(p.MinPort)
	} else {
		return json.Marshal(p.String())
	}
}

// String returns the string value.  If the min and max port are the same
// this returns a single string representation of the port number, otherwise
// if returns a colon separated range of ports.
func (p Port) String() string {
	if p.PortName != "" {
		return p.PortName
	} else if p.MinPort == p.MaxPort {
		return strconv.FormatUint(uint64(p.MinPort), 10)
	} else {
		return fmt.Sprintf("%d:%d", p.MinPort, p.MaxPort)
	}
}

// OpenAPISchemaType is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (_ Port) OpenAPISchemaType() []string { return []string{"string"} }

// OpenAPISchemaFormat is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (_ Port) OpenAPISchemaFormat() string { return "int-or-string" }
