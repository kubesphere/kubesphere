// Copyright (c) 2016-2020 Tigera, Inc. All rights reserved.

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

import "strings"

const (
	ProtocolUDP     = "UDP"
	ProtocolTCP     = "TCP"
	ProtocolICMP    = "ICMP"
	ProtocolICMPv6  = "ICMPv6"
	ProtocolSCTP    = "SCTP"
	ProtocolUDPLite = "UDPLite"

	ProtocolUDPV1  = "udp"
	ProtocolTCPV1  = "tcp"
	ProtocolSCTPV1 = "sctp"
)

var (
	allProtocolNames = []string{
		ProtocolUDP,
		ProtocolTCP,
		ProtocolICMP,
		ProtocolICMPv6,
		ProtocolSCTP,
		ProtocolUDPLite,
	}
)

type Protocol Uint8OrString

// ProtocolFromInt creates a Protocol struct from an integer value.
func ProtocolFromInt(p uint8) Protocol {
	return Protocol(
		Uint8OrString{Type: NumOrStringNum, NumVal: p},
	)
}

// ProtocolV3FromProtocolV1 creates a v3 Protocol from a v1 Protocol,
// while handling case conversion.
func ProtocolV3FromProtocolV1(p Protocol) Protocol {
	if p.Type == NumOrStringNum {
		return p
	}

	for _, n := range allProtocolNames {
		if strings.EqualFold(n, p.StrVal) {
			return Protocol(
				Uint8OrString{Type: NumOrStringString, StrVal: n},
			)
		}
	}

	return p
}

// ProtocolFromString creates a Protocol struct from a string value.
func ProtocolFromString(p string) Protocol {
	for _, n := range allProtocolNames {
		if strings.EqualFold(n, p) {
			return Protocol(
				Uint8OrString{Type: NumOrStringString, StrVal: n},
			)
		}
	}

	// Unknown protocol - return the value unchanged.  Validation should catch this.
	return Protocol(
		Uint8OrString{Type: NumOrStringString, StrVal: p},
	)
}

// ProtocolFromStringV1 creates a Protocol struct from a string value (for the v1 API)
func ProtocolFromStringV1(p string) Protocol {
	return Protocol(
		Uint8OrString{Type: NumOrStringString, StrVal: strings.ToLower(p)},
	)
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (p *Protocol) UnmarshalJSON(b []byte) error {
	return (*Uint8OrString)(p).UnmarshalJSON(b)
}

// MarshalJSON implements the json.Marshaller interface.
func (p Protocol) MarshalJSON() ([]byte, error) {
	return Uint8OrString(p).MarshalJSON()
}

// String returns the string value, or the Itoa of the int value.
func (p Protocol) String() string {
	return (Uint8OrString)(p).String()
}

// String returns the string value, or the Itoa of the int value.
func (p Protocol) ToV1() Protocol {
	if p.Type == NumOrStringNum {
		return p
	}
	return ProtocolFromStringV1(p.StrVal)
}

// NumValue returns the NumVal if type Int, or if
// it is a String, will attempt a conversion to int.
func (p Protocol) NumValue() (uint8, error) {
	return (Uint8OrString)(p).NumValue()
}

// SupportsProtocols returns whether this protocol supports ports.  This returns true if
// the numerical or string version of the protocol indicates TCP (6), UDP (17), or SCTP (132).
func (p Protocol) SupportsPorts() bool {
	num, err := p.NumValue()
	if err == nil {
		return num == 6 || num == 17 || num == 132
	} else {
		switch p.StrVal {
		case ProtocolTCP, ProtocolUDP, ProtocolTCPV1, ProtocolUDPV1, ProtocolSCTP, ProtocolSCTPV1:
			return true
		}
		return false
	}
}

// OpenAPISchemaType is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (_ Protocol) OpenAPISchemaType() []string { return []string{"string"} }

// OpenAPISchemaFormat is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (_ Protocol) OpenAPISchemaFormat() string { return "int-or-string" }
