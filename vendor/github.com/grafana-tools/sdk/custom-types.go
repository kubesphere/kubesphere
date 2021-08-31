package sdk

/*
   Copyright 2016 Alexander I.Grafov <grafov@gmail.com>
   Copyright 2016-2019 The Grafana SDK authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

	   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

   ॐ तारे तुत्तारे तुरे स्व
*/

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type BoolString struct {
	Flag  bool
	Value string
}

func (s *BoolString) UnmarshalJSON(raw []byte) error {
	if raw == nil || bytes.Equal(raw, []byte(`"null"`)) {
		return nil
	}
	var (
		tmp string
		err error
	)
	if raw[0] != '"' {
		if bytes.Equal(raw, []byte("true")) {
			s.Flag = true
			return nil
		}
		if bytes.Equal(raw, []byte("false")) {
			return nil
		}
		return errors.New("bad boolean value provided")
	}
	if err = json.Unmarshal(raw, &tmp); err != nil {
		return err
	}
	s.Value = tmp
	return nil
}

func (s BoolString) MarshalJSON() ([]byte, error) {
	if s.Value != "" {
		var buf bytes.Buffer
		buf.WriteRune('"')
		buf.WriteString(s.Value)
		buf.WriteRune('"')
		return buf.Bytes(), nil
	}
	return strconv.AppendBool([]byte{}, s.Flag), nil
}

type BoolInt struct {
	Flag  bool
	Value *int64
}

func (s *BoolInt) UnmarshalJSON(raw []byte) error {
	if raw == nil || bytes.Equal(raw, []byte(`"null"`)) {
		return nil
	}
	var (
		tmp int64
		err error
	)
	if tmp, err = strconv.ParseInt(string(raw), 10, 64); err != nil {
		if bytes.Equal(raw, []byte("true")) {
			s.Flag = true
			return nil
		}
		if bytes.Equal(raw, []byte("false")) {
			return nil
		}
		return errors.New("bad value provided")
	}
	s.Value = &tmp
	return nil
}

func (s BoolInt) MarshalJSON() ([]byte, error) {
	if s.Value != nil {
		return strconv.AppendInt([]byte{}, *s.Value, 10), nil
	}
	return strconv.AppendBool([]byte{}, s.Flag), nil
}

func NewIntString(i int64) *IntString {
	return &IntString{
		Value: i,
		Valid: true,
	}
}

// IntString represents special type for json values that could be strings or ints: 100 or "100"
type IntString struct {
	Value int64
	Valid bool
}

// UnmarshalJSON implements custom unmarshalling for IntString type
func (v *IntString) UnmarshalJSON(raw []byte) error {
	if raw == nil || bytes.Equal(raw, []byte(`"null"`)) || bytes.Equal(raw, []byte(`""`)) {
		return nil
	}

	strVal := string(raw)
	if rune(raw[0]) == '"' {
		strVal = strings.Trim(strVal, `"`)
	}

	i, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		return err
	}

	v.Value = i
	v.Valid = true

	return nil
}

// MarshalJSON implements custom marshalling for IntString type
func (v *IntString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		strVal := strconv.FormatInt(v.Value, 10)
		return []byte(strVal), nil
	}

	return []byte(`"null"`), nil
}

func NewFloatString(i float64) *FloatString {
	return &FloatString{
		Value: i,
		Valid: true,
	}
}

// FloatString represents special type for json values that could be strings or ints: 100 or "100"
type FloatString struct {
	Value float64
	Valid bool
}

// UnmarshalJSON implements custom unmarshalling for FloatString type
func (v *FloatString) UnmarshalJSON(raw []byte) error {
	if raw == nil || bytes.Equal(raw, []byte(`"null"`)) || bytes.Equal(raw, []byte(`""`)) {
		return nil
	}

	strVal := string(raw)
	if rune(raw[0]) == '"' {
		strVal = strings.Trim(strVal, `"`)
	}

	i, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		return err
	}

	v.Value = i
	v.Valid = true

	return nil
}

// MarshalJSON implements custom marshalling for FloatString type
func (v *FloatString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		strVal := strconv.FormatFloat(v.Value, 'g', -1, 64)
		return []byte(strVal), nil
	}

	return []byte(`"null"`), nil
}

// StringSliceString represents special type for json values that could be
// strings or slice of strings: "something" or ["something"].
type StringSliceString struct {
	Value []string
	Valid bool
}

// UnmarshalJSON implements custom unmarshalling for StringSliceString type.
func (v *StringSliceString) UnmarshalJSON(raw []byte) error {
	if raw == nil || bytes.Equal(raw, []byte(`"null"`)) {
		return nil
	}

	// First try with string.
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		v.Value = []string{str}
		v.Valid = true
		return nil
	}

	// Lastly try with string slice.
	var strSlice []string
	err := json.Unmarshal(raw, &strSlice)
	if err != nil {
		return err
	}

	v.Value = strSlice
	v.Valid = true
	return nil
}

// MarshalJSON implements custom marshalling for StringSliceString type.
func (v *StringSliceString) MarshalJSON() ([]byte, error) {
	if !v.Valid {
		return []byte(`"null"`), nil
	}

	return json.Marshal(v.Value)
}
