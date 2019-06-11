// Copyright 2015 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package gojenkins

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"
)

func makeJson(data interface{}) string {
	str, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(json.RawMessage(str))
}

func Reverse(s string) string {
	size := len(s)
	buf := make([]byte, size)
	for start := 0; start < size; {
		r, n := utf8.DecodeRuneInString(s[start:])
		start += n
		utf8.EncodeRune(buf[size-start:], r)
	}
	return string(buf)
}

type JenkinsBlueTime time.Time

func (t *JenkinsBlueTime) UnmarshalJSON(b []byte) error {
	if b == nil || strings.Trim(string(b), "\"") == "null" {
		*t = JenkinsBlueTime(time.Time{})
		return nil
	}
	j, err := time.Parse("2006-01-02T15:04:05.000-0700", strings.Trim(string(b), "\""))

	if err != nil {
		return err
	}
	*t = JenkinsBlueTime(j)
	return nil
}

func (t JenkinsBlueTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t))
}
