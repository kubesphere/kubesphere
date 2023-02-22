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
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ASNumber uint32

// ASNumberFromString creates an ASNumber struct from a string value.  The
// string value may simply be a number or may be the ASN in dotted notation.
func ASNumberFromString(s string) (ASNumber, error) {
	if num, err := strconv.ParseUint(s, 10, 32); err == nil {
		return ASNumber(num), nil
	}

	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		msg := fmt.Sprintf("invalid AS Number format (%s)", s)
		return 0, errors.New(msg)
	}

	if num1, err := strconv.ParseUint(parts[0], 10, 16); err != nil {
		msg := fmt.Sprintf("invalid AS Number format (%s)", s)
		return 0, errors.New(msg)
	} else if num2, err := strconv.ParseUint(parts[1], 10, 16); err != nil {
		msg := fmt.Sprintf("invalid AS Number format (%s)", s)
		return 0, errors.New(msg)
	} else {
		return ASNumber((num1 << 16) + num2), nil
	}
}

// UnmarshalJSON implements the json.Unmarshaller uinterface.
func (a *ASNumber) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, (*uint32)(a)); err == nil {
		return nil
	} else {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}

		if v, err := ASNumberFromString(s); err != nil {
			return err
		} else {
			*a = v
			return nil
		}
	}
}

// String returns the string value, or the Itoa of the uint value.
func (a ASNumber) String() string {
	return strconv.FormatUint(uint64(a), 10)
}
