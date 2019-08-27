// Copyright (c) 2016-2017 Tigera, Inc. All rights reserved.

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

package numorstring_test

import (
	"encoding/json"
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/projectcalico/libcalico-go/lib/numorstring"
)

func init() {

	asNumberType := reflect.TypeOf(numorstring.ASNumber(0))
	protocolType := reflect.TypeOf(numorstring.Protocol{})
	portType := reflect.TypeOf(numorstring.Port{})

	// Perform tests of JSON unmarshaling of the various field types.
	DescribeTable("NumOrStringJSONUnmarshaling",
		func(jtext string, typ reflect.Type, expected interface{}) {
			// Create a new field type and invoke the unmarshaller interface
			// directly (this covers a couple more error cases than calling
			// through json.Unmarshal.
			new := reflect.New(typ)
			u := new.Interface().(json.Unmarshaler)
			err := u.UnmarshalJSON([]byte(jtext))

			if expected != nil {
				Expect(err).To(BeNil(),
					"expected json unmarshal to not error")
				Expect(new.Elem().Interface()).To(Equal(expected),
					"expected value not same as json unmarshalled value")
			} else {
				Expect(err).ToNot(BeNil(),
					"expected json unmarshal to error")
			}
		},
		// ASNumber tests.
		Entry("should accept 0 AS number as int", "0", asNumberType, numorstring.ASNumber(0)),
		Entry("should accept 4294967295 AS number as int", "4294967295", asNumberType, numorstring.ASNumber(4294967295)),
		Entry("should accept 0 AS number as string", "\"0\"", asNumberType, numorstring.ASNumber(0)),
		Entry("should accept 4294967295 AS number as string", "\"4294967295\"", asNumberType, numorstring.ASNumber(4294967295)),
		Entry("should accept 1.10 AS number as string", "\"1.10\"", asNumberType, numorstring.ASNumber(65546)),
		Entry("should accept 00.00 AS number as string", "\"00.00\"", asNumberType, numorstring.ASNumber(0)),
		Entry("should accept 00.01 AS number as string", "\"00.01\"", asNumberType, numorstring.ASNumber(1)),
		Entry("should accept 65535.65535 AS number as string", "\"65535.65535\"", asNumberType, numorstring.ASNumber(4294967295)),
		Entry("should reject 1.1.1 AS number as string", "\"1.1.1\"", asNumberType, nil),
		Entry("should reject 65536.65535 AS number as string", "\"65536.65535\"", asNumberType, nil),
		Entry("should reject 65535.65536 AS number as string", "\"65535.65536\"", asNumberType, nil),
		Entry("should reject 0.-1 AS number as string", "\"0.-1\"", asNumberType, nil),
		Entry("should reject -1 AS number as int", "-1", asNumberType, nil),
		Entry("should reject 4294967296 AS number as int", "4294967296", asNumberType, nil),

		// Port tests.
		Entry("should accept 0 port as int", "0", portType, numorstring.SinglePort(0)),
		Entry("should accept 65535 port as int", "65535", portType, numorstring.SinglePort(65535)),
		Entry("should accept 0:65535 port range as string", "\"0:65535\"", portType, portFromRange(0, 65535)),
		Entry("should accept 1:10 port range as string", "\"1:10\"", portType, portFromRange(1, 10)),
		Entry("should accept foo-bar as named port", "\"foo-bar\"", portType, numorstring.NamedPort("foo-bar")),
		Entry("should reject -1 port as int", "-1", portType, nil),
		Entry("should reject 65536 port as int", "65536", portType, nil),
		Entry("should reject 0:65536 port range as string", "\"0:65536\"", portType, nil),
		Entry("should reject -1:65535 port range as string", "\"-1:65535\"", portType, nil),
		Entry("should reject 10:1 port range as string", "\"10:1\"", portType, nil),
		Entry("should reject 1:2:3 port range as string", "\"1:2:3\"", portType, nil),
		Entry("should reject bad named port string", "\"*\"", portType, nil),
		Entry("should reject bad port string", "\"1:2", portType, nil),

		// Protocol tests.  Invalid integer values will be stored as strings.
		Entry("should accept 0 protocol as int", "0", protocolType, numorstring.ProtocolFromInt(0)),
		Entry("should accept 255 protocol as int", "255", protocolType, numorstring.ProtocolFromInt(255)),
		Entry("should accept tcp protocol as string", "\"TCP\"", protocolType, numorstring.ProtocolFromString("TCP")),
		Entry("should accept tcp protocol as string", "\"TCP\"", protocolType, numorstring.ProtocolFromString("TCP")),
		Entry("should accept 0 protocol as string", "\"0\"", protocolType, numorstring.ProtocolFromInt(0)),
		Entry("should accept 0 protocol as string", "\"255\"", protocolType, numorstring.ProtocolFromInt(255)),
		Entry("should accept 256 protocol as string", "\"256\"", protocolType, numorstring.ProtocolFromString("256")),
		Entry("should reject bad protocol string", "\"25", protocolType, nil),
	)

	// Perform tests of JSON marshaling of the various field types.
	DescribeTable("NumOrStringJSONMarshaling",
		func(field interface{}, jtext string) {
			b, err := json.Marshal(field)
			if jtext != "" {
				Expect(err).To(BeNil(),
					"expected json marshal to not error")
				Expect(string(b)).To(Equal(jtext),
					"expected json not same as marshalled value")
			} else {
				Expect(err).ToNot(BeNil(),
					"expected json marshal to error")
			}
		},
		// ASNumber tests.
		Entry("should marshal ASN of 0", numorstring.ASNumber(0), "0"),
		Entry("should marshal ASN of 4294967295", numorstring.ASNumber(4294967295), "4294967295"),

		// Port tests.
		Entry("should marshal port of 0", numorstring.SinglePort(0), "0"),
		Entry("should marshal port of 65535", portFromRange(65535, 65535), "65535"),
		Entry("should marshal port of 10", portFromString("10"), "10"),
		Entry("should marshal port range of 10:20", portFromRange(10, 20), "\"10:20\""),
		Entry("should marshal port range of 20:30", portFromRange(20, 30), "\"20:30\""),
		Entry("should marshal named port", numorstring.NamedPort("foobar"), `"foobar"`),

		// Protocol tests.
		Entry("should marshal protocol of 0", numorstring.ProtocolFromInt(0), "0"),
		Entry("should marshal protocol of udp", numorstring.ProtocolFromString("UDP"), "\"UDP\""),
	)

	// Perform tests of Stringer interface various field types.
	DescribeTable("NumOrStringStringify",
		func(field interface{}, s string) {
			a := fmt.Sprint(field)
			Expect(a).To(Equal(s),
				"expected String() value to match")
		},
		// ASNumber tests.
		Entry("should stringify ASN of 0", numorstring.ASNumber(0), "0"),
		Entry("should stringify ASN of 4294967295", numorstring.ASNumber(4294967295), "4294967295"),

		// Port tests.
		Entry("should stringify port of 20", numorstring.SinglePort(20), "20"),
		Entry("should stringify port range of 10:20", portFromRange(10, 20), "10:20"),

		// Protocol tests.
		Entry("should stringify protocol of 0", numorstring.ProtocolFromInt(0), "0"),
		Entry("should stringify protocol of udp", numorstring.ProtocolFromString("UDP"), "UDP"),
	)

	// Perform tests of Protocols supporting ports.
	DescribeTable("NumOrStringProtocolsSupportingPorts",
		func(protocol numorstring.Protocol, supportsPorts bool) {
			Expect(protocol.SupportsPorts()).To(Equal(supportsPorts),
				"expected protocol port support to match")
		},
		Entry("protocol 6 supports ports", numorstring.ProtocolFromInt(6), true),
		Entry("protocol 17 supports ports", numorstring.ProtocolFromInt(17), true),
		Entry("protocol udp supports ports", numorstring.ProtocolFromString("UDP"), true),
		Entry("protocol udp supports ports", numorstring.ProtocolFromString("TCP"), true),
		Entry("protocol foo does not support ports", numorstring.ProtocolFromString("foo"), false),
		Entry("protocol 2 does not support ports", numorstring.ProtocolFromInt(2), false),
	)

	// Perform tests of Protocols FromString method.
	DescribeTable("NumOrStringProtocols FromString is not case sensitive",
		func(input, expected string) {
			Expect(numorstring.ProtocolFromString(input).StrVal).To(Equal(expected),
				"expected parsed protocol to match")
		},
		Entry("protocol udp -> UDP", "udp", "UDP"),
		Entry("protocol tcp -> TCP", "tcp", "TCP"),
		Entry("protocol updlite -> UDPLite", "udplite", "UDPLite"),
		Entry("unknown protocol xxxXXX", "xxxXXX", "xxxXXX"),
	)

	// Perform tests of Protocols FromStringV1 method.
	DescribeTable("NumOrStringProtocols FromStringV1 is lowercase",
		func(input, expected string) {
			Expect(numorstring.ProtocolFromStringV1(input).StrVal).To(Equal(expected),
				"expected parsed protocol to match")
		},
		Entry("protocol udp -> UDP", "UDP", "udp"),
		Entry("protocol tcp -> TCP", "TCP", "tcp"),
		Entry("protocol updlite -> UDPLite", "UDPLite", "udplite"),
		Entry("unknown protocol xxxXXX", "xxxXXX", "xxxxxx"),
	)

	// Perform tests of Protocols ToV1 method.
	DescribeTable("NumOrStringProtocols FromStringV1 is lowercase",
		func(input, expected numorstring.Protocol) {
			Expect(input.ToV1()).To(Equal(expected),
				"expected parsed protocol to match")
		},
		// Protocol tests.
		Entry("protocol udp -> UDP", numorstring.ProtocolFromInt(2), numorstring.ProtocolFromInt(2)),
		Entry("protocol tcp -> TCP", numorstring.ProtocolFromString("TCP"), numorstring.ProtocolFromStringV1("TCP")),
	)
}

func portFromRange(minPort, maxPort uint16) numorstring.Port {
	p, _ := numorstring.PortFromRange(minPort, maxPort)
	return p
}

func portFromString(s string) numorstring.Port {
	p, _ := numorstring.PortFromString(s)
	return p
}
