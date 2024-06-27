// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package uuid

import (
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
)

const (
	BILLION = 1000000000
)

// New Create a version 4 random UUID
func New(r io.Reader) (string, error) {
	bs := make([]byte, 16)
	n, err := io.ReadFull(r, bs)
	if n != len(bs) || err != nil {
		return "", err
	}
	bs[8] = bs[8]&^0xc0 | 0x80
	bs[6] = bs[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", bs[0:4], bs[4:6], bs[6:8], bs[8:10], bs[10:]), nil
}

// Parse will use the google/uuid library to parse the string into a uuid
// if parsing fails, it will return an empty map. It will fill the map
// with some decoded values with fillMap
// ref: https://datatracker.ietf.org/doc/html/rfc4122
func Parse(s string) (map[string]interface{}, error) {
	uuid, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	out := make(map[string]interface{}, getVersionLen(int(uuid.Version())))
	fillMap(out, uuid)
	return out, nil
}

// Fills the map with values from the uuid. Version and variant for every version.
// Version 1-2 has decodable values that could be of use, version 4 is random,
// and version 3,5 is not feasible to extract data. Generated with either MD5 or SHA1 hash
// ref: https://datatracker.ietf.org/doc/html/rfc4122 about creation of UUIDs
func fillMap(m map[string]interface{}, u uuid.UUID) {
	m["version"] = int(u.Version())
	m["variant"] = u.Variant().String()
	switch version := m["version"]; version {
	case 1, 2:
		m["time"] = nanoUnix(u.Time())
		m["nodeid"] = byteDecimalToHexMAC(u.NodeID(), "-")
		m["macvariables"] = macVars(u.NodeID()[0])
		m["clocksequence"] = u.ClockSequence()
		if version == 2 {
			m["id"] = int(u.ID())
			m["domain"] = u.Domain().String()
		}
	}
}

// macVars will take the first byte of a MAC-address and check for the
// local/global bit and check for the unicast/multicast bit of the byte,
// and return a string with this info.
// ref: https://datatracker.ietf.org/doc/html/rfc7042#section-2.1
func macVars(inpb byte) string {
	switch {
	case inpb&byte(0b11) == byte(0b11):
		return "local:multicast"
	case inpb&byte(0b01) == byte(0b01):
		return "global:multicast"
	case inpb&byte(0b10) == byte(0b10):
		return "local:unicast"
	}
	return "global:unicast"
}

// loops through the byte array to convert all bytes to hexes.
// It will also put the separator between every other to make it human-readable
func byteDecimalToHexMAC(bytes []byte, sep string) string {
	hexs := strings.Builder{}
	l := len(bytes)
	hexs.Grow((l * 3) - 1) // 1 byte -> 2 hexes + 1 separator (if one char)

	for i, b := range bytes {
		hexs.WriteString(fmt.Sprintf("%02x", b))
		if i < l-1 {
			hexs.WriteString(sep)
		}
	}

	return hexs.String()
}

// nanoUnix Converts the uuids encoded time into unix represented time in nanoseconds
func nanoUnix(t uuid.Time) int64 {
	unixsec, unixnsec := t.UnixTime()
	return unixsec*BILLION + unixnsec
}

// Helper function to make map with length based on version of uuid
// Most are 2 in length (version, variant), but version 1 and 2 have more.
func getVersionLen(version int) int {
	switch version {
	case 1:
		return 5
	case 2:
		return 7
	default:
		return 2
	}
}
