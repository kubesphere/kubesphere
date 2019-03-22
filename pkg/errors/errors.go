/*

 Copyright 2019 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.

*/
package errors

import (
	"encoding/json"
	"errors"
)

type Error struct {
	Message string `json:"message"`
}

var None = Error{Message: "success"}

func (e *Error) Error() string {
	return e.Message
}

func Wrap(err error) Error {
	return Error{Message: err.Error()}
}

func Parse(data []byte) error {
	var j map[string]string
	err := json.Unmarshal(data, &j)
	if err != nil {
		return errors.New(string(data))
	} else if message := j["message"]; message != "" {
		return errors.New(message)
	} else if message := j["Error"]; message != "" {
		return errors.New(message)
	} else {
		return errors.New(string(data))
	}
}
