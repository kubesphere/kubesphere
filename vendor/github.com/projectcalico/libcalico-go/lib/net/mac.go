// Copyright (c) 2016 Tigera, Inc. All rights reserved.

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

package net

import (
	"encoding/json"
	"net"
)

// Sub class net.HardwareAddr so that we can add JSON marshalling and unmarshalling.
type MAC struct {
	net.HardwareAddr
}

// MarshalJSON interface for a MAC
func (m MAC) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

// UnmarshalJSON interface for a MAC
func (m *MAC) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if mac, err := net.ParseMAC(s); err != nil {
		return err
	} else {
		m.HardwareAddr = mac
		return nil
	}
}
