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

package numorstring

import (
	"encoding/json"
	"strconv"
)

// UInt8OrString is a type that can hold an uint8 or a string.  When used in
// JSON or YAML marshalling and unmarshalling, it produces or consumes the
// inner type.  This allows you to have, for example, a JSON field that can
// accept a name or number.
type Uint8OrString struct {
	Type   NumOrStringType `json:"type"`
	NumVal uint8           `json:"numVal"`
	StrVal string          `json:"strVal"`
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (i *Uint8OrString) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}

		num, err := strconv.ParseUint(s, 10, 8)
		if err == nil {
			i.Type = NumOrStringNum
			i.NumVal = uint8(num)
		} else {
			i.Type = NumOrStringString
			i.StrVal = s
		}

		return nil
	}
	i.Type = NumOrStringNum
	return json.Unmarshal(b, &i.NumVal)
}

// MarshalJSON implements the json.Marshaller interface.
func (i Uint8OrString) MarshalJSON() ([]byte, error) {
	if num, err := i.NumValue(); err == nil {
		return json.Marshal(num)
	} else {
		return json.Marshal(i.StrVal)
	}
}

// String returns the string value, or the Itoa of the int value.
func (i Uint8OrString) String() string {
	if i.Type == NumOrStringString {
		return i.StrVal
	}
	return strconv.FormatUint(uint64(i.NumVal), 10)
}

// NumValue returns the NumVal if type Int, or if
// it is a String, will attempt a conversion to int.
func (i Uint8OrString) NumValue() (uint8, error) {
	if i.Type == NumOrStringString {
		num, err := strconv.ParseUint(i.StrVal, 10, 8)
		return uint8(num), err
	}
	return i.NumVal, nil
}

// OpenAPISchemaType is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (_ Uint8OrString) OpenAPISchemaType() []string { return []string{"string"} }

// OpenAPISchemaFormat is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (_ Uint8OrString) OpenAPISchemaFormat() string { return "int-or-string" }
